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

func (of *overflowFiles) overflowFilepath(bn binNum) string {
	sep := string(os.PathSeparator)
	folders, filename, _ := of.numberedFolders.NumberToFoldersAndFile(int64(bn))
	return of.folderPath + sep + "BinOverflows" + sep + folders + sep + filename + ".ovf"
}
