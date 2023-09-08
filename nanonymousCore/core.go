package main

import (
   "fmt"
   "time"
   "context"
   "net"
   "encoding/hex"
   "strings"
   "math/big"
   "strconv"
   "math"
   "math/rand"
   "sync"
   "log"
   "os"
   "crypto/tls"
   _"embed"

   // Local packages
   keyMan "nanoKeyManager"
   nt "nanoTypes"

   // 3rd party packages
   pgx "github.com/jackc/pgx/v4"
   pgxErr "github.com/jackc/pgerrcode"
   "github.com/jackc/pgconn"
   "golang.org/x/crypto/blake2b"
)

// TODO IP lock transactions 1 per 30 seconds??
// TODO blacklist pruning
// TODO Find out why website sometimes gets 000000000000000000000000000 for final hash.
// TODO add panic recovery
// TODO test backup internet
// TODO bad work completely halts the -S option
// TODO maybe for later but if there's too much funds tied up in current trascations, then wait for them to be available before starting a transaction
// TODO make rawtoNANAO exact by using shift and EXP like we do in core_test.go
// TODO blacklistHash() fails silently, so it can lead to a dirty address that's not mixed.

//go:embed embed.txt
var embeddedData string
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
// "fromEmail = [email address to send from]" in embed.txt to set this value
var fromEmail string
// "emailPass = [password of fromEmail]" in embed.txt to set this value
var emailPass string
// "toEmail = [email address to send to]" in embed.txt to set this value
var toEmail string

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

var betaMode bool

var wg sync.WaitGroup

var activeTransactionList = make(map[string][]byte)

var random *rand.Rand

const version = "1.0.0"

// Random info about used ports:
// 41721    Nanonymous request port
// 17076    RCP (test net)
// 17078    Web sockets (test net)
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
         err = initNanoymousCore(false)
         if (err != nil) {
            panic(err)
         }

         returnAllReceiveable()

      } else if (strings.ToLower(args[0]) == "-c") {
         // Command line interface
         // Default to no websocket subscriptions unless the -w option is also specified.
         var fullInstance = false
         if (len(args) > 1) {
            if (strings.ToLower(args[1]) == "-w") {
               fullInstance = true
               resetInUse()
            }
         }
         err = initNanoymousCore(fullInstance)
         if (err != nil) {
            panic(err)
         }

         CLI()

      } else if (strings.ToLower(args[0]) == "-w") {
         fmt.Println("-w option must be preceded by the -c option")
         return
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
      panic(err)
   }

   err = listen()
   if (err != nil) {
      fmt.Println(fmt.Errorf("main: %w", err))
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

   Info.Println("Started Nanonymous Core")


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
         case "fromEmail":
            fromEmail = strings.Trim(word[1], "\r\n")
         case "emailPass":
            emailPass = strings.Trim(word[1], "\r\n")
         case "toEmail":
            toEmail = strings.Trim(word[1], "\r\n")
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
      minPayment = nt.OneNano()
   } else {
      minPayment = nt.NewRaw(0)
   }

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
      go websocketListener(ch)
      // Wait until websockets are initialized
      <-ch
   }

   return nil
}

var safeExit bool
// listen is the default operation of nanonymousCore. It listens on port 41721
// for incoming requests from the front end and passes them off to the handler.
func listen() error {
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
   resetInUse()

   // Listen for eternity to incoming address requests
   if (verbosity < 3) {
      fmt.Println("Listening....")
   }

   for (!safeExit) {
      if (verbosity >= 3) {
         fmt.Println("Listening....")
      }
      conn, err := listener.Accept()
      if (err != nil) {
         Error.Println("Error with single instance port:", err.Error())
         if (verbosity >= 3) {
            fmt.Println("Error with single instance port:", err.Error())
         }
      }

      go handleRequest(conn)
   }

   return nil
}

// handleRequest takes an established connection and responds to it. There are
// only two types of expected requests.
//    (1) Get a new address: Retuns the next valid address in the database
//    (2) Register Transaction: Regesiters a callback for a particular
//        transaction. On completion of the transaction nanonymous will return
//        the final send's hash.
func handleRequest(conn net.Conn) error {
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
   var text = string(buff)
   var array = strings.Split(text, "&")
   if (len(array) >= 2 && array[0] == "newaddress") {
      var subArray = strings.Split(array[1], "=")
      if (len(subArray) >= 2 && subArray[0] == "address") {
         if (addressExsistsInDB(subArray[1]) && !addressIsReceiveOnly(subArray[1])) {
            // Cannont send to a Nanonymous wallet as the recipient.
            conn.Write([]byte("Invalid Request!"))
         } else {
            newKey, _, err := getNewAddress(subArray[1], false, false, 0)
            if (err != nil) {
               if (verbosity >= 3) {
                  fmt.Println("handleRequest2: ", err.Error())
               }
               conn.Write([]byte("There was an error, please try again later"))
               conn.Close()
               return fmt.Errorf("handleRequest: %w", err)
            }

            conn.Write([]byte("address="+ newKey.NanoAddress))
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

                     if (missedPolls > 10) {
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
   } else if (strings.Contains(text, "safeExit")) {
      if (conn.LocalAddr().String() == "127.0.0.1:41721") {
         // Only allow exit command to come from the server itself
         safeExit = true
      }
      if (safeExit) {
         httpHeader :=
         "HTTP/1.1 200 OK\n"+
         "Content-Type: text/plain\n"+
         "Connection: Closed\n"
         conn.Write([]byte(httpHeader +"\nAck"))
      }
   }

   return nil
}

// getNewAddress finds the next availalbe address given the keys stored in the
// database and returns address B. If "receivingAddress" A is not an empty
// string, then it will also place A->B into the blacklist.
func getNewAddress(receivingAddress string, receiveOnly bool, mixer bool, seedId int) (*keyMan.Key, int, error) {
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
   var id int
   if (rows.Next()) {
      err = rows.Scan(&id, &seed.Seed, &seed.Index)
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

   if (id == 0) {
      // No valid seeds in database. Generate a new one.
      err = keyMan.GenerateSeed(&seed)
      if (err != nil) {
         return nil, 0, fmt.Errorf("getNewAddress: %w ", err)
      }

      id, err = insertSeed(tx, seed.Seed)
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
   rowsAffected, err := tx.Exec(context.Background(), queryString, id, seed.Index, hash[:], receiveOnly, mixer)
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

   rowsAffected, err = tx.Exec(context.Background(), queryString, seed.Index, id)
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

      err = blacklist(tx, seed.PublicKey, receivingAddressByte)
      if (err != nil) {
         return nil, 0, fmt.Errorf("getNewAddress: Blacklist falied: %w", err)
      }

      // Track so that when we receive funds we know where to send it
      err = setRecipientAddress(id, seed.Index, receivingAddressByte)
      if (err != nil) {
         Warning.Println("getNewAddress: ", err.Error())
         //return nil, 0, fmt.Errorf("getNewAddress: %w", err)
      }

      // Make sure we don't keep this forever
      go timeoutTransaction(id, seed.Index)
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


   return &seed, id, nil
}

// timeoutTransaction simply deletes the link between address B and C after the
// deadline since we don't want to keep that information indefinitely even in
// memory. There is no problem with double deleting.
func timeoutTransaction(id int, seedIndex int) {

   if (inTesting) {
      return
   }

   time.Sleep(TRANSACTION_DEADLINE)

   err := setRecipientAddress(id, seedIndex, nil)
   if (err != nil) {
      Warning.Println("timeoutTransaction: Failed to delete transaction: %w", err)
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
func blacklist(conn psqlDB, sendingAddress []byte, receivingAddress []byte) error {

   concat := append(sendingAddress, receivingAddress[:]...)

   hash := blake2b.Sum256(concat)

   queryString :=
   "INSERT INTO " +
      "blacklist (hash)" +
   "VALUES "+
      "($1);"

   rowsAffected, err := conn.Exec(context.Background(), queryString, hash[:])
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

   blacklist(conn, sendingAddress, receivePubKey)


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
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return fmt.Errorf("receivedNano: %w", err)
   }
   defer conn.Close(context.Background())

   parentSeedId, index, err := getWalletFromAddress(nanoAddress)
   if (err != nil) {
      return fmt.Errorf("receivedNano: %w", err)
   }

   err = BlockUntilReceivable(nanoAddress, 5 * time.Minute)
   if (err != nil) {
      return fmt.Errorf("receivedNano, funds not receiveable: %w", err)
   }

   payment, receiveHash, _, err := Receive(nanoAddress)
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
      sendInfoToClient("info=amountTooLow", getRecipientAddress(parentSeedId, index))
      err := Refund(receiveHash)
      if (err != nil) {
         sendEmail("IMMEDIATE ATTENTION REQUIRED", "Non-transaction refund failed! "+ err.Error() +
               "\n\nPayment Hash: "+ receiveHash.String() +
               "\nID: "+ strconv.Itoa(parentSeedId) +","+ strconv.Itoa(index) +
               "\nAmount: "+ strconv.FormatFloat(rawToNANO(payment), 'f', -1, 64))
         Error.Println("non-transaction Refund failed!! %w", err)
         return fmt.Errorf("non-transaction Refund failed!! %w", err)
      }

      // Transaction aborted
      return nil
   }

   // If anything goes wrong the transactionManager will make sure to clean up
   // the mess.
   var t Transaction
   go transactionManager(&t)
   t.commChannel = make(chan transactionComm)
   t.errChannel = make(chan error)
   t.paymentParentSeedId = parentSeedId
   t.paymentIndex = index
   t.receiveHash = receiveHash
   t.dirtyAddress = -1
   defer func () {
      if (err != nil) {
         t.errChannel <- err
      }
   }()
   t.id, err = getNextTransactionId()
   if (err != nil) {
      err = fmt.Errorf("receivedNano: %w", err)
      return err
   }

   Info.Println("Transaction", t.id, "started")

   t.paymentAddress, err = keyMan.AddressToPubKey(nanoAddress)
   if (err != nil) {
      err = fmt.Errorf("receivedNano: %w", err)
      return err
   }
   setAddressInUse(nanoAddress)

   // TODO This is just for debugging
   //fmt.Println("If you're not debegging than something is wrong!!!!")
   //seed, _ := getSeedFromIndex(1, 8)
   //recipientPub, _ := keyMan.AddressToPubKey(seed.NanoAddress)
   //setRecipientAddress(t.paymentParentSeedId, t.paymentIndex, recipientPub)
   // TODO end of debugging code

   // Get recipient address for later use.
   t.recipientAddress = getRecipientAddress(t.paymentParentSeedId, t.paymentIndex)
   if (t.recipientAddress == nil) {
      err = fmt.Errorf("receivedNano: no active transaction available")
      return err
   }

   t.fee = calculateFee(payment)
   t.amountToSend = nt.NewRaw(0).Sub(payment, t.fee)
   if (verbosity >= 5) {
      fmt.Println("payment:        ", payment,
                  "\nfee:            ", t.fee,
                  "\namount to send: ", t.amountToSend)
   }

   err = findSendingWallets(&t, conn)
   if (err != nil) {
      err = fmt.Errorf("receivedNano: %w", err)
      return err
   }

   err = sendNanoToRecipient(&t)
   if (err != nil) {
      err = fmt.Errorf("receivedNano: %w", err)
      return err
   }

   return nil
}

var lock sync.Mutex
// findSendingWallets is just a sub function of receivedNano(). It's only here
// for readablity and Mutexing the database.
func findSendingWallets(t *Transaction, conn *pgx.Conn) error {
   var foundEntry bool
   var err error

   // Make sure this function is only accessed one at a time.
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
   rows, err = conn.Query(context.Background(), queryString, t.amountToSend, t.paymentParentSeedId, t.paymentIndex)
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
         t.sendingKeys = append(t.sendingKeys, tmpKey)
         t.walletSeed = append(t.walletSeed, tmpSeed)
         t.walletBalance = append(t.walletBalance, tmpBalance)
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
         return fmt.Errorf("findSendingWallets: Query: %w", err)
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
            t.sendingKeys = append(t.sendingKeys, tmpKey)
            t.walletSeed = append(t.walletSeed, tmpSeed)
            // tmpBalance contains a pointer, so we need a new address to add to the list
            newAddress := nt.NewFromRaw(tmpBalance)
            t.walletBalance = append(t.walletBalance, newAddress)
            totalBalance.Add(totalBalance, tmpBalance)
            setAddressInUse(tmpKey.NanoAddress)
            if (totalBalance.Cmp(t.amountToSend) >= 0) {
               // We've found enough
               enough = true
               break
            }
         }
      }
      rows.Close()
      if (!enough) {
         // Not enough in managed wallets. Check the mixer.
         _, _, mixerBalance, err := findTotalBalance()
         if (err != nil) {
            return fmt.Errorf("findSendingWallets: %w", err)
         }

         if (mixerBalance.Add(mixerBalance, totalBalance).Cmp(t.amountToSend) >= 0) {
            // There's enough; add the keys to the transaction manager
            keys, seeds, balances, err := getKeysFromMixer(nt.NewRaw(0).Sub(t.amountToSend, totalBalance))
            if (err != nil) {
               return fmt.Errorf("findSendingWallets: %w", err)
            }

            for _, key := range keys {
               setAddressInUse(key.NanoAddress)
            }

            t.sendingKeys = append(t.sendingKeys, keys...)
            t.walletSeed = append(t.walletSeed, seeds...)
            t.walletBalance = append(t.walletBalance, balances...)

         } else {
            // Not enough even if we add the mixer.
            return fmt.Errorf("findSendingWallets: not enough funds")
         }
      }
   }

   return nil
}

// sendNanoToRecipient is just a subfunction of receivedNano(). It's just here for
// readability.
func sendNanoToRecipient(t *Transaction) error {
   var err error

   // Send nano to recipient
   if (len(t.sendingKeys) == 1) {
      t.individualSendAmount = append(t.individualSendAmount, t.amountToSend)
      go Send(t.sendingKeys[0], t.recipientAddress, t.amountToSend, t.commChannel, t.errChannel, 0)
      t.commChannel <- *new(transactionComm)
   } else if (len(t.sendingKeys) > 1) {
      // Need to do a multi-send; Get a new wallet to combine all funds into
      t.transitionalKey, t.transitionSeedId, err = getNewAddress("", false, false, 0)
      if (err != nil) {
         return fmt.Errorf("sendNanoToRecipient: %w", err)
      }
      t.multiSend = true

      // Go through list of wallets and send to interim address
      var totalSent = nt.NewRaw(0)
      var currentSend = nt.NewRaw(0)
      for i, key := range t.sendingKeys {

         // if (total + balance) > payment
         var arithmaticResult = nt.NewRaw(0)
         if (arithmaticResult.Add(totalSent, t.walletBalance[i]).Cmp(t.amountToSend) > 0) {
            currentSend = arithmaticResult.Sub(t.amountToSend, totalSent)
            t.dirtyAddress = i
         } else {
            currentSend = t.walletBalance[i]
         }
         t.individualSendAmount = append(t.individualSendAmount, currentSend)
         go Send(key, t.transitionalKey.PublicKey, currentSend, t.commChannel, t.errChannel, i)
         if (i == 0) {
            t.commChannel <- *new(transactionComm)
         }
         totalSent.Add(totalSent, currentSend)
         if (verbosity >= 5) {
            fmt.Println("Sending", currentSend.Int, "from", t.walletSeed[i], key.Index, "to", t.transitionSeedId, t.transitionalKey.Index)
         }
      }

      // Now send to recipient
      if (verbosity >= 5) {
         fmt.Println("Sending", t.amountToSend, "from", t.transitionSeedId, t.transitionalKey.Index, "to recipient.")
      }
      go ReceiveAndSend(t.transitionalKey, t.recipientAddress, t.amountToSend, t.commChannel, t.errChannel, &t.receiveWg, &t.abort)
   } else {
      return fmt.Errorf("sendNanoToRecipient: not enough funds(2)")
   }

   return nil
}
// Send is intended to be used with receivedNano() (although it doesn't have to be). It's a wrapper to sendNano that gives callbacks to the transaction manager when done.
func Send(fromKey *keyMan.Key, toPublicKey []byte, amount *nt.Raw, commCh chan transactionComm, errCh chan error, i int) (nt.BlockHash, error) {
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
func getRecipientAddress(parentSeedId int, index int) []byte {
   key := strconv.Itoa(parentSeedId) + "-" + strconv.Itoa(index)
   return activeTransactionList[key]
}

// setRecipientAddress is an interface to work with the active transaction list.
// Adds or removes an entry that maps one of our internal wallets to a recipient
// address.
func setRecipientAddress(parentSeedId int, index int, recipientAddress []byte) error {
   key := strconv.Itoa(parentSeedId) + "-" + strconv.Itoa(index)
   if (activeTransactionList[key] != nil) {
      if (recipientAddress == nil) {
         delete(activeTransactionList, key)
      } else {
         return fmt.Errorf("setRecipientAddress: address already exists in active transaction list")
      }
   }

   if (recipientAddress != nil) {
      activeTransactionList[key] = recipientAddress
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

// needs improvment but can probaby be in CLI TODO
func ReceiveAll(account string) ([]nt.BlockHash, error) {
   var hashes []nt.BlockHash

   for {
      _, hash, numLeft, err := Receive(account)
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
func Receive(account string) (*nt.Raw, nt.BlockHash, int, error) {
   var block keyMan.Block
   var pendingInfo BlockInfo
   var newHash nt.BlockHash
   var numOfPendingHashes int

   key, _, _, err := getSeedFromAddress(account)
   if (err != nil) {
      return nil, nil, 0, fmt.Errorf("receive: %w", err)
   }

   if !(inTesting) {
      pendingHashes, _ := getReceivable(account, -1)
      numOfPendingHashes = len(pendingHashes)

      if (numOfPendingHashes > 0) {
         pendingHash := pendingHashes[0]
         pendingInfo, _ = getBlockInfo(pendingHash)
         accountInfo, err := getAccountInfo(account)
         if (err != nil) {
            // Filter out expected errors
            if !(strings.Contains(err.Error(), "Account not found")) {
               return nil, nil, 0, fmt.Errorf("Receive: %w", err)
            }
         }

         // Fill block with relavent information
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
            return nil, nil, 0, fmt.Errorf("receive: %w", err)
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
            return nil, nil, 0, fmt.Errorf("receive: %w", err)
         }

         // Send RCP request
         newHash, err = publishReceive(block, sig, PoW)
         if (err != nil){
            return nil, nil, 0, fmt.Errorf("receive: %w", err)
         }
         if (len(newHash) != 32){
            return nil, nil, 0, fmt.Errorf("receive: no block hash returned from node")
         }

         numOfPendingHashes--

         // We'ved used any stored PoW, clear it out for next use
         clearPoW(block.Account)
         go preCalculateNextPoW(block.Account, false)

         // Update database records
         err = updateBalance(block.Account, block.Balance)
         if (err != nil) {
            Error.Println("Balance update failed from receive:", err.Error())
            return pendingInfo.Amount, newHash, 0, fmt.Errorf("receive: updatebalance error %w", databaseError)
         }
      }
   } else {
      // Doing testing; behave as close as possible without calling RCP

      balance, _ := getBalance(account)

      if (testingPaymentExternal) {
         pendingInfo.Amount = testingPayment[0]
         testingPaymentExternal = false
      } else if (len(testingSends[account]) > 0) {
         pendingInfo.Amount = testingSends[account][len(testingSends[account])-1]
         testingSends[account] = testingSends[account][:len(testingSends[account])-1]
      } else {
         return pendingInfo.Amount, newHash, 0, fmt.Errorf("receive: No funds receiveable on %s", account)
      }

      newBalance := nt.NewRaw(0).Add(balance, pendingInfo.Amount)

      // Update database records
      err = updateBalance(account, newBalance)
      if (err != nil) {
         Error.Println("Balance update failed from receive:", err.Error())
         return pendingInfo.Amount, newHash, 0, fmt.Errorf("receive: updatebalance error %w", databaseError)
      }

      testingPendingHashesNum[testingReceiveAlls]--
      numOfPendingHashes = testingPendingHashesNum[testingReceiveAlls]
      if (testingPendingHashesNum[testingReceiveAlls] == 0) {
         testingReceiveAlls++
      }
   }

   return pendingInfo.Amount, newHash, numOfPendingHashes, err
}

// TODO chekc that all of these are still good active nodes before going live
// getNewRepresentative returns a random representative from a list of accounts
// that were recommended by mynano.ninja.
func getNewRepresentative() string {

   hardcodedList := []string {
      "nano_1my1snode8rwccjxkckjirj65zdxo6g5nhh16fh6sn7hwewxooyyesdsmii3", // My1s
      "nano_3msc38fyn67pgio16dj586pdrceahtn75qgnx7fy19wscixrc8dbb3abhbw6", // grOvity
      "nano_3pnanopr3d5g7o45zh3nmdkqpaqxhhp3mw14nzr41smjz8xsrfyhtf9xac77", // PlayNANO
      "nano_1wenanoqm7xbypou7x3nue1isaeddamjdnc3z99tekjbfezdbq8fmb659o7t", // WeNano
      "nano_3afmp9hx6pp6fdcjq96f9qnoeh1kiqpqyzp7c18byaipf48t3cpzmfnhc1b7", // Fast&Feeless
      "nano_396sch48s3jmzq1bk31pxxpz64rn7joj38emj4ueypkb9p9mzrym34obze6c", // SupeNode
      "nano_3kqdiqmqiojr1aqqj51aq8bzz5jtwnkmhb38qwf3ppngo8uhhzkdkn7up7rp", // ARaiNode
      "nano_18shbirtzhmkf7166h39nowj9c9zrpufeg75bkbyoobqwf1iu3srfm9eo3pz", // DE
      "nano_3uaydiszyup5zwdt93dahp7mri1cwa5ncg9t4657yyn3o4i1pe8sfjbimbas", // NANO Voting
      "nano_3n7ky76t4g57o9skjawm8pprooz1bminkbeegsyt694xn6d31c6s744fjzzz", // humble finland
   }

   randAddr := random.Intn(len(hardcodedList))
   return hardcodedList[randAddr]
}

// preCalculateNextPoW finds the proper hash, calcluates the PoW and saves it to
// the database for future use.
func preCalculateNextPoW(nanoAddress string, isReceiveBlock bool) {
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
   // TODO should receiveblocks request work from the node instead?
   // TODO do we need to make this work for multiple recalls? At the moment the channels would block

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
            case <-time.After(5 * time.Minute):
         }
      }

      delete(activePoW, nanoAddress)
      delete(workChannel, nanoAddress)

      return work

   } else {
      // PoW is already being computed by someone else. Wait for them to report it
      activePoW[nanoAddress] = 2
      workChannel[nanoAddress] = make(chan string)

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

// TODO
func checkBalance(nanoAddress string) error {

   balance, receiveable, _ := getAccountBalance(nanoAddress)

   if (receiveable.Cmp(nt.NewRaw(0)) != 0) {
      // Receive and update
      Receive(nanoAddress)
   } else {

      balanceInDB, err := getBalance(nanoAddress)
      if (err != nil) {
         return fmt.Errorf("checkBalance: %w", err)
      }

      if (balance.Cmp(balanceInDB) != 0) {
         updateBalance(nanoAddress, balance)
      }
   }

   return nil
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

   // Don't want the user to have to deal with dust so I'll round the fee down
   // to the nearest .001 * minimum
   minDust := nt.OneNano().Div(minPayment, nt.NewRaw(1000))

   _, dust := nt.OneNano().DivMod(fee, minDust)

   // Remove any dust from the fee
   fee.Sub(fee, dust)

   return fee
}

// returnAllReceivable checks all wallets in all seeds and finds any receiveable
// funds that are just lying around (Maybe someone accidentally re-sent funds to
// an address that I'm no longer actively monitoring). If the wallet is not
// marked as "in use" then it returns the funds to the original owner. This
// function is designed to be called occasionally by a seperate process to clean
// up any accidental sends from users.
func returnAllReceiveable() error {
   // TODO Might need to make sure it's not an internal send before refunding. I think that might be causing headache....

   rows, conn, err := getSeedRowsFromDatabase()
   if (err != nil) {
      Warning.Println("getSeedRowsFromDatabase failed on routine pending check:", err.Error())
      return fmt.Errorf("returnAllReceiveable: %w", err)
   }

   var seed keyMan.Key
   // for all our seeds
   for rows.Next() {
      rows.Scan(&seed.Seed, &seed.Index)

      // From max index to 0 return all funds
      for i := seed.Index; i >= 0; i-- {
         seed.Index = i
         keyMan.SeedToKeys(&seed)

         if (verbosity >= 5) {
            fmt.Println("  ", seed.NanoAddress)
            fmt.Println("  index", i)
         }
         // Check to make sure it's not being used in a current transaction
         inUse, err := isAddressInUse(seed.NanoAddress)
         if (err != nil) {
            Warning.Println("isAddresInUse failed on routine pending check:", err.Error())
            continue
         }
         if (inUse) {
            continue
         }

         hashes, err := getReceivable(seed.NanoAddress, -1)
         if (err != nil) {
            Warning.Println("getReceivable failed on routine pending check:", err.Error())
            return fmt.Errorf("returnAllReceiveable: %w", err)
         }
         numberOfHashes := len(hashes)

         for j := 0; j < numberOfHashes; j++ {
            if (verbosity >= 5) {
               fmt.Println("Receivable hash: ", hashes[j])
            }
            // Found funds. Receive them first and then refund them.
            _, receiveHash, _, _ := Receive(seed.NanoAddress)

            blockinfo, _ := getBlockInfo(hashes[j])
            block := blockinfo.Contents

            // TODO Test to make sure the internal send is being detected correctly
            // Don't refund if wallet is receive only or it's an internal send
            if (!addressIsReceiveOnly(seed.NanoAddress) && !addressExsistsInDB(block.Account)) {
               // Poll until block is confirmed
               for {
                  time.Sleep(5 * time.Second)

                  info, _ := getBlockInfo(receiveHash)
                  if (info.Confirmed) {
                     if (verbosity >= 6) {
                        fmt.Println(" Hash confirmed!")
                     }
                     break
                  } else if (verbosity >= 6) {
                     fmt.Println("Waiting on hash...")
                     if (verbosity >= 7) {
                        fmt.Println(info)
                     }
                  }
               }

               if (verbosity >= 5) {
                  fmt.Println("      Refunding")
               }
               err := Refund(receiveHash)
               if (err != nil) {
                  Warning.Println("Refund failed on", seed.NanoAddress, "during routine pending check:", err.Error())
               }
            }
         }
      }
   }

   conn.Close(context.Background())

   return nil
}
