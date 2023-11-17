package jsonblock

import (
	"encoding/json"
)

// In here are partial specifications of the JSON format for a block (and contained transactions etc) coming from Bitcoin core REST API.
// They are partial as we're not interested in everything that's in the JSON.
// Hence the word "essential"

type jsonBlockEssential struct {
	Height int64
	Hash   string
	Tx     []jsonTransEssential
}

type jsonTransEssential struct {
	Txid string
	Vin  []jsonTxiEssential
	Vout []jsonTxoEssential
}

type jsonTxiEssential struct {
	Txid string
	Vout int64
}

type jsonTxoEssential struct {
	Value float64
}

func parseJsonBlock(jsonBytes []byte) (*jsonBlockEssential, error) {
	var obj jsonBlockEssential
	err := json.Unmarshal(jsonBytes, &obj)
	if err != nil {
		return nil, err
	}
	return &obj, nil
}

func encodeJsonBlock(jsonBlock *jsonBlockEssential) ([]byte, error) {
	marshal, err := json.Marshal(jsonBlock)
	if err != nil {
		return nil, err
	}
	return marshal, nil
}
