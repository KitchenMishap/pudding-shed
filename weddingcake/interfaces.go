package weddingcake

// Define our specific fixed-size types
type Sha256 [32]byte
type Ripemd160 [20]byte

// HashArrayPointer is our type constraint (like a C++ concept).
// It ensures that whatever type H we pass in is a pointer to a byte array.
type HashArrayPointer interface {
	*Sha256 | *Ripemd160 | *[64]byte // Add any other fixed sizes you need
}

// Generic HashReader interface
type HashReader[H HashArrayPointer] interface {
	IndexOfHash(hash H) (int64, error)
	GetHashAtIndex(index int64, hash H) error
	CountHashes() (int64, error)
	Close() error
}

// Generic HashReadWriter interface
type HashReadWriter[H HashArrayPointer] interface {
	HashReader[H]
	AppendHash(hash H) (int64, error)
	Sync() error
}

// Generic HashStoreCreator interface
type HashStoreCreator[H HashArrayPointer] interface {
	HashStoreExists() bool
	CreateHashStore() error
	OpenHashStore() (HashReadWriter[H], error)
	OpenHashStoreReadOnly() (HashReader[H], error)
}

// Your legacy application sees these exact names, and they match perfectly!
type LegacyHashReader = HashReader[*Sha256]
type LegacyHashReadWriter = HashReadWriter[*Sha256]
type LegacyHashStoreCreator = HashStoreCreator[*Sha256]

// Because these are type aliases, any existing structural implementation in your application that satisfies
// the old HashReader will instantly satisfy LegacyHashReader without changing a line of code.
