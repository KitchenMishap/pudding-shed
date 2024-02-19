package artprojectacid

import "github.com/KitchenMishap/pudding-shed/chainreadinterface"

// The three colour bytes are the LSBs of the transaction hash of the coinbase transaction
func colourBytes(ibc chainreadinterface.IBlockChain,
	hBlock chainreadinterface.IBlockHandle) (colourByte0 int, colourByte1 int, colourByte2 int) {
	block, _ := ibc.BlockInterface(hBlock)
	hTrans, _ := block.NthTransaction(0)
	trans, _ := ibc.TransInterface(hTrans)
	hash, _ := trans.Hash()
	colourByte0 = int(hash[31])
	colourByte1 = int(hash[30])
	colourByte2 = int(hash[29])
	return colourByte0, colourByte1, colourByte2
}
