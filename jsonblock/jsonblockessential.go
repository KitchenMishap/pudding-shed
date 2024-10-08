package jsonblock

import (
	"encoding/json"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
)

// In here are partial specifications of the JSON format for a block (and contained transactions etc) coming from Bitcoin core REST API.
// They are partial as we're not interested in everything that's in the JSON.
// Hence the word "essential".
// Some non-essential fields are included; these are the integers which we can conveniently handle in a unified way

// SEE ALSO transaction.go, where these json types are furnished with functions to implement various interfaces

type JsonBlockEssential struct {
	J_height int                  `json:"height"`
	J_hash   string               `json:"hash"`
	J_tx     []JsonTransEssential `json:"tx"`
	hash     indexedhashes.Sha256 `json:"-"` // Does not appear in json. Calculated after parsing of whole block

	// Non essential integers
	// Changes here should be reflected in postJsonGatherNonEssentialInts()
	J_version        int64            `json:"version"`
	J_time           int64            `json:"time"`
	J_mediantime     int64            `json:"mediantime"`
	J_nonce          int64            `json:"nonce"`
	J_difficulty     float64          `json:"difficulty"`
	J_strippedsize   int64            `json:"strippedsize"`
	J_size           int64            `json:"size"`
	J_weight         int64            `json:"weight"`
	nonEssentialInts map[string]int64 `json:"-"` // Does not appear in json. Calculated after passing of whole block
}

type JsonTransEssential struct {
	J_txid string               `json:"txid"`
	J_vin  []JsonTxiEssential   `json:"vin"`
	J_vout []JsonTxoEssential   `json:"vout"`
	txid   indexedhashes.Sha256 `json:"-"` // Does not appear in json. Calculated after parsing of whole block
	handle TransHandle          `json:"-"` // Does not appear in json. "calculated" after parsing of whole block

	// Non essential integers
	J_version        int64            `json:"version"`
	J_size           int64            `json:"size"`
	J_vsize          int64            `json:"vsize"`
	J_weight         int64            `json:"weight"`
	J_locktime       int64            `json:"locktime"`
	nonEssentialInts map[string]int64 `json:"-"` // Does not appear in json. Calculated after passing of whole block
}

type JsonTxiEssential struct {
	J_txid       string               `json:"txid"`
	J_vout       int                  `json:"vout"`
	txid         indexedhashes.Sha256 `json:"-"` // Does not appear in json. Calculated after parsing of whole block
	parentTrans  TransHandle          `json:"-"` // Does not appear in json. "calculated" after parsing of whole block
	parentVIndex int64                `json:"-"` // Does not appear in json. "calculated" after parsing of whole block
	sourceTrans  TransHandle          `json:"-"` // Does not appear in json. Calculated after parsing of whole block
}

type JsonTxoEssential struct {
	J_value        float64                   `json:"value"`
	J_scriptPubKey JsonScriptPubKeyEssential `json:"scriptPubKey"`
	satoshis       int64                     `json:"-"` // Does not appear in json. Calculated after parsing of whole block
	parentTrans    TransHandle               `json:"-"` // Does not appear in json. "calculated" after parsing of whole block
	parentVIndex   int64                     `json:"-"` // Does not appear in json. "calculated" after parsing of whole block
}

type JsonScriptPubKeyEssential struct {
	J_hex       string               `json:"hex"`
	J_address   string               `json:"address"`
	J_type      string               `json:"type"`
	puddingHash indexedhashes.Sha256 `json:"-"` // Does not appear in json. Calculated after parsing of whole block
	// "pudding" because peculiar to pudding-shed
	// (This hash is not in use by bitcoiners generally)
}

func ParseJsonBlock(jsonBytes []byte) (*JsonBlockEssential, error) {
	var res JsonBlockEssential
	err := json.Unmarshal(jsonBytes, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func encodeJsonBlock(block *JsonBlockEssential) ([]byte, error) {
	jsonBytes, err := json.Marshal(block)
	if err != nil {
		return nil, err
	}
	return jsonBytes, nil
}

func (jb *JsonBlockEssential) BlockHash() indexedhashes.Sha256 {
	return jb.hash
}

func (jt *JsonTransEssential) TransHash() indexedhashes.Sha256 {
	return jt.txid
}
