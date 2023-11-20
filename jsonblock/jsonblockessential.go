package jsonblock

import (
	"encoding/json"
	"github.com/KitchenMishap/pudding-shed/indexedhashes"
)

// In here are partial specifications of the JSON format for a block (and contained transactions etc) coming from Bitcoin core REST API.
// They are partial as we're not interested in everything that's in the JSON.
// Hence the word "essential"

type jsonBlockEssential struct {
	Height int64
	Hash   indexedhashes.Sha256
	Tx     []jsonTransEssential
}

type jsonTransEssential struct {
	Txid indexedhashes.Sha256
	Vin  []jsonTxiEssential
	Vout []jsonTxoEssential
}

type jsonTxiEssential struct {
	Txid indexedhashes.Sha256
	Vout int64
}

type jsonTxoEssential struct {
	Value float64
}

func parseJsonBlock(jsonBytes []byte) (*jsonBlockEssential, error) {
	var res jsonBlockEssential
	err := json.Unmarshal(jsonBytes, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func encodeJsonBlock(jsonBlock *jsonBlockEssential) ([]byte, error) {
	jsonBytes, err := json.Marshal(jsonBlock)
	if err != nil {
		return nil, err
	}
	return jsonBytes, nil
}
