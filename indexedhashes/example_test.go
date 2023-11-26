package indexedhashes

import (
	"crypto/sha256"
	"log"
	"testing"
)

func TestExample(t *testing.T) {

	// 1. Instantiate a ConcreteHashStoreCreator (a factory)
	phbc := int64(30) // Partial hash bit count
	ebc := int64(4)   // Entry byte count
	cpc := int64(3)   // Collisions per chunk
	creator, err := NewConcreteHashStoreCreator("FileName", "Temp_Testing", phbc, ebc, cpc)
	if err != nil {
		log.Println(err)
	}

	// 2. Create the hash store files
	err = creator.CreateHashStore()
	if err != nil {
		log.Println(err)
	}

	// 3. Open the hash store for read/write
	store, err := creator.OpenHashStore()
	if err != nil {
		log.Println(err)
	}

	// 4. Create some hashes of strings using package crypto/sha256
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

	// 5. Store the hashes
	index, err := store.AppendHash(&h0)
	if err != nil {
		log.Println(err)
	}
	index, err = store.AppendHash(&h1)
	if err != nil {
		log.Println(err)
	}

	// 6. Read back the hash at index 0
	hash := Sha256{}
	err = store.GetHashAtIndex(0, &hash)
	if err != nil {
		log.Println(err)
	}

	// 7. Find the index of hash h1
	index, err = store.IndexOfHash(&h1)
	if err != nil {
		log.Println(err)
	}
	println(index)
}
