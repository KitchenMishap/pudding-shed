package weddingcakeback

import (
	"encoding/binary"
	"os"
	"path/filepath"
)

type TierTopCreator struct {
	folder string
}

func NewTierTopCreator(folder string) *TierTopCreator {
	result := TierTopCreator{}
	result.folder = folder
	return &result
}

func (ttc *TierTopCreator) Exists() bool {
	// Based on existence of <folder>/Tier0/Hashes.hsh
	filePath := filepath.Join(ttc.folder, "TierTop", "Hashes.hsh")
	file, err := os.Open(filePath)
	if err == nil {
		_ = file.Close()
		return true
	}
	return false
}

func (ttc *TierTopCreator) Create(firstGlobalPresentationIndex GlobalPiType) error {
	// Create an empty <folder>/Tier0/Hashes.hsh
	folderPath := filepath.Join(ttc.folder, "TierTop")
	err := os.MkdirAll(folderPath, 0777)
	if err != nil {
		return err
	}
	filePath := filepath.Join(ttc.folder, "TierTop", "Hashes.hsh")
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	// Also create a file <folder>/Tier0/FirstPresentationIndex.bin containing a uint64 of 0
	filePath2 := filepath.Join(ttc.folder, "TierTop", "FirstPresentationIndex.bin")
	file2, err := os.Create(filePath2)
	if err != nil {
		return err
	}
	defer func() { _ = file2.Close() }()
	err = binary.Write(file2, binary.LittleEndian, firstGlobalPresentationIndex)
	if err != nil {
		return err
	}

	return nil
}

func (ttc *TierTopCreator) Open() (*TierTop, error) {
	tierTop, err := NewTierTop(ttc.folder, false)
	if err != nil {
		return nil, err
	}
	return tierTop, nil
}

func (ttc *TierTopCreator) OpenReadOnly() (*TierTop, error) {
	tierTop, err := NewTierTop(ttc.folder, true)
	if err != nil {
		return nil, err
	}
	return tierTop, nil
}
