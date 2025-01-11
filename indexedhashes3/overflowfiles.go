package indexedhashes3

import (
	"github.com/KitchenMishap/pudding-shed/numberedfolders"
	"os"
)

type overflowFiles struct {
	folderPath      string
	numberedFolders numberedfolders.NumberedFolders
}

func newOverflowFiles(folderPath string, p *HashIndexingParams) *overflowFiles {
	result := overflowFiles{}
	result.folderPath = folderPath
	nf := numberedfolders.NewNumberedFolders(0, p.digitsPerNumberedFolder)
	result.numberedFolders = nf
	return &result
}

func (of *overflowFiles) overflowFolderpathFilepath(bn binNum) (folderpath string, filepath string) {
	sep := string(os.PathSeparator)
	folders, filename, _ := of.numberedFolders.NumberToFoldersAndFile(int64(bn))
	folderpath = of.folderPath + sep + "BinOverflows" + sep + folders
	filepath = folderpath + sep + filename + ".bes"
	return folderpath, filepath
}
