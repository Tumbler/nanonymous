package main


import (
   "fmt"
   "time"
   "strings"
   "strconv"
   "context"
   "bytes"

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
var removeWebSocketSubscription chan string

var registeredSendChannels = make(map[string]chan string)
var registeredReceiveChannels = make(map[string]chan string)

var numSubsribed int
const ACCOUNTS_TRACKED = 5000

var websocketRetries int

func websocketListener(ch chan int, fullSubscribe bool) {
   if (inTesting) {
      if (ch != nil) {
         ch <- 1
      }
      return
   }

   numSubsribed = 0

   if (verbosity >= 5) {
      fmt.Println("Started listening to websockets on:", websocketAddress)
   }

   defer func() {
      err := recover()
      if (err != nil) {
         Error.Println("websocketListener: ", err)
         if (websocketRetries < 3) {
            websocketRetries++
            go websocketListener(nil, fullSubscribe)
         } else {
            panic(fmt.Errorf("websocketListener panicking: %s", err))
         }
      }
   }()

   ws, err := websocket.Dial(websocketAddress, "", "http://localhost/")
   if (err != nil) {
      if (websocketRetries > 0) {
         time.Sleep(5 * time.Second)
      }
      panic(fmt.Errorf("websocketListener, Dial: %w", err))
   }
   defer ws.Close()

   if (fullSubscribe) {
      err = startSubscription(ws)
      if (err != nil) {
         if (websocketRetries > 0) {
            time.Sleep(5 * time.Second)
         }
         panic(fmt.Errorf("websocketListener: %w", err))
      }
   }
   // Tell caller that we're done initializing
   if (ch != nil) {
      ch <- 1
   }

   addWebSocketSubscription = make(chan string)
   removeWebSocketSubscription = make(chan string)

   var notification ConfirmationBlock
   wsChan := make(chan error)
   go websocketReceive(ws, &notification, wsChan)

   if (websocketRetries > 0) {
      // Successfully reconnected
      websocketRetries = 0

      checkActiveTransactions()
   }

   go pollActiveTransactions()

   // Listen for eternity
   for {
      select {
         case err := <-wsChan:
            if (err != nil) {
               Error.Println("Websocket receive error: ", err.Error())
               if (verbosity >= 5) {
                  fmt.Println(" err: ", err.Error())
               }

               // Try to reconnect. If the reconnect fails it will panic.
               Warning.Println("Lost websockets: %w", err)
               websocketRetries++
               go websocketListener(nil, fullSubscribe)
               return
            } else {
               go handleNotification(notification)
            }

            // Set up next receive
            go websocketReceive(ws, &notification, wsChan)

         case <-time.After(60 * time.Second):
            // If we haven't received anything in 60 seconds, send a KeepAlive packet.
            request :=
            `{
               "action": "ping"
            }`
            if (verbosity >= 10) {
               fmt.Println("ping")
            }
            ws.Write([]byte(request))
         case nanoAddress := <-addWebSocketSubscription:
            go addToSubscription(ws, nanoAddress)
         case nanoAddress := <-removeWebSocketSubscription:
            go removeFromSubscriptions(ws, nanoAddress)
      }
   }

   Warning.Println("Unreachable(2)")
}

func startSubscription(ws *websocket.Conn) error {

   var addressString string

   // Go through newest seeds first and find 5000 to track
   rows, seedConn, err := getSeedRowsFromDatabase()
   if (err != nil) {
      return fmt.Errorf("startSubscription: %w", err)
   }

   defer seedConn.Close(context.Background())

   var seed []byte
   var seedID int
   var maxIndex int
   // For all seeds find their accounts
   seedLoop:
   for rows.Next() {
      err = rows.Scan(&seed, &maxIndex, &seedID)
      if (err != nil || len(seed) == 0) {
         break
      }

      innerRows, walletConn, err := getManagedWalletsRowsFromDatabase(seedID)
      if (err != nil) {
         return fmt.Errorf("startSubscription: %w", err)
      }

      defer walletConn.Close(context.Background())

      for (innerRows.Next()) {
         var key keyMan.Key
         key.Seed = seed
         err = innerRows.Scan(&key.Index)
         if (err != nil) {
            return fmt.Errorf("startSubscription: %w", err)
         }
         err := keyMan.SeedToKeys(&key)
         if (err != nil) {
            return fmt.Errorf("startSubscription: %w", err)
         }

         addressString += `"`+ key.NanoAddress + `", `

         numSubsribed++

         if (numSubsribed >= ACCOUNTS_TRACKED) {
            break seedLoop
         }
      }
   }
   rows.Close()

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
      fmt.Println("\nSubscribed to "+ strconv.Itoa(numSubsribed) +" accounts!")
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
   defer func() {
      err := recover()
      if (err != nil) {
         Error.Println("websocketReceive panic: ", err)
         ch <- fmt.Errorf("websocketReceive panic: %s", err)
      }
   }()

   err := websocket.JSON.Receive(ws, r)
   ch <- err
}

func addToSubscription(ws *websocket.Conn, nanoAddress string) {
   defer func() {
      err := recover()
      if (err != nil) {
         Error.Println("addToSubscription panic: ", err)
      }
   }()

   if (verbosity >= 5) {
      fmt.Println("Adding to subscription: ", nanoAddress)
   }

   // unsub from oldest account
   var delString string
   if (numSubsribed >= ACCOUNTS_TRACKED) {
      rows, conn, err := getSeedRowsFromDatabase()
      if (err != nil) {
         Warning.Println("addToSubscription failed: unsub1: ", err)
         return
      }

      defer conn.Close(context.Background())

      var seed keyMan.Key
      var maxIndex int
      var countAccounts int
      // For all seeds find their accounts
      for rows.Next() {
         err := rows.Scan(&seed.Seed, &maxIndex, nil)
         if (err != nil) {
            Warning.Println("addToSubscription failed: unsub2: ", err)
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

   if (verbosity >= 10) {
      fmt.Println("request: ", request)
   }

   ws.Write([]byte(request))
   numSubsribed++
}

func removeFromSubscriptions(ws *websocket.Conn, nanoAddress string) {
   defer func() {
      err := recover()
      if (err != nil) {
         Error.Println("removeFromSubscriptions panic: ", err)
      }
   }()

   if (verbosity >= 5) {
      fmt.Println("Removing from subscriptions: ", nanoAddress)
   }

   request :=
   `{
      "action": "update",
      "topic": "confirmation",
      "options": {
         "confirmation_type": "active_quorum",
         "accounts_del" : ["`+ nanoAddress +`"]
      }
   }`

   if (verbosity >= 10) {
      fmt.Println("request: ", request)
   }

   ws.Write([]byte(request))
   numSubsribed--
}

var lastHashSeen nt.BlockHash
func handleNotification(cBlock ConfirmationBlock) {
   wg.Add(1)
   defer wg.Done()

   defer func() {
      err := recover()
      if (err != nil) {
         Error.Println("handleNotification panic: ", err)
      }
   }()

   if (bytes.Equal(cBlock.Message.Hash, lastHashSeen) && !inTesting) {
      // Just a KeepAlive response. Ignore.
      if (verbosity >= 10) {
         fmt.Println("pong")
      }
      return
   } else {
      lastHashSeen = cBlock.Message.Hash
   }

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
               err := BlockUntilReceivable(msg.Block.LinkAsAccount, 5 * time.Minute)
               if (err != nil) {
                  if (verbosity >= 6) {
                     fmt.Println("handleNotification: %w", err)
                  }
                  Warning.Println("handleNotification: %w", err)
               }
               Receive(msg.Block.LinkAsAccount)
            }
         } else {
            if (verbosity >= 5) {
               fmt.Println(" External Send")
            }
            if (addressIsReceiveOnly(msg.Block.LinkAsAccount)) {
               // Address flagged, don't start a transaction
               if (verbosity >= 5) {
                  fmt.Println("Receive only address.")
               }
               err := BlockUntilReceivable(msg.Block.LinkAsAccount, 5 * time.Minute)
               if (err != nil) {
                  if (verbosity >= 6) {
                     fmt.Println("handleNotification: %w", err)
                  }
                  Warning.Println("handleNotification: %w", err)
               }
               Receive(msg.Block.LinkAsAccount)
            } else {
               if (verbosity >= 5) {
                  fmt.Println("Starting Transaction!")
               }
               err := receivedNano(msg.Block.LinkAsAccount)
               if (err != nil) {
                  if (verbosity >= 5) {
                     fmt.Println("handleNotification: %w", err)
                  }
                  Warning.Println("handleNotification: %w", err)
               }
            }
         }
      } else {
         // A send to an address we don't own. Probably just the final send.
         // Ignore it.
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

// checkActiveTransactions just goes through the active transactions and sees if
// there are any receivable transactions available. If it finds any, it starts
// a transaction. It is essentially just a polling backup for the websockets.
func checkActiveTransactions() error {

   if (verbosity >= 6) {
      fmt.Println("Polling active transactions...")
   }

   var addresses []string
   for addressID, _ := range activeTransactionList {
      strings := strings.Split(addressID, "-")

      var seedID int
      var index int
      var err error

      if (len(strings) == 2) {
         seedID, err = strconv.Atoi(strings[0])
         if (err != nil) {
            continue
         }

         index, err = strconv.Atoi(strings[1])
         if (err != nil) {
            continue
         }
      } else {
         continue
      }

      key, err := getSeedFromIndex(seedID, index)
      if (err != nil) {
         continue
      }

      addresses = append(addresses, key.NanoAddress)
   }

   var accountsPending map[string][]nt.BlockHash
   var err error
   if (len(addresses) > 0) {
      accountsPending, err = getAccountsPending(addresses)
      if (err != nil) {
         return fmt.Errorf("checkActiveTransactions: %w", err)
      }
   }

   for address, _ := range accountsPending {
      if (verbosity >= 5) {
         fmt.Println("Starting Transaction from a poll!")
      }
      err := receivedNano(address)
      if (err != nil) {
         if (verbosity >= 5) {
            fmt.Println("checkActiveTransactions: ", err)
         }
         Warning.Println("checkActiveTransactions: ", err)
         return fmt.Errorf("checkActiveTransactions: %w", err)
      }
   }

   return nil
}

var pollSemaphore bool
func pollActiveTransactions() {
   defer func() {
      err := recover()
      if (err != nil) {
         Error.Println("pollActiveTransactions panic:", err)
         pollSemaphore = false

         go pollActiveTransactions()
      }
   }()

   if (pollSemaphore) {
      return
   }

   pollSemaphore = true

   for {
      time.Sleep(5 * time.Minute)

      err := checkActiveTransactions()
      if (err != nil) {
         Warning.Println("pollActiveTransactions: Failed during a poll:", err)
      }
   }
}
