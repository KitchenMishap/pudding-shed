package wordfile

import (
	"log"
	"os"
	"testing"
)

func TestConcreteWordFileCreator(t *testing.T) {
	var creator1 = NewConcreteWordFileCreator("CreatorTesting1", "", 1)
	var creator8 = NewConcreteWordFileCreator("CreatorTesting8", "", 8)

	// Delete the wordfile manually from any previous test
	os.Remove("CreatorTesting1.int")

	if creator1.WordFileExists() {
		t.Error("WordFileExists() should return false")
	}

	err := creator1.CreateWordFile()
	if err != nil {
		t.Error("Couldn't create wordfile")
	}

	rw, err := creator1.OpenWordFile()
	if err != nil {
		t.Error("Couldn't open wordfile")
	}
	err = rw.WriteWordAt(12, 4)
	if err != nil {
		t.Error("Couldn't write to wordfile")
	}

	count, err := rw.CountWords()
	if err != nil {
		t.Error("Couldn't count words in wordfile")
	}
	if count != 5 {
		t.Error("Writing to entry 4 should give a 5 word file")
	}

	val, err := rw.ReadWordAt(4)
	if err != nil {
		t.Error("Could not read from offset 4")
	}
	if val != 12 {
		t.Error("Expected to read back 12 from offset 4")
	}

	val, err = rw.ReadWordAt(3)
	if err != nil {
		t.Error("Could not read from offset 3")
	}
	if val != 0 {
		t.Error("Expected to read back 0 from offset 3")
	}

	err = rw.Close()
	if err != nil {
		t.Error("Could not close file")
	}

	log.Println("Note: A 'file already closed' error here is a PASS")
	err = rw.WriteWordAt(4, 12)
	if err == nil {
		t.Error("Write to closed file should give error")
	}

	r, err := creator1.OpenWordFileReadOnly()
	if err != nil {
		t.Error("Could not open file for read")
	}

	val, err = r.ReadWordAt(4)
	if err != nil {
		t.Error("Could not read from offset 4")
	}
	if val != 12 {
		t.Error("Expected to read back 12 from offset 4")
	}

	err = r.Close()
	if err != nil {
		t.Error("Could not close file")
	}

	exists := creator8.WordFileExists()
	if exists {
		t.Error("File should not exist")
	}
}
