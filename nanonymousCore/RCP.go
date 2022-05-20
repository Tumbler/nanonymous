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

   // Local packages
   keyMan "nanoKeyManager"
)

func getAccountBalance(nanoAddress string) (*keyMan.Raw, *keyMan.Raw, error) {

   url := "http://"+ nodeIP

   request :=
   `{
      "action":  "account_balance",
      "account": "`+ nanoAddress +`"
    }`

   response := struct {
      Balance *keyMan.Raw
      Receivable *keyMan.Raw
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
      Count keyMan.Raw
      Unchecked keyMan.Raw
      Cemented keyMan.Raw
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

func publishSend(block keyMan.Block, signature []byte, proofOfWork string) (keyMan.BlockHash, error) {

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
      Hash keyMan.BlockHash
   }{}

   err := rcpCallWithTimeout(request, &response, url, 5000)
   if (err != nil) {
      return nil, fmt.Errorf("publishSend: %w", err)
   }

   return response.Hash ,nil
}

func publishReceive(block keyMan.Block, signature []byte, proofOfWork string) (keyMan.BlockHash, error) {

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
      Hash keyMan.BlockHash
   }{}

   err := rcpCallWithTimeout(request, &response, url, 5000)
   if (err != nil) {
      return nil, fmt.Errorf("publishSend: %w", err)
   }

   return response.Hash, nil
}

type AccountInfo struct {
   Frontier keyMan.BlockHash
   OpenBlock keyMan.BlockHash         `json:"open_block"`
   RepresentativeBlock keyMan.HexData `json:"representative_block"`
   Representative string              `json:"representative"`
   Balance *keyMan.Raw
   ModifiedTimestamp keyMan.JInt      `json:"modified_timestamp"`
   BlockCount keyMan.JInt             `json:"block_count"`
   Account_Version keyMan.JInt        `json:"account_version"`
   ConfirmationHeight keyMan.JInt     `json:"confirmation_height"`
   ConfirmationHeightFrontier keyMan.BlockHash `json:"confirmation_height_frontier"`
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
      Amount *keyMan.Raw
      LocalTimestamp keyMan.JInt `json:"local_timestamp"`
      Height keyMan.JInt
      Hash keyMan.BlockHash
      Confirmed keyMan.JBool
   }
   Previous keyMan.BlockHash
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

func getPendingHash(nanoAddress string) map[string][]keyMan.BlockHash {

   url := "http://"+ nodeIP

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

   err := rcpCallWithTimeout(request, &response, url, 5000)
   if (err != nil) {
      //return fmt.Errorf("getAccountInfo: %w", err)
   }

   return response.Blocks
}

type BlockInfo struct {
   Amount *keyMan.Raw
   Contents keyMan.Block
   Height keyMan.JInt
   LocalTimestamp keyMan.JInt `json:"local_timestamp"`
   Successor keyMan.BlockHash
   Confirmed keyMan.JBool
   Subtype string
}

func getBlockInfo(hash keyMan.BlockHash) (BlockInfo, error) {

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

func generateWorkOnNode(hash keyMan.BlockHash, difficulty string) (string, error) {

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
    }{}

   err := rcpCall(request, &response, url, nil)
   if (err != nil) {
      return "", fmt.Errorf("generateWorkOnNode: %w", err)
   }

   return response.Work, nil

}

func generateWorkOnWorkServer(hash keyMan.BlockHash, difficulty string) (string, error) {

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
   defer func() {
      if (ch != nil) {
         ch <- err
      }
   }()

   // TODO capure error response from node??

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

   return nil
}

