package intrinsicobjects

import (
	"encoding/json"
	"strconv"

	"github.com/KitchenMishap/pudding-shed/indexedhashes"
)

func ParseJsonBlock(jsonBytes []byte, targetBlock *Block) error {
	parsed, err := parseJsonBlock(jsonBytes)
	if err != nil {
		return err
	}

	err = indexedhashes.HashHexToSha256(parsed.J_hash, &targetBlock.BlockHash)
	if err != nil {
		return err
	}
	if parsed.J_previousblockhash == "" {
		// Doesn't exist for genesis block!
		targetBlock.PrevHash = indexedhashes.Sha256{} // All zeros
	} else {
		err = indexedhashes.HashHexToSha256(parsed.J_previousblockhash, &targetBlock.PrevHash)
		if err != nil {
			return err
		}
	}
	err = indexedhashes.HashHexToSha256(parsed.J_merkleroot, &targetBlock.MerkleRoot)
	if err != nil {
		return err
	}
	targetBlock.Version = uint32(parsed.J_version)
	targetBlock.Weight = int(parsed.J_weight)
	targetBlock.Size = int(parsed.J_size)
	targetBlock.StrippedSize = int(parsed.J_strippedsize)
	targetBlock.Time = uint32(parsed.J_time)
	targetBlock.Difficulty = parsed.J_difficulty
	nBits, err := strconv.ParseUint(parsed.J_bits, 16, 32)
	if err != nil {
		return err
	}
	targetBlock.NBits = uint32(nBits)
	targetBlock.Nonce = uint32(parsed.J_nonce)

	txCount := len(parsed.J_tx)
	targetBlock.Transactions = make([]Transaction, txCount)
	for i := 0; i < txCount; i++ {
		err = parseJsonTransaction(&parsed.J_tx[i], &targetBlock.Transactions[i])
		if err != nil {
			return err
		}
	}
	return nil
}

// In here are partial specifications of the JSON format for a block coming from Bitcoin core REST API.
// They are partial as we're not interested in everything that's in the core JSON.
// Hence the word "essential".
// Some non-essential fields are included; these are the integers which we can conveniently handle in a unified way

type JsonBlockEssential struct {
	J_height            int                  `json:"height"`
	J_hash              string               `json:"hash"`
	J_previousblockhash string               `json:"previousblockhash"`
	J_tx                []JsonTransEssential `json:"tx"`
	J_merkleroot        string               `json:"merkleroot"`

	// Non essential integers
	J_version      int64 `json:"version"`
	J_time         int64 `json:"time"`
	J_mediantime   int64 `json:"mediantime"`
	J_nonce        int64 `json:"nonce"`
	J_strippedsize int64 `json:"strippedsize"`
	J_size         int64 `json:"size"`
	J_weight       int64 `json:"weight"`

	// Other non-essentials
	J_difficulty float64 `json:"difficulty"`
	J_bits       string  `json:"bits"`
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
