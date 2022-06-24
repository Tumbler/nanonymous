package nanoTypes

import (
   "fmt"
   "math/big"
   "database/sql/driver"
   "strings"
   "strconv"
)


// Raw represents an amount of raw nano.
type Raw struct {
    *big.Int
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
      return fmt.Errorf("Can't assign %s to Raw", src)
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

// WARNING: This function can have rounding errors if the Nano gets too small.
func NewRawFromNano(nano float64) *Raw {
   expanded := int64(nano * 1000000000000000)
   raw := NewRaw(0).Mul(NewRaw(expanded), NewRaw(0).Exp(NewRaw(10), NewRaw(15), nil))
   return raw
}

func (r *Raw) SetString(s string, base int) (*Raw, bool) {
   _, ok := r.Int.SetString(s, base)
   return r, ok
}

func OneNano() *Raw {
   return NewRaw(0).Exp(NewRaw(10), NewRaw(30), nil)
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

func (r *Raw) DivMod(x, y *Raw) (*Raw, *Raw) {

   z := NewRaw(0)
   r.Int, _ = r.Int.DivMod(x.Int, y.Int, z.Int)
   return r, z
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
// -1 if r <  x
//  0 if r == x
// +1 if r >  x
func (r *Raw) Cmp(x *Raw) int {
   return r.Int.Cmp(x.Int)
}

func (r *Raw) Mul(x, y *Raw) *Raw {
   r.Int = r.Int.Mul(x.Int, y.Int)
   return r
}
