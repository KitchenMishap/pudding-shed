package indexedhashes3

import (
	"bufio"
	"io"
	"os"
	"sync"
	"sync/atomic"

	"github.com/KitchenMishap/pudding-shed/wordfile"
)

// singlePassDetails holds details of one of the multiple passes
type singlePassDetails struct {
	firstBinNum       int64
	lastBinNumPlusOne int64
	binNumsWordFile   wordfile.WriterAtWord
	bins              *BinsArray
}

func newSinglePassDetails(firstBinNum int64, binsThisPass int64,
	binNumsWordFile wordfile.WriterAtWord, ba *BinsArray) *singlePassDetails {
	result := singlePassDetails{}
	result.firstBinNum = firstBinNum
	result.lastBinNumPlusOne = firstBinNum + binsThisPass
	if firstBinNum == 0 {
		result.binNumsWordFile = binNumsWordFile
	}
	result.bins = ba
	return &result
}

func (spd *singlePassDetails) readIn(mp *MultipassPreloader, threads int) error {

	// Clear out the bins (retaining capacity) from any previous use
	spd.bins.Reuse(mp.params.EntriesInBinStart(), mp.params.BytesPerBinEntry(), spd.lastBinNumPlusOne-spd.firstBinNum)

	sep := string(os.PathSeparator)
	hashesFilepath := mp.folderPath + sep + "Hashes.hsh"
	hashesFile, err := os.Open(hashesFilepath)
	if err != nil {
		return err
	}
	defer func() { _ = hashesFile.Close() }()

	reader := bufio.NewReaderSize(hashesFile, 8*1024*1024) // Google Gemini AI says this will be much faster

	// We will fan-out to some workers; each worker has exclusive access to a subset of the bins
	type workItem struct {
		// A work item relates to one hash
		aBin *bin
		sn   sortNum
		hi   int64
		th   truncatedHash
	}
	workerChans := make([]chan workItem, threads)
	var wg sync.WaitGroup
	for i := range threads {
		workerChans[i] = make(chan workItem, 1000)
	}
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func(workerIndex int) {
			defer wg.Done()
			for item := range workerChans[workerIndex] {
				spd.dealWithOneHash(item.aBin, item.sn, item.hi, item.th, mp.params)
			}
		}(i)
	}

	hi := int64(0)
	wi := workItem{}
	for {
		// We will prepare a work item
		// Prepare a work item
		wi.hi = hi

		// Read 32 bytes directly from the buffer
		// This is MUCH faster than manual chunking
		var hash [32]byte
		_, err = io.ReadFull(reader, hash[:])
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		hash3 := Hash(hash)
		abbr := hash3.toAbbreviatedHash()
		bn := abbr.toBinNum(mp.params)

		if spd.firstBinNum == 0 {
			// First pass, store the binNum in a wordfile
			err = spd.binNumsWordFile.WriteWordAt(int64(bn), hi)
			if err != nil {
				return err
			}
		}

		// This single pass only deals with a certain range of bin numbers
		if int64(bn) < spd.firstBinNum || int64(bn) >= spd.lastBinNumPlusOne {
			hi++
			continue
		}

		wi.th = hash3.toTruncatedHash() // This creates a new object
		wi.sn = abbr.toSortNum(mp.params)

		passBinNumber := int64(bn) - spd.firstBinNum
		wi.aBin = spd.bins.bins[passBinNumber]

		// It's crucial that each worker deals with bins that no other worker handles
		// So send the work to the CORRECT worker
		workerNum := passBinNumber % int64(threads)
		workerChans[workerNum] <- wi

		hi++
	}
	// Close all the channels into the workers
	for i := range threads {
		close(workerChans[i])
	}

	wg.Wait()

	return nil
}

func (spd *singlePassDetails) dealWithOneHash(theBin *bin,
	sn sortNum, hi int64, th truncatedHash, p *HashIndexingParams) {

	// Append blindly, we will sort and de-duplicate later
	theBin.appendBinEntry(sn, hashIndex(hi), th, p)
}

func (spd *singlePassDetails) writeFiles(mp *MultipassPreloader) error {
	for index, element := range spd.bins.bins {
		bn := spd.firstBinNum + int64(index)
		err := saveBinToFiles(binNum(bn), *element, mp.binStartsFile, mp.overflowFiles, mp.params)
		if err != nil {
			return err
		}
	}
	return nil
}

/*
// Google Gemini AI's rewrite, to alleviate the WriteAt()'s
func (spd *singlePassDetails) writeFiles(mp *MultipassPreloader) error {
	// 1. Pre-calculate the total size of the "BinStarts" block for this pass
	bytesPerBinTotal := mp.params.EntriesInBinStart() * mp.params.BytesPerBinEntry()
	passBufferSize := int64(len(spd.bins)) * bytesPerBinTotal

	// 2. Allocate one "Mega Buffer" for the whole pass
	// We use the 64GB grant logic here.
	megaBuffer := make([]byte, passBufferSize)

	for index, b := range spd.bins {
		bn := spd.firstBinNum + int64(index)

		// Calculate where this bin starts in our Mega Buffer
		destOffset := int64(index) * bytesPerBinTotal

		numEntries := int64(len(b))
		numEntriesBinStart := numEntries
		if numEntriesBinStart > mp.params.EntriesInBinStart() {
			numEntriesBinStart = mp.params.EntriesInBinStart()

			// --- Handle Overflows (Keep these as individual files for now) ---
			// (Your existing overflow logic is fine here, but use a reusable buffer if possible)
			saveOverflow(binNum(bn), b, numEntriesBinStart, mp.overflowFiles, mp.params)
		}

		// 3. Copy the bin data into the Mega Buffer (No syscall yet!)
		for entry := int64(0); entry < numEntriesBinStart; entry++ {
			copy(megaBuffer[destOffset+(entry*mp.params.BytesPerBinEntry()):], b[entry])
		}
	}

	// 4. ONE SINGLE SYSCALL for the entire pass
	globalStartOffset := spd.firstBinNum * bytesPerBinTotal
	_, err := mp.binStartsFile.WriteAt(megaBuffer, globalStartOffset)
	return err
}*/

func (spd *singlePassDetails) checkThereAreNonEmptyBins() {
	const verify = false
	if verify {
		for _, element := range spd.bins.bins {
			if len(element.bytes) > 0 {
				return // OK
			}
		}
		panic("There are no non-empty Bins")
	}
}

func (spd *singlePassDetails) sortAndDeduplicateBins(params *HashIndexingParams, threads int) {
	var wg sync.WaitGroup

	// 2. Use an atomic int64 counter to act as our thread-safe shared work index.
	// We initialize it to the first bin number.
	var currentBin int64 = int64(spd.firstBinNum)
	lastBin := int64(spd.lastBinNumPlusOne)

	// 3. Launch the fixed pool of concurrent workers
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				// Atomically increment the bin index so workers never collide or duplicate work.
				// AddInt64 returns the NEW value, so we subtract 1 to get the slot we actually claimed.
				b := atomic.AddInt64(&currentBin, 1) - 1

				// If we've processed all the bins, this worker thread exits gracefully.
				if b >= lastBin {
					break
				}

				// Execute the O(N log N) sorting and deduplication on this thread's core
				spd.bins.bins[b-spd.firstBinNum].SortAndDeduplicate(params)
			}
		}()
	}

	// 4. Block and wait here until every worker thread has completed its share of the bins
	wg.Wait()
}
