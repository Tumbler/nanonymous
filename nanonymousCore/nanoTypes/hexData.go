package nanoTypes

import (
   "encoding/hex"
   "encoding/json"
   "strings"
)

// HexData represents generic hex data.
type HexData []byte

// JSON HexData Marshaler
func (h HexData) MarshalJSON() ([]byte, error) {
   return json.Marshal(hex.EncodeToString(h))
}

// UnmarshalJSON sets *h to a copy of data.
func (h *HexData) UnmarshalJSON(data []byte) (err error) {
   var s string
   if err = json.Unmarshal(data, &s); err != nil {
      return
   }
   *h, err = hex.DecodeString(s)
   return
}

func (h HexData) String() string {
   return strings.ToUpper(hex.EncodeToString(h))
}

