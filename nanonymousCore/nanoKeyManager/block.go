package nanoKeyManager

import (
   "fmt"
   "strings"
   "math/big"
   "golang.org/x/crypto/blake2b"
   "encoding/hex"
   "encoding/json"
   "strconv"
   "database/sql/driver"

   // Third party packages
   "github.com/hectorchu/gonano/wallet/ed25519"

)

// Block corresponds to the JSON representation of a block.
type Block struct {
   Type           string     `json:"type"`
   Account        string     `json:"account"`
   Previous       BlockHash  `json:"previous"`
   Representative string     `json:"representative"`
   Balance        *Raw       `json:"balance"`
   Link           BlockHash  `json:"link"`
   LinkAsAccount  string     `json:"link_as_account"`
   Signature      HexData    `json:"signature"`
   Work           HexData    `json:"work"`
   Seed           Key
}

// BlockHash represents a block hash.
type BlockHash []byte
// HexData represents generic hex data.
type HexData []byte

// Raw represents an amount of raw nano.
type Raw struct {
    *big.Int
}


func (h BlockHash) String() string {
   return strings.ToUpper(hex.EncodeToString(h))
}

// Hash calculates the block hash.
func (b *Block) Hash() (hash BlockHash, err error) {
   h, err := blake2b.New256(nil)
   if err != nil {
      return
   }
   h.Write(make([]byte, 31))
   h.Write([]byte{6})
   pubkey, err := AddressToPubKey(b.Account)
   if err != nil {
      return
   }
   h.Write(pubkey)
   h.Write(b.Previous)
   pubkey, err = AddressToPubKey(b.Representative)
   if err != nil {
      return
   }
   h.Write(pubkey)
   h.Write(b.Balance.FillBytes(make([]byte, 16)))
   h.Write(b.Link)
   return h.Sum(nil), nil
}
func (b *Block) Sign() ([]byte, error) {

   if (b.Seed.KeyType > 1 || !b.Seed.Initialized) {
      return nil, fmt.Errorf("Sign: no private key in key struct")
   }

   hash , err := b.Hash()
   if (err != nil) {
      return nil, fmt.Errorf("Sign: %w", err)
   }

   keyPair := append(b.Seed.PrivateKey, b.Seed.PublicKey...)
   return ed25519.Sign(keyPair, hash), nil
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

// JSON Raw Marshaler
func (r Raw) MarshalJSON() ([]byte, error) {
    return []byte(r.String()), nil
}

func (r *Raw) UnmarshalJSON(src []byte) error {
    if string(src) == "null" {
        return nil
    }
    trimmed := strings.Trim(string(src), `"`)
    var z big.Int
    p, ok := z.SetString(trimmed, 10)
    if !ok {
        return fmt.Errorf("not a valid big integer: %s", src)
    }
    r.Int = p
    return nil
}

// Postgres Scan driver for Raw
func (r *Raw) Scan(src any) error {

   if str, ok := src.(string); ok {
      text := strings.Split(str, "e")
      numZeros, _ := strconv.Atoi(text[1])
      text[0] += strings.Repeat("0", numZeros)

      r.SetString(text[0], 10)
   } else {
      return fmt.Errorf("Can't assign", src, "to Raw")
   }

   return nil
}

// Postgres Insert driver for Raw
func (r *Raw) Value() (driver.Value, error) {
   return r.Int.String(), nil
}

// Wrapper functions for big.Int

func NewRaw(integer int64) *Raw {
   var r Raw
   r.Int = big.NewInt(integer)
   return &r
}

func NewFromRaw(raw *Raw) *Raw {
   var r = NewRaw(0)
   r.Int = new(big.Int).Set(raw.Int)
   return r
}

func (r Raw) String() string {
   return r.Int.String()
}

func (r *Raw) Exp(x, y, m *Raw) *Raw {
   if (m == nil) {
      r.Int = r.Int.Exp(x.Int, y.Int, nil)
   } else {
      r.Int = r.Int.Exp(x.Int, y.Int, m.Int)
   }
   return r
}

func (r *Raw) Div(x, y *Raw) *Raw {
   r.Int = r.Int.Div(x.Int, y.Int)
   return r
}

func (r *Raw) Sub(x, y *Raw) *Raw {
   r.Int = r.Int.Sub(x.Int, y.Int)
   return r
}

func (r *Raw) Add(x, y *Raw) *Raw {
   r.Int = r.Int.Add(x.Int, y.Int)
   return r
}

// Cmp compares r and x and returns:
// -1 if x <  y
//  0 if x == y
// +1 if x >  y
func (r *Raw) Cmp(x *Raw) int {
   return r.Int.Cmp(x.Int)
}

func (r *Raw) Mul(x, y *Raw) *Raw {
   r.Int = r.Int.Mul(x.Int, y.Int)
   return r
}
