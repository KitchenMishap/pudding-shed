package numberedfolders

import (
	"fmt"
	"os"
	"testing"
)

func Test_OneDigitForFile(t *testing.T) {
	nf := NewNumberedFolders(1, 2)

	folders, filename := nf.NumberToFoldersAndFile(0)
	if folders != "" {
		t.Fail()
	}
	if filename != "x" {
		t.Fail()
	}

	folders, filename = nf.NumberToFoldersAndFile(1)
	if folders != "" {
		t.Fail()
	}
	if filename != "x" {
		t.Fail()
	}

	folders, filename = nf.NumberToFoldersAndFile(10)
	if folders != "xxx" {
		t.Fail()
	}
	if filename != "01x" {
		t.Fail()
	}

	folders, filename = nf.NumberToFoldersAndFile(11)
	if folders != "xxx" {
		t.Fail()
	}
	if filename != "01x" {
		t.Fail()
	}

	folders, filename = nf.NumberToFoldersAndFile(234)
	if folders != "xxx" {
		t.Fail()
	}
	if filename != "23x" {
		t.Fail()
	}

	folders, filename = nf.NumberToFoldersAndFile(3456)
	if folders != "xxxxx\\03xxx" {
		t.Fail()
	}
	if filename != "0345x" {
		t.Fail()
	}
}

func Test_NoDigitsForFile(t *testing.T) {
	nf := NewNumberedFolders(0, 2)

	folders, filename := nf.NumberToFoldersAndFile(0)
	if folders != "xx" {
		t.Fail()
	}
	if filename != "00" {
		t.Fail()
	}

	folders, filename = nf.NumberToFoldersAndFile(1)
	if folders != "xx" {
		t.Fail()
	}
	if filename != "01" {
		t.Fail()
	}

	folders, filename = nf.NumberToFoldersAndFile(10)
	if folders != "xx" {
		t.Fail()
	}
	if filename != "10" {
		t.Fail()
	}

	folders, filename = nf.NumberToFoldersAndFile(11)
	if folders != "xx" {
		t.Fail()
	}
	if filename != "11" {
		t.Fail()
	}

	folders, filename = nf.NumberToFoldersAndFile(234)
	if folders != "xxxx\\02xx" {
		t.Fail()
	}
	if filename != "0234" {
		t.Fail()
	}

	folders, filename = nf.NumberToFoldersAndFile(3456)
	if folders != "xxxx\\34xx" {
		t.Fail()
	}
	if filename != "3456" {
		t.Fail()
	}
}

func Test_CreateFoldersFiles(t *testing.T) {
	os.RemoveAll("numfold2-1")
	nf := NewNumberedFolders(1, 2)
	prefix := "numfold2-1" + string(os.PathSeparator)
	for i := int64(0); i < int64(123456); i += 57 {
		folders, filename := nf.NumberToFoldersAndFile(i)
		folders = prefix + folders
		filePath := folders + string(os.PathSeparator) + filename + ".txt"
		f, err := os.OpenFile(filePath, os.O_APPEND, 0755)
		if err != nil {
			os.MkdirAll(folders, 0755)
			f, _ = os.Create(filePath)
		}
		str := fmt.Sprintf("%d", i)
		f.WriteString(str + "\n")
		println(str + "\t" + folders + "\t" + filename)
		f.Close()
	}

	os.RemoveAll("numfold2-0")
	nf = NewNumberedFolders(0, 2)
	prefix = "numfold2-0" + string(os.PathSeparator)
	for i := int64(0); i < int64(20000); i += 31 {
		folders, filename := nf.NumberToFoldersAndFile(i)
		folders = prefix + folders
		filePath := folders + string(os.PathSeparator) + filename + ".txt"
		f, err := os.Create(filePath)
		if err != nil {
			os.MkdirAll(folders, 0755)
			f, _ = os.Create(filePath)
		}
		str := fmt.Sprintf("%d", i)
		f.WriteString(str + "\n")
		println(str + "\t" + folders + "\t" + filename)
		f.Close()
	}
}
