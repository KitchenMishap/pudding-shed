package wordfile

import (
	"log"
	"os"
	"testing"
)

func TestConcreteWordFileCreatorNonOptimized(t *testing.T) {
	helperConcreteWordFileCreator(t, false)
}

func TestConcreteWordFileCreatorAppendOptimized(t *testing.T) {
	helperConcreteWordFileCreator(t, true)
}

func helperConcreteWordFileCreator(t *testing.T, appendOptimize bool) {
	var creator1 = NewConcreteWordFileCreator("CreatorTesting1", "Temp_Testing", 1, appendOptimize)
	var creator8 = NewConcreteWordFileCreator("CreatorTesting8", "Temp_Testing", 8, appendOptimize)

	// Delete the wordfile manually from any previous test
	os.Remove("Temp_Testing\\CreatorTesting1.int")

	if creator1.WordFileExists() {
		t.Error("WordFileExists() should return false")
	}

	err := creator1.CreateWordFile()
	if err != nil {
		t.Error("Coudn't create wordfile")
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
