package chainreadentities

import (
	"errors"
	"github.com/KitchenMishap/pudding-shed/chainreadinterface"
	"github.com/KitchenMishap/pudding-shed/genericentities"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
	"strconv"
)

type ChainReadEntityHandle struct {
	blockChain       chainreadinterface.IBlockChain
	maybeBlock       chainreadinterface.IBlockHandle
	maybeTransaction chainreadinterface.ITransHandle
	maybeTxi         chainreadinterface.ITxiHandle
	maybeTxo         chainreadinterface.ITxoHandle
}

func NewChainReadEntities(chain chainreadinterface.IBlockChain) genericentities.IEntity {
	res := ChainReadEntityHandle{}
	res.blockChain = chain
	res.maybeBlock = nil
	res.maybeTransaction = nil
	res.maybeTxi = nil
	res.maybeTxo = nil
	return res
}

func (c ChainReadEntityHandle) BlockChain() chainreadinterface.IBlockChain  { return c.blockChain }
func (c ChainReadEntityHandle) MaybeBlock() chainreadinterface.IBlockHandle { return c.maybeBlock }
func (c ChainReadEntityHandle) MaybeTransaction() chainreadinterface.ITransHandle {
	return c.maybeTransaction
}
func (c ChainReadEntityHandle) MaybeTxi() chainreadinterface.ITxiHandle { return c.maybeTxi }
func (c ChainReadEntityHandle) MaybeTxo() chainreadinterface.ITxoHandle { return c.maybeTxo }

func (c ChainReadEntityHandle) IntFieldNamesAvailable() ([]string, error) {
	var empty []string
	var res []string
	if c.maybeBlock != nil {
		block, err := c.blockChain.BlockInterface(c.maybeBlock)
		if err != nil {return empty, err}
		if block.HeightSpecified() {
			res = append(res, "height")
		}
		res = append(res, "transaction count")
		neis, _ := block.NonEssentialInts()
		for k, _ := range *neis {
			res = append(res, k)
		}
	}
	else if c.maybeTransaction != nil {
		trans, err := c.blockChain.TransInterface(c.maybeTransaction)
		if err != nil {return empty, err}
		if trans.IndicesPathSpecified() {
			res = append(res, "transaction within block")
		}
		res = append(res, "txi count", "txo count")
		neis, _ := trans.NonEssentialInts()
		for k, _ := range *neis {
			res = append(res, k)
		}
	}
	else if c.maybeTxi != nil {
		txi := c.blockChain.TxiInterface(c.maybeTxi)
		if txi.IndicesPathSpecified() {
			res = append(res, "txi within transaction")
		}
	}
	else if c.maybeTxo != nil {
		txo := c.blockChain.TxoInterface(c.maybeTxo)
		if txo.IndicesPathSpecified() {
			res = append(res, "txo within transaction")
		}
		res = append(res, "satoshis")
	}
	else {
		// blockchain
		// field "blocks" IF chainreadinterface supports it [  ] ToDo
		latest, err := c.blockChain.LatestBlock()
		if err != nil {
			return empty, err
		}
		if latest.HeightSpecified() {
			res = append(res, "blocks")
			}
	}
	return res
}

func (c ChainReadEntityHandle) StringFieldNamesAvailable() ([]string, error) {
	var empty []string
	var res []string
	if c.maybeBlock != nil {
		block, err := c.blockChain.BlockInterface(c.maybeBlock)
		if err != nil {return empty, err}
		if block.HashSpecified() {
			res = append(res, "hash")
		}
	}
	else if c.maybeTransaction != nil {
		trans, err := c.blockChain.TransInterface(c.maybeTransaction)
		if trans.HashSpecified() {
			res = append(res, "transaction hash")
		}
	}
	else if c.maybeTxi != nil {
		// No string fields for txi
	}
	else if c.maybeTxo != nil {
		// No string fields for txo
	}
	else {
		// blockchain
	}
	return res
}

func (c ChainReadEntityHandle) GetFieldTypeHint(fieldName string) string {
	switch fieldName {
		case "height", "transaction count", "transaction within block", "txi count", "txo count",
			"txi within transaction", "txo within transaction", "satoshis", "blocks",
			"size", "vsize", "strippedsize":
				return "count"
		case "hash":
			return "hash"
		case "time", "mediantime":
			return "time"
		case "difficulty":
			return "difficulty"
		default:
			return ""
	}
}

func (c ChainReadEntityHandle) GetIntField(fieldName string) (int64, error) {
	if c.maybeBlock != nil {
		block, err := c.blockChain.BlockInterface(c.maybeBlock)
		if err {
			return -1, err
		}
		if fieldName=="height" && block.HeightSpecified() {
			return block.Height(), nil
		}
		if fieldName=="transaction count" {
			count, err := block.TransactionCount()
			return count, err
		}
		nei, err := block.NonEssentialInts()
		if err {
			return -1, err
		}
		val, ok := (*nei)[fieldName]
		if ok {
			return val, nil
		}
		return -1, errors.New("ChainReadEntityHandle::GetIntField(): fieldName " + fieldName + " not supported for block entity")
	}
	else if c.maybeTransaction != nil {
		trans, err := c.blockChain.TransInterface(c.maybeTransaction)
		if err != nil {
			return -1, err
		}
		if fieldName == "transaction within block" && trans.IndicesPathSpecified() {
			_,trn := trans.IndicesPath()
			return trn, nil
		}
		if fieldName == "txi count" {
			txiCount, err := trans.TxiCount()
			if err != nil {
				return -1, err
			}
			return txiCount, nil
		}
		if fieldName == "txo count" {
			txoCount, err := trans.TxoCount()
			if err != nil {
				return -1, err
			}
			return txoCount, nil
		}
		nei, err := trans.NonEssentialInts()
		if err != nil {
			return -1, err
		}
		val, ok := (*nei)[fieldName]
		if ok {
			return val, nil
		}
		return -1, errors.New("ChainReadEntityHandle::GetIntField(): fieldName " + fieldName + " not supported for transaction entity")
	}
	else if c.maybeTxi != nil {
		txi, err := c.blockChain.TxiInterface(c.maybeTxi)
		if err != nil {
			return -1, err
		}
		if fieldName == "txi within transaction" && txi.IndicesPathSpecified() {
			_,_,vin := txi.IndicesPath()
			return vin, nil
		}
		return -1, errors.New("ChainReadEntityHandle::GetIntField(): fieldName " + fieldName + " not supported for txi entity")
	}
	else if c.maybeTxo != nil {
		txo, err := c.blockChain.TxoInterface(c.maybeTxo)
		if err != nil {
			return -1, err
		}
		if fieldName == "txi within transaction" && txi.IndicesPathSpecified() {
			_,_,vOut := txo.IndicesPath()
			return vOut, nil
		}
		if fieldName == "satoshis" {
			return txo.Satoshis()
		}
		return -1, errors.New("ChainReadEntityHandle::GetIntField(): fieldName " + fieldName + " not supported for txo entity")
	}
	else {
		// blockchain
		if fieldName=="blocks" {
			latest, err := c.blockChain.LatestBlock()
			if err != nil {
				return -1, nil
			}
			if latest.HeightSpecified() {
				return latest.Height(), nil
			} else {
				return -1, errors.New("ChainReadEntityHandle::GetIntField(): fieldName blocks not supported for this blockchain entity")
			}
		}
		return -1, errors.New("ChainReadEntityHandle::GetIntField(): fieldName " + fieldName + " not supported for blockchain entity")
	}
}

func (c ChainReadEntityHandle) GetStringField(fieldName string) (string, error) {
	if c.maybeBlock != nil {
		block, err := c.blockChain.BlockInterface(c.maybeBlock)
		if err != nil {
			return "", err
		}
		if fieldName=="block hash" && block.HashSpecified() {
			hash, err := block.Hash()
			if err != nil {
				return "", err
			}
			return indexedhashes.HashSha256ToHexString(&hash), nil
		}
		return -1, errors.New("ChainReadEntityHandle::GetStringField(): fieldName " + fieldName + " not supported for block entity")
	}
	else if c.maybeTransaction != nil {
		trans, err := c.blockChain.TransInterface(c.maybeTransaction)
		if err != nil {
			return "", err
		}
		if fieldName=="transaction hash" && trans.HashSpecified() {
			hash, err := trans.Hash()
			if err != nil {
				return "", err
			}
			return indexedhashes.HashSha256ToHexString(&hash), nil
		}
		return "", errors.New("ChainReadEntityHandle::GetStringField(): fieldName " + fieldName + " not supported for transaction entity")
	}
	else if c.maybeTxi != nil {
		return "", errors.New("ChainReadEntityHandle::GetStringField(): fieldName " + fieldName + " not supported for txi entity")
	}
	else if c.maybeTxo != nil {
		return "", errors.New("ChainReadEntityHandle::GetStringField(): fieldName " + fieldName + " not supported for txo entity")
	}
	else {
		// blockchain
		return "", errors.New("ChainReadEntityHandle::GetStringField(): fieldName " + fieldName + " not supported for blockchain entity")
	}
}

func (c ChainReadEntityHandle) EntityTypeName() string {
	if c.maybeBlock != nil { return "block" }
	else if c.maybeTransaction != nil { return "transaction" }
	else if c.maybeTxi != nil { return "txi" }
	else if c.maybeTxo != nil { return "txo" }
	else { return "blockchain" }
}

func (c ChainReadEntityHandle) IdentityRepresentations() ([]string, error) {
	var res []string
	if c.maybeBlock != nil {
		block, err := c.blockChain.BlockInterface(c.maybeBlock)
		if err != nil {
			return res, err
		}
		if block.HeightSpecified() {
			res = append(res, strconv.Itoa(int(block.Height()))
		}
		if block.HashSpecified() {
			hash, err := block.Hash()
			if err != nil {
				return []string{}, err
			}
			s := indexedhashes.HashSha256ToHexString(&hash)
			res = append(res, s[:6] + "...")
			res = append(res, s)
		}
		return res, nil
	}
	else if c.maybeTransaction != nil {
		trans, err := c.blockChain.TransInterface(c.maybeTransaction)
		if err != nil {
			return res, err
		}
		if trans.HashSpecified() {
			hash, err := block.Hash()
			if err != nil {
				return []string{}, err
			}
			s := indexedhashes.HashSha256ToHexString(&hash)
			res = append(res, s[:6] + "...")
			res = append(res, s)
		}
		if trans.IndicesPathSpecified() {
			b,t := trans.IndicesPath()
			res = append(res, "transaction " + strconv.Itoa(int(t)) + " of block " + strconv.Itoa(int(b)))
		}
		return res, nil
	}
	else if c.maybeTxi != nil {
		txi, err := c.blockChain.TxiInterface(c.maybeTxi)
		if err != nil {
			return res, err
		}
		if txi.IndicesPathSpecified()
		{
			b,t,vin := txi.IndicesPath()
			res = append(res, strconv.Itoa(int(vin)))
			res = append(res, "txi " + strconv.Itoa(int(vin)) + " of transaction " + strconv.Itoa(int(t)) + " of block " + strconv.Itoa(int(b)))
		}
		return res, nil
	}
	else if c.maybeTxo != nil {
		txo, err := c.blockChain.TxoInterface(c.maybeTxo)
		if err != nil {
			return res, err
		}
		if txo.IndicesPathSpecified()
		{
			b,t,vout := txo.IndicesPath()
			res = append(res, strconv.Itoa(int(vout)))
			res = append(res, "txi " + strconv.Itoa(int(vout)) + " of transaction " + strconv.Itoa(int(t)) + " of block " + strconv.Itoa(int(b)))
		}
		return res, nil
	}
	else {
		// blockchain
		return res, nil
	}
}

func (c ChainReadEntityHandle) PrevEntities() (map[string]genericentities.IEntityHandle, error) {
	var res map[string]genericentities.IEntityHandle
	if c.maybeBlock != nil {
		block, err := c.blockChain.BlockInterface(c.maybeBlock)
		if err != nil {
			return res, err
		}
		prevBlock := c.blockChain.ParentBlock(block)
		if !prevBlock.IsInvalid() {
			entity := ChainReadEntityHandle{
				blockChain:       c.blockChain,
				maybeBlock:       prevBlock,
				maybeTransaction: nil,
				maybeTxi:         nil,
				maybeTxo:         nil,
			}
			res["previous block"] = entity
		}
		return res, nil
	} else if c.maybeTransaction != nil {
		trans, err := c.blockChain.TransInterface(c.maybeTransaction)
		if err != nil {
			return res, err
		}
		prevTrans := c.blockChain.PreviousTransaction(trans)
		if !prevTrans.IsInvalid() {
			entity := ChainReadEntityHandle{
				blockChain:       c.blockChain,
				maybeBlock:       nil,
				maybeTransaction: prevTrans,
				maybeTxi:         nil,
				maybeTxo:         nil,
			}
			res["previous transaction"] = entity
		}
		return res, nil
	} else if c.maybeTxi != nil {
		txi, err := c.blockChain.TxiInterface(c.maybeTxi)
		if err != nil {
			return res, err
		}
		parentTrans, err := c.blockChain.TransInterface( txi.ParentTrans() )
		if err != nil {
			return res, err
		}
		if txi.IndicesPathSpecified()
		{
			b,t,vIn := txi.IndicesPath()
			if vIn > 0 {
				prevVin := vIn - 1
				prevTxi, err := parentTrans.NthTxi(prevVin)
				if err != nil {
					return res, err
				}
				entity := ChainReadEntityHandle{
					blockChain:       c.blockChain,
					maybeBlock:       nil,
					maybeTransaction: nil,
					maybeTxi:         prevTxi,
					maybeTxo:         nil,
				}
				res["previous txi"] = entity
			}
		}
		return res, nil
	} else if c.maybeTxo != nil {
		txo, err := c.blockChain.TxoInterface(c.maybeTxo)
		if err != nil {
			return res, err
		}
		parentTrans, err := c.blockChain.TransInterface( txo.ParentTrans() )
		if err != nil {
			return res, err
		}
		if txo.IndicesPathSpecified()
		{
			_,_,vOut := txo.IndicesPath()
			if vOut > 0 {
				prevVOut := vOut-1
				prevTxo, err := parentTrans.NthTxo(prevVOut)
				if err != nil {
					return res, err
				}
				entity := ChainReadEntityHandle{
					blockChain:       c.blockChain,
					maybeBlock:       nil,
					maybeTransaction: nil,
					maybeTxi:         nil,
					maybeTxo:         prevTxo,
				}
				res["previous txo"] = entity
			}
		}
		return res, nil
	}
	else {
		// blockchain
		return res, nil
	}
}

func (c ChainReadEntityHandle) NextEntities() (map[string]genericentities.IEntityHandle, error) {
	var res map[string]genericentities.IEntityHandle
	if c.maybeBlock != nil {
		block, err := c.blockChain.BlockInterface(c.maybeBlock)
		if err != nil {
			return res, err
		}
		nextBlock, err := c.blockChain.NextBlock(block)
		if err != nil {
			return res, err
		}
		if !nextBlock.IsInvalid() {
			entity := ChainReadEntityHandle{
				blockChain:       c.blockChain,
				maybeBlock:       nextBlock,
				maybeTransaction: nil,
				maybeTxi:         nil,
				maybeTxo:         nil,
			}
			res["next block"] = entity
		}
		return res, nil
	}
	else if c.maybeTransaction != nil {
		trans, err := c.blockChain.TransInterface(c.maybeTransaction)
		if err != nil {
			return res, err
		}
		nextTrans, err := c.blockChain.NextTransaction(trans)
		if err != nil {
			return res, err
		}
		if !nextTrans.IsInvalid() {
			entity := ChainReadEntityHandle{
				blockChain:       c.blockChain,
				maybeBlock:       nil,
				maybeTransaction: nextTrans,
				maybeTxi:         nil,
				maybeTxo:         nil,
			}
			res["next transaction"] = entity
		}
		return res, nil
	}
	else if c.maybeTxi != nil {
		txi, err := c.blockChain.TxiInterface(c.maybeTxi)
		if err != nil {
			return res, err
		}
		parentTrans, err := c.blockChain.TransInterface( txi.ParentTrans() )
		if err != nil {
			return res, err
		}
		vIns, err := parentTrans.TxiCount()
		if err != nil {
			return res, err
		}
		if txi.IndicesPathSpecified()
		{
			_,_,vIn := txi.IndicesPath()
			if vIn < vIns - 1 {
				nextVIn := vIn+1
				nextTxi, err := parentTrans.NthTxi(nextVIn)
				if err != nil {
					return res, err
				}
				entity := ChainReadEntityHandle{
					blockChain:       c.blockChain,
					maybeBlock:       nil,
					maybeTransaction: nil,
					maybeTxi:         nextTxi,
					maybeTxo:         nil,
				}
				res["next txi"] = entity
			}
		}
		return res, nil
	}
	else if c.maybeTxo != nil {
		txo, err := c.blockChain.TxoInterface(c.maybeTxo)
		if err != nil {
			return res, err
		}
		parentTrans, err := c.blockChain.TransInterface( txo.ParentTrans() )
		if err != nil {
			return res, err
		}
		vOuts, err := parentTrans.TxoCount()
		if err != nil {
			return res, err
		}
		if txo.IndicesPathSpecified()
		{
			_,_,vOut := txo.IndicesPath()
			if vOut < vOuts - 1 {
				nextVOut := vOut + 1
				nextTxo, err := parentTrans.NthTxo(nextVOut)
				if err != nil {
					return res, err
				}
				entity := ChainReadEntityHandle{
					blockChain:       c.blockChain,
					maybeBlock:       nil,
					maybeTransaction: nil,
					maybeTxi:         nil,
					maybeTxo:         nextTxo,
				}
				res["next txo"] = entity
			}
		}
		return res, nil
	}
	else {
		// blockchain
		return res, nil
	}
}

func (c ChainReadEntityHandle) ParentEntityCounts() (map[string]int64, error) {
	var empty map[string]int64
	var res map[string]int64
	if c.maybeBlock != nil {
		res["blockchain"] = 1		// One parent blockchain
	} else if c.maybeTransaction != nil {
		// One parent block... but chainreadinterface won't let us find it! [  ] ToDo

		// Txis are considered parents of transaction
		trans, err := c.blockChain.TransInterface(c.maybeTransaction)
		if err != nil {
			return empty, err
		}
		res["txis"], err = trans.TxiCount()
		if err != nil {
			return empty, err
		}
	} else if c.maybeTxi != nil {
		// One parent transaction... IF chainreadinterface will let us find it! [  ] ToDo
		txi, err := c.blockChain.TxiInterface(c.maybeTxi)
		if err != nil {
			return empty, err
		}
		if txi.ParentSpecified() {
			res["transaction"] = 1
		}

		res["txo spent from"] = 1	// Txi comes from one Txo
	} else if c.maybeTxo != nil {
		// One parent transaction... IF chainreadinterface will let us find it! [  ] ToDo
		txo, err := c.blockChain.TxoInterface(c.maybeTxo)
		if err != nil {
			return empty, err
		}
		if txo.ParentSpecified() {
			res["transaction"] = 1
		}
	} else {
		// blockchain has no parent
	}
	return res, nil
}

func (c ChainReadEntityHandle) ParentEntity(parentName string, index int64) (genericentities.IEntityHandle, error) {
	empty := ChainReadEntityHandle{
		blockChain:       nil,
		maybeBlock:       nil,
		maybeTransaction: nil,
		maybeTxi:         nil,
		maybeTxo:         nil,
	}
	if c.maybeBlock != nil {
		// Blockchain parent of block
		if parentName=="blockchain" && index==0 {
			return ChainReadEntityHandle{
				blockChain:       c.blockChain,
				maybeBlock:       nil,
				maybeTransaction: nil,
				maybeTxi:         nil,
				maybeTxo:         nil,
			}, nil
		}
	} else if c.maybeTransaction != nil {
		// Block parent of transaction... but chainreadinterface won't let us find it [  ] ToDo

		// Txi's are considered parents of transaction
		trans, err := c.blockChain.TransInterface(c.maybeTransaction)
		if err != nil {
			return empty, err
		}
		txiCount, err := trans.TxiCount()
		if err != nil {
			return empty, err
		}
		if parentName=="txis" && index < txiCount {
			txiHandle, err := trans.NthTxi(index)
			if err != nil {
				return empty, err
			}
			return ChainReadEntityHandle{
				blockChain:       c.blockChain,
				maybeBlock:       nil,
				maybeTransaction: nil,
				maybeTxi:         txiHandle,
				maybeTxo:         nil,
			}, nil
		}
	} else if c.maybeTxi != nil {
		// One parent transaction of txi... IF chainreadinterface will let us find it! [  ] ToDo
		if parentName=="transaction" && index==0
		{
			txi, err := c.blockChain.TxiInterface(c.maybeTxi)
			if err != nil {
				return empty, err
			}
			if txi.ParentSpecified() {
				trans := txi.ParentTrans()
				return ChainReadEntityHandle{
					blockChain:       c.blockChain,
					maybeBlock:       nil,
					maybeTransaction: trans,
					maybeTxi:         nil,
					maybeTxo:         nil,
				}, nil
			}
		}

		// One txo is considered a parent of a txi
		if parentName=="txo spent from" && index==0 {
			txi, err := c.blockChain.TxiInterface(c.maybeTxi)
			if err != nil {
				return empty, err
			}
			txo, err := txi.SourceTxo()
			if err != nil {
				return empty, err
			}
			return ChainReadEntityHandle{
				blockChain:       c.blockChain,
				maybeBlock:       nil,
				maybeTransaction: nil,
				maybeTxi:         nil,
				maybeTxo:         txo,
			}, nil
		}
	}
	else if c.maybeTxo != nil {
		// One parent transaction of Txo, IF chainreadinterface will let us find it [  ] ToDo
		if parentName=="transaction" && index==0
		{
			txo, err := c.blockChain.TxoInterface(c.maybeTxo)
			if err != nil {
				return empty, err
			}
			if txo.ParentSpecified() {
				trans := txo.ParentTrans()
				return ChainReadEntityHandle{
					blockChain:       c.blockChain,
					maybeBlock:       nil,
					maybeTransaction: trans,
					maybeTxi:         nil,
					maybeTxo:         nil,
				}, nil
			}
		}
	}
	else {
		// blockchain has no parent
	}
	err := errors.New("Parent " + parentName + " index " + strconv.Itoa(int(index)) + " could not be found")
	return empty, err
}

func (c ChainReadEntityHandle) ChildEntityCounts() (map[string]int64, error) {
	var empty map[string]int64
	var res map[string]int64
	if c.maybeBlock != nil {
		// Transactions are children of a block
		block, err := c.blockChain.BlockInterface(c.maybeBlock)
		if err != nil {
			return empty, err
		}
		count, err := block.TransactionCount()
		if err != nil {
			return empty, err
		}
		res["transactions"] = count
	} else if c.maybeTransaction != nil {
		// Txos are considered children of transaction
		trans, err := c.blockChain.TransInterface(c.maybeTransaction)
		if err != nil {
			return empty, err
		}
		res["txos"], err = trans.TxoCount()
		if err != nil {
			return empty, err
		}
	} else if c.maybeTxi != nil {
		// Because Txis are considered parents of a transaction, a transaction must be a child of each Txi

		// One child transaction... IF chainreadinterface will let us find it! [  ] ToDo
		txi, err := c.blockChain.TxiInterface(c.maybeTxi)
		if err != nil {
			return empty, err
		}
		if txi.ParentSpecified() {
			res["transaction"] = 1
		}
	} else if c.maybeTxo != nil {
		// The child of a Txo is the Txi it is spent to (if any, and if available)
		// This "coin spent when" info is not yet coded [  ] ToDo
	} else {
		// Blockchain
		// Blocks are children of the blockchain
		// BUT chainreadinterface won't give us the nth block
	}
	return res, nil
}

func (c ChainReadEntityHandle) ChildEntity(childName string, index int64) (genericentities.IEntityHandle, error) {
	empty := ChainReadEntityHandle{
		blockChain:       nil,
		maybeBlock:       nil,
		maybeTransaction: nil,
		maybeTxi:         nil,
		maybeTxo:         nil,
	}
	if c.maybeBlock != nil {
		// Transactions are children of a block
		if childName=="transactions" {
			block, err := c.blockChain.BlockInterface(c.maybeBlock)
			if err != nil {
				return empty, err
			}
			count, err := block.TransactionCount()
			if err != nil {
				return empty, err
			}
			if index < count {
				transHandle, err := block.NthTransaction(index)
				if err != nil {
					return empty, err
				}
				return ChainReadEntityHandle{
					blockChain:       c.blockChain,
					maybeBlock:       nil,
					maybeTransaction: transHandle,
					maybeTxi:         nil,
					maybeTxo:         nil,
				}, nil
			}
		}
	} else if c.maybeTransaction != nil {
		// Txos are considered children of transaction
		if childName=="txos" {
			trans, err := c.blockChain.TransInterface(c.maybeTransaction)
			if err != nil {
				return empty, err
			}
			count, err := trans.TxoCount()
			if err != nil {
				return empty, err
			}
			if index < count {
				txoHandle, err := trans.NthTxo(index)
				if err != nil {
					return empty, err
				}
				return ChainReadEntityHandle{
					blockChain:       c.blockChain,
					maybeBlock:       nil,
					maybeTransaction: nil,
					maybeTxi:         nil,
					maybeTxo:         txoHandle,
				}, nil
			}
		}
	} else if c.maybeTxi != nil {
		// Because Txis are considered parents of a transaction, a transaction must be a child of each Txi

		// One child transaction... IF chainreadinterface will let us find it! [  ] ToDo
		if childName=="transaction" {
			txi, err := c.blockChain.TxiInterface(c.maybeTxi)
			if err != nil {
				return empty, err
			}
			if txi.ParentSpecified() {
				trans := txi.ParentTrans()
				return ChainReadEntityHandle{
					blockChain:       c.blockChain,
					maybeBlock:       nil,
					maybeTransaction: trans,
					maybeTxi:         nil,
					maybeTxo:         nil,
				}, nil
			}
		}
	} else if c.maybeTxo != nil {
		// The child of a Txo is the Txi it is spent to (if any, and if available)
		// This "coin spent when" info is not yet coded [  ] ToDo
	} else {
		// Blockchain
		// Blocks are children of the blockchain
		// BUT chainreadinterface won't give us the nth block [  ] ToDo
	}
	err := errors.New("ChainReadEntityHandle.ChildEntity(): index " + strconv.Itoa(int(index)) + " of child " + childName + " could not be found" )
	return empty, err
}
