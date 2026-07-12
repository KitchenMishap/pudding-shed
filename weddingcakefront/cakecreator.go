package weddingcakefront

import (
	"github.com/KitchenMishap/pudding-shed/weddingcakeback"
)

type CakeCreator struct {
	folder         string
	tierTopCreator *weddingcakeback.TierTopCreator
}

// Check that implements
var _ LegacyHashStoreCreator = (*CakeCreator)(nil)

func NewCakeCreator(folder string) *CakeCreator {
	result := CakeCreator{}
	result.folder = folder
	result.tierTopCreator = weddingcakeback.NewTierTopCreator(folder)
	return &result
}

func (cc *CakeCreator) HashStoreExists() bool {
	// Exists if tier zero exists
	return cc.tierTopCreator.Exists()
}

func (cc *CakeCreator) CreateHashStore() error {
	// Create an empty tier zero
	return cc.tierTopCreator.Create(0)
}

func (cc *CakeCreator) OpenHashStore() (HashReadWriter[*Sha256], error) {
	// Open tier zero
	tierTop, err := cc.tierTopCreator.Open()
	if err != nil {
		return nil, err
	}
	// Make a cake out of tier zero
	return newCake(tierTop), nil
}

func (cc *CakeCreator) OpenHashStoreReadOnly() (HashReader[*Sha256], error) {
	// Open tier zero read only
	tierTop, err := cc.tierTopCreator.OpenReadOnly()
	if err != nil {
		return nil, err
	}
	// Make a cake out of tier zero
	return newCake(tierTop), nil
}
