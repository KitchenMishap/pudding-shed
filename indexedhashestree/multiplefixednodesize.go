package indexedhashestree

import "math"

// Every node format maps 256 byte values to either:
// A) A null, meaning no hash exists with this prefix, or
// B) A leaf, giving a presentation index, or
// C) A node ID, for deeper examination into the tree

// There are three node formats, but two of them have a sizing parameter

// A node format along with any sizing parameter, comprises a node spec

// Crucially, a given node spec has a fixed byte size. There are likely to be about 10 node specs.

// Node IDs increase from 1 through ranges of integers, each range capturing nodes of a certain node spec

// The node container specifies the node ID ranges for each sequential node spec.
// It also specifies the byte offset for the first node in each node spec.
// It is therefore possible, given a container and a node ID, to quickly arrive at the following
// with just a very small number of loookups, additions, and multiplications:
// 1) Byte offset for the node
// 2) Format for the node
// 3) Sizing param for the node
// 4) Bytesize of the node
// In short, everything.

type nodeDetails struct {
	byteOffset int
	nodeFormat int
	sizeParam  int
	byteSize   int
}

type nodeContainer struct {
	nodeIdStartingEachSpec []uint64
}

type nodeSpec struct {
	nodeFormat int
	sizeParam  int
	byteSize   int
}

func newNodeSpec(nodeFormat int, sizeParam int) *nodeSpec {
	result := nodeSpec{}
	result.nodeFormat = nodeFormat
	result.sizeParam = sizeParam
	if nodeFormat == formatTiny {
		result.byteSize = 1              // For the hash byte index 0..31
		result.byteSize += sizeParam * 3 // 3 bytes for a byteHashVal and a lookup
	} else if nodeFormat == formatMedium {
		result.byteSize = 1              // For a hash byte index
		result.byteSize += 32            // 256 bits which say whether each possible byteHashVal is represented below
		result.byteSize += sizeParam * 2 // sizeParam (padded) list of lookups for every bit that's a one above
	} else if nodeFormat == formatFull {
		result.byteSize = 1        // For a hash byte index
		result.byteSize += 256 * 2 // Lookup for every conceivable value of byteHashVal
	} else {
		panic("Don't know this format")
	}
	return &result
}

type containerParams struct {
	nodeSpecs []*nodeSpec
}

const formatTiny = 1
const formatMedium = 2
const formatFull = 3

// An example config
// This config is manually optimized for a count of approximately 32,768 hashes
func newContainerParamsConfigA() *containerParams {
	result := containerParams{}
	result.nodeSpecs = make([]*nodeSpec, 0)
	result.nodeSpecs = append(result.nodeSpecs, newNodeSpec(formatTiny, 2)) // Important - 86% of nodes!
	result.nodeSpecs = append(result.nodeSpecs, newNodeSpec(formatTiny, 3)) // 9% of nodes
	result.nodeSpecs = append(result.nodeSpecs, newNodeSpec(formatTiny, 4)) // 0.5% of nodes
	result.nodeSpecs = append(result.nodeSpecs, newNodeSpec(formatTiny, 5)) // A tiny five is vaguely worthwhile
	// For 50_000 hashes (nearly filling our 16 bit budget), there's a big gap
	// with virtually no nodes having between 5 and 127 slots. Let's not waste time there,
	// and instead give a finer grained size choice between 120 and 165 slots which is fairly busy.
	// For 32_768 hashes (a nice number), we find a gap between
	// 5 and 90 slots per node. Then we have a fair number of cases between 90 and 130 slots,
	// with a peak between 105 and 120. We therefore concentrate our range of available sizes more finely there.
	// (There's about 100-86-9-0.5 = 4.5% of nodes spread among the following):
	result.nodeSpecs = append(result.nodeSpecs, newNodeSpec(formatMedium, 90))
	result.nodeSpecs = append(result.nodeSpecs, newNodeSpec(formatMedium, 100))
	result.nodeSpecs = append(result.nodeSpecs, newNodeSpec(formatMedium, 105)) // 103,104,105 typically make up 0.8% of nodes (out of 4.5%)
	result.nodeSpecs = append(result.nodeSpecs, newNodeSpec(formatMedium, 108)) // 106,107,108 typically make up 0.9% of nodes (out of 4.5%)
	result.nodeSpecs = append(result.nodeSpecs, newNodeSpec(formatMedium, 112)) // 109,110,111 typically make up 0.8% of nodes (out of 4.5%)
	result.nodeSpecs = append(result.nodeSpecs, newNodeSpec(formatMedium, 118))
	result.nodeSpecs = append(result.nodeSpecs, newNodeSpec(formatMedium, 126))
	result.nodeSpecs = append(result.nodeSpecs, newNodeSpec(formatMedium, 130))
	result.nodeSpecs = append(result.nodeSpecs, newNodeSpec(formatFull, 256))
	return &result
}

func (cp *containerParams) nodeBytesizeSuitableFor(slots int) int {
	bestByteSize := math.MaxInt
	foundOne := false
	for _, spec := range cp.nodeSpecs {
		if spec.sizeParam >= slots {
			foundOne = true
			byteSize := spec.byteSize
			if byteSize < bestByteSize {
				bestByteSize = byteSize
			}
		}
	}
	if !foundOne {
		panic("Not found")
	}
	return bestByteSize
}
