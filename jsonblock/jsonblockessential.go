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
	J_tx     []jsonTransEssential `json:"tx"`
	hash     indexedhashes.Sha256 `json:"-"` // Does not appear in json. Calculated after parsing of whole block

	// Non essential integers
	// Changes here should be reflected in postJsonGatherNonEssentialInts()
	J_version        int64            `json:"version"`
	J_time           int64            `json:"time"`
	J_mediantime     int64            `json:"mediantime"`
	J_nonce          int64            `json:"nonce"`
	J_difficulty     int64            `json:"difficulty"`
	J_strippedsize   int64            `json:"strippedsize"`
	J_size           int64            `json:"size"`
	J_weight         int64            `json:"weight"`
	nonEssentialInts map[string]int64 `json:"-"` // Does not appear in json. Calculated after passing of whole block
}

type jsonTransEssential struct {
	J_txid string               `json:"txid"`
	J_vin  []jsonTxiEssential   `json:"vin"`
	J_vout []jsonTxoEssential   `json:"vout"`
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

type jsonTxiEssential struct {
	J_txid       string               `json:"txid"`
	J_vout       int                  `json:"vout"`
	txid         indexedhashes.Sha256 `json:"-"` // Does not appear in json. Calculated after parsing of whole block
	parentTrans  TransHandle          `json:"-"` // Does not appear in json. "calculated" after parsing of whole block
	parentVIndex int64                `json:"-"` // Does not appear in json. "calculated" after parsing of whole block
	sourceTrans  TransHandle          `json:"-"` // Does not appear in json. Calculated after parsing of whole block
}

type jsonTxoEssential struct {
	J_value      float64     `json:"value"`
	satoshis     int64       `json:"-"` // Does not appear in json. Calculated after parsing of whole block
	parentTrans  TransHandle `json:"-"` // Does not appear in json. "calculated" after parsing of whole block
	parentVIndex int64       `json:"-"` // Does not appear in json. "calculated" after parsing of whole block
}

func parseJsonBlock(jsonBytes []byte) (*JsonBlockEssential, error) {
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
