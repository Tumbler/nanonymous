package main


import (
   "fmt"
   "time"
   "strings"

   "golang.org/x/net/websocket"

   // Local packages
   keyMan "nanoKeyManager"
   nt "nanoTypes"
)

// Port 80 (unsecured) and 443 (secure)

type ConfirmationBlock struct {
   Time nt.JInt
   Message struct {
      Block keyMan.Block
      Account string
      Amount *nt.Raw
      Hash nt.BlockHash
      ConfirmationType string `json:"confirmation_type"`
   }
   Ack string
   Error string
}

var addWebSocketSubscription chan string

var registeredSendChannels map[string]chan string
var registeredReceiveChannels map[string]chan string

var numSubsribed int
const ACCOUNTS_TRACKED = 5000

func websocketListener(ch chan int) {
   // TODO needs to work over secure connection (wss)
   if (verbosity >= 5) {
      fmt.Println("Started listening to websockets on:", websocketAddress)
   }

   ws, err := websocket.Dial(websocketAddress, "", "http://localhost/")
   if (err != nil) {
      panic(fmt.Errorf("websocketListener, Dial: %w", err))
   }
   defer ws.Close()

   err = startSubscription(ws)
   if (err != nil) {
      panic(fmt.Errorf("websocketListener: %w", err))
   }
   // Tell caller that we're done initializing
   if (ch != nil) {
      ch <- 1
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
               Error.Println("Websocket receive error: ", err.Error())
               if (verbosity >= 5) {
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

   Warning.Println("Unreachable(2)")
}

func startSubscription(ws *websocket.Conn) error {

   var addressString string

   // Go through newest seeds first and find 5000 to track
   rows, err := getSeedRowsFromDatabase()
   var seed []byte
   var maxIndex int
   // For all seeds find their accounts
   for rows.Next() {
      err = rows.Scan(&seed, &maxIndex)
      if (err != nil || len(seed) == 0) {
         break
      }

      startingPoint := maxIndex - (ACCOUNTS_TRACKED - numSubsribed)
      if (startingPoint < 0) {
         startingPoint = 0
      }
      for i := startingPoint; i <= maxIndex; i++ {
         var key keyMan.Key
         key.Seed = seed
         key.Index = i
         err := keyMan.SeedToKeys(&key)
         if (err != nil) {
            return fmt.Errorf("startSubscription: %w", err)
         }

         addressString += `"`+ key.NanoAddress + `", `

         numSubsribed++
      }

      if (numSubsribed >= ACCOUNTS_TRACKED) {
         break
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

   if (verbosity >= 7) {
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
         if (verbosity >= 5) {
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

   if (verbosity >= 5) {
      fmt.Println("Adding to subscription: ", nanoAddress)
   }

   // unsub from oldest account
   var delString string
   if (numSubsribed >= ACCOUNTS_TRACKED) {
      rows, _ := getSeedRowsFromDatabase()
      var seed keyMan.Key
      var maxIndex int
      var countAccounts int
      // For all seeds find their accounts
      for rows.Next() {
         err := rows.Scan(&seed.Seed, &maxIndex)
         if (err != nil) {
            Warning.Println("addToSubscription failed:", err)
            return
         }

         if (countAccounts + maxIndex >= ACCOUNTS_TRACKED) {
            // Found the seed to use, stop searching
            break
         } else {
            countAccounts += maxIndex
         }
      }
      rows.Close()

      if (maxIndex - (ACCOUNTS_TRACKED - countAccounts) < 0) {
         seed.Index = 0
      } else {
         seed.Index = maxIndex - (ACCOUNTS_TRACKED - countAccounts)
      }
      keyMan.SeedToKeys(&seed)

      delString = ",\n         "+`"accounts_del" : ["`+ seed.NanoAddress +`"]`

      numSubsribed--
   }

   request :=
   `{
      "action": "update",
      "topic": "confirmation",
      "options": {
         "confirmation_type": "active_quorum",
         "accounts_add" : ["`+ nanoAddress +`"]` +
         delString +`
      }
   }`

   ws.Write([]byte(request))
   numSubsribed++
}

func handleNotification(cBlock ConfirmationBlock) {
   wg.Add(1)
   defer wg.Done()

   msg := cBlock.Message
   // Send to one of our tracked addresses
   if (msg.Block.Subtype == "send") {
      if (addressExsistsInDB(msg.Block.LinkAsAccount)) {
         // Check if any transaction manager is expecting this and give to them instead
         if (addressExsistsInDB(msg.Account)) {
            if (verbosity >= 5) {
               fmt.Println(" Internal Send")
            }
            // Internal network send, don't trigger a transaction
            if (registeredSendChannels[msg.Block.LinkAsAccount] != nil) {
               select {
                  case registeredSendChannels[msg.Block.LinkAsAccount] <- msg.Hash.String():
                  case <-time.After(5 * time.Minute):
                     Warning.Println("Registered send channel timeout")
               }
            } else {
               // Internal send, but no transaction is requesting it so just
               // receive the funds
               Receive(msg.Block.LinkAsAccount)
            }
         } else {
            if (verbosity >= 5) {
               fmt.Println(" External Send")
               fmt.Println("Starting Transaction!")
            }
            if (addressIsReceiveOnly(msg.Block.LinkAsAccount)) {
               // Address flagged, don't start a transaction
               Receive(msg.Block.LinkAsAccount)
            } else {
               receivedNano(msg.Block.LinkAsAccount)
            }
         }
      } else {
         // Tracking an address that we don't own?
         Warning.Println("Tracked address not in DB")
      }
   } else if (msg.Block.Subtype == "receive") {
      if (verbosity >= 5) {
         fmt.Println(" Receive")
      }
      if (registeredReceiveChannels[msg.Account] != nil) {
         select {
            case registeredReceiveChannels[msg.Account] <- msg.Hash.String():
            case <-time.After(5 * time.Minute):
               Warning.Println("Registered receive channel timeout")
         }
      }
   } else if (cBlock.Ack != "") {
      Info.Println("Websocket Ack:", cBlock.Ack)
      if (verbosity >= 5 ) {
         fmt.Println("Ack:", cBlock.Ack)
      }
   } else if (cBlock.Error != "") {
      Warning.Println("Websocket action came back with an error:", cBlock.Error)
      if (verbosity >= 5 ) {
         fmt.Println("Websocket action came back with an error:", cBlock.Error)
      }
   }
}

func registerConfirmationListener(nanoAddress string, ch chan string, operation string) {
   if (inTesting) {
      return
   }

   if (operation == "send") {
      registeredSendChannels[nanoAddress] = ch
   } else {
      registeredReceiveChannels[nanoAddress] = ch
   }
}

func unregisterConfirmationListener(nanoAddress string, operation string) {
   if (inTesting) {
      return
   }

   if (operation == "send") {
      delete(registeredSendChannels, nanoAddress)
   } else {
      delete(registeredReceiveChannels, nanoAddress)
   }
}
