package model

import (
	"encoding/json"
	"fmt"
)

// CID represents a Filecoin Content Identifier.
// Lotus encodes CIDs in JSON as {"/": "<cid-string>"}.
type CID string

func (c *CID) UnmarshalJSON(data []byte) error {
	var obj struct {
		Slash string `json:"/"`
	}
	if err := json.Unmarshal(data, &obj); err == nil {
		*c = CID(obj.Slash)
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("cannot unmarshal CID from %s", data)
	}
	*c = CID(s)
	return nil
}

func (c CID) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{"/": string(c)})
}

// TipSetKey is an ordered list of CIDs identifying a tipset.
type TipSetKey []CID
