package indexedhashes3

import (
	"os"
)

func loadBinFromFiles(bn binNum, binStartsFile *os.File, ovf *overflowFiles, p *HashIndexingParams) (bin, error) {
	// Start with just the binStart
	theBinBytes := make([]byte, p.BytesPerBinEntry()*p.EntriesInBinStart())
	_, err := binStartsFile.ReadAt(theBinBytes, int64(bn)*p.BytesPerBinEntry()*p.EntriesInBinStart())
	if err != nil {
		return nil, err
	}

	// See how many zero entries are at the end
	zeroes := countZeroBytesAtEnd(theBinBytes)
	zeroEntries := zeroes / p.BytesPerBinEntry()

	if zeroEntries > 0 {
		// If there are some zero entries, truncate
		theBinBytes = theBinBytes[0 : p.BytesPerBinEntry()*(p.EntriesInBinStart()-zeroEntries)]
	} else {
		// Otherwise try to append any entries from overflow file
		overflowsFilepath := ovf.overflowFilepath(bn)
		overflowBytes, err := os.ReadFile(overflowsFilepath)
		if err != nil {
			// Do nothing when file doesn't exist
		} else {
			theBinBytes = append(theBinBytes, overflowBytes...)
		}
	}
	// Now split into entries and construct a bin
	theBin := make([]binEntryBytes, len(theBinBytes)/int(p.BytesPerBinEntry()))
	for i := int64(0); i < int64(len(theBin)); i++ {
		theBin[i] = theBinBytes[i*p.BytesPerBinEntry() : (i+1)*p.BytesPerBinEntry()]
	}
	return theBin, nil
}

func countZeroBytesAtEnd(bytes []byte) int64 {
	l := int64(len(bytes))
	for result := l; result > 0; result-- {
		if bytes[result-1] != 0 {
			return result
		}
	}
	return 0
}

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
		overflowFilepath := ovf.overflowFilepath(bn)
		err := os.WriteFile(overflowFilepath, overflowBytes, 0644)
		if err != nil {
			return err
		}
	}
	// Write the binStarts (not bothering with zeroes after)
	binStartByteCount := numEntriesBinStart * p.BytesPerBinEntry()
	binStartBytes := make([]byte, binStartByteCount)
	for entry := int64(0); entry < numEntriesBinStart; entry++ {
		copy(binStartBytes[entry*p.BytesPerBinEntry():], b[entry])
	}
	_, err := binStartsFile.WriteAt(binStartBytes, int64(bn)*p.EntriesInBinStart()*p.BytesPerBinEntry())
	return err
}
