package nanoTypes

import (
   "encoding/json"
   "strconv"
)
// This is just a wrapper to an int64 because the nano nodes use nonstandard JSON and wrap their integers in double quotes for no reason in particular.
type JInt int64

// JSON int Marshaler
func (j JInt) MarshalJSON() ([]byte, error) {
   return json.Marshal(int64(j))
}

func (j *JInt) UnmarshalJSON(data []byte) (err error) {
   var s string
   if err = json.Unmarshal(data, &s); err != nil {
      return
   }
   i, err := strconv.Atoi(s)
   *j = JInt(i)
   return
}

func (j JInt) String() string {
   return strconv.Itoa(int(j))
}

