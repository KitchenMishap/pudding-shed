package weddingcakefront

import (
	"errors"
	"fmt"

	"github.com/KitchenMishap/pudding-shed/weddingcakeback"
)

type Cake struct {
	// A tierZero on its own (as things stand) supports the interfaces required of a Cake
	tierTop    *weddingcakeback.TierTop
	tierReader weddingcakeback.TierReadable
}

// Check that implements
var _ LegacyHashReader = (*Cake)(nil)
var _ LegacyHashReadWriter = (*Cake)(nil)

func newCake(tierTop *weddingcakeback.TierTop) *Cake {
	result := Cake{}
	result.tierTop = tierTop
	result.tierReader = tierTop
	return &result
}

func (c *Cake) IndexOfHash(hash *Sha256) (int64, error) {
	found := false
	reader := c.tierReader
	var result GlobalPiType
	var err error
	tierCount := 0
	for !found && reader != nil {
		tierCount++
		result, found, err = reader.TryIndexOfHash((*hash)[:])
		if err != nil {
			return -1, err
		}
		if !found {
			reader = reader.GetNextTier()
		}
	}
	if !found {
		//fmt.Printf("Hash not found after %d tiers\n", tierCount)
		return -1, nil
	}
	if tierCount == 1 {
		fmt.Printf("Found in tier %d\n", tierCount)
	}
	return int64(result), nil
}

func (c *Cake) GetHashAtIndex(index int64, hash *Sha256) error {
	found := false
	reader := c.tierReader
	var err error
	for !found && reader != nil {
		found, err = reader.TryGetHashAtIndex(GlobalPiType(index), (*hash)[:])
		if err != nil {
			return err
		}
		if !found {
			reader = reader.GetNextTier()
		}
	}
	if !found {
		return errors.New("Hash not found for index")
	}
	return nil
}

func (c *Cake) CountHashes() (int64, error) {
	panic("Not implemented")
}

func (c *Cake) Close() error {
	err := c.tierTop.CloseAll()
	if err != nil {
		return err
	}
	c.tierTop = nil
	c.tierReader = nil
	return nil
}

func (c *Cake) AppendHash(hash *Sha256) (int64, error) {
	result, err := c.tierTop.AppendHash((*hash)[:])
	if err != nil {
		return -1, err
	}
	return int64(result), nil
}

func (c *Cake) Sync() error {
	return c.tierTop.Sync()
}
