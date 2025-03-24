package memfile

import (
	"bufio"
	"os"
	"testing"
)

func TestBufferedWriteRead(t *testing.T) {
	f, err := os.Create("Buffered")
	if err != nil {
		t.Error(err)
	}
	b := bufio.NewWriter(f)

	// Write lots of 10 digit strings
	const count = 100000
	for i := 0; i < count; i++ {
		b.WriteString("0123456789")
	}

	// Flush and sync it to disk
	err = b.Flush()
	if err != nil {
		t.Error(err)
	}
	err = f.Sync()
	if err != nil {
		t.Error(err)
	}

	// Read back the last 10 digits
	byts := make([]byte, 10)
	_, err = f.ReadAt(byts, (count-1)*10)
	if err != nil {
		t.Error(err)
	}
}
