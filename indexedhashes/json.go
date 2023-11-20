package indexedhashes

import (
	"encoding/json"
)

// Enable json encoding of indexedhashes.Sha256

func (sh *Sha256) UnmarshalJSON(b []byte) error {
	var hexAscii string
	if err := json.Unmarshal(b, &hexAscii); err != nil {
		return err
	}
	err := hashHexToSha256(hexAscii, sh)
	return err
}

func (sh *Sha256) MarshalJSON() ([]byte, error) {
	hexAscii := hashSha256ToHexString(sh)
	return json.Marshal(hexAscii)
}
