package jsonblock

import (
	"encoding/json"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
)

// In here are partial specifications of the JSON format for a block (and contained transactions etc) coming from Bitcoin core REST API.
// They are partial as we're only interested in gathering hashes (blocks, transactions, and addresses) in this code

type JsonBlockHashes struct {
	J_height int                  `json:"height"`
	J_hash   string               `json:"hash"`
	J_tx     []JsonTransHashes    `json:"tx"`
	hash     indexedhashes.Sha256 `json:"-"` // Does not appear in json. Calculated after parsing of whole block
}

type JsonTransHashes struct {
	J_txid string               `json:"txid"`
	J_vout []JsonTxoHashes      `json:"vout"` // We're not interested in vins as they don't define any hashes
	txid   indexedhashes.Sha256 `json:"-"`    // Does not appear in json. Calculated after parsing of whole block
}

type JsonTxoHashes struct {
	J_scriptPubKey JsonScriptPubKeyEssential2 `json:"scriptPubKey"`
}

type JsonScriptPubKeyEssential2 struct {
	J_hex       string               `json:"hex"`
	J_address   string               `json:"address"`
	J_type      string               `json:"type"`
	puddingHash indexedhashes.Sha256 `json:"-"` // Does not appear in json. Calculated after parsing of whole block
	// "pudding" because peculiar to pudding-shed
	// (This hash is not in use by bitcoiners generally)
}

func ParseJsonBlockHashes(jsonBytes []byte) (*JsonBlockHashes, error) {
	var res JsonBlockHashes
	err := json.Unmarshal(jsonBytes, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (jb *JsonBlockHashes) BlockHash() indexedhashes.Sha256 {
	return jb.hash
}

func (jt *JsonTransHashes) TransHash() indexedhashes.Sha256 {
	return jt.txid
}

func (jt *JsonTxoHashes) AddrHash() indexedhashes.Sha256 {
	return jt.J_scriptPubKey.puddingHash
}
