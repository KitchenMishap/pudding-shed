package weddingcakeback

import (
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

func (tzc *TierTopCreator) Exists() bool {
	// Based on existence of <folder>/Tier0/Hashes.hsh
	filePath := filepath.Join(tzc.folder, "Tier0", "Hashes.hsh")
	file, err := os.Open(filePath)
	if err == nil {
		_ = file.Close()
		return true
	}
	return false
}

func (tzc *TierTopCreator) Create() error {
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

func (tzc *TierTopCreator) Open() (*TierTop, error) {
	filePath := filepath.Join(tzc.folder, "Tier0")
	tierTop, err := NewTierTop(filePath, false)
	if err != nil {
		return nil, err
	}
	return tierTop, nil
}

func (tzc *TierTopCreator) OpenReadOnly() (*TierTop, error) {
	filePath := filepath.Join(tzc.folder, "Tier0")
	tierTop, err := NewTierTop(filePath, true)
	if err != nil {
		return nil, err
	}
	return tierTop, nil
}
