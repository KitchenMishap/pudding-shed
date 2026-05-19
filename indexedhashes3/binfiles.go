package indexedhashes3

import (
	"bufio"
	"encoding/binary"
	"io"
	"math/bits"
	"os"
)

func loadBinFromFiles(bn binNum, binStartsFile *os.File, ovf *overflowFiles, p *HashIndexingParams) (*bin, error) {
	theBinBytes := make([]byte, p.BytesPerBinEntry()*p.EntriesInBinStart())
	_, err := binStartsFile.ReadAt(theBinBytes, int64(bn)*p.BytesPerBinEntry()*p.EntriesInBinStart())
	if err != nil {
		return nil, err
	}

	return loadBinFromSlice(theBinBytes, bn, ovf, p)
}

func loadBinFromSlice(theBinBytes []byte, bn binNum, ovf *overflowFiles, p *HashIndexingParams) (*bin, error) {
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

	// Now copy into new contiguous memory of the exact appropriate size
	bytes := make([]byte, len(theBinBytes))
	copy(bytes, theBinBytes)

	// The bin object itself will be a slice of slices that index into that memory
	theBin := bin{}
	theBin.bytes = bytes
	return &theBin, nil
}

func loadBinFromSliceIntoSlot(theBinBytes []byte, bn binNum, ovf *overflowFiles, p *HashIndexingParams, targetBin *bin) error {
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

	// Now copy into new contiguous memory of the exact appropriate size
	bytes := make([]byte, len(theBinBytes))
	copy(bytes, theBinBytes)

	targetBin.bytes = bytes
	return nil
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
		err = loadBinFromSliceIntoSlot(theBinBytes, bn, ovf, p, &result[int64(bn)])
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

// Faster function by Google Gemini AI:
func countZeroBytesAtEnd(data []byte) int64 {
	var count int64 = 0
	n := len(data)

	// 1. Step backward in fast 8-byte chunks
	for n >= 8 {
		// Grab the last 8 bytes
		chunk := binary.LittleEndian.Uint64(data[n-8 : n])

		if chunk == 0 {
			count += 8
			n -= 8
			continue
		}

		// If it's not entirely zero, find exactly how many trailing bytes are zero
		// Since it's Little Endian, the last bytes in the slice are the most significant bits.
		// Leading zeros in the uint64 / 8 tells us how many whole bytes at the end are 0.
		zeroBits := bits.LeadingZeros64(chunk)
		count += int64(zeroBits / 8)
		return count
	}

	// 2. Clean up any remaining bytes if the total size wasn't a multiple of 8
	for i := n - 1; i >= 0; i-- {
		if data[i] == 0 {
			count++
		} else {
			break
		}
	}

	return count
}

// This fn is still used somewhere despite a re-write...
func saveBinToFiles(bn binNum, b bin, binStartsFile *os.File, ovf *overflowFiles, p *HashIndexingParams) error {
	// Zeroes at the end of binStarts only ever get overwritten as bin gets bigger, the bin never gets smaller.
	// So we don't have to write zeroes after the bins, as they are already in the file.
	bytesPerBinEntry := p.BytesPerBinEntry()
	numEntries := b.length(bytesPerBinEntry)
	numEntriesBinStart := numEntries
	if numEntriesBinStart > p.EntriesInBinStart() {
		numEntriesBinStart = p.EntriesInBinStart()
		// Write the overflows file (to bytes first)
		overflowBytes := b.bytes[bytesPerBinEntry*numEntriesBinStart:]
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
	binStartByteCount := numEntriesBinStart * bytesPerBinEntry
	binStartBytes := b.bytes[:binStartByteCount]
	_, err := binStartsFile.WriteAt(binStartBytes, int64(bn)*p.EntriesInBinStart()*p.BytesPerBinEntry())
	return err
}

/*
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
}*/
