package nanoTypes

import (
   "strings"
   "encoding/hex"
   "encoding/json"
)

// BlockHash represents a block hash.
type BlockHash []byte

func (h BlockHash) String() string {
   if (len(h) == 0) {
      return "0000000000000000000000000000000000000000000000000000000000000000"
   } else {
      return strings.ToUpper(hex.EncodeToString(h))
   }
}

// JSON BlockHash Marshaler
func (h BlockHash) MarshalJSON() ([]byte, error) {
   return json.Marshal(h.String())
}

func (h *BlockHash) UnmarshalJSON(data []byte) (err error) {
   var s string
   if err = json.Unmarshal(data, &s); err != nil {
      return
   }
   *h, err = hex.DecodeString(s)
   return
}

