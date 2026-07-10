package weddingcakeback

import "os"

// BakingSourceTier is the interface that a tier must expose before being baked into a DonutForest in the
// next tier down
type BakingSourceTier interface {
	// GetIndicesCount gives the 256^n count of trees that exist in EACH DonutForest of the source tier
	GetIndicesCount() uint64
	// GetHashesAtIndex returns the hashes of the "tree at index" within each DonutForest of the source tier
	// It must therefore loop through each DonutForest in the source tier, concatenating the hashes from each tree
	GetHashesAtIndex(uint64, *CakeConfig) []SingleTreeHash // Repeatedly call GetHashes(uint(0)) ... GetHashes(GetIndicesCount()-1)
	// GetNextTierPrefixBytesCount returns "n" for the next tier (the destination tier)
	GetNextTierPrefixBytesCount() byte
	// GetNextTierIndex returns the index of the next tier (the destination tier). We think it is identical to "n".
	GetNextTierIndex() byte
	// AppendHashesFile appends the hashes file from the Source Tier to the specified file
	AppendHashesFile(*os.File) error
	// GetFirstPresentationIndex returns the first presentation index of the source tier
	GetFirstPresentationIndex() GlobalPiType
}
