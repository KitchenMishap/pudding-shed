package chainstorage

import (
	"errors"
	"github.com/KitchenMishap/pudding-shed/memfile"
	"github.com/KitchenMishap/pudding-shed/wordfile"
	"os"
	"path"
)

type ConcreteHashesChainCreator struct {
	blocksFolder       string
	transactionsFolder string
	addressesFolder    string
}

// Check that implements
var _ IAppendableHashesChainFactory = (*ConcreteHashesChainCreator)(nil)

func NewConcreteHashesChainCreator(folder string) *ConcreteHashesChainCreator {
	result := ConcreteHashesChainCreator{}

	sep := string(os.PathSeparator)

	// These folder names are purposefully identical to a
	// subset of those in ConcreteAppendableChainFactory
	result.blocksFolder = path.Join(folder, "Blocks"+sep+"Hashes")
	result.transactionsFolder = path.Join(folder, "Transactions"+sep+"Hashes")
	result.addressesFolder = path.Join(folder, "Addresses"+sep+"Hashes")

	return &result
}

func (chcc *ConcreteHashesChainCreator) Exists() bool {
	// Exists if block hashes file exists
	filename := path.Join(chcc.blocksFolder, "Hashes.hsh")
	file, err := os.Open(filename)
	if err != nil {
		return false
	}
	_ = file.Close()
	return true
}

func (chcc *ConcreteHashesChainCreator) Create() error {
	if chcc.Exists() {
		return errors.New("HashesChain already exists")
	}

	err := os.MkdirAll(chcc.blocksFolder, os.ModePerm)
	if err != nil {
		return err
	}
	err = os.MkdirAll(chcc.transactionsFolder, os.ModePerm)
	if err != nil {
		return err
	}
	err = os.MkdirAll(chcc.addressesFolder, os.ModePerm)
	if err != nil {
		return err
	}

	file, err := os.Create(path.Join(chcc.blocksFolder, "Hashes.hsh"))
	if err != nil {
		return err
	}
	_ = file.Close()

	file, err = os.Create(path.Join(chcc.transactionsFolder, "Hashes.hsh"))
	if err != nil {
		return err
	}
	_ = file.Close()

	file, err = os.Create(path.Join(chcc.addressesFolder, "Hashes.hsh"))
	if err != nil {
		return err
	}
	_ = file.Close()

	return nil
}

func (chcc *ConcreteHashesChainCreator) Open() (IAppendableHashesChain, error) {
	cac, err := chcc.openPrivate()
	return cac, err
}

func (chcc *ConcreteHashesChainCreator) openPrivate() (*concreteHashesChain, error) {
	result := concreteHashesChain{}

	file, err := os.OpenFile(path.Join(chcc.blocksFolder, "Hashes.hsh"), os.O_RDWR, os.ModePerm)
	if err != nil {
		return nil, err
	}
	opt, err := memfile.NewAppendOptimizedFile(file)
	if err != nil {
		return nil, err
	}
	result.blkHashList = *wordfile.NewHashFile(opt, 0)

	file, err = os.OpenFile(path.Join(chcc.transactionsFolder, "Hashes.hsh"), os.O_RDWR, os.ModePerm)
	if err != nil {
		return nil, err
	}
	opt, err = memfile.NewAppendOptimizedFile(file)
	if err != nil {
		return nil, err
	}
	result.trnHashList = *wordfile.NewHashFile(opt, 0)

	file, err = os.OpenFile(path.Join(chcc.addressesFolder, "Hashes.hsh"), os.O_RDWR, os.ModePerm)
	if err != nil {
		return nil, err
	}
	opt, err = memfile.NewAppendOptimizedFile(file)
	if err != nil {
		return nil, err
	}
	result.adrHashList = *wordfile.NewHashFile(opt, 0)

	return &result, nil
}
