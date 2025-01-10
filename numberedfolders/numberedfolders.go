package numberedfolders

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

// "bigit" means "big digit". For example "456" when a big digit is 3 decimal digits

type NumberedFolders interface {
	NumberToFoldersAndFile(num int64) (string, string, int64)
}

type numberedFolders struct {
	digitsPerBigitFile     int
	digitsPerBigitFolder   int
	bigitStringsFile       []string
	bigitStringsFolder     []string
	bigitPlaceholderFile   string
	bigitPlaceholderFolder string
}

func bigitStrings(decimalDigitsPerBigit int) []string {
	format := "%0" + strconv.Itoa(decimalDigitsPerBigit) + "d"
	count := int(math.Pow10(decimalDigitsPerBigit))
	var result []string
	for i := 0; i < count; i++ {
		str := fmt.Sprintf(format, i)
		result = append(result, str)
	}
	return result
}

func NewNumberedFolders(fileDigits int, folderDigits int) NumberedFolders {
	numbered := numberedFolders{}
	numbered.digitsPerBigitFile = fileDigits
	numbered.bigitStringsFile = bigitStrings(fileDigits)
	numbered.digitsPerBigitFolder = folderDigits
	numbered.bigitStringsFolder = bigitStrings(folderDigits)
	numbered.bigitPlaceholderFile = ""
	for i := 0; i < fileDigits; i++ {
		numbered.bigitPlaceholderFile += "x"
	}
	numbered.bigitPlaceholderFolder = ""
	for i := 0; i < folderDigits; i++ {
		numbered.bigitPlaceholderFolder += "x"
	}
	return &numbered
}

func (nf *numberedFolders) NumberToFoldersAndFile(num int64) (folders string, filename string, filenum int64) {
	bigitCount := 0
	var bigitArray []string
	// bigitArray[0] will be the least significant bigit
	// First, gather the big digits (bigits)
	// Like a do..while loop:
	for keepGoing := true; keepGoing; keepGoing = num > 0 {
		digitsPerBigit := nf.digitsPerBigitFolder
		if bigitCount == 0 && nf.digitsPerBigitFile > 0 {
			digitsPerBigit = nf.digitsPerBigitFile
		}
		entriesPerBigit := int64(math.Pow10(digitsPerBigit))

		bigit := num % entriesPerBigit
		var bigitStr string
		if bigitCount == 0 && nf.digitsPerBigitFile > 0 {
			bigitStr = nf.bigitStringsFile[bigit]
		} else {
			bigitStr = nf.bigitStringsFolder[bigit]
		}
		bigitArray = append(bigitArray, bigitStr)

		// Move on ready for any bigger digit position
		num = num - bigit
		num /= entriesPerBigit
		bigitCount++
	}

	// Second, construct the filename
	filename = ""
	for bigitNum := 0; bigitNum < bigitCount; bigitNum++ {
		bigit := bigitArray[bigitNum]

		if bigitNum == 0 && nf.digitsPerBigitFile > 0 {
			// This is the least significant bigit, and each file holds many entries
			bigit = nf.bigitPlaceholderFile
		}
		// remember the first bigit in the array is least significant (ie at the end of the filename)
		filename = bigit + filename
	}

	// Third, construct the nested folder path
	folderPath := ""
	if nf.digitsPerBigitFile == 0 {
		// Case where there is a unique file for each num

		// For a particular num, and when we have n bigits, there are n folders nested.
		// This outer loop loops through these n folders which contain each other, starting with the outer folder
		for folderLevel := bigitCount - 1; folderLevel >= 0; folderLevel-- {
			// So for a three bigit number (for example), folderLevel goes 2,1,0
			// For a three bigit number (for example),
			// every folder in the nested structure has three bigit placeholders,
			// some bigits will be something like "xx", other bigits will be specific numbers like "75"

			// For this particular folder level, construct the folder name
			folderName := ""
			// We loop through the available digits (most significant digit first, for example 2)
			for bigitSignificance := bigitCount - 1; bigitSignificance >= 0; bigitSignificance-- {
				// The least significant big-digit never goes into a foldername; it only goes into the filename
				if bigitSignificance == 0 {
					folderName += nf.bigitPlaceholderFolder // Without a separator, because it's at the end
				} else if bigitSignificance <= folderLevel {
					folderName += nf.bigitPlaceholderFolder
				} else {
					folderName += bigitArray[bigitSignificance]
				}
			}
			// We have a folder name, add it to the path
			folderPath += folderName
			if folderLevel > 0 {
				folderPath += string(os.PathSeparator)
			}
		}
		fileNum, _ := strconv.Atoi(filename)
		return folderPath, filename, int64(fileNum)
	} else {
		// Case where multiple values of num are represented in the same file

		// For a particular num, and when we have n bigits, there are n MINUS ONE folders nested.
		// This outer loop loops through these n-1 folders which contain each other, starting with the outer folder
		for folderLevel := bigitCount - 1; folderLevel > 0; folderLevel-- {
			// So for a three bigit number (for example), folderLevel goes 2,1
			// For a three bigit number (for example),
			// every folder in the nested structure has three bigit placeholders,
			// some bigits will be something like "xx", other bigits will be specific numbers like "75"

			// For this particular folder level, construct the folder name
			folderName := ""
			// We loop through the available digits (most significant digit first, for example 2)
			for bigitSignificance := bigitCount - 1; bigitSignificance >= 0; bigitSignificance-- {
				// The least significant big-digit never goes into a foldername; it is an index into a single file
				if bigitSignificance == 0 {
					folderName += nf.bigitPlaceholderFile // Without a separator, because it's at the end
				} else if bigitSignificance <= folderLevel {
					folderName += nf.bigitPlaceholderFolder
				} else {
					folderName += bigitArray[bigitSignificance]
				}
			}
			// We have a folder name, add it to the path
			folderPath += folderName
			if folderLevel > 1 {
				folderPath += string(os.PathSeparator)
			}
		}
		fileNameForNum := filename
		fileNameForNum = strings.ReplaceAll(fileNameForNum, "x", "0")
		fileNum, _ := strconv.Atoi(fileNameForNum)
		return folderPath, filename, int64(fileNum)
	}
}
