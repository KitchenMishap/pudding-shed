# package indexedhashes
indexedhashes is a key-value database mapping unique sha256 hashes to counting integers.
- indexedhashes holds a sequence of unique sha256 hashes for you.
- They are indexed 0,1,2...
- You can (of course) supply an index and lookup the corresponding hash previously stored (ie, an array)
- But you can also supply a hash and lookup the index it is stored at (quickly) (ie, a dictionary)
### Files
- The hashes are stored in three files
- **FileName.hsh** - the growing raw hashes in sequence
- **FileName.lkp** - A fixed size lookup file indexed by the "partial hash" (the LSBs) of each full hash
- **FileName.cls** - A growing linked list of chunks dealing with collisions: cases where two hashes have the same partial hashes
### Parameters
- partialHashBitCount - The number of LSBs of a 256 bit hash that constitute a partial hash
- entryByteCount - The number of bytes that constitute an entry in the lookup file, or in a chunk
- collisionsPerChunk - The number of collisions stored in a chunk

I've calculated/chosen some parameters for use with Bitcoin,
suitable for the number of block hashes and transaction hashes currently in the blockchain as of 2023:

**For Block Hashes:**
partialHashBitCount = 30,
entryByteCount = 3,
collisionsPerChunk = 3

**For Transaction Hashes:**
partialHashBitCount = 30,
entryByteCount = 4,
collisionsPerChunk = 3

## How to use indexedhashes package
### 1. Instantiate a ConcreteHashStoreCreator (a factory)
```
import (
		"github.com/KitchenMishap/pudding-shed/indexedhashes"
        "crypto/sha256"
		"log"
)

phbc := 30      // Partial hash bit count
ebc := 4        // Entry byte count
cpc := 3        // Collisions per chunk
creator, err := NewConcreteHashStoreCreator("FileName", "FolderName", phbc, ebc, cpc)
if err != nil {
    log.Println(err)
}
```
### 2. Create the hash store files
```
err = creator.CreateHashStore()
if err != nil {
    log.Println(err)
}
```
### 3. Open the hash store for read/write
```
store, err := creator.OpenHashStore()
if err != nil {
    log.Println(err)
}
```
### 4. Create some hashes of strings using package crypto/sha256
```
h0 := Sha256{}
h := sha256.New()
h.Write([]byte("Hello"))
o := h.Sum(nil)
for i := 0; i < len(o); i++ {
    h0[i] = o[i]
}
h1 := Sha256{}
h = sha256.New()
h.Write([]byte("World"))
o = h.Sum(nil)
for i := 0; i < len(o); i++ {
    h1[i] = o[i]
}
```
### 5. Store the hashes
```
index, err := store.AppendHash(&h0)
if err != nil {
    log.Println(err)
}
index, err = store.AppendHash(&h1)
if err != nil {
    log.Println(err)
}
```
### 6. Read back the hash at index 0
```
hash := Sha256{}
err = store.GetHashAtIndex(0, &hash)
if err != nil {
    log.Println(err)
}
```
### 7. Find the index of hash h1
```
index, err = store.IndexOfHash(&h1)
if err != nil {
    log.Println(err)
}
println(index)
```
# How it works - file formats
### 1. Hashes file, FileName.hsh
This is simply the binary 256 bit full hashes stored in sequence
### 2. Lookup file, FileName.lkp
* This is a large fixed size file, initially full of zeroes.
* Each entry is entryByteCount bytes long
* It is indexed by partial hash, which is the LSBs of a full 256 bit hash
* There are partialHashBitCount bits in a partial hash
* Thus the filesize is 2 ^ partialHashBitCount * entryByteCount
* An entry of zero means there are no stored hashes having the given partial hash (LSBs)
* If the MSB of an entry us unset, then the entry is the only matching hash's index PLUS ONE
* If the MSB is set, then there are multiple stored hashes matching the given partial hash
* In this case, the lower bits are a chunk index PLUS ONE and you must then consult the collisions file, starting at that chunk
### 3. Collisions file, FileName.cls
* This is a linked list of chunks, each containing collisionsPerChunk entries, followed by a link
* Each chunk pertains to a collision; that is, one partial hash that matches multiple stored full hashes in the Hashes file
* If a chunk overflows, it will end with a link index to another chunk that continues documenting the collision 
* Each entry in a chunk, as for the Lookup file, and including the final link, is entryByteCount bytes long
* Each entry in a chunk is either a hash index PLUS ONE, or a zero
* A non-zero entry is the hash index PLUS ONE for one of the hashes that matches the partial hash
* We must continue through the chunk, and any linked chunks, as there may be more than one full hash in the collision
* If we encounter a zero entry, then we have encountered all the matching hashes; we must check which full hash matches
* Once we've traversed collisionsPerChunk entries in a chunk, we must read and follow the link that follows them
* A link is a chunk index PLUS ONE, and we can follow it by calculating a file Seek
* A link of zero represents the end of the linked list, so all potential hash indices have been encountered