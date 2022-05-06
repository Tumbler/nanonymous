package main

import (
   "fmt"
   "net/http"
   "strings"
   "io/ioutil"
   "encoding/json"
   "math/big"

   // 3rd party packages
   //"github.com/shopspring/decimal"
   //curl "github.com/andelf/go-curl"
)

func getAddressBalance(nanoAddress string) (*big.Int, error) {

   request :=
   `{
      "action":  "account_balance",
      "account": "`+ nanoAddress +`"
    }`

   response := struct {
      Balance BigInt
      Receivable BigInt
   }{}

   url := "http://"+ nodeIP

   err := rcpCall(url, request, &response)
   if (err != nil) {
      return nil, fmt.Errorf("getAddressBalance: %w", err)
   }

   if (verbose) {
      fmt.Println("Balance: ", response.Balance.Int.String())
      fmt.Println("receivable: ", response.Receivable.Int.String())
   }

   return &response.Balance.Int, nil
}

func rcpCall(url string, request string, response any) error {
   // TODO error out if connection takes too long

   payload := strings.NewReader(request)
   req, err := http.NewRequest("POST", url, payload)
   if (err != nil) {
      return fmt.Errorf("NewRequest: %w", err)
   }

   res, err := http.DefaultClient.Do(req)
   if (err != nil) {
      return fmt.Errorf("DefaultClient: %w", err)
   }
   defer res.Body.Close()

   body, err := ioutil.ReadAll(res.Body)
   if (err != nil) {
      return fmt.Errorf("readAll: %w", err)
   }

   if (verbose) {
      fmt.Println(string(body))
   }

   err = json.Unmarshal(body, &response)
   if (err != nil) {
      return fmt.Errorf("error: %w", err)
   }

   return nil
}



// JSON BigInt Marshaler
type BigInt struct {
    big.Int
}

func (b BigInt) MarshalJSON() ([]byte, error) {
    return []byte(b.String()), nil
}

func (b *BigInt) UnmarshalJSON(p []byte) error {
    if string(p) == "null" {
        return nil
    }
    trimmed := strings.Trim(string(p), `"`)
    var z big.Int
    _, ok := z.SetString(trimmed, 10)
    if !ok {
        return fmt.Errorf("not a valid big integer: %s", p)
    }
    b.Int = z
    return nil
}
