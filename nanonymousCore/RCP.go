package main

import (
   "fmt"
   "net/http"
   "strings"
   "io/ioutil"
   "encoding/json"
   "encoding/hex"
   "math/big"

   // Local packages
   keyMan "nanoKeyManager"

   // 3rd party packages
   //"github.com/shopspring/decimal"
   //curl "github.com/andelf/go-curl"
)

func getAccountBalance(nanoAddress string) (*keyMan.Raw, *keyMan.Raw, error) {

   request :=
   `{
      "action":  "account_balance",
      "account": "`+ nanoAddress +`"
    }`

   response := struct {
      Balance *keyMan.Raw
      Receivable *keyMan.Raw
   }{}

   err := rcpCall(request, &response)
   if (err != nil) {
      return nil, nil, fmt.Errorf("getAddressBalance: %w", err)
   }

   return response.Balance, response.Receivable, nil
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

func publishSend(block keyMan.Block, signature []byte) error {
   proofOfWork := "00bfb848"

   request :=
   `{
      "action":  "process",
      "json_block":  "true",
      "subype": "send",
      "block": {
         "type": "state",
         "account": "`+ block.Account +`",
         "previous": "`+ block.Previous.String() +`",
         "representative": "`+ block.Representative +`",
         "balance": "`+ block.Balance.String() +`",
         "link": "`+ block.Link.String() +`",
         "link_as_account"": "`+ block.LinkAsAccount +`",
         "signature": "`+ string(signature) +`",
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

func openAccount(block keyMan.Block, signature []byte, proofOfWork string) error {

   sig := strings.ToUpper(hex.EncodeToString(signature))

   request :=
   `{
      "action":  "process",
      "json_block":  "true",
      "subype": "receive",
      "block": {
         "type": "state",
         "account": "`+ block.Account +`",
         "previous": "0000000000000000000000000000000000000000000000000000000000000000",
         "representative": "`+ block.Representative +`",
         "balance": "`+ block.Balance.String() +`",
         "link": "`+ block.Link.String() +`",
         "signature": "`+ sig +`",
         "work": "`+ proofOfWork +`"
      }
    }`

   fmt.Println("request:\r\n", request)

   response := struct {
      Hash string
   }{}

   err := rcpCall(request, &response)
   if (err != nil) {
      return fmt.Errorf("publishSend: %w", err)
   }

   return nil
}

func getAccountInfo(nanoAddress string) error {

   request :=
   `{
      "action": "account_info",
      "account": "`+ nanoAddress +`"
    }`

   response := struct {
      Frontier keyMan.HexData
      OpenBlock keyMan.HexData           `json:"open_block"`
      RepresentativeBlock keyMan.HexData `json:"representative_block"`
      Balance *keyMan.Raw
      ModifiedTimestamp int              `json:"modified_timestamp"`
      BlockCount int64                   `json:"block_count"`
      Account_Version int64              `json:"account_version"`
      ConfirmationHeight int64           `json:"confirmation_height"`
      ConfirmationHeightFrontier int64   `json:"confirmation_height_frontier"`
   }{}

   verbose = true
   err := rcpCall(request, &response)
   if (err != nil) {
      return fmt.Errorf("getAccountInfo: %w", err)
   }

   return nil
}

func getAccountHistory(nanoAddress string) error {

   request :=
   `{
      "action": "account_history",
      "account": "`+ nanoAddress +`",
      "count": "1"
    }`

   response := struct {
      Frontier keyMan.HexData
      OpenBlock keyMan.HexData           `json:"open_block"`
      RepresentativeBlock keyMan.HexData `json:"representative_block"`
      Balance *keyMan.Raw
      ModifiedTimestamp int              `json:"modified_timestamp"`
      BlockCount int64                   `json:"block_count"`
      Account_Version int64              `json:"account_version"`
      ConfirmationHeight int64           `json:"confirmation_height"`
      ConfirmationHeightFrontier int64   `json:"confirmation_height_frontier"`
   }{}

   verbose = true
   err := rcpCall(request, &response)
   if (err != nil) {
      return fmt.Errorf("getAccountInfo: %w", err)
   }

   return nil
}

func getPendingHash(nanoAddress string) map[string][]keyMan.BlockHash {

   request :=
   `{
      "action": "accounts_pending",
      "accounts": ["`+ nanoAddress +`"],
      "count": "1"
    }`

   response := struct {
      Blocks map[string][]keyMan.BlockHash
   }{}

   //var response map[string]interface{}

   verbose = true
   err := rcpCall(request, &response)
   if (err != nil) {
      //return fmt.Errorf("getAccountInfo: %w", err)
   }

   return response.Blocks
}

func getBlockInfo(hash keyMan.BlockHash) (keyMan.Block, error) {

   request :=
   `{
      "action": "block_info",
      "json_block": "true",
      "hash": "`+ hash.String() +`"
    }`
    fmt.Println("request: ", request)

    response := struct {
       Contents keyMan.Block
    }{}

   verbose = true
   err := rcpCall(request, &response)
   if (err != nil) {
      return response.Contents, fmt.Errorf("getAccountInfo: %w", err)
   }

   return response.Contents, nil
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

