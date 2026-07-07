package weddingcakefront

import (
	"github.com/KitchenMishap/pudding-shed/weddingcakeback"
)

type CakeCreator struct {
	folder          string
	tierZeroCreator *weddingcakeback.TierZeroCreator
}

// Check that implements
var _ LegacyHashStoreCreator = (*CakeCreator)(nil)

func NewCakeCreator(folder string) *CakeCreator {
	result := CakeCreator{}
	result.folder = folder
	result.tierZeroCreator = weddingcakeback.NewTierZeroCreator(folder)
	return &result
}

func (cc *CakeCreator) HashStoreExists() bool {
	// Exists if tier zero exists
	return cc.tierZeroCreator.Exists()
}

func (cc *CakeCreator) CreateHashStore() error {
	// Create an empty tier zero
	return cc.tierZeroCreator.Create()
}

func (cc *CakeCreator) OpenHashStore() (HashReadWriter[*Sha256], error) {
	// Open tier zero
	tierZero, err := cc.tierZeroCreator.Open()
	if err != nil {
		return nil, err
	}
	// Make a cake out of tier zero
	return newCake(tierZero), nil
}

func (cc *CakeCreator) OpenHashStoreReadOnly() (HashReader[*Sha256], error) {
	// Open tier zero read only
	tierZero, err := cc.tierZeroCreator.OpenReadOnly()
	if err != nil {
		return nil, err
	}
	// Make a cake out of tier zero
	return newCake(tierZero), nil
}
