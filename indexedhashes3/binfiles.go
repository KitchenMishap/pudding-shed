package indexedhashes3

import (
	"bufio"
	"io"
	"os"
)

func loadBinFromFiles(bn binNum, binStartsFile *os.File, ovf *overflowFiles, p *HashIndexingParams) (bin, error) {
	theBinBytes := make([]byte, p.BytesPerBinEntry()*p.EntriesInBinStart())
	_, err := binStartsFile.ReadAt(theBinBytes, int64(bn)*p.BytesPerBinEntry()*p.EntriesInBinStart())
	if err != nil {
		return nil, err
	}

	return loadBinFromSlice(theBinBytes, bn, ovf, p)
}

func loadBinFromSlice(theBinBytes []byte, bn binNum, ovf *overflowFiles, p *HashIndexingParams) (bin, error) {
	// Start with just the binStart

	// See how many zero entries are at the end
	zeroes := countZeroBytesAtEnd(theBinBytes)
	zeroEntries := zeroes / p.BytesPerBinEntry()

	if zeroEntries > 0 {
		// If there are some zero entries, truncate
		theBinBytes = theBinBytes[0 : p.BytesPerBinEntry()*(p.EntriesInBinStart()-zeroEntries)]
	} else {
		// Otherwise try to append any entries from overflow file
		_, overflowsFilepath := ovf.overflowFolderpathFilepath(bn)
		overflowBytes, err := os.ReadFile(overflowsFilepath)
		if err != nil {
			// Do nothing when file doesn't exist
		} else {
			theBinBytes = append(theBinBytes, overflowBytes...)
		}
	}

	entrySize := int(p.BytesPerBinEntry())
	entries := len(theBinBytes) / entrySize
	// Now copy into new contiguous memory of the exact appropriate size
	bytes := make([]byte, len(theBinBytes))
	copy(bytes, theBinBytes)

	// The bin object itself will be a slice of slices that index into that memory
	theBin := make([]binEntryBytes, entries)
	for i := 0; i < entries; i++ {
		theBin[i] = bytes[i*entrySize : (i+1)*entrySize]
	}
	return theBin, nil
}

func loadAllBinsFromFiles(binStartsFile *os.File, ovf *overflowFiles, p *HashIndexingParams) ([]bin, error) {
	result := make([]bin, p.NumberOfBins())
	theBinBytes := make([]byte, p.BytesPerBinEntry()*p.EntriesInBinStart())

	reader := bufio.NewReaderSize(binStartsFile, 64*1024*1024)

	for bn := binNum(0); int64(bn) < p.NumberOfBins(); bn++ {
		_, err := io.ReadFull(reader, theBinBytes)
		if err != nil {
			return nil, err
		}
		result[int64(bn)], err = loadBinFromSlice(theBinBytes, bn, ovf, p)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func countZeroBytesAtEnd(bytes []byte) int64 {
	l := int64(len(bytes))
	for result := int64(0); result < l; result++ {
		if bytes[l-result-1] != 0 {
			return result
		}
	}
	return l
}

// This fn is still used somewhere despite a re-write...
func saveBinToFiles(bn binNum, b bin, binStartsFile *os.File, ovf *overflowFiles, p *HashIndexingParams) error {
	// Zeroes at the end of binStarts only ever get overwritten as bin gets bigger, the bin never gets smaller.
	// So we don't have to write zeroes after the bins, as they are already in the file.
	numEntries := int64(len(b))
	numEntriesBinStart := numEntries
	if numEntriesBinStart > p.EntriesInBinStart() {
		numEntriesBinStart = p.EntriesInBinStart()
		// Write the overflows file (to bytes first)
		overflowByteCount := (numEntries - numEntriesBinStart) * p.BytesPerBinEntry()
		overflowBytes := make([]byte, overflowByteCount)
		for entry := numEntriesBinStart; entry < numEntries; entry++ {
			copy(overflowBytes[(entry-numEntriesBinStart)*p.BytesPerBinEntry():], b[entry])
		}
		// (now to file)
		overflowFolderpath, overflowFilepath := ovf.overflowFolderpathFilepath(bn)
		err := os.MkdirAll(overflowFolderpath, os.ModePerm)
		if err != nil {
			return err
		}
		err = os.WriteFile(overflowFilepath, overflowBytes, 0644)
		if err != nil {
			return err
		}
	}
	// Write the binStarts (not bothering with zeroes after)
	binStartByteCount := numEntriesBinStart * p.BytesPerBinEntry()
	binStartBytes := make([]byte, binStartByteCount)
	for entry := int64(0); entry < numEntriesBinStart; entry++ {
		copy(binStartBytes[entry*p.BytesPerBinEntry():], b[entry][:])
	}
	_, err := binStartsFile.WriteAt(binStartBytes, int64(bn)*p.EntriesInBinStart()*p.BytesPerBinEntry())
	return err
}

// For Gemini's rewrite
func saveOverflow(bn binNum, b bin, numEntriesBinStart int64, ovf *overflowFiles, p *HashIndexingParams) error {
	numEntries := int64(len(b))
	// Write the overflows file (to bytes first)
	overflowByteCount := (numEntries - numEntriesBinStart) * p.BytesPerBinEntry()
	overflowBytes := make([]byte, overflowByteCount)
	for entry := numEntriesBinStart; entry < numEntries; entry++ {
		copy(overflowBytes[(entry-numEntriesBinStart)*p.BytesPerBinEntry():], b[entry])
	}
	// (now to file)
	overflowFolderpath, overflowFilepath := ovf.overflowFolderpathFilepath(bn)
	err := os.MkdirAll(overflowFolderpath, os.ModePerm)
	if err != nil {
		return err
	}
	err = os.WriteFile(overflowFilepath, overflowBytes, 0644)
	if err != nil {
		return err
	}
	return nil
}
