package main

import (
   "fmt"
   "net/http"
   "strings"
   "io/ioutil"
   "encoding/json"
   "encoding/hex"
   "math/big"
   "context"
   "time"
   "strconv"
   "reflect"

   // Local packages
   keyMan "nanoKeyManager"
   nt "nanoTypes"
)

func getAccountBalance(nanoAddress string) (*nt.Raw, *nt.Raw, error) {

   url := "http://"+ nodeIP

   request :=
   `{
      "action":  "account_balance",
      "account": "`+ nanoAddress +`"
    }`

   response := struct {
      Balance *nt.Raw
      Receivable *nt.Raw
      Error string
   }{}

   err := rcpCallWithTimeout(request, &response, url, 5000)
   if (err != nil) {
      return nil, nil, fmt.Errorf("getAddressBalance: %w", err)
   }

   return response.Balance, response.Receivable, nil
}

func getOwnerOfBlock(hash string) (string, error) {

   url := "http://"+ nodeIP

   hash = strings.ToUpper(hash)

   request :=
   `{
      "action":  "block_account",
      "account": "`+ hash +`"
    }`

   response := struct {
      Account string
      Error string
   }{}

   err := rcpCallWithTimeout(request, &response, url, 5000)
   if (err != nil) {
      return "", fmt.Errorf("getAddressBalance: %w", err)
   }

   return response.Account, nil
}

func confirmBlock(hash string) error {

   url := "http://"+ nodeIP

   hash = strings.ToUpper(hash)

   request :=
   `{
      "action":  "block_confirm",
      "account": "`+ hash +`"
    }`

   response := struct {
      Started string
      Error string
   }{}

   err := rcpCallWithTimeout(request, &response, url, 5000)
   if (err != nil) {
      return fmt.Errorf("getAddressBalance: %w", err)
   }
   if (response.Started != "1") {
      return fmt.Errorf("unknown error, block confirm not started")
   }

   return nil
}

func getBlockCount() (*big.Int, *big.Int, *big.Int, error) {

   url := "http://"+ nodeIP

   request :=
   `{
      "action":  "block_count"
    }`

   response := struct {
      Count nt.Raw
      Unchecked nt.Raw
      Cemented nt.Raw
      Error string
   }{}

   err := rcpCallWithTimeout(request, &response, url, 5000)
   if (err != nil) {
      return nil, nil, nil, fmt.Errorf("getAddressBalance: %w", err)
   }

   return response.Count.Int, response.Unchecked.Int, response.Cemented.Int, nil
}

func printPeers() error {

   url := "http://"+ nodeIP

   request :=
   `{
      "action":       "peers",
      "peer_details": "true"
    }`

   response := struct {
      peers string
      Error string
   }{}

   verboseSave := verbosity
   verbosity = 9
   err := rcpCallWithTimeout(request, &response, url, 5000)
   verbosity = verboseSave
   if (err != nil) {
      return fmt.Errorf("getAddressBalance: %w", err)
   }

   return nil
}

func telemetry() error {

   url := "http://"+ nodeIP

   request :=
   `{
      "action": "telemetry"
    }`

   verboseSave := verbosity
   verbosity = 9
   err := rcpCallWithTimeout(request, nil, url, 5000)
   verbosity = verboseSave
   if (err != nil) {
      return fmt.Errorf("getAddressBalance: %w", err)
   }

   return nil
}

func publishSend(block keyMan.Block, signature []byte, proofOfWork string) (nt.BlockHash, error) {

   url := "http://"+ nodeIP

   sig := strings.ToUpper(hex.EncodeToString(signature))

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
         "signature": "`+ sig +`",
         "work": "`+ proofOfWork +`"
      }
    }`

   response := struct {
      Hash nt.BlockHash
      Error string
   }{}

   err := rcpCallWithTimeout(request, &response, url, 5000)
   if (err != nil) {
      return nil, fmt.Errorf("publishSend: %w", err)
   }

   return response.Hash ,nil
}

func publishReceive(block keyMan.Block, signature []byte, proofOfWork string) (nt.BlockHash, error) {

   url := "http://"+ nodeIP

   sig := strings.ToUpper(hex.EncodeToString(signature))

   request :=
   `{
      "action":  "process",
      "json_block":  "true",
      "subype": "receive",
      "block": {
         "type": "state",
         "account": "`+ block.Account +`",
         "previous": "`+ block.Previous.String() +`",
         "representative": "`+ block.Representative +`",
         "balance": "`+ block.Balance.String() +`",
         "link": "`+ block.Link.String() +`",
         "signature": "`+ sig +`",
         "work": "`+ proofOfWork +`"
      }
    }`

   response := struct {
      Hash nt.BlockHash
      Error string
   }{}

   err := rcpCallWithTimeout(request, &response, url, 5000)
   if (err != nil) {
      return nil, fmt.Errorf("publishSend: %w", err)
   }

   return response.Hash, nil
}

type AccountInfo struct {
   Frontier nt.BlockHash
   OpenBlock nt.BlockHash         `json:"open_block"`
   RepresentativeBlock nt.HexData `json:"representative_block"`
   Representative string          `json:"representative"`
   Balance *nt.Raw
   ModifiedTimestamp nt.JInt      `json:"modified_timestamp"`
   BlockCount nt.JInt             `json:"block_count"`
   Account_Version nt.JInt        `json:"account_version"`
   ConfirmationHeight nt.JInt     `json:"confirmation_height"`
   ConfirmationHeightFrontier nt.BlockHash `json:"confirmation_height_frontier"`
   Error string
}

func getAccountInfo(nanoAddress string) (AccountInfo, error) {

   url := "http://"+ nodeIP

   request :=
   `{
      "action": "account_info",
      "representative": "true",
      "account": "`+ nanoAddress +`"
    }`

   var response AccountInfo

   err := rcpCallWithTimeout(request, &response, url, 5000)
   if (err != nil) {
      return response, fmt.Errorf("getAccountInfo: %w", err)
   }

   return response, nil
}

type AccountHistory struct {
   History []struct {
      Type string
      Account string
      Amount *nt.Raw
      LocalTimestamp nt.JInt `json:"local_timestamp"`
      Height nt.JInt
      Hash nt.BlockHash
      Confirmed nt.JBool
   }
   Previous nt.BlockHash
   Error string
}

func getAccountHistory(nanoAddress string, num int) (AccountHistory, error) {

   url := "http://"+ nodeIP

   request :=
   `{
      "action": "account_history",
      "account": "`+ nanoAddress +`",
      "count": "`+ strconv.Itoa(num) +`"
    }`

   var response AccountHistory
   err := rcpCallWithTimeout(request, &response, url, 5000)
   if (err != nil) {
      return response, fmt.Errorf("getAccountInfo: %w", err)
   }

   return response, nil
}

func getPendingHashes(nanoAddress string) (map[string][]nt.BlockHash, error) {

   url := "http://"+ nodeIP

   request :=
   `{
      "action": "accounts_pending",
      "accounts": ["`+ nanoAddress +`"],
      "count": "-1"
    }`

   response := struct {
      Blocks map[string][]nt.BlockHash
      Error string
   }{}

   err := rcpCallWithTimeout(request, &response, url, 5000)
   if (err != nil) {
      return nil, fmt.Errorf("getPendingHash: %w", err)
   }

   return response.Blocks, nil
}

type BlockInfo struct {
   Amount *nt.Raw
   Contents keyMan.Block
   Height nt.JInt
   LocalTimestamp nt.JInt `json:"local_timestamp"`
   Successor nt.BlockHash
   Confirmed nt.JBool
   Subtype string
   Error string
}

func getBlockInfo(hash nt.BlockHash) (BlockInfo, error) {

   url := "http://"+ nodeIP

   request :=
   `{
      "action": "block_info",
      "json_block": "true",
      "hash": "`+ hash.String() +`"
    }`

   var response BlockInfo
   err := rcpCallWithTimeout(request, &response, url, 5000)
   if (err != nil) {
      return response, fmt.Errorf("getAccountInfo: %w", err)
   }

   return response, nil
}

func generateWorkOnNode(hash nt.BlockHash, difficulty string) (string, error) {

   url := "http://"+ nodeIP

   request :=
   `{
      "action": "work_generate",
      "difficulty": "`+ difficulty +`",
      "use_peers": "true",
      "hash": "`+ hash.String() +`"
    }`

    response := struct {
       Work string
       Difficulty string
       Multiplier string
       Error string
    }{}

   err := rcpCall(request, &response, url, nil)
   if (err != nil) {
      return "", fmt.Errorf("generateWorkOnNode: %w", err)
   }

   return response.Work, nil

}

func generateWorkOnWorkServer(hash nt.BlockHash, difficulty string) (string, error) {

   request :=
   `{
      "action": "work_generate",
      "difficulty": "`+ difficulty +`",
      "hash": "`+ hash.String() +`"
    }`

    response := struct {
       Work string
       Difficulty string
       Multiplier string
       Error string
    }{}

   err := rcpCall(request, &response, workServer, nil)
   if (err != nil) {
      return "", fmt.Errorf("generateWorkOnNode: %w", err)
   }

   return response.Work, nil
}

func rcpCallWithTimeout(request string, response any, url string, ms time.Duration) error {

   ctx, _ := context.WithTimeout(context.Background(), ms*time.Millisecond)
   ch := make(chan error)

   go rcpCall(request, response, url, ch)

   select {
      case <-ctx.Done():
         return fmt.Errorf("rcpCallWithTimeout: rcp call took too long (%d ms)", ms)
      case err := <-ch:
         return err
   }
}

func rcpCall(request string, response any, url string, ch chan error) error {
   var err error
   var checkResponse bool
   defer func() {
      if (ch != nil) {
         ch <- err
      }
   }()

   // Check to make sure response is in the correct format
   if (response != nil) {
      val := reflect.Indirect(reflect.ValueOf(response))
      i := val.NumField() - 1
      if (val.Type().Field(i).Name != "Error") {
         err = fmt.Errorf("rcpCall: response requires 'Error' as last field")
         return err
      }
      checkResponse = true
   }


   if (verbosity >= 10) {
      fmt.Println("request: ", request)
   }

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

   if (verbosity >= 9) {
      fmt.Println(string(body))
   }

   err = json.Unmarshal(body, &response)
   if (err != nil) {
      return fmt.Errorf("Unmarshal: %w", err)
   }

   // Check if there was an error
   if (checkResponse) {
      val := reflect.Indirect(reflect.ValueOf(response))
      i := val.NumField() - 1
      errString := val.Field(i).String()
      if (errString != "") {
         err = fmt.Errorf("rcpCall: node returned error: %s", errString)
         return err
      }
   }

   return nil
}

