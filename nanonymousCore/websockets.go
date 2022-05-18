package main


import (
   "fmt"
   "time"
   "strings"

   "golang.org/x/net/websocket"

   // Local packages
   keyMan "nanoKeyManager"
)

// Port 80 (unsecured) and 443 (secure)

type ConfirmationBlock struct {
   Time keyMan.JInt
   Message struct {
      Block keyMan.Block
      Account string
      Amount *keyMan.Raw
      Hash keyMan.BlockHash
      ConfirmationType string `json:"confirmation_type"`
   }
}

var addWebSocketSubscription chan string

// TODO this is called in a go routine so returned errors are ignored
func websocketListener() error {
   if (verbose) {
      fmt.Println("Started listening to websockets on:", websocketAddress)
   }

   ws, err := websocket.Dial(websocketAddress, "", "http://localhost/")
   if (err != nil) {
      return fmt.Errorf("websocketListener, Dial: %w", err)
   }
   defer ws.Close()

   err = startSubscription(ws)
   if (err != nil) {
      return fmt.Errorf("websocketListener: %w", err)
   }

   addWebSocketSubscription = make(chan string)

   var notification ConfirmationBlock
   wsChan := make(chan error)
   go websocketReceive(ws, &notification, wsChan)

   for {
      select {
         case err := <-wsChan:
            if (err != nil) {
               // log
               fmt.Println(" err: ", err.Error())
            } else {
               fmt.Println("notification!")
               handleNotification(notification)
            }

            // Set up next receive
            go websocketReceive(ws, &notification, wsChan)
         case nanoAddress := <-addWebSocketSubscription:
            addToSubscription(ws, nanoAddress)
      }
   }

   // Should never get here TODO log
   return nil
}

func startSubscription(ws *websocket.Conn) error {

   var addressString string

   for i := 1 ;; i++{
      seed, _ := getSeedFromDatabase(i)
      if (len(seed) == 0) {
         // no more seeds
         // TODO make single DB call?
         break
      }
      maxIndex, _ := getCurrentIndexFromDatabase(i)
      for j := 0; j <= maxIndex; j++ {
         var key keyMan.Key
         key.Seed = seed
         key.Index = j
         err := keyMan.SeedToKeys(&key)
         if (err != nil) {
            return fmt.Errorf("startSubscription: %w", err)
         }

         addressString += `"`+ key.NanoAddress + `", `
      }
   }

   addressString = strings.Trim(addressString, ", ")

   request :=
   `{
      "action": "subscribe",
      "topic": "confirmation",
      "ack": true,
      "options": {
         "confirmation_type": "active_quorum",
         "accounts" : [`+ addressString +`]
      }
   }`

   if (verbose) {
      fmt.Print(request)
   }

   _, err := ws.Write([]byte(request))
   if (err != nil) {
      return fmt.Errorf("websocketListener, Send: %w", err)
   }

   response := struct {
      Ack string
   }{}

   ch := make(chan error)
   go websocketReceive(ws, &response, ch)

   select {
      case err := <-ch:
         if (err != nil) {
            return fmt.Errorf("startSubscription: %w", err)
         }
         if (response.Ack != "subscribe") {
            return fmt.Errorf("startSubscription: no ack")
         }
         if (verbose) {
            fmt.Println("Websocket sucessfully opened!")
         }
         return nil
      case <-time.After(5 * time.Second):
        return fmt.Errorf("startSubscription: timeout")
   }

   return nil
}

func websocketReceive(ws *websocket.Conn, r any, ch chan error) {
   err := websocket.JSON.Receive(ws, r)
   ch <- err
}

func addToSubscription(ws *websocket.Conn, nanoAddress string) {

   request :=
   `{
      "action": "update",
      "topic": "confirmation",
      "options": {
         "accounts_add" : ["`+ nanoAddress +`"]
      }
   }`

   ws.Write([]byte(request))
}

func handleNotification(cBlock ConfirmationBlock) {

   msg := cBlock.Message
   if (msg.Block.SubType == "send") {
      fmt.Println("Send")
      if (addressExsistsInDB(msg.Block.LinkAsAccount)) {
         // TODO check if any transaction manager is expecting this and give to them instead
         fmt.Println("Managed address")
         if (addressExsistsInDB(msg.Account)) {
            fmt.Println("Internal Send")
            // Internal network send, don't trigger a transaction
            //TODO debugging code
            seed, _ := getSeedFromIndex(1, 4)
            if (strings.Compare(msg.Account, seed.NanoAddress) == 0) {
               fmt.Println("Debug transaction!")
               receivedNano(msg.Block.LinkAsAccount)
            }
            //TODO end of debugging code
         } else {
            fmt.Println("External Send")
            if (verbose) {
               fmt.Println("Starting Transaction!")
            }
            receivedNano(msg.Block.LinkAsAccount)
         }
      }
   }
}
