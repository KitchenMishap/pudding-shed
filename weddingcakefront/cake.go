package weddingcakefront

import "github.com/KitchenMishap/pudding-shed/weddingcakeback"

type Cake struct {
	// A tierZero on its own (as things stand) supports the interfaces required of a Cake
	tierTop *weddingcakeback.TierTop
}

// Check that implements
var _ LegacyHashReader = (*Cake)(nil)
var _ LegacyHashReadWriter = (*Cake)(nil)

func newCake(tierTop *weddingcakeback.TierTop) *Cake {
	result := Cake{}
	result.tierTop = tierTop
	return &result
}

func (c *Cake) IndexOfHash(hash *Sha256) (int64, error) {
	result, err := c.tierTop.IndexOfHash(hash)
	return int64(result), err
}

func (c *Cake) GetHashAtIndex(index int64, hash *Sha256) error {
	return c.tierTop.GetHashAtIndex(index, hash)
}

func (c *Cake) CountHashes() (int64, error) {
	return c.tierTop.CountHashes()
}

func (c *Cake) Close() error {
	return c.tierTop.Close()
}

func (c *Cake) AppendHash(hash *Sha256) (int64, error) {
	result, err := c.tierTop.AppendHash(hash)
	if err != nil {
		return -1, err
	}
	return result, nil
}

func (c *Cake) Sync() error {
	return c.tierTop.Sync()
}
