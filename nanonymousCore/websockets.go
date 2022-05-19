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

var registeredSendChannels map[string]chan string
var registeredReceiveChannels map[string]chan string

// TODO this is called in a go routine so returned errors are ignored
func websocketListener() error {
   // TODO needs to work over secure connection (wss)
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
   registeredSendChannels = make(map[string]chan string)
   registeredReceiveChannels = make(map[string]chan string)

   var notification ConfirmationBlock
   wsChan := make(chan error)
   go websocketReceive(ws, &notification, wsChan)

   // Listen for eternity
   for {
      select {
         case err := <-wsChan:
            if (err != nil) {
               // TODO log
               if (verbose) {
                  fmt.Println(" err: ", err.Error())
               }
            } else {
               go handleNotification(notification)
            }

            // Set up next receive
            go websocketReceive(ws, &notification, wsChan)

         case nanoAddress := <-addWebSocketSubscription:
            go addToSubscription(ws, nanoAddress)
      }
   }

   // Should never get here TODO log
   return nil
}

func startSubscription(ws *websocket.Conn) error {

   var addressString string

   // Get list of all accounts we've opened so far
   rows, err := getSeedRowsFromDatabase()
   var seed []byte
   var maxIndex int
   // For all seeds find their acounts
   for rows.Next() {
      err = rows.Scan(&seed, &maxIndex)
      if (err != nil || len(seed) == 0) {
         break
      }

      // From 0 to the last account we've opened
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

   _, err = ws.Write([]byte(request))
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

   if (verbose) {
      fmt.Println("Adding to subscription: ", nanoAddress)
   }

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
   wg.Add(1)
   defer wg.Done()

   msg := cBlock.Message
   if (msg.Block.SubType == "send") {
      if (addressExsistsInDB(msg.Block.LinkAsAccount)) {
         // Check if any transaction manager is expecting this and give to them instead
         if (addressExsistsInDB(msg.Account)) {
            if (verbose) {
               fmt.Println("Internal Send")
            }
            // Internal network send, don't trigger a transaction
            //TODO debugging code
            seed, _ := getSeedFromIndex(1, 0)
            if (strings.Compare(msg.Account, seed.NanoAddress) == 0) {
               fmt.Println("Debug transaction!")
               receivedNano(msg.Block.LinkAsAccount)
            } else {
               //TODO end of debugging code
               if (registeredSendChannels[msg.Block.LinkAsAccount] != nil) {
                  select {
                     case registeredSendChannels[msg.Block.LinkAsAccount] <- msg.Hash.String():
                     case <-time.After(5 * time.Minute):
                        // TODO log
                  }
               }
            }
         } else {
            if (verbose) {
               fmt.Println("External Send")
               fmt.Println("Starting Transaction!")
            }
            receivedNano(msg.Block.LinkAsAccount)
         }
      }
   } else {
      if (verbose) {
         fmt.Println(" Receive")
      }
      if (registeredReceiveChannels[msg.Account] != nil) {
         select {
            case registeredReceiveChannels[msg.Account] <- msg.Hash.String():
            case <-time.After(5 * time.Minute):
               // TODO log
         }
      }
   }
}

func registerConfirmationListener(nanoAddress string, ch chan string, operation string) {

   if (operation == "send") {
      registeredSendChannels[nanoAddress] = ch
   } else {
      registeredReceiveChannels[nanoAddress] = ch
   }
}

func unregisterConfirmationListener(nanoAddress string, operation string) {

   if (operation == "send") {
      delete(registeredSendChannels, nanoAddress)
   } else {
      delete(registeredReceiveChannels, nanoAddress)
   }
}
