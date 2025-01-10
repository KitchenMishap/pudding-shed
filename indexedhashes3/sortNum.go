package indexedhashes3

// A sortNum is used for sorting and searching hashes within a bin.
// It can also be used to reconstruct a hash if you also have the truncatedHash and binNum.
// It is the remainder when you divide the abbreviatedHash by the number of bins for this store.

type sortNum uint64
