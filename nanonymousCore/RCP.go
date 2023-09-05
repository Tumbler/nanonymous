package main

import (
   "fmt"
   "net"
   "net/http"
   "net/smtp"
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

   url := "https://"+ nodeIP

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

func getAccountBlockCount(nanoAddress string) (int, error) {

   url := "https://"+ nodeIP

   request :=
   `{
      "action":  "account_block_count",
      "account": "`+ nanoAddress +`"
    }`

   response := struct {
      BlockCount nt.JInt `json:"block_count"`
      Error string
   }{}

   err := rcpCallWithTimeout(request, &response, url, 5000)
   if (err != nil) {
      return 0, fmt.Errorf("getAddressBalance: %w", err)
   }

   return int(response.BlockCount), nil
}

func getOwnerOfBlock(hash string) (string, error) {

   url := "https://"+ nodeIP

   hash = strings.ToUpper(hash)

   request :=
   `{
      "action":  "block_account",
      "hash": "`+ hash +`"
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

   url := "https://"+ nodeIP

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

   url := "https://"+ nodeIP

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

   url := "https://"+ nodeIP

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
   verbosity = verboseSave

   return nil
}

func telemetry() error {

   url := "https://"+ nodeIP

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

   url := "https://"+ nodeIP

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

   url := "https://"+ nodeIP

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
   AccountVersion nt.JInt         `json:"account_version"`
   ConfirmationHeight nt.JInt     `json:"confirmation_height"`
   ConfirmationHeightFrontier nt.BlockHash `json:"confirmation_height_frontier"`
   Error string
}

func getAccountInfo(nanoAddress string) (AccountInfo, error) {

   url := "https://"+ nodeIP

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

func getAccountRep(nanoAddress string) (AccountInfo, error) {
   url := "https://"+ nodeIP

   request :=
   `{
      "action": "account_representative",
      "account": "`+ nanoAddress +`"
    }`

   var response AccountInfo

   err := rcpCallWithTimeout(request, &response, url, 5000)
   if (err != nil) {
      return response, fmt.Errorf("getAccountRep: %w", err)
   }

   return response, nil
}

func getAccountWeight(nanoAddress string) (*nt.Raw, error) {
   url := "https://"+ nodeIP

   request :=
   `{
      "action": "account_weight",
      "account": "`+ nanoAddress +`"
    }`

   response := struct {
      Weight *nt.Raw
      Error string
   }{}

   err := rcpCallWithTimeout(request, &response, url, 5000)
   if (err != nil) {
      return nil, fmt.Errorf("getAccountRep: %w", err)
   }

   return response.Weight, nil
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

   url := "https://"+ nodeIP

   request :=
   `{
      "action": "account_history",
      "account": "`+ nanoAddress +`",
      "count": "`+ strconv.Itoa(num) +`"
    }`

   var response AccountHistory
   err := rcpCallWithTimeout(request, &response, url, 5000)
   if (err != nil) {
      return response, fmt.Errorf("getAccountHistory: %w", err)
   }

   return response, nil
}

func getAccountsPending(nanoAddresses []string) (map[string][]nt.BlockHash, error) {

   url := "https://"+ nodeIP

   var addressString string
   for _, address := range nanoAddresses {
      addressString += `"`+ address + `", `
   }
   addressString = strings.Trim(addressString, ", ")

   request :=
   `{
      "action": "accounts_pending",
      "accounts": [`+ addressString +`],
      "count": "-1"
    }`

   response := struct {
      Blocks map[string][]nt.BlockHash
      Error string
   }{}

   err := rcpCallWithTimeout(request, &response, url, 5000)
   if (err != nil) {
      // Filter out blank block error
      if !(strings.Contains(err.Error(), "cannot unmarshal string into Go struct field .Blocks")) {
         return nil, fmt.Errorf("getAccountsPending: %w", err)
      }
   }

   return response.Blocks, nil
}

// Same as above, but for only one account
func getReceivable(nanoAddress string, count int) ([]nt.BlockHash, error) {

   url := "https://"+ nodeIP

   request :=
   `{
      "action": "receivable",
      "account": "`+ nanoAddress +`",
      "count": "`+ strconv.Itoa(count) +`"
    }`

   response := struct {
      Blocks []nt.BlockHash
      Error string
   }{}

   err := rcpCallWithTimeout(request, &response, url, 5000)
   if (err != nil) {
      // Filter out blank block error
      if !(strings.Contains(err.Error(), "cannot unmarshal string into Go struct field .Blocks")) {
         return nil, fmt.Errorf("getReceivable: %w", err)
      }
   }

   return response.Blocks, nil
}

func republish(hash nt.BlockHash) error {
   url := "https://"+ nodeIP

   request :=
   `{
      "action": "republish",
      "hash": "`+ hash.String() +`"
    }`

   response := struct {
      Error string
   }{}

   verboseSave := verbosity
   verbosity = 9
   err := rcpCallWithTimeout(request, &response, url, 5000)
   if (err != nil) {
      return fmt.Errorf("republish: %w", err)
   }
   verbosity = verboseSave

   return nil
}

func printStats(argument string) error {

   url := "https://"+ nodeIP

   request :=
   `{
      "action": "stats",
      "type": "`+ argument +`"
    }`

   response := struct {
      Error string
   }{}

   verboseSave := verbosity
   verbosity = 9
   err := rcpCallWithTimeout(request, &response, url, 5000)
   if (err != nil) {
      return fmt.Errorf("printStats: %w", err)
   }
   verbosity = verboseSave

   return nil
}

func getSuccessors(block nt.BlockHash, count int) ([]nt.BlockHash, error) {

   url := "https://"+ nodeIP

   request :=
   `{
      "action": "successors",
      "block": "`+ block.String() +`",
      "count": "`+ strconv.Itoa(count) +`"
    }`

   response := struct {
      Blocks []nt.BlockHash
      Error string
   }{}

   err := rcpCallWithTimeout(request, &response, url, 5000)
   if (err != nil) {
      // Filter out blank block error
      if !(strings.Contains(err.Error(), "cannot unmarshal string into Go struct field .Blocks")) {
         return nil, fmt.Errorf("getSuccessors: %w", err)
      }
   }

   return response.Blocks, nil

}

func printTelemetry() error {

   url := "https://"+ nodeIP

   request :=
   `{
      "action": "telemetry"
    }`

   response := struct {
      Error string
   }{}

   verboseSave := verbosity
   verbosity = 9
   err := rcpCallWithTimeout(request, &response, url, 5000)
   if (err != nil) {
      return fmt.Errorf("printTelemetry: %w", err)
   }
   verbosity = verboseSave

   return nil
}

func printVersion() error {

   url := "https://"+ nodeIP

   request :=
   `{
      "action": "version"
    }`

   response := struct {
      Error string
   }{}

   verboseSave := verbosity
   verbosity = 9
   err := rcpCallWithTimeout(request, &response, url, 5000)
   if (err != nil) {
      return fmt.Errorf("printVersion: %w", err)
   }
   verbosity = verboseSave

   return nil
}

func getUncheckedBlocks(count int) (map[string]keyMan.Block, error) {

   url := "https://"+ nodeIP

   request :=
   `{
      "action": "unchecked",
      "json_block": "true",
      "count": "`+ strconv.Itoa(count) +`"
    }`

   response := struct {
      Blocks map[string]keyMan.Block
      Error string
   }{}

   err := rcpCallWithTimeout(request, &response, url, 5000)
   if (err != nil) {
      // Filter out blank block error
      if !(strings.Contains(err.Error(), "cannot unmarshal string into Go struct field .Blocks")) {
         return nil, fmt.Errorf("getUncheckedBlocks: %w", err)
      }
   }

   return response.Blocks, nil
}

func getUptime() (int, error) {

   url := "https://"+ nodeIP

   request :=
   `{
      "action": "uptime"
    }`

   response := struct {
      Seconds nt.JInt
      Error string
   }{}

   err := rcpCallWithTimeout(request, &response, url, 5000)
   if (err != nil) {
      return 0, fmt.Errorf("getUptime: %w", err)
   }

   return int(response.Seconds), nil
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

   url := "https://"+ nodeIP

   request :=
   `{
      "action": "block_info",
      "json_block": "true",
      "hash": "`+ hash.String() +`"
    }`

   var response BlockInfo
   err := rcpCallWithTimeout(request, &response, url, 5000)
   if (err != nil) {
      return response, fmt.Errorf("getBlockInfo: %w", err)
   }

   return response, nil
}

func getAvailableSupply () (*nt.Raw, error) {
   url := "https://"+ nodeIP

   request :=
   `{
      "action": "available_supply"
    }`

    response := struct {
       Available *nt.Raw
       Error string
    }{}

   err := rcpCall(request, &response, url, nil)
   if (err != nil) {
      return nil, fmt.Errorf("getAvailableSupply: %w", err)
   }

   return response.Available, nil
}

func getNumberOfDelegators(nanoAddress string) (int, error) {
   url := "https://"+ nodeIP

   request :=
   `{
      "action": "delegators_count",
      "account": "`+ nanoAddress +`"
    }`

   response := struct {
      Count nt.JInt
      Error string
   }{}

   err := rcpCallWithTimeout(request, &response, url, 5000)
   if (err != nil) {
      return -1, fmt.Errorf("getAccountHistory: %w", err)
   }

   return int(response.Count), nil
}

func getFrontierCount() (int, error) {
   url := "https://"+ nodeIP

   request :=
   `{
      "action": "frontier_count"
    }`

   response := struct {
      Count nt.JInt
      Error string
   }{}

   err := rcpCallWithTimeout(request, &response, url, 5000)
   if (err != nil) {
      return -1, fmt.Errorf("getAccountHistory: %w", err)
   }

   return int(response.Count), nil
}

func printBootstrapStatus() error {

   url := "https://"+ nodeIP

   request :=
   `{
      "action": "bootstrap_status"
    }`

   response := struct {
      Error string
   }{}

   verboseSave := verbosity
   verbosity = 9
   err := rcpCall(request, &response, url, nil)
   if (err != nil) {
      return fmt.Errorf("printBootstrapStatus: %w", err)
   }
   verbosity = verboseSave

   return nil
}

func printConfirmationQuorum() error {

   url := "https://"+ nodeIP

   request :=
   `{
      "action": "confirmation_quorum"
    }`

   response := struct {
      Error string
   }{}

   verboseSave := verbosity
   verbosity = 9
   err := rcpCall(request, &response, url, nil)
   if (err != nil) {
      return fmt.Errorf("printBootstrapStatus: %w", err)
   }
   verbosity = verboseSave

   return nil
}

func getBlocksInChain(block nt.BlockHash, count int) ([]nt.BlockHash, error) {

   url := "https://"+ nodeIP

   request :=
   `{
      "action": "chain",
      "block": "`+ block.String() +`",
      "count": "`+ strconv.Itoa(count) +`"
    }`

   response := struct {
      Blocks []nt.BlockHash
      Error string
   }{}

   err := rcpCall(request, &response, url, nil)
   if (err != nil) {
      return nil, fmt.Errorf("getBlocksInChain: %w", err)
   }

   return response.Blocks, nil

}

func getActiveConfirmations() (string, int, int, error) {

   url := "https://"+ nodeIP

   request :=
   `{
      "action": "confirmation_active"
    }`

   response := struct {
      Confirmations string
      Unconfirmed nt.JInt
      Confirmed nt.JInt
      Error string
   }{}

   err := rcpCall(request, &response, url, nil)
   if (err != nil) {
      return "", 0, 0, fmt.Errorf("getActiveConfirmations: %w", err)
   }

   return response.Confirmations, int(response.Unconfirmed), int(response.Confirmed), nil

}

func printConfirmationHistory() error {

   url := "https://"+ nodeIP

   request :=
   `{
      "action": "confirmation_history"
    }`

   response := struct {
      Error string
   }{}

   verboseSave := verbosity
   verbosity = 9
   err := rcpCall(request, &response, url, nil)
   if (err != nil) {
      return fmt.Errorf("printBootstrapStatus: %w", err)
   }
   verbosity = verboseSave

   return nil

}

func generateWorkOnNode(hash nt.BlockHash, difficulty string) (string, error) {

   url := "https://"+ nodeIP

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
      "use_peers": "true",
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
      return "", fmt.Errorf("generateWorkOnWorkServer: %w", err)
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

// Use a custom client to exend time of TLS handshake timeout.
var longTLStransport = &http.Transport{
   Dial: (&net.Dialer{
      Timeout: 60 * time.Second,
      KeepAlive: 30 * time.Second,
   }).Dial,
   TLSHandshakeTimeout: 30 * time.Second,
}
var nanonymousHttpClient = &http.Client{
   Transport: longTLStransport,
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

   // If not json this will need to change.
   req.Header.Add("Content-Type", "application/json")

   res, err := nanonymousHttpClient.Do(req)
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

// getNanoUSDValue uses the coingecko API to find the current price of nano in
// USD. This function isn't actually RCP, but it doesn't fit better anywhere
// else.
func getNanoUSDValue() (float64, error) {
   url := `https://api.coingecko.com/api/v3/simple/price?ids=nano&vs_currencies=usd`
   res, err := http.Get(url)
   if (err != nil) {
      return 0.0, fmt.Errorf("getNanoUSDValue: %w", err)
   }
   defer res.Body.Close()

   body, err := ioutil.ReadAll(res.Body)
   if (err != nil) {
      return 0.0, fmt.Errorf("getNanoUSDValue: %w", err)
   }

   response := struct {
      Nano struct {
         Usd float64
      }
   }{}
   err = json.Unmarshal(body, &response)
   if (err != nil) {
      return 0.0, fmt.Errorf("getNanoUSDValue: %w", err)
   }

   return response.Nano.Usd, nil
}

func sendEmail(contents string) error {
   from := fromEmail
   password := emailPass

   to := []string {
      toEmail,
   }

   smtpHost := "smtp.gmail.com"
   smtpPort := "587"

   message := []byte(contents)

   auth := smtp.PlainAuth("", from, password, smtpHost)

   err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, message)
   if (err != nil) {
      return fmt.Errorf("sendEmail: %w", err)
   }

   return nil
}
