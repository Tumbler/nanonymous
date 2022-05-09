package main

import (
   "fmt"
   "net/http"
   "strings"
   "io/ioutil"
   "encoding/json"
   "math/big"

   // Local packages
   keyMan "nanoKeyManager"

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
      Balance keyMan.Raw
      Receivable keyMan.Raw
   }{}

   err := rcpCall(request, &response)
   if (err != nil) {
      return nil, fmt.Errorf("getAddressBalance: %w", err)
   }

   return response.Balance.Int, nil
}

// TODO test
func getOwnerOfBlock(hash string) (string, error) {

   hash = strings.ToUpper(hash)

   request :=
   `{
      "action":  "block_account",
      "account": "`+ hash +`"
    }`

   response := struct {
      Account string
   }{}

   err := rcpCall(request, &response)
   if (err != nil) {
      return "", fmt.Errorf("getAddressBalance: %w", err)
   }

   return response.Account, nil
}

// TODO test
func confirmBlock(hash string) error {

   hash = strings.ToUpper(hash)

   request :=
   `{
      "action":  "block_confirm",
      "account": "`+ hash +`"
    }`

   response := struct {
      Started string
   }{}

   err := rcpCall(request, &response)
   if (err != nil) {
      return fmt.Errorf("getAddressBalance: %w", err)
   }
   if (response.Started != "1") {
      return fmt.Errorf("unknown error, block confirm not started")
   }

   return nil
}

func getBlockCount() (*big.Int, *big.Int, *big.Int, error) {

   request :=
   `{
      "action":  "block_count"
    }`

   response := struct {
      Count keyMan.Raw
      Unchecked keyMan.Raw
      Cemented keyMan.Raw
   }{}

   err := rcpCall(request, &response)
   if (err != nil) {
      return nil, nil, nil, fmt.Errorf("getAddressBalance: %w", err)
   }

   return response.Count.Int, response.Unchecked.Int, response.Cemented.Int, nil
}

func printPeers() error {

   request :=
   `{
      "action":       "peers",
      "peer_details": "true"
    }`

   response := struct {
      peers string
   }{}

   verboseSave := verbose
   verbose = true
   err := rcpCall(request, &response)
   verbose = verboseSave
   if (err != nil) {
      return fmt.Errorf("getAddressBalance: %w", err)
   }

   return nil
}

func telemetry() error {

   request :=
   `{
      "action": "telemetry"
    }`

   verboseSave := verbose
   verbose = true
   err := rcpCall(request, nil)
   verbose = verboseSave
   if (err != nil) {
      return fmt.Errorf("getAddressBalance: %w", err)
   }

   return nil
}

func publishSend(nanoAddress string) error {

   previousHash := "4567896456"
   rep := "nano_33t5by1653nt196hfwm5q3wq7oxtaix97r7bhox5zn8eratrzoqsny49ftsd"
   newBalance := big.NewInt(41)
   receiveAddressPub := "345678976546789900003250235023502350"
   receiveAddress := "nano133895sd6645876ds875s8dfsd8fas8df75"
   blockSignature := "45678987656789875456789"
   proofOfWork := "00bfb848"

   request :=
   `{
      "action":  "process",
      "json_block":  "true",
      "subype": "send",
      "block": {
         "type": "state",
         "account": "`+ nanoAddress +`",
         "previous": "`+ previousHash +`",
         "representative": "`+ rep +`",
         "balance": "`+ newBalance.String() +`",
         "link": "`+ receiveAddressPub +`",
         "link_as_account"": "`+ receiveAddress +`",
         "signature": "`+ blockSignature +`",
         "work": "`+ proofOfWork +`"
      }
    }`

   response := struct {
      Hash string
   }{}

   err := rcpCall(request, &response)
   if (err != nil) {
      return fmt.Errorf("publishSend: %w", err)
   }

   return nil
}

func rcpCall(request string, response any) error {
   // TODO error out if connection takes too long

   url := "http://"+ nodeIP
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

