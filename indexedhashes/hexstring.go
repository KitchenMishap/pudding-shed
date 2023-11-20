package indexedhashes

import (
	"encoding/hex"
	"errors"
)

// Yukky global
var hexLookup = buildHexLookup()

func hashHexToBin(hexAscii string) ([32]byte, error) {
	res, err := hashHexToBinInternal(hexAscii, hexLookup)
	return res, err
}

func hashHexToSha256(hexAscii string, sha256 *Sha256) error {
	err := hashHexToSha256Internal(hexAscii, hexLookup, sha256)
	return err
}

// Attempt to make this quick
func hashHexToBinInternal(hexAscii string, lookup *[256 * 256]byte) ([32]byte, error) {
	if len(hexAscii) != 64 {
		errors.New("hash must be 64 hex digits")
	}
	result := [32]byte{
		// To get the ascii of a rune from a string, we cast to an int
		lookup[int(hexAscii[63])|(int(hexAscii[62])<<8)],
		lookup[int(hexAscii[61])|(int(hexAscii[60])<<8)],
		lookup[int(hexAscii[59])|(int(hexAscii[58])<<8)],
		lookup[int(hexAscii[57])|(int(hexAscii[56])<<8)],
		lookup[int(hexAscii[55])|(int(hexAscii[54])<<8)],
		lookup[int(hexAscii[53])|(int(hexAscii[52])<<8)],
		lookup[int(hexAscii[51])|(int(hexAscii[50])<<8)],
		lookup[int(hexAscii[49])|(int(hexAscii[48])<<8)],
		lookup[int(hexAscii[47])|(int(hexAscii[46])<<8)],
		lookup[int(hexAscii[45])|(int(hexAscii[44])<<8)],
		lookup[int(hexAscii[43])|(int(hexAscii[42])<<8)],
		lookup[int(hexAscii[41])|(int(hexAscii[40])<<8)],
		lookup[int(hexAscii[39])|(int(hexAscii[38])<<8)],
		lookup[int(hexAscii[37])|(int(hexAscii[36])<<8)],
		lookup[int(hexAscii[35])|(int(hexAscii[34])<<8)],
		lookup[int(hexAscii[33])|(int(hexAscii[32])<<8)],
		lookup[int(hexAscii[31])|(int(hexAscii[30])<<8)],
		lookup[int(hexAscii[29])|(int(hexAscii[28])<<8)],
		lookup[int(hexAscii[27])|(int(hexAscii[26])<<8)],
		lookup[int(hexAscii[25])|(int(hexAscii[24])<<8)],
		lookup[int(hexAscii[23])|(int(hexAscii[22])<<8)],
		lookup[int(hexAscii[21])|(int(hexAscii[20])<<8)],
		lookup[int(hexAscii[19])|(int(hexAscii[18])<<8)],
		lookup[int(hexAscii[17])|(int(hexAscii[16])<<8)],
		lookup[int(hexAscii[15])|(int(hexAscii[14])<<8)],
		lookup[int(hexAscii[13])|(int(hexAscii[12])<<8)],
		lookup[int(hexAscii[11])|(int(hexAscii[10])<<8)],
		lookup[int(hexAscii[9])|(int(hexAscii[8])<<8)],
		lookup[int(hexAscii[7])|(int(hexAscii[6])<<8)],
		lookup[int(hexAscii[5])|(int(hexAscii[4])<<8)],
		lookup[int(hexAscii[3])|(int(hexAscii[2])<<8)],
		lookup[int(hexAscii[1])|(int(hexAscii[0])<<8)]}
	return result, nil
}

// Attempt to make this quick
func hashHexToSha256Internal(hexAscii string, lookup *[256 * 256]byte, sha256 *Sha256) error {
	if len(hexAscii) != 64 {
		return errors.New("hash should be 64 hex digits")
	}
	// To get the ascii of a rune from a string, we cast to an int
	(*sha256)[0] = lookup[int(hexAscii[63])|(int(hexAscii[62])<<8)]
	(*sha256)[1] = lookup[int(hexAscii[61])|(int(hexAscii[60])<<8)]
	(*sha256)[2] = lookup[int(hexAscii[59])|(int(hexAscii[58])<<8)]
	(*sha256)[3] = lookup[int(hexAscii[57])|(int(hexAscii[56])<<8)]
	(*sha256)[4] = lookup[int(hexAscii[55])|(int(hexAscii[54])<<8)]
	(*sha256)[5] = lookup[int(hexAscii[53])|(int(hexAscii[52])<<8)]
	(*sha256)[6] = lookup[int(hexAscii[51])|(int(hexAscii[50])<<8)]
	(*sha256)[7] = lookup[int(hexAscii[49])|(int(hexAscii[48])<<8)]
	(*sha256)[8] = lookup[int(hexAscii[47])|(int(hexAscii[46])<<8)]
	(*sha256)[9] = lookup[int(hexAscii[45])|(int(hexAscii[44])<<8)]
	(*sha256)[10] = lookup[int(hexAscii[43])|(int(hexAscii[42])<<8)]
	(*sha256)[11] = lookup[int(hexAscii[41])|(int(hexAscii[40])<<8)]
	(*sha256)[12] = lookup[int(hexAscii[39])|(int(hexAscii[38])<<8)]
	(*sha256)[13] = lookup[int(hexAscii[37])|(int(hexAscii[36])<<8)]
	(*sha256)[14] = lookup[int(hexAscii[35])|(int(hexAscii[34])<<8)]
	(*sha256)[15] = lookup[int(hexAscii[33])|(int(hexAscii[32])<<8)]
	(*sha256)[16] = lookup[int(hexAscii[31])|(int(hexAscii[30])<<8)]
	(*sha256)[17] = lookup[int(hexAscii[29])|(int(hexAscii[28])<<8)]
	(*sha256)[18] = lookup[int(hexAscii[27])|(int(hexAscii[26])<<8)]
	(*sha256)[19] = lookup[int(hexAscii[25])|(int(hexAscii[24])<<8)]
	(*sha256)[20] = lookup[int(hexAscii[23])|(int(hexAscii[22])<<8)]
	(*sha256)[21] = lookup[int(hexAscii[21])|(int(hexAscii[20])<<8)]
	(*sha256)[22] = lookup[int(hexAscii[19])|(int(hexAscii[18])<<8)]
	(*sha256)[23] = lookup[int(hexAscii[17])|(int(hexAscii[16])<<8)]
	(*sha256)[24] = lookup[int(hexAscii[15])|(int(hexAscii[14])<<8)]
	(*sha256)[25] = lookup[int(hexAscii[13])|(int(hexAscii[12])<<8)]
	(*sha256)[26] = lookup[int(hexAscii[11])|(int(hexAscii[10])<<8)]
	(*sha256)[27] = lookup[int(hexAscii[9])|(int(hexAscii[8])<<8)]
	(*sha256)[28] = lookup[int(hexAscii[7])|(int(hexAscii[6])<<8)]
	(*sha256)[29] = lookup[int(hexAscii[5])|(int(hexAscii[4])<<8)]
	(*sha256)[30] = lookup[int(hexAscii[3])|(int(hexAscii[2])<<8)]
	(*sha256)[31] = lookup[int(hexAscii[1])|(int(hexAscii[0])<<8)]
	return nil
}

func buildHexLookup() *[256 * 256]byte {
	var result [256 * 256]byte
	for i := 0; i < 256*256; i++ {
		lsbAscii := byte(i & 0xFF)
		msbAscii := byte((i & 0xFF00) >> 8)
		lsbNibble := asciiToNibble(lsbAscii)
		msbNibble := asciiToNibble(msbAscii)
		result[i] = (msbNibble << 4) | lsbNibble
	}
	return &result
}

func asciiToNibble(ascii byte) byte {
	if ascii >= 0x30 && ascii <= 0x39 {
		return ascii - 0x30 // '0' - '9'
	}
	if ascii >= 0x41 && ascii <= 0x46 {
		return ascii - (0x41 - 0x0A) // 'A' - 'F'
	}
	if ascii >= 0x61 && ascii <= 0x66 {
		return ascii - (0x61 - 0x0a) // 'a' - 'f'
	}
	return 0
}

func hashBinToHexString(hashBin *[32]byte) string {
	var reversed [32]byte
	for i := 0; i < 32; i++ {
		reversed[i] = hashBin[31-i]
	}
	return hex.EncodeToString(reversed[0:32])
}

func hashSha256ToHexString(hash *Sha256) string {
	var reversed [32]byte
	for i := 0; i < 32; i++ {
		reversed[i] = hash[31-i]
	}
	return hex.EncodeToString(reversed[0:32])
}
