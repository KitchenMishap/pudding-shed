package indexedhashes3

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"sync"

	"github.com/KitchenMishap/pudding-shed/wordfile"
)

// "Adaptive pass" is a replacement for the multi-pass / single pass pattern.

// Stage one reads in all the hashes (with duplicates), determines the bin number from each hash,
// and counts the hashes (with duplicates) heading to each bin.

// The bins are sorted on size, so the biggest can go to a worker in Stage two soonest.

// Stage two "designs" "the next pass" using a knapsack algorithm.
// A pass deals with a specified non-sequential set of bins. A pass is designed to fit into the memory grant (knapsack).
// In the design stage, we reset "available memory" to the grant size.
// The unprocessed bin with the biggest count gets included first.
// The memory needed for that gets subtracted from the available memory for the pass.
// If this first bin exceeds the memory budget on its own, tough! But the pass is now designed.
// We continue with the design, repeatedly adding "the biggest unprocessed bin that fits",
// until all have been attempted.

// Stage three runs the designed pass using a number of workers based on the PC's cpu count.
// We we now "design" the workload for each worker:
// Repeat {assign the biggest remaining bin to the worker with the smallest workload}
// So each worker deals with a specified set of bins.
// A single thread reads hashes from file in sequence, determines the bin number and
// passes it to the input of the appropriate worker.
// When the hashes are exhausted, it closes the each of the workers input channel.
// Each worker receives each hash, and appends to its (duplicated) hash list for the appropriate bin.
// When the worker sees its input channel close, it sorts and de-duplicates its list, and passses
// a pointer to the resulting bin to its output channel.

// I need to think about writing the results to disk? The resultant bin data isn't in bin-number order!

type binWorkInfo struct {
	binNumber int64
	hashCount int64
}

// Remember to pass in a buffered file for efficiency, and to seek to zero
func stageOneCountBinWork(hashesFilepath string, binNumsFile wordfile.WriterAtWord, params *HashIndexingParams) ([]binWorkInfo, error) {
	fmt.Println("Reading hashes for pass planning...")
	numberOfBins := params.NumberOfBins()
	result := make([]binWorkInfo, numberOfBins)
	for b := range numberOfBins {
		result[b].binNumber = b
		result[b].hashCount = 0
	}
	hash := [32]byte{}

	fil, err := os.Open(hashesFilepath)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReaderSize(fil, 64*1024*1024)

	// Read first hash
	hi := int64(0)

	var bytesRead int
	bytesRead, err = reader.Read(hash[:])
	if err != nil && err != io.EOF {
		return nil, err
	}
	if bytesRead != 32 && err != io.EOF {
		return nil, errors.New("stageOneCountBinHashes(): Couldn't read 32 bytes")
	}
	for err != io.EOF {
		h := Hash(hash[:])
		abbr := h.toAbbreviatedHash()
		bn := abbr.toBinNum(params)

		result[bn].hashCount++

		err = binNumsFile.WriteWordAt(int64(bn), hi)

		// Read next hash
		bytesRead, err = reader.Read(hash[:])
		if err != nil && err != io.EOF {
			return nil, err
		}
		if bytesRead != 32 && err != io.EOF {
			return nil, errors.New("stageOneCountBinHashes(): Couldn't read 32 bytes")
		}
		hi++
	}
	err = fil.Close()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func stageTwoStageThreeHandleHashBins(hashesFilepath string,
	binStartsFile *os.File, ovf *overflowFiles,
	params *HashIndexingParams,
	workInfo []binWorkInfo, nGbBudget int64, threads int) error {

	hashesWorkBudget := 1024 * 1024 * 1024 * nGbBudget / params.BytesPerBinEntry()
	// The following are re-used between passes
	passWorkItems := make([]binWorkInfo, 0, params.NumberOfBins()) // This is actually more than we need
	workerHashWork := make([]int64, 0, threads)

	// Sort by descending hashCount
	sort.Slice(workInfo, func(i, j int) bool { return workInfo[i].hashCount > workInfo[j].hashCount })

	numberOfBins := int64(len(workInfo))
	workItemScheduled := make([]bool, numberOfBins)
	for w := range workItemScheduled {
		workItemScheduled[w] = false
	}

	binsAllocated := int64(0)

	for binsAllocated < numberOfBins {
		// Stage Two: Design a "pass" (being a collection of bins)
		fmt.Println("New pass...")
		passWorkItems = passWorkItems[:0]
		passHashesWork := int64(0) // <-- Move inside the loop here!

		// Repeatedly add the biggest bin that will fit the budget
		// (Or the biggest bin if it alone is bigger than the budget)
		// This is a greedy knapsack algorithm
		thisIsBiggest := true

		for w := 0; w < len(workInfo); w++ {
			if workItemScheduled[w] {
				continue
			}

			hashesLeft := hashesWorkBudget - passHashesWork
			willFit := hashesLeft > workInfo[w].hashCount

			if !workItemScheduled[w] && (willFit || thisIsBiggest) {
				passWorkItems = append(passWorkItems, workInfo[w])
				workItemScheduled[w] = true
				passHashesWork += workInfo[w].hashCount
				thisIsBiggest = false
				binsAllocated++
			}
		}
		// We have a design for a "pass" through the hashes (ie a collection of bin numbers)

		allScheduled := true
		for w := 0; w < len(workInfo); w++ {
			if !workItemScheduled[w] {
				allScheduled = false
			}
		}
		if allScheduled {
			fmt.Println("This will be the last pass")
		} else {
			fmt.Println("There will be another pass")
		}

		// Stage Three: Run the Pass
		// We will use a number of workers (threads).
		// We need to design the allocation of bins to workers.
		// We do this by keeping track of how much work (hashes) is assigned to each worker
		workerHashWork = workerHashWork[:0] // Reset for re-use
		for _ = range threads {
			workerHashWork = append(workerHashWork, 0)
		}
		binToWorkerMap := make(map[int64]int64)
		for _, wi := range passWorkItems {
			// Find the worker who so far has the least work on his plate
			leastWork := int64(math.MaxInt64)
			availableWorker := int64(0)
			for worker := range threads {
				if workerHashWork[worker] < leastWork {
					leastWork = workerHashWork[worker]
					availableWorker = int64(worker)
				}
			}

			// Assign the worker this work item
			workerHashWork[availableWorker] += wi.hashCount
			binToWorkerMap[wi.binNumber] = availableWorker
		}

		// Get the runtime size of a single bin entry (e.g., 34 bytes)
		bytesPerEntry := params.BytesPerBinEntry()

		// 1. One massive flat byte slice for the entire pass footprint
		fmt.Printf("Allocating %d Mb for pass\n", passHashesWork*bytesPerEntry/1024/1024)
		masterPool := make([]byte, passHashesWork*bytesPerEntry)

		// 2. Carve out a distinct byte-slice window for each worker
		workerBuffers := make([][]byte, threads)
		poolByteOffset := int64(0)

		for worker := 0; worker < threads; worker++ {
			allocatedEntries := workerHashWork[worker]
			if allocatedEntries > 0 {
				allocatedBytes := allocatedEntries * bytesPerEntry
				// Slice out this worker's exact raw byte territory
				workerBuffers[worker] = masterPool[poolByteOffset : poolByteOffset+allocatedBytes : poolByteOffset+allocatedBytes]
				poolByteOffset += allocatedBytes
			}
		}

		type workerAssignment struct {
			buffer      []byte          // Flat byte slice assigned to this worker
			binOffsets  map[int64]int64 // Map: binNumber -> Starting BYTE offset inside buffer
			binMaxBytes map[int64]int64 // Map: binNumber -> Maximum allowed BYTES for this bin's segment
		}

		// Inside Stage Three (initializing workers):
		assignments := make([]workerAssignment, threads)
		for t := range threads {
			assignments[t] = workerAssignment{
				buffer:      workerBuffers[t],
				binOffsets:  make(map[int64]int64),
				binMaxBytes: make(map[int64]int64),
			}
		}

		// Track current byte allocation pointer per worker to distribute bins contiguously
		workerCurrentBytePointer := make([]int64, threads)

		for _, wi := range passWorkItems {
			targetWorker := binToWorkerMap[wi.binNumber]

			// Set the starting byte position for this bin
			assignments[targetWorker].binOffsets[wi.binNumber] = workerCurrentBytePointer[targetWorker]

			binBytesNeeded := wi.hashCount * bytesPerEntry
			assignments[targetWorker].binMaxBytes[wi.binNumber] = binBytesNeeded

			// Advance the worker's allocation pointer by the total bytes needed
			workerCurrentBytePointer[targetWorker] += binBytesNeeded
		}

		// Here are the workers' input channels, waiting for their work
		type hashWork struct {
			binNumber int64
			hash      [32]byte
			hi        hashIndex
			sn        sortNum
		}

		inChans := make([]chan hashWork, threads)
		for t := range threads {
			inChans[t] = make(chan hashWork)
		}
		type binResult struct {
			binNumber int64
			binBytes  []byte
		}
		outChan := make(chan binResult)

		// Now launch your workers safely!
		wg := sync.WaitGroup{}
		for t := range threads {
			// Check if this worker actually has bins assigned to it
			if len(assignments[t].binOffsets) == 0 {
				close(inChans[t]) // Close it immediately, no one is listening
				continue
			}

			wg.Add(1)
			go func(workerID int, assign workerAssignment, inputChan <-chan hashWork, p *HashIndexingParams) {
				defer wg.Done()

				// ADD THIS RECOVER BLOCK TO EXPOSE SILENT PANICS:
				defer func() {
					if r := recover(); r != nil {
						fmt.Printf("CRITICAL: Worker %d panicked and died! Error: %v\n", workerID, r)
					}
				}()

				bPerEntry := p.BytesPerBinEntry()

				// Track the next BYTE insertion index for each bin
				writeByteOffsets := make(map[int64]int64)
				for bn, startBytePos := range assign.binOffsets {
					writeByteOffsets[bn] = startBytePos
				}

				workerBinEntryBytes := make([]byte, bytesPerEntry)
				// We assume your channel or main reader passes the `hashIndex` (hi) and `sortNum` (sn)
				// alongside the hash, or you calculate it here. Let's assume hashWork contains them:

				for work := range inputChan {
					bytePos := writeByteOffsets[work.binNumber]

					// 1. Extract your native custom truncated hash
					h := Hash(work.hash[:])
					th := h.toTruncatedHash()

					// 2. Generate the dynamic N-byte serialized binEntry byte slice
					writeBinEntryBytes(workerBinEntryBytes, th, work.hi, work.sn, p)

					// 3. Absolute direct byte copy into our flat pool segment
					copy(assign.buffer[bytePos:bytePos+bPerEntry], workerBinEntryBytes)

					// Advance the byte write pointer by N bytes
					writeByteOffsets[work.binNumber] = bytePos + bPerEntry
				}

				// --- Post-Processing Phase ---
				// When inputChan closes, each bin's data sits perfectly isolated in raw bytes
				for trueBinNumber, startBytePos := range assign.binOffsets {
					maxBytes := assign.binMaxBytes[trueBinNumber]

					// Slice the active byte segment for this specific bin
					rawBinBytes := assign.buffer[startBytePos : startBytePos+maxBytes]

					// Create a temporary bin wrapper structure to run your optimized batch code
					tempBin := &bin{bytes: rawBinBytes}

					// Execute the O(N log N) sorting and deduplication on this thread's core!
					// This runs the byte-agnostic code we fixed that eliminates the ghost-byte bug!
					tempBin.sortAndDeduplicate(p)

					// Now pass this clean `tempBin.bytes` slice out to your Disk Coordinator thread!
					output := binResult{}
					output.binNumber = trueBinNumber
					output.binBytes = tempBin.bytes
					outChan <- output
				}
			}(t, assignments[t], inChans[t], params)
		} // for t = threads

		// Now read the hashes and send to the workers
		go func() {
			hash := [32]byte{}

			hashesFile, ferr := os.Open(hashesFilepath)
			if ferr != nil {
				panic(ferr) // ToDo
			}
			defer hashesFile.Close()
			hashesReader := bufio.NewReaderSize(hashesFile, 64*1024*1024)

			hi := int64(0)

			// Read first hash
			var bytesRead int
			bytesRead, ferr = hashesReader.Read(hash[:])
			if ferr != nil && ferr != io.EOF {
				panic(ferr) // ToDo
			}
			if bytesRead != 32 && ferr != io.EOF {
				panic("stageOneCountBinHashes(): Couldn't read 32 bytes") // ToDo
			}
			for ferr != io.EOF {
				h := Hash(hash[:])
				abbr := h.toAbbreviatedHash()
				bn := abbr.toBinNum(params)
				sn := abbr.toSortNum(params)

				worker, workerExists := binToWorkerMap[int64(bn)]

				if workerExists {
					workPackage := hashWork{}
					workPackage.hash = h
					workPackage.binNumber = int64(bn)
					workPackage.sn = sn
					workPackage.hi = hashIndex(hi)
					// Send to worker
					inChans[worker] <- workPackage
				}

				// Read next hash
				bytesRead, ferr = hashesReader.Read(hash[:])
				if ferr != nil && ferr != io.EOF {
					panic(ferr) // ToDo
				}
				if bytesRead != 32 && ferr != io.EOF {
					panic("stageOneCountBinHashes(): Couldn't read 32 bytes")
				}
				hi++
			}
			// We have sent all the hashes, close all the workers' chans
			for ch := range threads {
				close(inChans[ch])
			}
		}()

		// We are currently sending all the hashes to the workers
		// Read out the results from the workers and save to files
		// 1. Monitor the WaitGroup in the background to safely close outChan
		go func() {
			wg.Wait()
			close(outChan)
		}()

		// 2. Consume the channel safely until it is drained and closed naturally
		fmt.Println("Starting to save to files...")
		// 1. Collect all pass results into a slice in memory
		passResults := make([]binResult, 0, len(passWorkItems))
		for out := range outChan {
			passResults = append(passResults, out)
		}

		// 2. Sort the completed results by bin number (ASCENDING)
		sort.Slice(passResults, func(i, j int) bool {
			return passResults[i].binNumber < passResults[j].binNumber
		})

		// 3. Blast them to disk sequentially! Sequential WriteAt's are MUCH faster (even with gaps)
		fmt.Println("Streaming sorted bins sequentially to disk...")
		for _, out := range passResults {
			aBin := bin{bytes: out.binBytes}
			err := saveBinToFiles(binNum(out.binNumber), aBin, binStartsFile, ovf, params)
			if err != nil {
				return err
			}
		}

		// At this point, ALL workers are dead and outChan is completely empty.
		fmt.Println("...Finished a pass")
	} // for binsAllocated < numberOfBins (ie, next pass)

	return nil
}
