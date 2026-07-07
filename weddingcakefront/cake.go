package weddingcakefront

import "github.com/KitchenMishap/pudding-shed/weddingcakeback"

type Cake struct {
	// A tierZero on its own (as things stand) supports the interfaces required of a Cake
	tierZero *weddingcakeback.TierZero
}

// Check that implements
var _ LegacyHashReader = (*Cake)(nil)
var _ LegacyHashReadWriter = (*Cake)(nil)

func newCake(tierZero *weddingcakeback.TierZero) *Cake {
	result := Cake{}
	result.tierZero = tierZero
	return &result
}

func (c *Cake) IndexOfHash(hash *Sha256) (int64, error) {
	result, err := c.tierZero.IndexOfHash(hash)
	return int64(result), err
}

func (c *Cake) GetHashAtIndex(index int64, hash *Sha256) error {
	return c.tierZero.GetHashAtIndex(index, hash)
}

func (c *Cake) CountHashes() (int64, error) {
	return c.tierZero.CountHashes()
}

func (c *Cake) Close() error {
	return c.tierZero.Close()
}

func (c *Cake) AppendHash(hash *Sha256) (int64, error) {
	result, err := c.tierZero.AppendHash(hash)
	if err != nil {
		return -1, err
	}
	return int64(result), nil
}

func (c *Cake) Sync() error {
	return c.tierZero.Sync()
}
