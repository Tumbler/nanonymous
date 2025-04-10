package main

import (
   "fmt"
   "time"
   "context"
   "net"
   "encoding/hex"
   "bytes"
   "strings"
   "math/big"
   "strconv"
   "math"
   "math/rand"
   "sync"
   "log"
   "os"
   "crypto/tls"

   // Local packages
   keyMan "github.com/Tumbler/nanoKeyManager"
   nt "github.com/Tumbler/nanoTypes"

   // 3rd party packages
   pgx "github.com/jackc/pgx/v4"
   pgxErr "github.com/jackc/pgerrcode"
   "github.com/jackc/pgconn"
   "golang.org/x/crypto/blake2b"
)

//go:generate go run github.com/c-sto/encembed -i embed.txt -decvarname embeddedByte
var embeddedData = string(embeddedByte)
// "db = [url]" in embed.txt to set this value
var databaseUrl string
// "pass = [pass]" in embed.txt to set this value
var databasePassword string
// "nodeIP = [ip]" in embed.txt to set this value
var nodeIP string
// "websocket = [address]" in embed.txt to set this value
var websocketAddress string
// "work = [work_server_address]" in embed.txt to set this value
var workServer string
// "network = [network]" in embed.txt to set this value (main/beta/test)
var network string
// "fromEmail = [email address to send from]" in embed.txt to set this value
var fromEmail string
// "emailPass = [password of fromEmail]" in embed.txt to set this value
var emailPass string
// "toEmail = [email address to send to]" in embed.txt to set this value
var toEmail string
// "whitelist = [comma sperated list of addresses]" in embed.txt to set this value
var whitelist = make(map[string]bool)

// Should only be set to true in test functions
var inTesting = false
var testingPayment []*nt.Raw
var testingPaymentExternal bool
var testingPendingHashesNum []int
var testingReceiveAlls int
var testingSends = make(map[string][]*nt.Raw)

const MAX_INDEX = 4294967295

const TRANSACTION_DEADLINE = time.Hour

// Fee in %
const FEE_PERCENT = float64(0.2)
var feeDividend int64
var minPayment *nt.Raw
var maxPayment *nt.Raw

var betaMode bool

var wg sync.WaitGroup

type activeTransaction struct {
   recipient []byte
   bridge bool
   percents []int
   delays []int
}

var activeTransactionList = make(map[string]activeTransaction)

var random *rand.Rand

const version = "1.1.5"

// Random info about used ports:
// 41721    Nanonymous request port
// 17076    RCP (test net)
// 17078    Web sockets (test net)
// 7076     RCP (main net)
// 7078     Web sockets (main net)
// 7076     Work server

// interface to allow pgx to pass around comms and txs interchangeably.
type psqlDB interface {
   Begin(ctx context.Context) (pgx.Tx, error)
   Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
   Query(ctx context.Context, sql string, optionsAndArgs ...interface{}) (pgx.Rows, error)
   QueryRow(ctx context.Context, sql string, optionsAndArgs ...interface{}) pgx.Row
}

var Info *log.Logger
var Warning *log.Logger
var Error *log.Logger

var args []string

var verbosity int
func main() {
   var err error
   args = os.Args[1:]

   if (len(args) > 0) {
      verbosity, _ = strconv.Atoi(args[len(args)-1])
      if (strings.ToLower(args[0]) == "-s") {
         // Scan for receivables that haven't been taken care of
         // Only scans last two seeds unless -a option is specified
         var fullSearch = false
         if (len(args) > 1) {
            if (strings.ToLower(args[1]) == "-a") {
               fullSearch = true
            }
         }
         err = initNanoymousCore(false)
         if (err != nil) {
            panic(err)
         }

         err := returnAllReceiveable(fullSearch)
         if (err != nil) {
            panic(err)
         }

      } else if (strings.ToLower(args[0]) == "-c") {
         // Command line interface
         // Default to no websocket subscriptions unless the -w option is also specified.
         var fullInstance = false
         if (len(args) > 1) {
            if (strings.ToLower(args[1]) == "-w") {
               fullInstance = true
            }
         }
         err = initNanoymousCore(fullInstance)
         if (fullInstance) {
            err = resetInUse()
            if (err != nil) {
               Warning.Println("Problem with reset: ", err)
            }
         }
         if (err != nil) {
            panic(err)
         }

         CLI()

      } else if (strings.ToLower(args[0]) == "-w") {
         fmt.Println("-w option must be preceded by the -c option")
         return
      } else if (strings.ToLower(args[0]) == "-r") {
         // Report on the week

         err = initNanoymousCore(false)
         if (err != nil) {
            panic(err)
         }

         err := lastWeekSummary()
         if (err != nil) {
            panic(err)
         }
      } else if (strings.ToLower(args[0]) == "-v" || strings.ToLower(args[0]) == "--version" ) {
         fmt.Println("Version: "+ version)
      } else if (strings.ToLower(args[0]) == "-h" || strings.ToLower(args[0]) == "--help" ) {
         fmt.Println("Nanonymous Core version "+ version +
                   "\n\n  no optoins: This is the default operation. Starts listening on port 41721"+
                     "\n     for TLS connections. It will only respond to new address requests and" +
                     "\n     transactions subscriptions requests." +
                   "\n\n  -c [-w] Start the CLI. The CLI has manual wallet access, an RCP client, and"+
                     "\n     a rudimentary database explorer. Specify the -w option if you want to"+
                     "\n     subscribe to websockets. (Will interfere with main instance if it's"+
                     "\n     running)"+
                   "\n\n  -s Go through all known wallets and check for receivable funds." +
                   "\n\n  -r Give a report on how things went last week. (Prints and emails)" +
                   "\n\n  -v Print version information." +
                   "\n\n  -beta Run in beta mode. (No fees)"+
                   "\n\n  # If the last argument is a number, the verbosity is changed to that number"+
                     "\n     (1-10)")
      } else {
         // Default operation, but with changed verbosity or beta
         if (strings.ToLower(args[0]) == "-beta") {
            betaMode = true
         }
         defaultOperation()
      }
   } else {
      // Default operation
      defaultOperation()
   }

   wg.Wait()
}

func defaultOperation() {
   err := initNanoymousCore(true)
   if (err != nil) {
      Error.Println("defaultOperation: Failed initialization: ", err)
      panic(err)
   }

   err = listen()
   if (err != nil) {
      Error.Println(fmt.Errorf("main: %w", err))
   }
}

// initNanoymousCore sets up our variables that need to preexist before other
// functions can be called.
func initNanoymousCore(mainInstance bool) error {
   // Init loggers
   _, err := os.ReadDir("./logs")
   if (err != nil ){
      if (strings.Contains(err.Error(), "no such file or directory")) {
         os.Mkdir("./logs", 0755)
      } else {
         return fmt.Errorf("initNanoymousCore: %w", err)
      }
   }

   var logFile *os.File
   if (len(args) > 0) {
      logFile, err = os.OpenFile("./logs/" + args[0] + "_logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
      if err != nil {
         return fmt.Errorf("initNanoymousCore: %w", err)
      }
   } else if (inTesting){
      logFile, err = os.OpenFile("./logs/test_logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
      if err != nil {
         return fmt.Errorf("initNanoymousCore: %w", err)
      }
   } else {
      logFile, err = os.OpenFile("./logs/logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
      if err != nil {
         return fmt.Errorf("initNanoymousCore: %w", err)
      }
   }

   Info = log.New(logFile, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)
   Warning = log.New(logFile, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)
   Error = log.New(logFile, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)

   Info.Println("Started Nanonymous Core version", version)


   // Grab embedded data
   for _, line := range strings.Split(embeddedData, "\n") {
      word := strings.Split(line, " = ")

      switch word[0] {
         case "db":
            databaseUrl = strings.Trim(word[1], "\r\n")
         case "pass":
            databasePassword = strings.Trim(word[1], "\r\n")
         case "nodeIP":
            nodeIP = strings.Trim(word[1], "\r\n")
         case "websocket":
            websocketAddress = strings.Trim(word[1], "\r\n")
         case "work":
            workServer = strings.Trim(word[1], "\r\n")
         case "network":
            network = strings.Trim(word[1], "\r\n")
         case "fromEmail":
            fromEmail = strings.Trim(word[1], "\r\n")
         case "emailPass":
            emailPass = strings.Trim(word[1], "\r\n")
         case "toEmail":
            toEmail = strings.Trim(word[1], "\r\n")
         case "whitelist":
            list := strings.Split(strings.Trim(word[1], "\r\n"), ",")
            for _, address := range list {
               whitelist[address] = true
            }
      }
   }

   // Check all data is as expected
   if (databaseUrl == "") {
      return fmt.Errorf("initNanoymousCore: database Url not found! (Use \"db = {yourdb}\" in embed.txt)")
   }
   if (databasePassword == "") {
      return fmt.Errorf("initNanoymousCore: database password not found! (Use \"pass = {yourpassword}\" in embed.txt)")
   }
   if (nodeIP == "") {
      return fmt.Errorf("initNanoymousCore: node IP nout found! (Use \"nodeIP = {IP_Address}\" in embed.txt)")
   }
   if (workServer == "") {
      return fmt.Errorf("initNanoymousCore: work server address not found! (Use \"work = {work_server_address}\" in embed.txt)")
   }

   if (!betaMode) {
      feeDividend = int64(math.Trunc(100/FEE_PERCENT))
      // 500 is the minimum we can get and still charge at least 1 raw for a fee.
      minPayment = nt.NewRaw(500)
   } else {
      minPayment = nt.NewRaw(0)
   }

   maxPayment = findMaxPayment()

   // Seed randomness
   if !(inTesting) {
      random = rand.New(rand.NewSource(time.Now().UnixNano()))
   } else {
      random = rand.New(rand.NewSource(41))
   }

   // Some things to do for only the main instance
   if (mainInstance) {

      activePoW = make(map[string]int)
      workChannel = make(map[string]chan string)

      registeredClientComunicationPipes = make(map[string]chan string)

      ch := make(chan int)
      go websocketListener(ch, true)
      // Wait until websockets are initialized
      <-ch
   } else {
      ch := make(chan int)
      go websocketListener(ch, false)
      // Wait until websockets are initialized
      <-ch
   }

   return nil
}

var safeExit bool
var safeExitChan = make(chan int)
// listen is the default operation of nanonymousCore. It listens on port 41721
// for incoming requests from the front end and passes them off to the handler.
func listen() error {

   defer func() {
      err := recover()
      if (err != nil) {
         Error.Println("listen panic: ", err)
         go listen()
      }
   }()

   const INSTANCE_PORT = 41721

   cer, err := tls.LoadX509KeyPair("tls/server.cert.pem", "tls/server.key.pem")
   if err != nil {
       log.Println(err)
       return fmt.Errorf("listen: %w", err)
   }

   config := &tls.Config{Certificates: []tls.Certificate{cer}}
   listener, err := tls.Listen("tcp", fmt.Sprintf(":%d", INSTANCE_PORT), config)
   if (err != nil) {
      if (strings.Index(err.Error(), "in use") != -1) {
         return fmt.Errorf("listen: Another instance was detected")
      } else {
         return fmt.Errorf("listen: %w", err)
      }
   }
   defer listener.Close()

   // This is an init thing, but we REALLY don't want to do it on another
   // instance accidentally. So it's here behind the instance check.
   err = oneInstanceInits()
   if (err != nil) {
      return fmt.Errorf("listen: %w", err)
   }

   // Listen for eternity to incoming address requests
   if (verbosity < 3) {
      fmt.Println("Listening....")
   }

   for (!safeExit) {
      if (verbosity >= 3) {
         fmt.Println("Listening....")
      }
      go accept(listener)
      var conn net.Conn
      select {
         case ret := <- listenChan:
            if (ret.err != nil) {
               Error.Println("Error with single instance port:", err.Error())
               if (verbosity >= 3) {
                  fmt.Println("Error with single instance port:", err.Error())
               }
            }
            conn = ret.conn
         case <- safeExitChan:
      }

      if (!safeExit) {
         go handleRequest(conn)
      }
   }

   fmt.Println(" Leaving due to safe exit")

   return nil
}

// This whole thing is simply so that listen.Accept() can be interrupted if we need.
type listenRet struct{
   conn net.Conn
   err error
}
var listenChan = make(chan listenRet)
func accept(listener net.Listener) {
   conn, err := listener.Accept()
   listenChan <- listenRet{conn, err}
}

func oneInstanceInits() (err error) {
   resetInUse()

   // Check if there are any delayed transactions that still need to be completed.
   ids, paymentAddresses, err := getDelayedIds()
   if (err != nil) {
      return fmt.Errorf("oneInstanceInits: %w", err)
   }

   // Make sure that any received funds in pending transactions are locked until the transaction completes.
   for _, paymentAddress := range paymentAddresses {
      fmt.Println("setting in use:", paymentAddress)
      setAddressInUse(paymentAddress)
   }

   for _, id := range ids {
      // Load the stored transaction and start it back up.
      var t Transaction
      getTranscationRecord(id, &t)
      fmt.Println("Loaded transaction:", t)

      // make new channels
      for i := 0; i < t.numSubSends; i++ {
         t.commChannel = append(t.commChannel, make(chan transactionComm))
         t.errChannel = append(t.errChannel, make(chan error))
      }
      t.confirmationChannel = make([]chan string, t.numSubSends)
      t.abortchan = make(chan int)

      // Start it up.
      wg.Add(1)
      go transactionManager(&t)

      for i := 0; i < t.numSubSends; i++ {
         if (t.transactionSuccessful[i]) {
            // Already completed
            fmt.Println("Skipping subsend", i)
            continue
         }
         fmt.Println("Starting up subsend", i)

         var subSend = i

         go func() {
            defer func() {
               recoverMessage := recover()
               if (recoverMessage != nil) {
                  Error.Println("oneInstanceInits: Delayed subsend panic: ", recoverMessage)
               }
            }()

            // Delay for amount specified.
            if (len(t.delays) > subSend && t.delays[subSend] > 0) {
               select {
                  case <-time.After(time.Duration(t.delays[subSend]) * time.Second):
                     // Normal delay, proceed as normal.
                  case <-safeExitChan:
                     fmt.Println("Safe Exit from delayed send")
                     // Exiting early
                     return
               }
            }

            if (verbosity >= 5) {
               fmt.Println("Startingup subsend from delayed record", subSend)
            }

            err = findSendingWallets(&t, subSend)
            if (err != nil) {
               err = fmt.Errorf("oneInstanceInits: %w", err)
               select {
                  case t.errChannel[subSend] <- err:
                  case <-time.After(5 * time.Second):
               }
               return
            }

            err = sendNanoToRecipient(&t, subSend)
            if (err != nil) {
               err = fmt.Errorf("oneInstanceInits: %w", err)
               select {
                  case t.errChannel[subSend] <- err:
                  case <-time.After(5 * time.Second):
               }
               return
            }
         }()
      }
   }

   return
}

// handleRequest takes an established connection and responds to it. There are
// only two types of expected requests.
//    (1) Get a new address: Retuns the next valid address in the database
//    (2) Register Transaction: Regesiters a callback for a particular
//        transaction. On completion of the transaction nanonymous will return
//        the final send's hash.
func handleRequest(conn net.Conn) error {

   defer func() {
      err := recover()
      if (err != nil) {
         Error.Println("handleRequest panic: ", err)
      }
   }()

   buff := make([]byte, 1024)
   conn.SetDeadline(time.Now().Add(12 * time.Hour))

   defer conn.Close();

   _, err := conn.Read(buff)
   if (err != nil) {
      if (verbosity >= 3) {
         fmt.Println("handleRequest1:", err.Error())
      }
      return fmt.Errorf("handleRequest: %w", err)
   }
   // Trim any nulls that php handed us.
   var text = string(bytes.Trim(buff, "\x00"))
   var array = strings.Split(text, "&")

   // Parse the whole url
   var percents []int
   var delays []int
   var valueCheck = nt.NewRaw(0)
   var apiResponse bool
   for _, optionString := range array {
      if (verbosity >= 6) {
         fmt.Println(optionString)
      }
      optionArray := strings.Split(optionString, "=")
      switch (optionArray[0]) {
         case "percents":
            if (len(optionArray) > 1) {
               for _, num := range strings.Split(optionArray[1], ",") {
                  integer, err := strconv.Atoi(num)
                  if (err != nil) {
                     continue
                  }
                  percents = append(percents, integer)
               }
            }
            // Max of 5 subsends
            if (len(percents) > 5) {
               percents = percents[:5]
            }
            // Make sure input properly adds up to 100
            var total int
            for _, num := range percents {
               total += num
            }
            for (total > 100 && len(percents) > 0) {
               // Reduce the last value to one before dropping it entirely
               if ((total - percents[len(percents)-1] + 1) < 100) {
                  total = total - percents[len(percents)-1] + 1
                  percents[len(percents)-1] = 1
               } else {
                  total -= percents[len(percents)-1]
                  percents = percents[:len(percents)-1]
               }
            }
            if (len(percents) == 0) {
               percents = append(percents, 100)
            }
            // Increase the final value to take up rest of the 100%.
            if (total < 100) {
               calculatedLastValue := 100 - (total - percents[len(percents)-1])
               percents[len(percents)-1] = calculatedLastValue
            }
         case "delays":
            if (len(optionArray) > 1) {
               for _, num := range strings.Split(optionArray[1], ",") {
                  integer, err := strconv.Atoi(num)
                  if (err != nil) {
                     fmt.Println("Error:", err)
                     continue
                  }
                  // Max delay of 1 hour.
                  if (integer > 3600) {
                     integer = 3600
                  } else if (integer < 0) {
                     integer = 0
                  }
                  delays = append(delays, integer)
               }
            }
         case "amount":
            if (len(optionArray) > 1) {
               valueCheck.SetString(optionArray[1], 10)
               fmt.Println("valueCheck:", valueCheck)
            }
         case "api":
            if (len(optionArray) > 1) {
               if (optionArray[1] == "true") {
                  apiResponse = true
               }
            } else {
                  apiResponse = true
            }
      }
   }
   // Delays must match percents
   if (len(delays) > len(percents)) {
      // If we have no percents, then make at least one.
      if (len(percents) == 0) {
         percents = append(percents, 100)
      }
      delays = delays[:len(percents)]
   }
   for (len(delays) < len(percents)) {
      delays = append(delays, 0)
   }

   if (len(array) >= 2 && array[0] == "newaddress") {
      var subArray = strings.Split(array[1], "=")
      if (len(subArray) >= 2 && subArray[0] == "address") {
         if (addressExsistsInDB(subArray[1]) && !addressIsReceiveOnly(subArray[1])) {
            // They've selected a nanonymous wallet. Double check there isn't an
            // active transaction waiting on it. If there is, "bridge the gap."
            // If not, then it's invalid.
            seedID, index, err := getWalletFromAddress(subArray[1])
            if (err != nil) {
               conn.Write([]byte("Invalid Request!"))
               return fmt.Errorf("handleRequest3: %w", err)
            }

            recipientPub, _, _, _ := getRecipientAddress(seedID, index)

            if (len(recipientPub) == 32) {
               // Theres an active transaction

               recipientAddress, err := keyMan.PubKeyToAddress(recipientPub)
               if (err != nil) {
                  return fmt.Errorf("handleRequest0: %w", err)
               }

               // Bridging to recipient address instead of the address
               // specified.
               err = respondWithNewAddress(recipientAddress, conn, true, percents, delays, valueCheck, apiResponse)
               if (err != nil) {
                  return fmt.Errorf("handleRequest2: %w", err)
               }
            } else {
               conn.Write([]byte("Invalid Request!"))
            }
         } else {
            err := respondWithNewAddress(subArray[1], conn, false, percents, delays, valueCheck, apiResponse)
            if (err != nil) {
               return fmt.Errorf("handleRequest4: %w", err)
            }
         }
      } else {
         conn.Write([]byte("Invalid Request!"))
      }
   } else if (len(array) >= 2 && array[0] == "trRequest") {
      var subArray = strings.Split(array[1], "=")
      if (len(subArray) >= 2 && subArray[0] == "address") {
         ch := make(chan string)
         registerClientComunicationPipe(subArray[1], ch)
         defer unregisterClientComunicationPipe(subArray[1])

         var response string

         var missedPolls int

         deadline := time.Now().Add(TRANSACTION_DEADLINE)

         commloop:
         for (time.Now().Before(deadline)) {
            select {
               case response = <- ch:
                  var subArray = strings.Split(response, "=")
                  if (len(subArray) > 1) {
                     if (subArray[0] == "hash") {
                        break commloop
                     } else {
                        conn.Write([]byte(response +"\n"))
                     }
                  } else {
                     // invalid communication
                     if (verbosity > 1) {
                        fmt.Println("Warning: Invalid communication:", response)
                     }
                     Warning.Println("Warning: Invalid communication:", response)
                  }
               case <-time.After(20 * time.Second):
                  _, err := conn.Write([]byte("keepAlive\n"))
                  if (err != nil) {
                     missedPolls++
                     if (verbosity > 1) {
                        fmt.Println("Warning! keepAlive: (Client may have just closed the tab)", err)
                     }

                     if (missedPolls > 10 || strings.Contains(err.Error(), "broken pipe")) {
                        return fmt.Errorf("Warning! keepAlive: %w (Client may have just closed the tab)", err)
                     }
                  }
            }
         }

         if (time.Now().After(deadline)) {
            // Timeout.
            Info.Println("handleRequest: Hash request timeout")
            return fmt.Errorf("handleRequest: Hash request timeout")
         }

         conn.Write([]byte(response))
      } else {
         conn.Write([]byte("Invalid Request!"))
      }
   } else if (len(array) >= 1 && strings.ToLower(array[0]) == "feecheck") {
      if (betaMode) {
         conn.Write([]byte("fee=0.0"))
      } else if (apiResponse) {
         // API expects a flat number not given in percent
         uri := "fee="+ strconv.FormatFloat(FEE_PERCENT / 100, 'f', 3, 64)

         if (valueCheck.Cmp(nt.NewRaw(0)) > 0) {
            fee := calculateInverseFee(valueCheck)
            total := nt.NewRaw(0).Add(valueCheck, fee)

            uri += "&amountToSend="+ total.String()
         }

         conn.Write([]byte(uri))
      } else {
         conn.Write([]byte("fee="+ strconv.FormatFloat(FEE_PERCENT, 'f', 2, 64)))
      }
   } else if (conn.LocalAddr().String() == "127.0.0.1:41721") {
      // Local commands for controlling the core
      if (strings.Contains(text, "safeExit")) {
         Info.Println("Got safe Exit request.")
         safeExit = true
         sendSafeExit()

         httpHeader :=
         "HTTP/1.1 200 OK\n"+
         "Content-Type: text/plain\n"+
         "Connection: Closed\n"
         conn.Write([]byte(httpHeader +"\nAck"))
      } else if (strings.Contains(text, "retireSeed")) {
         Info.Println("Got retire seed request.")
         err := retireCurrentSeed()
         if (err != nil) {
            Warning.Println("handleRequest: ", err)
            if (verbosity >= 5) {
               fmt.Println("handleRequest: ", err)
            }
         }

         httpHeader :=
         "HTTP/1.1 200 OK\n"+
         "Content-Type: text/plain\n"+
         "Connection: Closed\n"
         conn.Write([]byte(httpHeader +"\nAck"))
      } else if (strings.Contains(text, "refreshMaxPayment")) {
         Info.Println("Got refresh max payment request.")
         maxPayment = findMaxPayment()

         httpHeader :=
         "HTTP/1.1 200 OK\n"+
         "Content-Type: text/plain\n"+
         "Connection: Closed\n"
         conn.Write([]byte(httpHeader +"\nAck"))
      }
   }

   return nil
}

// getNewAddress finds the next available address given the keys stored in the
// database and returns address B. If "receivingAddress" A is not an empty
// string, then it will also place A->B into the blacklist and register it as
// an active transaction.
func getNewAddress(receivingAddress string, receiveOnly bool, mixer bool, bridge bool, percents []int, delays []int, seedId int) (*keyMan.Key, int, error) {
   var seed keyMan.Key

   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return nil, 0, fmt.Errorf("getNewAddress: %w", err)
   }
   defer conn.Close(context.Background())

   tx, err := conn.Begin(context.Background())
   if err != nil {
      return nil, 0, fmt.Errorf("getNewAddress: %w", err)
   }
   defer tx.Rollback(context.Background())

   var queryString string
   if (seedId == 0) {
      queryString =
      "SELECT " +
         "id, " +
         "pgp_sym_decrypt_bytea(seed, $1), " +
         "current_index " +
      "FROM " +
         "seeds " +
      "WHERE " +
         "current_index < $2 AND " +
         "active = true " +
      "ORDER BY " +
         "id;"
   } else {
      queryString =
      "SELECT " +
         "id, " +
         "pgp_sym_decrypt_bytea(seed, $1), " +
         "current_index " +
      "FROM " +
         "seeds " +
      "WHERE " +
         "current_index < $2 AND " +
         "active = true AND " +
         "id = "+ strconv.Itoa(seedId) +" " +
      "ORDER BY " +
         "id;"
   }

   rows, err := tx.Query(context.Background(), queryString, databasePassword, MAX_INDEX)
   if (err != nil) {
      return nil, 0, fmt.Errorf("getNewAddress: Query seed: %w", err)
   }

   // Get a current seed. If it fails, generate a new one.
   var seedID int
   if (rows.Next()) {
      err = rows.Scan(&seedID, &seed.Seed, &seed.Index)
      if (err != nil) {
         return nil, 0, fmt.Errorf("getNewAddress: %w ", err)
      } else {
         rows.Close()

         // Get next index
         seed.Index += 1
         err = keyMan.SeedToKeys(&seed)
         if (err != nil) {
            return nil, 0, fmt.Errorf("getNewAddress: %w", err)
         }
      }
   }

   if (seedID == 0) {
      // No valid seeds in database. Generate a new one.
      err = keyMan.GenerateSeed(&seed)
      if (err != nil) {
         return nil, 0, fmt.Errorf("getNewAddress: %w ", err)
      }

      seedID, err = insertSeed(tx, seed.Seed)
      if (err != nil) {
         return nil, 0, fmt.Errorf("getNewAddress: %w ", err)
      }
   }

   // Add to list of managed wallets
   queryString =
   "INSERT INTO "+
      "wallets(parent_seed, index, balance, hash, receive_only, mixer) " +
   "VALUES " +
      "($1, $2, 0, $3, $4, $5)"

   hash := blake2b.Sum256(seed.PublicKey)
   rowsAffected, err := tx.Exec(context.Background(), queryString, seedID, seed.Index, hash[:], receiveOnly, mixer)
   if (err != nil) {
      return nil, 0, fmt.Errorf("getNewAddress: %w", err)
   }
   if (rowsAffected.RowsAffected() < 1) {
      return nil, 0, fmt.Errorf("getNewAddress: no rows affected in insert")
   }

   queryString =
   "UPDATE " +
      "\"seeds\"" +
   "SET " +
      "\"current_index\" = $1 " +
   "WHERE " +
      "\"id\" = $2;"

   rowsAffected, err = tx.Exec(context.Background(), queryString, seed.Index, seedID)
   if (err != nil) {
      return nil, 0, fmt.Errorf("getNewAddress: %w", err)
   }
   if (rowsAffected.RowsAffected() < 1) {
      return nil, 0, fmt.Errorf("getNewAddress: no rows affected during index incrament")
   }

   // Blacklist new address with the receiving address
   if (receivingAddress != "") {
      receivingAddressByte, err := keyMan.AddressToPubKey(receivingAddress)
      if (err != nil) {
         return nil, 0, fmt.Errorf("getNewAddress: %w", err)
      }

      err = blacklist(tx, seed.PublicKey, receivingAddressByte, seedID)
      if (err != nil) {
         return nil, 0, fmt.Errorf("getNewAddress: Blacklist falied: %w", err)
      }

      // Track so that when we receive funds we know where to send it
      err = setRecipientAddress(seedID, seed.Index, receivingAddressByte, bridge, percents, delays)
      if (err != nil) {
         Warning.Println("getNewAddress: ", err.Error())
         //return nil, 0, fmt.Errorf("getNewAddress: %w", err)
      }

      // Make sure we don't keep this forever
      go timeoutTransaction(seedID, seed.Index)
   }

   if (verbosity >= 10) {
      fmt.Println("Commiting new address")
   }
   err = tx.Commit(context.Background())
   if (err != nil) {
      return nil, 0, fmt.Errorf("getNewAddress: %w", err)
   }

   // Generate work for first use
   go preCalculateNextPoW(seed.NanoAddress, true)

   if !(inTesting || mixer) {
      // Track confirmations on websocket
      select {
         case addWebSocketSubscription <- seed.NanoAddress:
         case <-time.After(3 * time.Second):
            if (verbosity >= 5) {
               fmt.Println("Subscription add failed!")
            }
            Warning.Println("Add subscription timeout")
      }
   }


   return &seed, seedID, nil
}

// timeoutTransaction simply deletes the link between address B and C after the
// deadline since we don't want to keep that information indefinitely even in
// memory. There is no problem with double deleting.
func timeoutTransaction(seedID int, index int) {

   defer func() {
      err := recover()
      if (err != nil) {
         Error.Println("timeoutTransaction panic: ", err)
      }
   }()

   if (inTesting) {
      return
   }

   time.Sleep(TRANSACTION_DEADLINE)

   err := setRecipientAddress(seedID, index, nil, false, []int{}, []int{})
   if (err != nil) {
      Warning.Println("timeoutTransaction: Failed to delete transaction: %w", err)
   }

   key, err := getSeedFromIndex(seedID, index)

   if (err == nil) {
      sendInfoToClient("info=Transaction timed out. Please acquire a new address.", key.PublicKey)
   }
}

// blacklist takes two public addresses, hashes them and stores them in the
// database. The purpose of the blacklist is to securely store a transaction
// pair that should never happen. If, for example, Alice sends nano from address
// A to address B in order to receive it anonymously at another address C, then
// Alice wants to be sure that her address A and address C are never associated.
// However, if she later orders another transaction to address C, nanonymous
// would run the risk of using address B to send to C. Before doing such a
// transaction, nanonymous regenerates the blacklist hash and checks the
// blacklist. If it doesn't exist, then we can be sure that there will be no
// unintentional associations.
func blacklist(conn psqlDB, sendingAddress []byte, receivingAddress []byte, seedID int) error {

   concat := append(sendingAddress, receivingAddress[:]...)

   hash := blake2b.Sum256(concat)

   queryString :=
   "INSERT INTO " +
      "blacklist (hash, seed_id) " +
   "VALUES "+
      "($1, $2);"

   rowsAffected, err := conn.Exec(context.Background(), queryString, hash[:], seedID)
   if (err != nil || rowsAffected.RowsAffected() < 1) {
      // We don't care if it's a duplicate entry
      if !(strings.Contains(err.Error(), pgxErr.UniqueViolation)) {
         if (err != nil) {
            return fmt.Errorf("blacklist: %w", err)
         } else {
            return fmt.Errorf("blacklist: no rows affected")
         }
      }
   }

   return nil
}

// blacklistHash is just a wrapper to blacklist() that takes a receiving hash
// instead of a receiving pub key.
func blacklistHash(sendingAddress []byte, receivingHash nt.BlockHash) error {
   if (inTesting) {
      return nil
   }

   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return fmt.Errorf("blacklistHash: %w", err)
   }
   defer conn.Close(context.Background())

   rInfo, err := getBlockInfo(receivingHash)
   if (err != nil) {
      return fmt.Errorf("blacklistHash: %w", err)
   }

   sInfo, err := getBlockInfo(rInfo.Contents.Link)
   if (err != nil) {
      return fmt.Errorf("blacklistHash: %w", err)
   }

   receivePubKey, err := keyMan.AddressToPubKey(sInfo.Contents.Account)
   if (err != nil) {
      return fmt.Errorf("blacklistHash: %w", err)
   }

   nanoAddress, _ := keyMan.PubKeyToAddress(sendingAddress)
   seedID, _, _ := getWalletFromAddress(nanoAddress)

   err = blacklist(conn, sendingAddress, receivePubKey, seedID)
   if (err != nil) {
      return fmt.Errorf("blacklistHash: %w", err)
   }


   return nil
}

// receivedNano is a large function that does most of the back-end work for
// nanonymous.  Upon receiving nano it does 5 distinct things:
//    (1) Updates the database with the newly recived nano
//    (2) Checks if we were expecting the tranaction
//    (3) Calculates the fee
//    (4) Finds the wallet(s) with enough funds to support the transaction
//        (minus the blacklisted ones)
//    (5) Sends the funds to the recipient
func receivedNano(nanoAddress string) error {
   var err error

   parentSeedId, index, err := getWalletFromAddress(nanoAddress)
   if (err != nil) {
      return fmt.Errorf("receivedNano: %w", err)
   }

   err = BlockUntilReceivable(nanoAddress, 5 * time.Minute)
   if (err != nil) {
      return fmt.Errorf("receivedNano, funds not receiveable: %w", err)
   }

   payment, receiveHash, _, sender, err := Receive(nanoAddress)
   if (err != nil) {
      payment, receiveHash, err = retryOrigReceive(nanoAddress, err)
      if (err != nil) {
         Error.Println("Original Receive problem:", err)
         return fmt.Errorf("receivedNano: %w", err)
      }
   }
   if (payment == nil || payment.Cmp(nt.NewRaw(0)) == 0) {
      return fmt.Errorf("receivedNano: No payment received")
   }

   if (minPayment.Cmp(payment) > 0) {
      // Less than the minimum. Refund it.
      pubKey, _ := keyMan.AddressToPubKey(nanoAddress)
      sendInfoToClient("info=The minimum transaction supported is 500 raw. Your transaction has been refunded.", pubKey)
      nonTransactionRefund(receiveHash, parentSeedId, index, payment)
      // Transaction aborted
      return nil
   } else if (maxPayment.Cmp(payment) < 0) {
      // More than the maximum. Refund it.
      pubKey, _ := keyMan.AddressToPubKey(nanoAddress)
      if (betaMode) {
         sendInfoToClient("info=During the beta, the maximum transaction is "+ strconv.FormatFloat(rawToNANO(maxPayment), 'f', -1, 64) +" Nano. Your transaction has been refunded.", pubKey)
      } else {
         sendInfoToClient("info=The maximum transaction currently supported is "+ strconv.FormatFloat(rawToNANO(maxPayment), 'f', -1, 64) +" Nano. Your transaction has been refunded.", pubKey)
      }
      nonTransactionRefund(receiveHash, parentSeedId, index, payment)
      // Transaction aborted
      return nil
   }


   // If anything goes wrong the transactionManager will make sure to clean up
   // the mess.
   var t Transaction
   t.paymentParentSeedId = parentSeedId
   t.paymentIndex = index
   t.payment = payment
   t.receiveHash = receiveHash

   t.id, err = getNextTransactionId()
   if (err != nil) {
      err = fmt.Errorf("receivedNano: %w", err)
      nonTransactionRefund(receiveHash, parentSeedId, index, payment)
      // Transaction aborted
      return err
   }

   Info.Println("Transaction", t.id, "started")

   t.paymentAddress, err = keyMan.AddressToPubKey(nanoAddress)
   if (err != nil) {
      err = fmt.Errorf("receivedNano: %w", err)
      nonTransactionRefund(receiveHash, parentSeedId, index, payment)
      // Transaction aborted
      return err
   }
   setAddressInUse(nanoAddress)

   // Get recipient address for later use.
   t.recipientAddress, t.bridge, t.percents, t.delays = getRecipientAddress(t.paymentParentSeedId, t.paymentIndex)
   if (t.recipientAddress == nil) {
      err = fmt.Errorf("receivedNano: no active transaction available")
      nonTransactionRefund(receiveHash, parentSeedId, index, payment)
      // Transaction aborted
      return err
   }

   // Nothing in percents means only one send.
   if (len(t.percents) == 0) {
      t.percents = append(t.percents, 100)
   }

   // Now that we know how many subsends there are, init all subsend arrays.
   t.numSubSends = len(t.percents)
   t.confirmationChannel = make([]chan string, t.numSubSends)
   t.transitionalKey = make([]*keyMan.Key, t.numSubSends)
   t.commChannel = make([]chan transactionComm, t.numSubSends)
   t.errChannel = make([]chan error, t.numSubSends)
   for i := 0; i < t.numSubSends; i++ {
      t.commChannel[i] = make(chan transactionComm)
      t.errChannel[i] = make(chan error)
      t.sendingKeys = append(t.sendingKeys, make([]*keyMan.Key, 0))
      t.walletSeed = append(t.walletSeed, make([]int, 0))
      t.walletBalance = append(t.walletBalance, make([]*nt.Raw, 0))
      t.multiSend = append(t.multiSend, false)
      t.individualSendAmount = append(t.individualSendAmount, make([]*nt.Raw, 0))
      t.transitionSeedId = append(t.transitionSeedId, 0)
      t.dirtyAddress = append(t.dirtyAddress, -1)
   }

   wg.Add(1)
   go transactionManager(&t)

   // No fee for those in the whitelist.
   if (whitelist[sender]) {
      t.fee = nt.NewRaw(0)
   } else {
      t.fee = calculateFee(payment)
   }

   amountToSend := nt.NewRaw(0).Sub(payment, t.fee)
   if (verbosity >= 5) {
      fmt.Println("payment:        ", payment,
                  "\nfee:            ", t.fee,
                  "\namount to send: ",   amountToSend)
   }

   // Split the send up into its smaller pieces
   totalAdded := nt.NewRaw(0)
   for i, percent := range t.percents {
      t.amountToSend = append(t.amountToSend, nt.NewRaw(0))
      if (i == len(t.percents) -1) {
         // Final send is just whatever is left over
         t.amountToSend[i] = nt.NewRaw(0).Sub(amountToSend, totalAdded)
      } else {
         onePercent := nt.NewRaw(0).Div(amountToSend, nt.NewRaw(100))
         xPercent := onePercent.Mul(onePercent, nt.NewRaw(int64(percent)))
         t.amountToSend[i] = xPercent

         totalAdded.Add(totalAdded, t.amountToSend[i])
      }
   }
   if (verbosity >= 5 && len(t.amountToSend) > 1) {
      fmt.Println("Split send:", t.amountToSend)
   }

   var communicatedWithClient bool
   if (len(t.delays) > 0) {
      sendInfoToClient("update=Funds received! Final send(s) waiting until delay specified....", t.paymentAddress)
      communicatedWithClient = true
   }

   updateDelayRecords(&t)
   for i := 0; i < t.numSubSends; i++ {
      var subSend = i

      go func() {
         defer func() {
            recoverMessage := recover()
            if (recoverMessage != nil) {
               Error.Println("receivedNano: Subsend panic: ", recoverMessage)
            }
         }()

         // Delay for amount specified.
         if (len(t.delays) > subSend) {
            time.Sleep(time.Duration(t.delays[subSend]) * time.Second)
         }

         if (verbosity >= 5) {
            fmt.Println("Startingup subsend", subSend)
         }

         err = findSendingWallets(&t, subSend)
         if (err != nil) {
            err = fmt.Errorf("receivedNano: %w", err)
            select {
               case t.errChannel[subSend] <- err:
               case <-time.After(5 * time.Second):
            }
            return
         }

         // If it's a complicated send let the client know we've received the funds.
         if (!communicatedWithClient && (t.numSubSends > 1 || t.multiSend[0])) {
            sendInfoToClient("update=Funds received! Generating final send...", t.paymentAddress)
            communicatedWithClient = true
         }

         err = sendNanoToRecipient(&t, subSend)
         if (err != nil) {
            err = fmt.Errorf("receivedNano: %w", err)
            select {
               case t.errChannel[subSend] <- err:
               case <-time.After(5 * time.Second):
            }
            return
         }
      }()
   }

   return nil
}

var lock sync.Mutex
// findSendingWallets is just a sub function of receivedNano(). It's only here
// for readablity and Mutexing the database.
func findSendingWallets(t *Transaction, i int) error {
   var err error
   var foundEntry bool

   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return fmt.Errorf("receivedNano: %w", err)
   }
   defer conn.Close(context.Background())


   // Make sure this function is only accessed one at a time so that we never
   // choose the same wallet for multiple transactions.
   lock.Lock()
   defer lock.Unlock()

   // Find all wallets that have enough funds to send out the payment that
   // aren't the wallet we just received in.
   queryString :=
   "SELECT " +
      "parent_seed, " +
      "index, " +
      "balance " +
   "FROM " +
      "wallets " +
   "WHERE " +
      "balance >= $1 AND NOT" +
      "( parent_seed = $2 AND " +
      "  index = $3 ) AND " +
      "in_use = FALSE AND " +
      "receive_only = FALSE AND " +
      "mixer = FALSE " +
   "ORDER BY " +
      "balance, " +
      "index;"

   var rows pgx.Rows
   rows, err = conn.Query(context.Background(), queryString, t.amountToSend[i], t.paymentParentSeedId, t.paymentIndex)
   if (err != nil) {
      return fmt.Errorf("findSendingWallets: Query: %w", err)
   }

   var foundAddress bool
   var tmpSeed int
   var tmpIndex int
   var tmpBalance = nt.NewRaw(0)
   for rows.Next() {
      err = rows.Scan(&tmpSeed, &tmpIndex, tmpBalance)
      if (err != nil) {
         return fmt.Errorf("findSendingWallets: Scan: %w", err)
      }

      // Check the blacklist before accepting
      var tmpKey *keyMan.Key
      foundEntry, tmpKey, err = checkBlackList(tmpSeed, tmpIndex, t.recipientAddress)
      if (err != nil) {
         return fmt.Errorf("findSendingWallets: %w", err)
      }
      if (!foundEntry) {
         // Use this address
         t.sendingKeys[i] = append(t.sendingKeys[i], tmpKey)
         t.walletSeed[i] = append(t.walletSeed[i], tmpSeed)
         t.walletBalance[i] = append(t.walletBalance[i], tmpBalance)
         setAddressInUse(tmpKey.NanoAddress)
         foundAddress = true
         if (verbosity >= 5) {
            fmt.Println("sending from:", tmpSeed, tmpIndex, "to recipient")
         }
         break
      }
   }
   rows.Close()

   if (!foundAddress) {
      // No single wallet has enough, try to combine several.
      queryString =
      "SELECT " +
         "parent_seed, " +
         "index, " +
         "balance " +
      "FROM " +
         "wallets " +
      "WHERE " +
         "balance > 0 AND NOT (" +
         "parent_seed = $1 AND " +
         "index = $2) AND " +
         "in_use = FALSE AND " +
         "receive_only = FALSE AND " +
         "mixer = FALSE " +
      "ORDER BY " +
         "balance, " +
         "index;"

      rows, err := conn.Query(context.Background(), queryString, t.paymentParentSeedId, t.paymentIndex)
      if (err != nil) {
         return fmt.Errorf("findSendingWallets: Query(2): %w", err)
      }

      var enough bool
      var totalBalance = nt.NewRaw(0)
      for rows.Next() {
         err = rows.Scan(&tmpSeed, &tmpIndex, tmpBalance)
         if (err != nil) {
            return fmt.Errorf("findSendingWallets: Scan(2): %w", err)
         }

         // Check the blacklist before adding to the list
         var tmpKey *keyMan.Key
         foundEntry, tmpKey, err = checkBlackList(tmpSeed, tmpIndex, t.recipientAddress)
         if (err != nil) {
            return fmt.Errorf("findSendingWallets: %w", err)
         }
         if (!foundEntry) {
            t.sendingKeys[i] = append(t.sendingKeys[i], tmpKey)
            t.walletSeed[i] = append(t.walletSeed[i], tmpSeed)
            // tmpBalance contains a pointer, so we need a new address to add to the list
            newAddress := nt.NewFromRaw(tmpBalance)
            t.walletBalance[i] = append(t.walletBalance[i], newAddress)
            totalBalance.Add(totalBalance, tmpBalance)
            setAddressInUse(tmpKey.NanoAddress)
            if (totalBalance.Cmp(t.amountToSend[i]) >= 0) {
               // We've found enough
               enough = true
               break
            }
         }
      }
      rows.Close()
      if (!enough) {
         // Not enough in managed wallets. Check the mixer.
         mixerBalance, err := getReadyMixerFunds()
         if (err != nil) {
            return fmt.Errorf("findSendingWallets: %w", err)
         }

         if (mixerBalance.Add(mixerBalance, totalBalance).Cmp(t.amountToSend[i]) >= 0) {
            // There's enough; add the keys to the transaction manager
            keys, seeds, balances, err := getKeysFromMixer(nt.NewRaw(0).Sub(t.amountToSend[i], totalBalance))
            if (err != nil) {
               return fmt.Errorf("findSendingWallets: %w", err)
            }

            for _, key := range keys {
               setAddressInUse(key.NanoAddress)
            }

            t.sendingKeys[i] = append(t.sendingKeys[i], keys...)
            t.walletSeed[i] = append(t.walletSeed[i], seeds...)
            t.walletBalance[i] = append(t.walletBalance[i], balances...)

         } else {
            // Not enough even if we add the mixer.
            return fmt.Errorf("findSendingWallets: not enough funds")
         }
      }
   }

   if (len(t.sendingKeys[i]) > 1) {
      t.multiSend[i] = true
   }

   return nil
}

// sendNanoToRecipient is just a subfunction of receivedNano(). It's just here for
// readability.
func sendNanoToRecipient(t *Transaction, i int) error {
   var err error

   if (t.abort) {
      return fmt.Errorf("Aborted by another send")
   }

   // Send nano to recipient
   if (len(t.sendingKeys[i]) == 1) {
      t.individualSendAmount[i] = append(t.individualSendAmount[i], t.amountToSend[i])
      go Send(t.sendingKeys[i][0], t.recipientAddress, t.amountToSend[i], t.commChannel[i], t.errChannel[i], 0)
      t.commChannel[i] <- *new(transactionComm)
   } else if (len(t.sendingKeys[i]) > 1) {
      // Need to do a multi-send; Get a new wallet to combine all funds into
      t.transitionalKey[i], t.transitionSeedId[i], err = getNewAddress("", false, false, false, []int{}, []int{}, 0)
      if (err != nil) {
         return fmt.Errorf("sendNanoToRecipient: %w", err)
      }

      // Go through list of wallets and send to interim address
      var totalSent = nt.NewRaw(0)
      var currentSend = nt.NewRaw(0)
      for j, key := range t.sendingKeys[i] {

         // if (total + balance) > payment
         var arithmaticResult = nt.NewRaw(0)
         if (arithmaticResult.Add(totalSent, t.walletBalance[i][j]).Cmp(t.amountToSend[i]) > 0) {
            currentSend = arithmaticResult.Sub(t.amountToSend[i], totalSent)
            t.dirtyAddress[i] = j
         } else {
            currentSend = t.walletBalance[i][j]
         }
         t.individualSendAmount[i] = append(t.individualSendAmount[i], currentSend)
         go Send(key, t.transitionalKey[i].PublicKey, currentSend, t.commChannel[i], t.errChannel[i], j)
         if (j == 0) {
            t.commChannel[i] <- *new(transactionComm)
         }
         totalSent.Add(totalSent, currentSend)
         if (verbosity >= 5) {
            fmt.Println("Sending", currentSend.Int, "from", t.walletSeed[i][j], key.Index, "to", t.transitionSeedId[i], t.transitionalKey[i].Index)
         }
      }

      // Now send to recipient
      if (verbosity >= 5) {
         fmt.Println("Sending", t.amountToSend[i], "from", t.transitionSeedId[i], t.transitionalKey[i].Index, "to recipient.")
      }
      go ReceiveAndSend(t.transitionalKey[i], t.recipientAddress, t.amountToSend[i], t.commChannel[i], t.errChannel[i], &t.receiveWg[i], &t.abort)
   } else {
      return fmt.Errorf("sendNanoToRecipient: not enough funds(2)")
   }

   return nil
}
// Send is intended to be used with receivedNano() (although it doesn't have to be). It's a wrapper to sendNano that gives callbacks to the transaction manager when done.
func Send(fromKey *keyMan.Key, toPublicKey []byte, amount *nt.Raw, commCh chan transactionComm, errCh chan error, i int) (nt.BlockHash, error) {

   defer func() {
      err := recover()
      if (err != nil) {
         Error.Println("Send panic: ", err)
         if (errCh != nil) {
            errCh <- fmt.Errorf("Send panic: %s)", err)
         }
      }
   }()

   var tComm transactionComm

   newHash, err := sendNano(fromKey, toPublicKey, amount)
   if (err != nil) {
      if (errCh != nil) {
         if (verbosity >= 5) {
            fmt.Println("Error with send!!!")
         }
         if (i >= 0) {
            err = fmt.Errorf(">>%d<< %w", i, err)
         }
         errCh <- err
      }
      return nil, fmt.Errorf("Send: %w", err)
   }

   setAddressNotInUse(fromKey.NanoAddress)

   if (commCh != nil) {
      tComm.i = i
      tComm.hashes = []nt.BlockHash{newHash}
      commCh <- tComm
   }

   return newHash, nil
}

// ReceiveAndSend is a function that is intended to be used with receivedNano().
// It receives all funds to an internal wallet, and then, with the direction of
// the transaction manager sends the funds to the recipient.
func ReceiveAndSend(transitionalKey *keyMan.Key, toPublicKey []byte, amount *nt.Raw, commCh chan transactionComm, errCh chan error, transactionWg *sync.WaitGroup, abort *bool) {

   defer func() {
      err := recover()
      if (err != nil) {
         Error.Println("ReceiveAndSend panic: ", err)
         if (errCh != nil) {
            errCh <- fmt.Errorf("ReceiveAndSend panic: %s)", err)
         }
      }
   }()

   if (*abort) {
      return
   }

   wg.Add(1)
   defer wg.Done()

   var tComm transactionComm

   // Wait until all sends have processed and confirmed
   transactionWg.Wait()
   transactionWg.Add(1)
   if (*abort) {
      return
   }

   // All transactions are still pending. Receive the funds.
   receiveHashes, err := ReceiveAll(transitionalKey.NanoAddress)
   if (err != nil) {
      err = fmt.Errorf("ReceiveAndSend: %w", err)
      errCh <- err
   } else {
      tComm.i = 2
      tComm.hashes = receiveHashes
      commCh <- tComm
   }

   // Wait until all receives have processed and confirmed
   transactionWg.Wait()
   transactionWg.Add(1)
   if (*abort) {
      return
   }

   // Finally, send to recipient.
   newHash, err := Send(transitionalKey, toPublicKey, amount, nil, nil, -1)
   if (err != nil) {
      err = fmt.Errorf("ReceiveAndSend: %w", err)
      errCh <- err
   } else {
      tComm.i = 3
      tComm.hashes = []nt.BlockHash{newHash}
      commCh <- tComm
   }
}

// SendEasy is just a wrapper for Send(). If you have the info already Send()
// is more efficient so don't overuse this function.
func SendEasy(from string, to string, amount *nt.Raw, all bool) {

   fromKey, _, _, err := getSeedFromAddress(from)
   toKey, _, _, err := getSeedFromAddress(to)
   if (err != nil) {
      toKey.PublicKey, _ = keyMan.AddressToPubKey(to)
   }

   if (all) {
      err = sendAllNano(&fromKey, toKey.PublicKey)
   } else {
      _, err = Send(&fromKey, toKey.PublicKey, amount, nil, nil, -1)
      if (err != nil && verbosity >= 5) {
         fmt.Println("Error: ", err.Error())
      }
   }
}

// rawToNANO is used to convert raw to NANO AKA Mnano (the communnity just calls
// this a nano). We don't have a conversion to go the other way as all
// operations should be done in raw to avoid rounding errors. We only want to
// convert when outputing for human readable format.
func rawToNANO(raw *nt.Raw) float64 {
   // 1 NANO is 10^30 raw
   if (raw == nil) {
      return float64(0)
   }
   rawConv := new(big.Int).Exp(big.NewInt(10), big.NewInt(30), nil)
   rawConvFloat := new(big.Float).SetInt(rawConv)
   rawFloat := new(big.Float).SetInt(raw.Int)

   NanoFloat := new(big.Float).Quo(rawFloat, rawConvFloat)

   NanoFloat64, _ := NanoFloat.Float64()

   return NanoFloat64
}

// checkBlackList back referances our maintainted wallets, hashes them with with
// a given wallet and finds out if the result is already in the blacklist.
func checkBlackList(parentSeedId int, index int, clientAddress []byte) (bool, *keyMan.Key, error) {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return true, nil, fmt.Errorf("checkBlackList: %w", err)
   }
   defer conn.Close(context.Background())

   // Generate the hash
   queryString :=
   "SELECT " +
      "pgp_sym_decrypt_bytea(seed, $1)" +
   "FROM " +
      "seeds " +
   "WHERE " +
      "id = $2;"

   var seed keyMan.Key
   row, err := conn.Query(context.Background(), queryString, databasePassword, parentSeedId)
   if (err != nil) {
      return true, nil, fmt.Errorf("checkBlackList: seed query: %w", err)
   }
   if (row.Next()) {
      err = row.Scan(&seed.Seed)
      row.Close()
      if (err != nil) {
         return true, nil, fmt.Errorf("checkBlacklist: %w", err)
      }
   } else {
      return true, nil, fmt.Errorf("checkBlacklist: No such seed found: %d", parentSeedId)
   }
   row.Close()

   seed.Index = index
   err = keyMan.SeedToKeys(&seed)
   if (err != nil) {
      return true, nil, fmt.Errorf("checkBlackList: %w", err)
   }

   concat := append(seed.PublicKey, clientAddress[:]...)
   blackListHash := blake2b.Sum256(concat)

   if (verbosity >= 5) {
      fmt.Println("check blacklist for:", hex.EncodeToString(blackListHash[:]))
   }

   // Check hash against the blacklist
   queryString =
   "SELECT " +
      "hash " +
   "FROM " +
      "blacklist " +
   "WHERE " +
      "\"hash\" = $1;"

   blacklistRows, err := conn.Query(context.Background(), queryString, blackListHash[:])
   if (err != nil) {
      return true, nil, fmt.Errorf("checkBlackList: blacklist query: %w", err)
   }

   if (blacklistRows.Next()) {
      // Found entry in blacklist
      return true, nil, nil
   } else {
      return false, &seed, nil
   }
}

// getRecipientAddress is an interface to work with the active transaction list.
// Takes an internal wallet and returns the registered recipient address.
func getRecipientAddress(parentSeedId int, index int) ([]byte, bool, []int, []int) {
   key := strconv.Itoa(parentSeedId) + "-" + strconv.Itoa(index)
   return activeTransactionList[key].recipient, activeTransactionList[key].bridge, activeTransactionList[key].percents, activeTransactionList[key].delays
}

// setRecipientAddress is an interface to work with the active transaction list.
// Adds or removes an entry that maps one of our internal wallets to a recipient
// address.
func setRecipientAddress(parentSeedId int, index int, recipientAddress []byte, bridge bool, percents []int, delays []int) error {
   key := strconv.Itoa(parentSeedId) + "-" + strconv.Itoa(index)
   _, exists := activeTransactionList[key]
   if (exists) {
      if (recipientAddress == nil) {
         delete(activeTransactionList, key)
         return nil
      } else {
         return fmt.Errorf("setRecipientAddress: address already exists in active transaction list")
      }
   }

   if (recipientAddress != nil) {
      activeTransactionList[key] = activeTransaction {
                                     recipientAddress,
                                     bridge,
                                     percents,
                                     delays,
                                  }
   }

   return nil
}

// sendNano is the base-level function to send nano. Takes an internal key
// object to send from, a public key to send to, and of cource the amount to
// send in raw. Returns the block hash of the resulting block.
func sendNano(fromKey *keyMan.Key, toPublicKey []byte, amountToSend *nt.Raw) (nt.BlockHash, error) {
   var block keyMan.Block
   var newHash nt.BlockHash

   if !(inTesting) {
      accountInfo, err := getAccountInfo(fromKey.NanoAddress)
      if (err != nil) {
         return nil, fmt.Errorf("sendNano: %s %w", fromKey.NanoAddress, err)
      }

      // if (Balance < amountToSend)
      if (accountInfo.Balance.Cmp(amountToSend) < 0) {
         return nil, fmt.Errorf("sendNano: not enough funds in account.\n have: %s\n need: %s", accountInfo.Balance, amountToSend)
      }

      // Create send block
      block.Previous = accountInfo.Frontier
      block.Seed = *fromKey
      block.Account = block.Seed.NanoAddress
      block.Representative = accountInfo.Representative
      block.Balance = nt.NewRaw(0).Sub(accountInfo.Balance, amountToSend)
      block.Link = toPublicKey

      sig, err := block.Sign()
      if (err != nil) {
         return nil, fmt.Errorf("sendNano: %w", err)
      }

      if (verbosity >= 6) {
         fmt.Println("account:", block.Account)
         fmt.Println("representative:", block.Representative)
         fmt.Println("balance:", block.Balance)
         fmt.Println("link:", strings.ToUpper(hex.EncodeToString(block.Link)))
         h, _ := block.Hash()
         fmt.Println("hash:", strings.ToUpper(hex.EncodeToString(h)))
         fmt.Println("private:", strings.ToUpper(hex.EncodeToString(block.Seed.PrivateKey)))
         fmt.Println("Sig:", strings.ToUpper(hex.EncodeToString(sig)))
      }

      PoW, err := getPoW(block.Account, false)
      if (err != nil) {
         return nil, fmt.Errorf("sendNano: %w", err)
      }

      // Send RCP request
      newHash, err = publishSend(block, sig, PoW)
      if (err != nil) {
         return nil, fmt.Errorf("sendNano: %w", err)
      }
      if (len(newHash) != 32) {
         return nil, fmt.Errorf("sendNano: no block hash returned from node")
      }

      clearPoW(block.Account)
      // If there's no nano left in the account, it's a dead account; no need to
      // calculate PoW for it.
      if (block.Balance.Cmp(nt.NewRaw(0)) != 0) {
         go preCalculateNextPoW(block.Account, false)
      }

      // Update database records
      err = updateBalance(block.Account, block.Balance)
      if (err != nil) {
         Error.Println("Balance update failed from send:", err.Error())
         return nil, fmt.Errorf("sendNano: updatebalance error %w", databaseError)
      }
   } else {
      // Doing tests; behave as close as possible without calling RCP

      balance, _ := getBalance(fromKey.NanoAddress)
      newBalance := nt.NewRaw(0).Sub(balance, amountToSend)

      // if (Balance < amountToSend)
      if (balance.Cmp(amountToSend) < 0) {
         return nil, fmt.Errorf("sendNano: not enough funds in account.\n have: %s\n need: %s", balance, amountToSend)
      }

      // Setup the funds be received later if needed (If it's an external
      // address theres no harm in adding it).
      sendAddress, err := keyMan.PubKeyToAddress(toPublicKey)
      if (err != nil) {
         Error.Println("Failed to get address", err.Error())
         return nil, fmt.Errorf("sendNano: Failed to get address")
      }
      testingSends[sendAddress] = append(testingSends[sendAddress], amountToSend)

      // Update database records
      err = updateBalance(fromKey.NanoAddress, newBalance)
      if (err != nil) {
         Error.Println("Balance update failed from send:", err.Error())
         return nil, fmt.Errorf("sendNano: updatebalance error %w", databaseError)
      }
   }

   return newHash, nil
}

// Wrapper to sendNano() that sends the total balance of the wallet.
func sendAllNano(fromKey *keyMan.Key, toPublicKey []byte) error {
   defer setAddressNotInUse(fromKey.NanoAddress)

   balance, _, err := getAccountBalance(fromKey.NanoAddress)
   if (err != nil) {
      return fmt.Errorf("sendAllNano: %w", err)
   }

   _, err = sendNano(fromKey, toPublicKey, balance)
   if (err != nil) {
      return fmt.Errorf("sendAllNano: %w", err)
   }

   return nil
}

func ReceiveAll(account string) ([]nt.BlockHash, error) {
   var hashes []nt.BlockHash

   for {
      _, hash, numLeft, _, err := Receive(account)
      if (err != nil) {
         return hashes, fmt.Errorf("ReceiveAll: %w", err)
      }
      if (len(hash) > 0) {
         hashes = append(hashes, hash)
      }
      if (numLeft <= 0) {
         break
      }
   }

   return hashes, nil
}

func BlockUntilReceivable(account string, d time.Duration) error {

   if (inTesting) {
      return nil
   }

   deadline := time.Now().Add(d)

   for {
      hashArray, err := getReceivable(account, 1)
      if (err != nil) {
         return fmt.Errorf("BlockUntilReceivable: %w", err)
      }

      if (len(hashArray) > 0) {
         break
      }

      if (time.Now().After(deadline)) {
         break
      }

      time.Sleep(2 * time.Second)
   }

   return nil
}

// Receive takes the next available receivable block and receives it. Returns
// the amount received, the block hash of the created block, and the number of
// remaining pending/receivable hashes.
func Receive(account string) (*nt.Raw, nt.BlockHash, int, string, error) {
   var amountReceived *nt.Raw
   var newHash nt.BlockHash
   var numOfPendingHashes int
   var sender string
   var err error

   if !(inTesting) {
      pendingHashes, _ := getReceivable(account, -1)
      numOfPendingHashes = len(pendingHashes)

      if (numOfPendingHashes > 0) {
         pendingHash := pendingHashes[0]
         amountReceived, newHash, sender, err = receiveHash(pendingHash, numOfPendingHashes)

         numOfPendingHashes--
      }
   } else {
      // Doing testing; behave as close as possible without calling RCP

      balance, _ := getBalance(account)

      if (testingPaymentExternal) {
         amountReceived = testingPayment[0]
         testingPaymentExternal = false
      } else if (len(testingSends[account]) > 0) {
         amountReceived = testingSends[account][len(testingSends[account])-1]
         testingSends[account] = testingSends[account][:len(testingSends[account])-1]
      } else {
         return amountReceived, newHash, 0, "", fmt.Errorf("receive: No funds receiveable on %s", account)
      }

      newBalance := nt.NewRaw(0).Add(balance, amountReceived)

      // Update database records
      err = updateBalance(account, newBalance)
      if (err != nil) {
         Error.Println("Balance update failed from receive:", err.Error())
         return amountReceived, newHash, 0, "", fmt.Errorf("receive: updatebalance error %w", databaseError)
      }

      testingPendingHashesNum[testingReceiveAlls]--
      numOfPendingHashes = testingPendingHashesNum[testingReceiveAlls]
      if (testingPendingHashesNum[testingReceiveAlls] == 0) {
         testingReceiveAlls++
      }
   }

   return amountReceived, newHash, numOfPendingHashes, sender, err
}

func receiveHash(pendingHash nt.BlockHash, numReceivable int) (*nt.Raw, nt.BlockHash, string, error) {
   var block keyMan.Block
   var pendingInfo BlockInfo
   var newHash nt.BlockHash
   var err error

   pendingInfo, err = getBlockInfo(pendingHash)
   if (err != nil) {
      return nil, nil, "", fmt.Errorf("receive: %w", err)
   }
   if (pendingInfo.Subtype != "send") {
      return nil, nil, "", fmt.Errorf("receive: Not a receivable block!")
   }
   sender := pendingInfo.Contents.Account
   account := pendingInfo.Contents.LinkAsAccount
   accountInfo, err := getAccountInfo(account)

   key, _, _, err := getSeedFromAddress(account)
   if (err != nil) {
      return nil, nil, "", fmt.Errorf("receive: %w", err)
   }

   if (err != nil) {
      // Filter out expected errors
      if !(strings.Contains(err.Error(), "Account not found")) {
         return nil, nil, "", fmt.Errorf("Receive: %w", err)
      }
   }

   // Fill block with relevant information
   if (len(accountInfo.Frontier) == 0) {
      // New account. Structure as an open.
      block.Previous = make([]byte, 32)
      block.Representative = getNewRepresentative()
      block.Balance = pendingInfo.Amount
   } else {
      // Old account. Do a standard receive.
      block.Previous = accountInfo.Frontier
      block.Representative = accountInfo.Representative
      block.Balance = nt.NewRaw(0).Add(accountInfo.Balance, pendingInfo.Amount)
   }
   block.Account = account
   block.Link = pendingHash
   block.Seed = key

   sig, err := block.Sign()
   if (err != nil) {
      return nil, nil, "", fmt.Errorf("receive: %w", err)
   }

   if (verbosity >= 6) {
      fmt.Println("account:", block.Account)
      fmt.Println("representative:", block.Representative)
      fmt.Println("balance:", block.Balance)
      fmt.Println("link:", strings.ToUpper(hex.EncodeToString(block.Link)))
      h, _ := block.Hash()
      fmt.Println("hash:", strings.ToUpper(hex.EncodeToString(h)))
      fmt.Println("private:", strings.ToUpper(hex.EncodeToString(block.Seed.PrivateKey)))
      fmt.Println("Sig:", strings.ToUpper(hex.EncodeToString(sig)))
   }

   PoW, err := getPoW(block.Account, true)
   if (err != nil) {
      return nil, nil, "", fmt.Errorf("receive: %w", err)
   }

   // Send RCP request
   newHash, err = publishReceive(block, sig, PoW)
   if (err != nil){
      return nil, nil, "", fmt.Errorf("receive: %w", err)
   }
   if (len(newHash) != 32){
      return nil, nil, "", fmt.Errorf("receive: no block hash returned from node")
   }

   // We've used any stored PoW, clear it out for next use
   clearPoW(block.Account)
   go preCalculateNextPoW(block.Account, numReceivable > 1)

   // Update database records
   err = updateBalance(block.Account, block.Balance)
   if (err != nil) {
      Error.Println("Balance update failed from receive:", err.Error())
      return pendingInfo.Amount, newHash, sender, fmt.Errorf("receive: updatebalance error %w", databaseError)
   }

   return pendingInfo.Amount, newHash, sender, err
}

// getNewRepresentative returns a random representative from a list of accounts
// that are up to date and found on https://nanolooker.com/node-monitors.
func getNewRepresentative() string {

   hardcodedList := []string {
      "nano_1my1snode8rwccjxkckjirj65zdxo6g5nhh16fh6sn7hwewxooyyesdsmii3", // My1s
      "nano_3msc38fyn67pgio16dj586pdrceahtn75qgnx7fy19wscixrc8dbb3abhbw6", // grOvity
      "nano_3g6ue89jij6bxaz3hodne1c7gzgw77xawpdz4p38siu145u3u17c46or4jeu", // Madora
      "nano_1wenanoqm7xbypou7x3nue1isaeddamjdnc3z99tekjbfezdbq8fmb659o7t", // WeNano
      "nano_1iuz18n4g4wfp9gf7p1s8qkygxw7wx9qfjq6a9aq68uyrdnningdcjontgar", // NanoTicker
      "nano_396sch48s3jmzq1bk31pxxpz64rn7joj38emj4ueypkb9p9mzrym34obze6c", // SupeNode
      "nano_3kqdiqmqiojr1aqqj51aq8bzz5jtwnkmhb38qwf3ppngo8uhhzkdkn7up7rp", // ARaiNode
      "nano_1ec5optppmndqsb3rxu1qa4hpo39957s7mfqycpbd547jga4768o6xz8gfie", // Nano Bank
      "nano_318uu1tsbios3kp4dts5b6zy1y49uyb88jajfjyxwmozht8unaxeb43keork", // Scandi Node
      "nano_3n7ky76t4g57o9skjawm8pprooz1bminkbeegsyt694xn6d31c6s744fjzzz", // humble finland
   }

   randAddr := random.Intn(len(hardcodedList))
   return hardcodedList[randAddr]
}

// preCalculateNextPoW finds the proper hash, calcluates the PoW and saves it to
// the database for future use.
func preCalculateNextPoW(nanoAddress string, isReceiveBlock bool) {

   defer func() {
      err := recover()
      if (err != nil) {
         Error.Println("preCalculateNextPoW panic: ", err)
      }
   }()
   wg.Add(1)
   defer wg.Done()
   if (inTesting) {
      return
   }

   work := calculateNextPoW(nanoAddress, isReceiveBlock)

   if (work != "") {

      conn, _ := pgx.Connect(context.Background(), databaseUrl)

      queryString :=
      "UPDATE " +
         "wallets " +
      "SET " +
         "\"pow\" = $1 " +
      "WHERE " +
         "\"hash\" = $2;"

      pubKey,  _ := keyMan.AddressToPubKey(nanoAddress)
      nanoAddressHash := blake2b.Sum256(pubKey)
      conn.Exec(context.Background(), queryString, work, nanoAddressHash[:])
   }
}

var activePoW = make(map[string]int, 0)
var workChannel = make(map[string]chan string, 0)

// calculateNextPoW finds the hash and calculates the PoW using any number of
// different work servers setup at runtime. It returns the work generated as a
// string. It also will make sure not to request work that is already being
// worked on, and so can safely be called multiple times on the same account.
func calculateNextPoW(nanoAddress string, isReceiveBlock bool) string {
   if (inTesting) {
      return ""
   }

   // Check if PoW is already being calculated
   if (activePoW[nanoAddress] == 0) {
      // PoW is not being calculated, so do it!
      activePoW[nanoAddress] = 1

      var hash nt.BlockHash
      accountInfo, _ := getAccountInfo(nanoAddress)

      if (len(accountInfo.Frontier) == 32) {
         hash = accountInfo.Frontier
      } else {
         // New account, just use pubKey
         var err error
         hash, err = keyMan.AddressToPubKey(nanoAddress)
         if (err != nil) {
            return ""
         }

      }

      var difficulty string
      if (isReceiveBlock) {
         difficulty = "fffffe0000000000"
      } else {
         difficulty = "fffffff800000000"
      }

      work, err := generateWorkOnWorkServer(hash, difficulty)
      //work, err := generateWorkOnNode(hash, difficulty)
      if (err != nil) {
         if (verbosity >= 2) {
            fmt.Println("Failed to connect to work server", err.Error())
         }
         Warning.Println("Failed to connect to work server")

         // Fall back server
         work, err = generateWorkOnNode(hash, difficulty)
         if (err != nil) {
            if (verbosity >= 1) {
               fmt.Println("Failed to generate work")
            }
            Error.Println("Failed to generate work")
            return ""
         }
      }
      if (verbosity >= 7) {
         fmt.Println("Work generated for address", nanoAddress, ":", work)
      }

      // See if anyone has requested PoW since we started
      if (activePoW[nanoAddress] == 2) {
         select {
            case workChannel[nanoAddress] <- work:
               // Sent to whoever is going to use it, so no need to write to DB
               work = ""
               // If anyone else happens to be wanting this work, let em know they're too late.
               go func() {
                  // Don't block the rest of the operation while we take care of any stragglers.
                  toLateLoop:
                  for {
                     select {
                        case workChannel[nanoAddress] <- "":
                        case <-time.After(250 * time.Millisecond):
                           // No one's here move along.
                           break toLateLoop
                        }
                  }
               }()
            case <-time.After(5 * time.Minute):
         }
      }

      delete(activePoW, nanoAddress)
      delete(workChannel, nanoAddress)

      return work

   } else {
      // PoW is already being computed by someone else. Wait for them to report it
      if (activePoW[nanoAddress] != 2) {
         activePoW[nanoAddress] = 2
         workChannel[nanoAddress] = make(chan string)
      }

      select {
         case work := <-workChannel[nanoAddress]:
            if (verbosity >= 5) {
               fmt.Println("Got", work, "from channel")
            }
            return work
         case <-time.After(5 * time.Minute):
            Warning.Println("Work never generated")
            return ""
      }
   }
}

// getPoW tries to acquire PoW from database first. If that fails, it then
// calculates it from scratch and returns the work generated.
func getPoW(nanoAddress string, isReceiveBlock bool) (string, error) {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return "", fmt.Errorf("getPoW: %w", err)
   }
   defer conn.Close(context.Background())

   // Check to see if we have pre computed PoW stored on the database
   queryString :=
   "SELECT " +
      "pow " +
   "FROM " +
      "wallets " +
   "WHERE " +
      "\"hash\" = $1;"

   pubKey,  _ := keyMan.AddressToPubKey(nanoAddress)
   nanoAddressHash := blake2b.Sum256(pubKey)
   var PoW string
   err = conn.QueryRow(context.Background(), queryString, nanoAddressHash[:]).Scan(&PoW)
   if (err != nil) {
      if !(strings.Contains(err.Error(), "cannot scan null into *string")) {
         return "", fmt.Errorf("getPoW: %w", err)
      }
   }

   if (PoW == "") {
      // none stored; need to generate it
      PoW = calculateNextPoW(nanoAddress, isReceiveBlock)

      if (PoW == "") {
         return "", fmt.Errorf("getPoW: PoW could not be generated")
      }
   }

   return PoW, nil
}

// checkBalance finds the actual balance of an account from the node and updates
// the database if it doesn't match. Will also receive funds if it finds
// receivable funds just sitting on the account.
func checkBalance(nanoAddress string) (error, bool) {

   var updated bool
   balance, receiveable, _ := getAccountBalance(nanoAddress)

   if (receiveable.Cmp(nt.NewRaw(0)) != 0) {
      // Receive and update
      Receive(nanoAddress)
   } else {

      balanceInDB, err := getBalance(nanoAddress)
      if (err != nil) {
         return fmt.Errorf("checkBalance: %w", err), false
      }

      fmt.Println("BalanceInDB:", balanceInDB, "\nbalance     :", balance)
      if (balance.Cmp(balanceInDB) != 0) {
         updated = true
         err := updateBalance(nanoAddress, balance)
         if (err != nil) {
            return fmt.Errorf("checkBalance: %w", err), false
         }
      }
   }

   return nil, updated
}

// calculateFee applies the stored fee percent, but takes out any resulting dust
// because ain't nobody got time for that.
func calculateFee(payment *nt.Raw) *nt.Raw {

   if (betaMode) {
      // No fee in beta mode
      return nt.NewRaw(0)
   }

   // Find base fee simply by taking the percentage
   fee := nt.NewRaw(0).Div(payment, nt.NewRaw(feeDividend))

   // Find the most significant digit
   var mostSig = nt.NewRaw(1)
   var paymentCopy = nt.NewFromRaw(payment)
   for (paymentCopy.Cmp(nt.NewRaw(9)) > 0) {
      paymentCopy.Div(paymentCopy, nt.NewRaw(10))

      mostSig.Mul(mostSig, nt.NewRaw(10))
   }
   // Maxes out at 1 NANO
   if (mostSig.Cmp(nt.OneNano()) > 0) {
      mostSig = nt.OneNano()
   }

   // Don't want the user to have to deal with dust so I'll round the fee down
   // to the nearest .001 * minimum
   minDust := nt.OneNano().Div(mostSig, nt.NewRaw(1000))

   _, dust := nt.OneNano().DivMod(fee, minDust)

   // Remove any dust from the fee
   fee.Sub(fee, dust)

   return fee
}

// calculateInverseFee figures out what the fee would be when given the final
// amount that will be sent to the recipient.
func calculateInverseFee(payment *nt.Raw) *nt.Raw {

   // Find original value
   // Some setup required since we can't use fractions.
   adjustedPayment := nt.NewRaw(0).Mul(payment, nt.NewRaw(1000))
   originalWithDust := nt.NewRaw(0).Div(adjustedPayment, nt.NewRaw(int64((1-FEE_PERCENT/100)*1000)))

   fee := nt.NewRaw(0).Sub(originalWithDust, payment)

   // Find the most significant digit
   var mostSig = nt.NewRaw(1)
   var origCopy = nt.NewFromRaw(originalWithDust)
   for (origCopy.Cmp(nt.NewRaw(9)) > 0) {
      origCopy.Div(origCopy, nt.NewRaw(10))

      mostSig.Mul(mostSig, nt.NewRaw(10))
   }
   // Maxes out at 1 NANO
   if (mostSig.Cmp(nt.OneNano()) > 0) {
      mostSig = nt.OneNano()
   }

   // Don't want the user to have to deal with dust so I'll round the fee down
   // to the nearest .001 * minimum
   minDust := nt.OneNano().Div(mostSig, nt.NewRaw(1000))

   _, dust := nt.OneNano().DivMod(fee, minDust)

   // Remove any dust from the fee
   fee.Sub(fee, dust)

   return fee
}

// returnAllReceivable checks all wallets in all active seeds and finds any
// receiveable funds that are just lying around (Maybe someone accidentally
// re-sent funds to an address that I'm no longer actively monitoring). If the
// wallet is not marked as "in use" then it returns the funds to the original
// owner. This function is designed to be called occasionally by a seperate
// process to clean up any accidental sends from users.
func returnAllReceiveable(allMeansAll bool) error {

   rows, conn, err := getSeedRowsFromDatabase()
   if (err != nil) {
      Warning.Println("getSeedRowsFromDatabase failed on routine pending check:", err.Error())
      return fmt.Errorf("returnAllReceiveable: %w", err)
   }

   var accounts []string

   defer conn.Close(context.Background())

   var seed keyMan.Key
   var seedID int
   var searched int
   // For all our active seeds (Last two are defined as the active seeds)
   for rows.Next() {
      if (searched >= 2 && !allMeansAll) {
         break
      }

      rows.Scan(&seed.Seed, &seed.Index, &seedID)

      innerRows, conn2, err := getWalletRowsFromDatabaseFromSeed(seedID)
      if (err != nil) {
         Warning.Println("getWalletRowsFromDatabaseFromSeed failed on routine pending check:", err.Error())
         return fmt.Errorf("returnAllReceiveable: %w", err)
      }
      defer conn2.Close(context.Background())

      // Get all accounts
      for (innerRows.Next()) {
         var i int

         err := innerRows.Scan(&i)
         if (err != nil) {
            Warning.Println("innerRows.Scan failed on routine pending check:", err.Error())
            return fmt.Errorf("returnAllReceiveable: %w", err)
         }

         seed.Index = i
         keyMan.SeedToKeys(&seed)

         accounts = append(accounts, seed.NanoAddress)

         if (verbosity >= 5) {
            fmt.Println("  ", seed.NanoAddress)
            fmt.Println("  index", i)
         }
      }

      searched++
   }

   // For all our accounts get all receivable hashes
   hashes, err := getAccountsPending(accounts)
   if (err != nil) {
      Warning.Println("getReceivable failed on routine pending check:", err.Error())
      return fmt.Errorf("returnAllReceiveable: %w", err)
   }
   numberOfHashes := len(hashes)

   if (numberOfHashes > 0) {
      Info.Println("Recievable hashes found during routine check: ", numberOfHashes)
   }
   if (verbosity >= 5) {
      fmt.Println(numberOfHashes, "accounts with receivable hash(es) found!")
   }

   for address, receivableHashes := range hashes {
      for _, receivableHash := range receivableHashes {
         if (verbosity >= 5) {
            fmt.Println("account: ", address)
            fmt.Println("Receivable hash: ", receivableHash)
         }
         // Found funds. Receive them first and then refund them.
         _, receiveHash, _, err := receiveHash(receivableHash, 1)
         if (err != nil) {
            Error.Println("Receive Failed: ", err.Error())
            if (verbosity >= 5) {
               fmt.Println("Receive Failed: ", err.Error())
            }
            continue
         }

         // Don't refund if wallet is receive only or it's an internal send
         if (!addressIsReceiveOnly(address) && !addressExsistsInDB(address)) {
            // Poll until block is confirmed
            waitForConfirmations([]nt.BlockHash{receiveHash})

            if (verbosity >= 5) {
               fmt.Println("      Refunding")
            }
            err := Refund(receiveHash, nt.NewRaw(0))
            if (err != nil) {
               Warning.Println("Refund failed on", seed.NanoAddress, "during routine pending check:", err.Error())
            }
         }
      }
   }

   return nil
}

// waitForConfirmations will block until the specified hashes are confirmed or
// until 5 minutes, whichever comes first. Returns an error if the timeout
// triggered. It subscribes to the websocket listeners and will confirm them
// there, but if they don't it also polls every 5 seconds. Therefore, this
// function can be used without websocket subscriptions without any fear.
func waitForConfirmations(hashList []nt.BlockHash) error{
   if (inTesting) {
      return nil
   }

   var listener = make(chan string)

   // Go through all hashes and register them if someone else hasn't already.
   for i := len(hashList)-1; i >= 0; i-- {
      hash := hashList[i]
      blockInfo, err := getBlockInfo(hash)
      if (err != nil) {
         if (verbosity >= 5) {
            fmt.Println("waitForConfirmations: Can't get block:", err)
         }
         continue
      }
      if (blockInfo.Confirmed) {
         // That was fast. No need to watch.
         hashList = removeHash(hashList, i)
      } else {
         if (blockInfo.Subtype == "send") {
            if (registeredSendChannels[blockInfo.Contents.LinkAsAccount] == nil) {
               registerConfirmationListener(blockInfo.Contents.LinkAsAccount, listener, "send")
               defer unregisterConfirmationListener(blockInfo.Contents.LinkAsAccount, "send")
            }
         } else {
            if (registeredReceiveChannels[blockInfo.Contents.Account] == nil) {
               registerConfirmationListener(blockInfo.Contents.Account, listener, "receive")
               defer unregisterConfirmationListener(blockInfo.Contents.Account, "receive")
            }
         }
      }
   }

   deadline := time.Now().Add(5 * time.Minute)

   // Remove them as they come in on the websocket or poll for them in case we
   // weren't subscribed.
   for (len(hashList) > 0) {
      select {
         case hash := <- listener:
            for i, h := range hashList {
               if (hash == h.String()) {
                  if (verbosity >= 5) {
                     fmt.Println("Found hash... removing")
                  }
                  hashList = removeHash(hashList, i)
                  break
               }
            }
         // Also poll just to make sure we haven't missed anything.
         case <-time.After(5 * time.Second):
            for i := len(hashList)-1; i >= 0; i-- {
               blockInfo, err := getBlockInfo(hashList[i])
               if (err != nil) {
                  if (verbosity >= 5) {
                     fmt.Println(fmt.Errorf("waitForConfirmations warning: %w", err))
                  }
               }
               if (blockInfo.Confirmed) {
                  hashList = removeHash(hashList, i)
                  if (verbosity >= 6) {
                     fmt.Println(" Hash confirmed!")
                  }
               } else if (verbosity >= 6) {
                  fmt.Println("Waiting on hash...")
                  if (verbosity >= 7) {
                     fmt.Println(blockInfo)
                  }
               }
            }
         case <-time.After(deadline.Sub(time.Now())):
            return fmt.Errorf("waitForConfirmations: Wait timed out: %s", deadline)
      }
   }

   return nil
}

func removeHash(hashList []nt.BlockHash, i int) []nt.BlockHash {
   if (i >= len(hashList)) {
      return hashList
   }
   hashList[i] = hashList[len(hashList)-1]
   return hashList[:len(hashList)-1]
}

func respondWithNewAddress(recipientAddress string, conn net.Conn, bridge bool, percents []int, delays []int, valueCheck *nt.Raw, api bool) error {
   newKey, _, err := getNewAddress(recipientAddress, false, false, bridge, percents, delays, 0)
   if (err != nil) {
      if (verbosity >= 3) {
         fmt.Println("respondWithNewAddress: ", err.Error())
      }
      if (api) {
         conn.Write([]byte("error=bad address"))
      } else {
         conn.Write([]byte("There was an error, please try again later"))
      }
      conn.Close()
      Warning.Println("respondWithNewAddress: ", err)
      return fmt.Errorf("respondWithNewAddress: %w", err)
   }

   if (api) {
      var uri = "address="+ newKey.NanoAddress
      for i, percent := range percents {
         if (i == 0) {
            uri += "&percents="
         } else {
            uri += ","
         }

         uri += strconv.Itoa(percent)
      }
      for i, delay := range delays {
         if (i == 0) {
            uri += "&delays="
         } else {
            uri += ","
         }

         uri += strconv.Itoa(delay)
      }
      uri += "&fee=0.00"+ strconv.Itoa((int)(FEE_PERCENT * 10))

      if (valueCheck.Cmp(nt.NewRaw(0)) > 0) {
         fee := calculateInverseFee(valueCheck)
         total := nt.NewRaw(0).Add(valueCheck, fee)

         uri += "&amountToSend="+ total.String()
      }

      conn.Write([]byte(uri))
   } else {
      conn.Write([]byte("address="+ newKey.NanoAddress +"&bridge="+ fmt.Sprint(bridge)))
   }

   return nil
}

func findMaxPayment() *nt.Raw {
   var max *nt.Raw
   var oneHundred = nt.OneNano().Mul(nt.OneNano(), nt.NewRaw(100))
   totalFunds, _, _, err := findTotalBalance()
   if (err != nil) {
      Warning.Println("findMaxPayment: ", err)
      return oneHundred
   }

   if (!betaMode) {
      // Stop payments from more than 10% of total balance to avoid too few wallets.
      max = nt.NewRaw(0)
      max.Div(totalFunds, nt.NewRaw(10))
      // Round down to nearest 100
      _, leftover := nt.OneNano().DivMod(max, oneHundred)
      max.Sub(max, leftover)

      oneThousand := nt.OneNano().Mul(oneHundred, nt.NewRaw(10))
      if (max.Cmp(oneHundred) < 0) {
         Warning.Println("Max payment was less than 100.")
         // Has to be at least 100
         max.Add(nt.NewRaw(0), oneHundred)
      } else if (max.Cmp(oneThousand) > 0) {
         // Has to be at most 1000
         max.Add(nt.NewRaw(0), oneThousand)
      }
   } else {
      // Max payment during beta is 1 Nano.
      max = nt.OneNano()
   }

   return max
}

func nonTransactionRefund(receiveHash nt.BlockHash, parentID int, index int, payment *nt.Raw) {
   err := Refund(receiveHash, nt.NewRaw(0))
   if (err != nil) {
      sendEmail("IMMEDIATE ATTENTION REQUIRED", "Non-transaction refund failed! "+ err.Error() +
            "\n\nPayment Hash: "+ receiveHash.String() +
            "\nID: "+ strconv.Itoa(parentID) +","+ strconv.Itoa(index) +
            "\nAmount: "+ strconv.FormatFloat(rawToNANO(payment), 'f', -1, 64))
      Error.Println("non-transaction Refund failed!!", err)
      if (verbosity >= 1) {
         fmt.Println("non-transaction Refund failed!!", err)
      }
   }
}

func sendSafeExit() {

   for {
      select {
         case safeExitChan <- 1:
         case <-time.After(1 * time.Second):
            break
      }
   }
}

func testTransaction() {
   deleteTransactionRecord(41)

   feeDividend = int64(math.Trunc(100/FEE_PERCENT))
   minPayment = nt.OneNano()
   var t Transaction
   t.paymentParentSeedId = 1
   t.paymentIndex = 72
   t.payment = nt.OneNano()
   t.fee = calculateFee(t.payment)
   t.receiveHash, _ = hex.DecodeString("DCC89776B79E96C778EAECB4C2A6C4EE8EE22DE9388651FD5C57DDE32E181D96")
   t.id = 41
   t.paymentAddress, _ = keyMan.AddressToPubKey("nano_3mhrc9czyfzzok7xeoeaknq6w5ok9horo7d4a99m8tbtbyogg8apz491pkzt")
   t.recipientAddress, _ = keyMan.AddressToPubKey("nano_3rksbipm1b1g64gw6t36ufc77q7mtw1uybnto4xyn1e7ae5aikyknb9fg4su")
   hash, _ := hex.DecodeString("AC684A9447E47340DFC2C0F92208F1EBC58798FDAB9D1A40A46ACBAD73BE3314")
   t.finalHash = append([]nt.BlockHash{}, hash)
   t.bridge = false
   t.percents = []int{40, 60}
   t.delays = []int{15, 349}
   t.numSubSends = len(t.percents)
   t.confirmationChannel = make([]chan string, t.numSubSends)
   t.transitionalKey = make([]*keyMan.Key, t.numSubSends)
   t.commChannel = make([]chan transactionComm, t.numSubSends)
   t.errChannel = make([]chan error, t.numSubSends)
   t.transactionSuccessful = []bool{false, false}
   for i := 0; i < t.numSubSends; i++ {
      t.commChannel[i] = make(chan transactionComm)
      t.errChannel[i] = make(chan error)
      t.sendingKeys = append(t.sendingKeys, make([]*keyMan.Key, 0))
      tmp, _ := getSeedFromIndex(1, i)
      tmp2, _ := getSeedFromIndex(1, i+1)
      t.sendingKeys[i] = append(t.sendingKeys[i], tmp)
      if (i == 0) {
         t.sendingKeys[i] = append(t.sendingKeys[i], tmp2)
      }
      t.walletSeed = append(t.walletSeed, make([]int, 0))
      //t.walletSeed[i] = []int{}
      t.walletBalance = append(t.walletBalance, make([]*nt.Raw, 0))
      //t.walletBalance[i] = []*nt.Raw{}
      t.multiSend = append(t.multiSend, false)
      t.individualSendAmount = append(t.individualSendAmount, make([]*nt.Raw, 0))
      //t.individualSendAmount[i] = []*nt.Raw{}
      t.transitionSeedId = append(t.transitionSeedId, 0)
      t.dirtyAddress = append(t.dirtyAddress, -1)

      var err error
      t.transitionalKey[i], err = getSeedFromIndex(1, i)
      if (err != nil) {
         fmt.Println("error? ", err)
      }
   }
   // Split the send up into its smaller pieces
   totalAdded := nt.NewRaw(0)
   var amountToSend = nt.OneNano()
   for i, percent := range t.percents {
      t.amountToSend = append(t.amountToSend, nt.NewRaw(0))
      if (i == len(t.percents) -1) {
         // Final send is just whatever is left over
         t.amountToSend[i] = nt.NewRaw(0).Sub(amountToSend, totalAdded)
      } else {
         onePercent := nt.NewRaw(0).Div(amountToSend, nt.NewRaw(100))
         xPercent := onePercent.Mul(onePercent, nt.NewRaw(int64(percent)))
         t.amountToSend[i] = xPercent

         totalAdded.Add(totalAdded, t.amountToSend[i])
      }
   }

   deleteTransactionRecord(41)
}
