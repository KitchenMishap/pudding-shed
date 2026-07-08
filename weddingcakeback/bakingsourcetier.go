package weddingcakeback

// BakingSourceTier is the interface that a tier must expose before being baked into a DonutForest in the
// next tier down
type BakingSourceTier interface {
	GetIndicesCount() uint64
	GetHashesAtIndex(uint64, *CakeConfig) []SingleTreeHash // Repeatedly call GetHashes(uint(0)) ... GetHashes(GetIndicesCount()-1)
	GetNextTierPrefixBytesCount() byte
	GetNextTierIndex() byte
}
