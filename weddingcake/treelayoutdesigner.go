package weddingcake

// TreeLayoutDesigner is introduced in the effort to consolidate/bake a (smaller/newer) tier n-1 of the "cake" into a
// new ring/chunk in the (larger/older) tier n underneath it. Tier n-1 will become empty ready for injection of
// new hashes (if it is tier 0), or ready for new rings/chunks to be consolidated/baked into it from the (even
// smaller/newer) tier n-2.

// TreeLayoutDesigner coordinates the measurement phase across multiple streaming inputs
type TreeLayoutDesigner struct {
	Metrics TierMetrics
}

// RecordSubtreeDistribution analyzes a raw list of hashes representing a future
// downstream subtree, measures it, and discards it to protect memory.
func (tld *TreeLayoutDesigner) RecordSubtreeDistribution(hashes []ShallowTreeHash, config StoreConfig, baseLevel byte) {
	if len(hashes) == 0 {
		return
	}

	// 1. Generate the temporary ShallowTree for this specific prefix bucket
	// We use the config's standard metrics
	tempTree := GenerateShallowTree(hashes, config.HashLength, config.ReassuranceBytesCount)

	// 2. Measure its nodes into our persistent, lightweight metrics matrix
	if tempTree.RootSlot.IsEmpty() {
		panic("Empty root slot should have been caught by len(hashes)==0")
	}
	tld.Metrics.RecurseAccumulateSubtreeMetrics(tempTree.RootSlot.NextNode, baseLevel)

	// 3. By returning, tempTree is dropped and cleared for Go's Garbage Collector!
}

// SlotCount represents how many active paths fan out from a node (0 to 256).
type SlotCount uint16

// LevelMetrics captures the statistical distribution of nodes on a single level.
// It tracks exactly how many nodes have a specific number of active slots.
type LevelMetrics struct {
	// Index represents the active slot count (0 to 256).
	// Value represents the total number of nodes matching that active slot density.
	ActiveSlotHistogram [257]NodeCountType
}

// TierMetrics Accumulates structural profiles across the entire ring/chunk layer.
type TierMetrics struct {
	// Max supported levels
	// (64 for each byte of the max supported hash size, plus 1 level for potential final leaf nodes)
	Levels [65]LevelMetrics
}

// RecurseAccumulateSubtreeMetrics analyzes a single hydrated subtree and aggregates its
// layout fingerprint before the subtree is dropped from memory.
func (tm *TierMetrics) RecurseAccumulateSubtreeMetrics(root *ShallowTreeNode, subtreeBaseLevel byte) {
	if root == nil {
		return
	}

	// Track the structural density of this specific node
	activeSlots := root.activeSlotsCount()
	targetLevel := root.Level + subtreeBaseLevel
	tm.Levels[targetLevel].ActiveSlotHistogram[activeSlots]++

	// Recurse down through this local subtree's children
	if root.SlotsNode != nil {
		for s := 0; s < 256; s++ {
			if !root.SlotsNode.Slots[s].IsEmpty() {
				tm.RecurseAccumulateSubtreeMetrics(root.SlotsNode.Slots[s].NextNode, subtreeBaseLevel)
			}
		}
	}
}
