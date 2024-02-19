package artprojectacid

type JsonBlockType struct {
	Height     int
	SizeBytes  int
	MedianTime int
}

type JsonBlockArray struct {
	Blocks []JsonBlockType
}
