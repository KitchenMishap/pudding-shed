package artprojectacid

type JsonBlockType struct {
	Height      int
	SizeBytes   int
	MedianTime  int
	ColourByte0 int
	ColourByte1 int
	ColourByte2 int
}

type JsonBlockArray struct {
	Blocks []JsonBlockType
}
