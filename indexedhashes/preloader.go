package indexedhashes

import (
	"encoding/binary"
	"os"
)

type UniformHashPreLoader struct {
	uniform *UniformHashStore
}

// createEmptyFiles is a LONG operation suitable for a goroutine
func (pl *UniformHashPreLoader) createEmptyFiles() error {
	top := pl.biggestAddress()
	// We iterate backwards to see progress easily in file manager
	for i := int64(top); i >= 0; i-- {
		folders, filename, _ := pl.uniform.numberedFolders.NumberToFoldersAndFile(int64(i))
		folderPath, filePath := pl.uniform.folderPathFilePathFromFoldersFilename(folders, filename)
		err := os.MkdirAll(folderPath, 0755)
		if err != nil {
			return err
		}
		file, err := os.Create(filePath)
		if err != nil {
			return err
		}
		err = file.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (pl *UniformHashPreLoader) biggestAddress() uint64 {
	var hash Sha256 = [32]byte{255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255,
		255, 255, 255, 255, 255, 255, 255, 255}
	return pl.dividedAddressForHash(&hash)
}

func (pl *UniformHashPreLoader) dividedAddressForHash(hash *Sha256) uint64 {
	hashLSBs := binary.LittleEndian.Uint64(hash[0:8])
	dividedHash := hashLSBs / pl.uniform.hashDivider
	return dividedHash
}
