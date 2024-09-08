package indexedhashes

import (
	"encoding/binary"
	"errors"
	"github.com/KitchenMishap/pudding-shed/memfile"
	"log"
	"os"
)

type HashStore struct {
	partialHashBitCount int64
	entryByteCount      int64
	collisionsPerChunk  int64
	hashesFile          *BasicHashStore
	lookupsFile         memfile.SparseLookupFile
	collisionsFile      *os.File // wordfile.ReadWriteAtWordCounter someday?
}

func NewHashStore(partialHashBitCount int64, entryByteCount int64, collisionsPerChunk int64,
	hashesFile *BasicHashStore,
	lookupsFile memfile.SparseLookupFile, collisionsFile *os.File) *HashStore {
	result := HashStore{}
	result.partialHashBitCount = partialHashBitCount
	result.entryByteCount = entryByteCount
	result.collisionsPerChunk = collisionsPerChunk
	result.hashesFile = hashesFile
	result.lookupsFile = lookupsFile
	result.collisionsFile = collisionsFile
	return &result
}

const ZEROBUF = 32 // Arbitrary number. Should be enough (we do check)

func (hs *HashStore) AppendHash(hash *Sha256) (int64, error) {
	// Write to hashes file
	index, err := hs.hashesFile.AppendHash(hash)
	if err != nil {
		// AppendHash() has already printed error
		return -1, err
	}

	hashV := binary.LittleEndian.Uint64(hash[0:8])
	valnum := index

	NBYTES := hs.entryByteCount
	MSB := int64(0x80) << (8 * (NBYTES - 1))
	NBITS := hs.partialHashBitCount
	NMASK := (int64(1) << NBITS) - 1
	CPC := hs.collisionsPerChunk

	if valnum >= MSB {
		err := errors.New("fatal: size of hashes file is now too large to index into")
		log.Println("AppendHash(): FATAL: Size of hashes file is now too large to index into")
		return -1, err
	}

	// Encode valnum+1 as NBYTES bytes
	var valnumBytes [8]byte
	binary.LittleEndian.PutUint64(valnumBytes[0:8], uint64(valnum+1)) // MUST be LittleEndian

	// Is something already stored in lookup[partialKey] ?
	partialKey := int64(hashV) & NMASK
	lookupBytes := [8]byte{0, 0, 0, 0, 0, 0, 0, 0}
	_, err = hs.lookupsFile.ReadAt(lookupBytes[0:NBYTES], partialKey*NBYTES) // Assumes LittleEndian
	if err != nil {
		log.Println(err)
		log.Println("AppendHash(): Couldn't ReadAt() from lookups file")
		return -1, err
	}
	lookup := int64(binary.LittleEndian.Uint64(lookupBytes[0:8]))
	if lookup == 0 {
		// No, vacant, put it in there
		_, err := hs.lookupsFile.WriteAt(valnumBytes[0:NBYTES], partialKey*NBYTES)
		if err != nil {
			log.Println(err)
			log.Println("AppendHash(): Couldn't WriteAt() into lookups file")
			return -1, err
		}
	} else if lookup&MSB == 0 {
		// Yes, MSB clear, so we have a new first collision

		//println("Benign code coverage check(1 of 4): Collision occurred and handled")

		// Need a new linked list
		// Append a chunk with the two colliding entries
		filestat, err := hs.collisionsFile.Stat()
		if err != nil {
			log.Println(err)
			log.Println("AppendHash() Couldn't call Stat() on collisions file")
			return -1, err
		}
		availableChunks := filestat.Size() / (NBYTES * (CPC + 1))
		if availableChunks+1 >= MSB {
			err = errors.New("fatal: size of collisions file is now too large to index into")
			log.Println("AppendHash(): FATAL: Size of collisions file is now too large to index into")
			return -1, err
		}
		chunkIndexBytes := [8]byte{0, 0, 0, 0, 0, 0, 0, 0}
		binary.LittleEndian.PutUint64(chunkIndexBytes[0:8], uint64((availableChunks+1)|MSB))
		_, err = hs.lookupsFile.WriteAt(chunkIndexBytes[0:NBYTES], partialKey*NBYTES)
		if err != nil {
			log.Println(err)
			log.Println("AppendHash(): Could not WriteAt() into lookups file")
			return -1, err
		}

		// Append lookup to collisions file
		_, err = hs.collisionsFile.WriteAt(lookupBytes[0:NBYTES], filestat.Size())
		if err != nil {
			log.Println(err)
			log.Println("AppendHash: Could not WriteAt() into collisions file")
			return -1, err
		}
		// Append valnum+1 to collisions file
		_, err = hs.collisionsFile.WriteAt(valnumBytes[0:NBYTES], filestat.Size()+NBYTES)
		if err != nil {
			log.Println(err)
			log.Println("AppendHash: Could not WriteAt() into collisions file")
			return -1, err
		}
		var zeroBytes [ZEROBUF]byte
		_, err = hs.collisionsFile.WriteAt(zeroBytes[0:(CPC+1-2)*NBYTES], filestat.Size()+NBYTES*2)
		if err != nil {
			log.Println(err)
			log.Println("AppendHash: Could not WriteAt() into collisions file")
			return -1, err
		}
	} else {
		// MSB set, so append to an existing linked list in collisions file

		//println("Benign code coverage check(2 of 4): Second collision occurred and handled")

		lookup = lookup & ^MSB
		chunk := lookup - 1
		j := int64(0)
		var colliderBytes [8]byte
		_, err := hs.collisionsFile.ReadAt(colliderBytes[0:NBYTES], (chunk*(CPC+1)+j)*NBYTES)
		if err != nil {
			log.Println(err)
			log.Println("AppendHash(): Could not ReadAt() from collisions file")
			return -1, err
		}
		collider := int64(binary.LittleEndian.Uint64(colliderBytes[:]))
		for collider != 0 {
			// Need to skip past occupied slots
			j++
			if j == CPC {
				// No more slots in chunk
				j = 0
				// Read linked list offset
				var nextChunkBytes [8]byte
				_, err = hs.collisionsFile.ReadAt(nextChunkBytes[0:NBYTES], (chunk*(CPC+1)+CPC)*NBYTES)
				if err != nil {
					log.Println(err)
					log.Println("AppendHash(): Could not ReadAt() from collisions file")
					return -1, err
				}
				nextChunk := int64(binary.LittleEndian.Uint64(nextChunkBytes[:]))
				if nextChunk == 0 {
					// End of linked list

					//println("Benign code coverage check(3 of 4): Appending second chunk to linked list in collisions file")

					// Append an empty chunk
					filestat, err := hs.collisionsFile.Stat()
					if err != nil {
						log.Println(err)
						log.Println("AppendHash(): Could not call Stat() on collisions file")
						return -1, err
					}
					availableChunks := filestat.Size() / (NBYTES * (CPC + 1))
					// Append a chunk of zeroes to the collisions file
					var zeroBytes [ZEROBUF]byte
					_, err = hs.collisionsFile.WriteAt(zeroBytes[0:NBYTES*(CPC+1)], filestat.Size())
					if err != nil {
						log.Println(err)
						log.Println("AppendHash(): Could not call WriteAt() on collisions file")
						return -1, err
					}
					// Link to the new empty chunk
					binary.LittleEndian.PutUint64(nextChunkBytes[:], uint64(availableChunks+1))
					_, err = hs.collisionsFile.WriteAt(nextChunkBytes[0:NBYTES], (chunk*(CPC+1)+CPC)*NBYTES)
					if err != nil {
						log.Println(err)
						log.Println("AppendHash(): Could not call WriteAt() on collisions file")
						return -1, err
					}
					chunk = availableChunks
					collider = 0
				} else {
					// End of chunk but not end of linked list
					// println("Benign code coverage check(4 of 4): Reached end of chunk that's not end of linked list")

					chunk = nextChunk - 1
					_, err = hs.collisionsFile.ReadAt(colliderBytes[0:NBYTES], (chunk*(CPC+1)+j)*NBYTES)
					if err != nil {
						log.Println(err)
						log.Println("AppendHash(): Could not call ReadAt() on collisionsFile")
						return -1, err
					}
					collider = int64(binary.LittleEndian.Uint64(colliderBytes[:]))
				}
			} else {
				_, err = hs.collisionsFile.ReadAt(colliderBytes[0:NBYTES], (chunk*(CPC+1)+j)*NBYTES)
				if err != nil {
					log.Println(err)
					log.Println("AppendHash(): Could not call ReadAt() on collisionsFile")
					return -1, err
				}
				collider = int64(binary.LittleEndian.Uint64(colliderBytes[:]))
			}
		} // for collider != 0

		// Write collisions[chunk][j] = valnum + 1
		_, err = hs.collisionsFile.WriteAt(valnumBytes[0:NBYTES], (chunk*(CPC+1)+j)*NBYTES)
		if err != nil {
			log.Println(err)
			log.Println("AppendHash(): Could not call WriteAt() on collisionsFile")
			return -1, err
		}
	}
	return index, nil
}

// IndexOfHash -1 indicates "Not Present" but error will be nil if that's all that is wrong
func (hs *HashStore) IndexOfHash(hash *Sha256) (int64, error) {
	NBYTES := hs.entryByteCount
	MSB := int64(0x80) << (8 * (NBYTES - 1))
	NBITS := hs.partialHashBitCount
	NMASK := (int64(1) << NBITS) - 1
	CPC := hs.collisionsPerChunk

	// 64 LSBs of hash
	// We prefer to use LSBs as MSBs of block hashes have leading zeros due to mining proof of work
	// Hashes are stored as very big LittleEndian 256 values in .hsh file
	// Therefore the 64 LSBs are in the FIRST EIGHT bytes
	v := int64(binary.LittleEndian.Uint64(hash[0:8]))

	lookupIndex := v & NMASK

	// First try the lookup file
	lookupBytes := [8]byte{0, 0, 0, 0, 0, 0, 0, 0}
	_, er := hs.lookupsFile.ReadAt(lookupBytes[0:NBYTES], lookupIndex*NBYTES)
	if er != nil {
		log.Println(er)
		log.Println("IndexOfHash(): Could not call ReadAt() on lookups file")
		return -1, er
	}
	lookup := int64(binary.LittleEndian.Uint64(lookupBytes[0:8]))

	if (lookup & MSB) == 0 {
		if lookup == 0 {
			// No record whatsoever
			return -1, nil
		}

		// We have a potential index, no collisions
		var potentialHash Sha256
		er := hs.GetHashAtIndex(lookup-1, &potentialHash)
		if er != nil {
			// GetHashAtIndex() has printed error
			return -1, er
		}
		if potentialHash == *hash {
			// Found directly in lookup file, no collisions
			return lookup - 1, nil
		} else {
			// The only lead we had did not lead to the required hash
			return -1, nil
		}
	} else {
		// Collision of the partial hash. Refer to the collisions file
		link := lookup & ^MSB
		var collision int64
		for link != 0 {
			for j := int64(0); j <= CPC; j++ {
				// Read a collision (j<CPC) or a link (j==CPC)
				// Note lookup is a chunk index +1
				collisionBytes := [8]byte{0, 0, 0, 0, 0, 0, 0, 0}
				_, er := hs.collisionsFile.ReadAt(collisionBytes[0:NBYTES], ((link-1)*(CPC+1)+j)*NBYTES)
				if er != nil {
					log.Println(er)
					log.Println("IndexOfHash(): Could not call ReadAt() on collisions file")
					return -1, er
				}
				collision = int64(binary.LittleEndian.Uint64(collisionBytes[0:8]))
				if collision == 0 {
					// Either a zero entry (end of entries in chunk)
					// or a zero link (end of linked list of chunks)
					// Either way we did not find our hash (not an error)
					return -1, nil
				}
				if j < CPC {
					// We have a potential index, from the collisions file
					var potentialHash Sha256
					er = hs.GetHashAtIndex(collision-1, &potentialHash)
					if er != nil {
						// GetHashAtIndex() has printed error
						return -1, er
					}
					if potentialHash == *hash {
						// Hash found using collisions file
						return collision - 1, nil
					} else {
						// Didn't match. Keep looking. Next j
					}
				} else {
					// j == CPC
					// collision is a link to next collisions chunk
					// Follow the link to Keep looking
					link = collision
				}
			} // Next j
		}
		// End of linked list of chunks. Not an error
		return -1, nil
	}
}

func (hs *HashStore) GetHashAtIndex(index int64, hash *Sha256) error {
	// This call will print any error that occurs
	return hs.hashesFile.GetHashAtIndex(index, hash)
}

func (hs *HashStore) CountHashes() (int64, error) {
	// This call will print any error that occurs
	count, err := hs.hashesFile.CountHashes()
	return count, err
}

func (hs *HashStore) Close() error {
	err := hs.hashesFile.Close()
	if err != nil {
		// Close() has printed any error
		return err
	}
	err = hs.lookupsFile.Close()
	if err != nil {
		log.Println(err)
		log.Println("HashStore::Close() could not call Close() on lookups file")
		return err
	}
	err = hs.collisionsFile.Close()
	if err != nil {
		log.Println(err)
		log.Println("HashStore::Close() could not call Close() on collisions file")
	}
	return err
}

/*
func (hs *HashStore) WholeFileAsInt32() ([]uint32, error) {
	array, err := hs.hashesFile.WholeFileAsInt32()
	// WholeFileAsInt32() has printed any error that occurred
	return array, err
}
*/

func (hs *HashStore) Sync() error {
	//err := hs.lookupsFile.Sync()	Really big so we'd rather not!
	//if err != nil {return err}
	err := hs.collisionsFile.Sync()
	if err != nil {
		return err
	}
	err = hs.hashesFile.Sync()
	if err != nil {
		return err
	}
	return nil
}

func (hs *HashStore) SelfTest() error {
	println("Hash Store is self testing...")
	println("This test will correctly FAIL if blocks 91812,91842")
	println("are included!")
	count, err := hs.CountHashes()
	if err != nil {
		return err
	}
	for i := int64(0); i < count; i++ {
		hash := Sha256{}
		err := hs.GetHashAtIndex(i, &hash)
		if err != nil {
			return err
		}
		j, err := hs.IndexOfHash(&hash)
		if err != nil {
			return err
		}
		if j == -1 {
			println("Hash in question: ", hashBinToHexString((*[32]byte)(&hash)))
			return errors.New("Hash which is present could not be found!")
		}
		if j != i {
			println("Hash in question: ", hashBinToHexString((*[32]byte)(&hash)))
			hash2 := Sha256{}
			err = hs.GetHashAtIndex(i, &hash2)
			if err != nil {
				return err
			}
			println("Hash at ", i, ": ", hashBinToHexString((*[32]byte)(&hash2)))
			hash3 := Sha256{}
			err = hs.GetHashAtIndex(j, &hash3)
			if err != nil {
				return err
			}
			println("Hash at ", j, ": ", hashBinToHexString((*[32]byte)(&hash3)))
			println("Note that this test highlights the genuine issue that")
			println("the coinbase transactions of blocks 91812 and 91842")
			println("have identical hashes! (identical transactions, BIP37)")
			return errors.New("Hash detected at multiple indices!")
		}
	}
	println("...passed")
	return nil
}
