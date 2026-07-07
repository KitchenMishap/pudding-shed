package weddingcakeback

import (
	"os"
	"path/filepath"
)

type TierZeroCreator struct {
	folder string
}

func NewTierZeroCreator(folder string) *TierZeroCreator {
	result := TierZeroCreator{}
	result.folder = folder
	return &result
}

func (tzc *TierZeroCreator) Exists() bool {
	// Based on existence of <folder>/Tier0/Hashes.hsh
	filePath := filepath.Join(tzc.folder, "Tier0", "Hashes.hsh")
	file, err := os.Open(filePath)
	if err == nil {
		_ = file.Close()
		return true
	}
	return false
}

func (tzc *TierZeroCreator) Create() error {
	// Create an empty <folder>/Tier0/Hashes.hsh
	folderPath := filepath.Join(tzc.folder, "Tier0")
	err := os.MkdirAll(folderPath, 0777)
	if err != nil {
		return err
	}
	filePath := filepath.Join(tzc.folder, "Tier0", "Hashes.hsh")
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	return nil
}

func (tzc *TierZeroCreator) Open() (*TierZero, error) {
	filePath := filepath.Join(tzc.folder, "Tier0")
	tierZero, err := NewTierZero(filePath, false)
	if err != nil {
		return nil, err
	}
	return tierZero, nil
}

func (tzc *TierZeroCreator) OpenReadOnly() (*TierZero, error) {
	filePath := filepath.Join(tzc.folder, "Tier0")
	tierZero, err := NewTierZero(filePath, true)
	if err != nil {
		return nil, err
	}
	return tierZero, nil
}
