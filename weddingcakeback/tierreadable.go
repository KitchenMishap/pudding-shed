package weddingcakeback

// TierReadable supports hash lookup and hash at index, but in a way that acknowledges the possible presence of
// subsequent tiers
type TierReadable interface {
	// TryIndexOfHash The second return is true for "found"
	TryIndexOfHash(hash []byte) (GlobalPiType, bool, error)
	// TryGetHashAtIndex The second return is true for "found"
	TryGetHashAtIndex(index GlobalPiType, hash []byte) (bool, error)
	// GetNextTier returns nil to indicate no more tiers
	GetNextTier() TierReadable
	// Close closes the readable tier
	Close() error
}
