package indexedhashes3

// A truncated hash is just the 24 MSBytes of a hash

type truncatedHash [24]byte

func (tr *truncatedHash) equals(tr2 *truncatedHash) bool {
	return *tr == *tr2
}
