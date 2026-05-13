package intrinsicobjects

import (
	"encoding/hex"
	"encoding/json"

	"github.com/KitchenMishap/pudding-shed/indexedhashes"
)

func ParseJsonTransaction(jsonBytes []byte, targetTrans *Transaction) error {
	parsed, err := parseJsonTrans(jsonBytes)
	if err != nil {
		return err
	}
	err = parseJsonTransaction(parsed, targetTrans)
	if err != nil {
		return err
	}
	return nil
}

func parseJsonTransaction(parsed *JsonTransEssential, targetTrans *Transaction) error {
	targetTrans.Version = uint32(parsed.J_version)
	targetTrans.Size = int(parsed.J_size)
	targetTrans.Weight = int(parsed.J_weight)
	targetTrans.VSize = int(parsed.J_vsize)
	targetTrans.IsSegWit = parsed.J_vsize < parsed.J_size
	err := indexedhashes.HashHexToSha256(parsed.J_txid, &targetTrans.TxId)
	if err != nil {
		return err
	}

	txiCount := len(parsed.J_vin)
	targetTrans.BitcoinCoreTxis = make([]Txi, txiCount)
	for i := 0; i < txiCount; i++ {
		if parsed.J_vin[i].J_txid == "" {
			// Must be a coinbase transaction
			targetTrans.BitcoinCoreTxis[i].TxId = indexedhashes.Sha256{} // All zeroes
		} else {
			err = indexedhashes.HashHexToSha256(parsed.J_vin[i].J_txid, &targetTrans.BitcoinCoreTxis[i].TxId)
			if err != nil {
				return err
			}
		}
		targetTrans.BitcoinCoreTxis[i].VOut = int64(parsed.J_vin[i].J_vout)
	}
	txoCount := len(parsed.J_vout)
	targetTrans.Txos = make([]Txo, txoCount)
	for i := 0; i < txoCount; i++ {
		targetTrans.Txos[i].Value = int64(parsed.J_vout[i].J_value)
		var scriptPubKey []byte
		scriptPubKey, err = hex.DecodeString(parsed.J_vout[i].J_scriptPubKey.J_hex)
		if err != nil {
			return err
		}
		targetTrans.Txos[i].ScriptPubKey = scriptPubKey
	}
	// Google Gemini AI told me this bit
	targetTrans.StrippedSize = (int(parsed.J_weight) - int(parsed.J_size)) / 3
	if parsed.J_weight == 0 || targetTrans.StrippedSize == 0 {
		targetTrans.StrippedSize = int(parsed.J_size)
	}

	return nil
}

// In here are partial specifications of the JSON format for a transaction (and contained txis etc)
// coming from Bitcoin core REST API.
// They are partial as we're not interested in everything that's in the JSON.
// Hence the word "essential".
// Some non-essential fields are included; these are the integers which we can conveniently handle in a unified way

type JsonTransEssential struct {
	J_txid string             `json:"txid"`
	J_vin  []JsonTxiEssential `json:"vin"`
	J_vout []JsonTxoEssential `json:"vout"`

	// Non essential integers
	J_version  int64 `json:"version"`
	J_size     int64 `json:"size"`
	J_vsize    int64 `json:"vsize"`
	J_weight   int64 `json:"weight"`
	J_locktime int64 `json:"locktime"`
}

type JsonTxiEssential struct {
	J_txid string `json:"txid"`
	J_vout int    `json:"vout"`
}

type JsonTxoEssential struct {
	J_value        float64                   `json:"value"`
	J_scriptPubKey JsonScriptPubKeyEssential `json:"scriptPubKey"`
}

type JsonScriptPubKeyEssential struct {
	J_hex     string `json:"hex"`
	J_address string `json:"address"`
	J_type    string `json:"type"`
}

func parseJsonTrans(jsonBytes []byte) (*JsonTransEssential, error) {
	var res JsonTransEssential
	err := json.Unmarshal(jsonBytes, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func encodeJsonTrans(block *JsonTransEssential) ([]byte, error) {
	jsonBytes, err := json.Marshal(block)
	if err != nil {
		return nil, err
	}
	return jsonBytes, nil
}
