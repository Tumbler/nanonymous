package nanoTypes

import (
   "encoding/json"
)

// This is just a wrapper to a bool because the nano nodes use nonstandard JSON
// and wrap their bools in double quotes for no reason in particular.
type JBool bool

// JSON bool Marshaler
func (b JBool) MarshalJSON() ([]byte, error) {
   return json.Marshal(bool(b))
}

func (b *JBool) UnmarshalJSON(data []byte) (err error) {
   var s string
   if err = json.Unmarshal(data, &s); err != nil {
      return
   }
   if (s == "true") {
      *b = true
   } else {
      *b = false
   }
   return
}

func (b JBool) String() string {
   if (b) {
      return "true"
   } else {
      return "false"
   }
}

