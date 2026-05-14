package indexedhashes3

import (
	"bufio"
	"io"
	"os"
	"sync"

	"github.com/KitchenMishap/pudding-shed/wordfile"
)

// singlePassDetails holds details of one of the multiple passes
type singlePassDetails struct {
	firstBinNum       int64
	lastBinNumPlusOne int64
	binNumsWordFile   wordfile.WriterAtWord
	bins              []bin
}

func newSinglePassDetails(firstBinNum int64, binsCount int64,
	binNumsWordFile wordfile.WriterAtWord, expectedEntriesPerBin int64) *singlePassDetails {
	result := singlePassDetails{}
	result.firstBinNum = firstBinNum
	result.lastBinNumPlusOne = firstBinNum + binsCount
	if firstBinNum == 0 {
		result.binNumsWordFile = binNumsWordFile
	}
	result.bins = make([]bin, binsCount)
	for i := int64(0); i < binsCount; i++ {
		result.bins[i] = newEmptyBin(expectedEntriesPerBin)
	}
	return &result
}

func (spd *singlePassDetails) readIn(mp *MultipassPreloader, threads int) error {
	sep := string(os.PathSeparator)
	hashesFilepath := mp.folderPath + sep + "Hashes.hsh"
	hashesFile, err := os.Open(hashesFilepath)
	if err != nil {
		return err
	}
	defer hashesFile.Close()

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
	for {
		// We will prepare a work item
		// Prepare a work item
		wi := workItem{}
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
		wi.aBin = &(spd.bins[passBinNumber])

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

	// Is it in the bin already?
	if theBin.lookupByHash(th, sn, p) != -1 {
		return
	}

	theBin.insertBinEntry(sn, hashIndex(hi), th, p)
}

func (spd *singlePassDetails) writeFiles(mp *MultipassPreloader) error {
	for index, element := range spd.bins {
		bn := spd.firstBinNum + int64(index)
		err := saveBinToFiles(binNum(bn), element, mp.binStartsFile, mp.overflowFiles, mp.params)
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
		for _, element := range spd.bins {
			if len(element) > 0 {
				return // OK
			}
		}
		panic("There are no non-empty Bins")
	}
}
