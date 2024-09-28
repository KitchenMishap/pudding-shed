package indexedhashes

func CreateHashIndexFiles() error {
	creator := newUniformHashStoreCreatorPrivate(1000000000, "E:/Data/Hashes", "TransHashes", 2)
	err := creator.CreateHashStore()
	if err != nil {
		return err
	}
	hs, err := creator.openHashStorePrivate()
	if err != nil {
		return err
	}
	preloader := UniformHashPreLoader{}
	preloader.uniform = hs
	err = preloader.createEmptyFiles()
	return err
}
