package main

import (
   "fmt"
   "time"
   "context"
   "encoding/hex"
   "strings"
   "math/big"
   "strconv"
   "math"
   "math/rand"
   "sync"
   "log"
   "os"
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

// Should only be set to true in test functions
var testing = false

const MAX_INDEX = 4294967295

// Fee in %
const FEE_PERCENT = float64(0.2)
var feeDividend int64
var minPayment *nt.Raw

var wg sync.WaitGroup

var activeTransactionList = make(map[string][]byte)

var random *rand.Rand

type psqlDB interface {
   Begin(ctx context.Context) (pgx.Tx, error)
   Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
   Query(ctx context.Context, sql string, optionsAndArgs ...interface{}) (pgx.Rows, error)
   QueryRow(ctx context.Context, sql string, optionsAndArgs ...interface{}) pgx.Row
}

var Info *log.Logger
var Warning *log.Logger
var Error *log.Logger

var verbosity int
func main() {
   err := initNanoymousCore()
   if (err != nil) {
      panic(err)
   }

   var usr string
   var seed keyMan.Key

   menu:
   for {
      fmt.Print("1. Generate Seed\n",
                "2. Get Account info\n",
                "3. Insert into database\n",
                "4. Send pretend request for new address\n",
                "5. Find total balance\n",
                "6. Pretend nano receive\n",
                "7. Get Wallet Info\n",
                "8. Add nano to wallet\n",
                "9. Get block count\n",
                "A. Clear PoW\n",
                "B. Telemetry\n",
                "C. Get Account Info\n",
                "D. Sign Block\n",
                "E. Block Info\n",
                "H. OpenAccount\n",
                "I. GenerateWork\n",
                "J. Send\n",
                "K. Recive All\n",
                "L. Black list\n",
                "M. Get Pending\n",
                "N. Check Balance\n",
                "O. Channel Test\n",
                "P. Test\n",
             )
      fmt.Scan(&usr)

      switch (strings.ToUpper(usr)) {
      case "1":
         keyMan.WalletVerbose(true)

         err := keyMan.GenerateSeed(&seed)
         if (err != nil){
            fmt.Println(err.Error())
         }

         keyMan.WalletVerbose(false)
      case "2":
         verbosity = 10
         fmt.Print("Seed: ")
         fmt.Scan(&usr)
         seed, _ := strconv.Atoi(usr)
         fmt.Print("Index: ")
         fmt.Scan(&usr)
         index, _ := strconv.Atoi(usr)
         seedKey, _ := getSeedFromIndex(seed, index)
         _, err := getAccountInfo(seedKey.NanoAddress)
         if (err != nil) {
            fmt.Println("error: ", err.Error())
         }
         _, _, err = getAccountBalance(seedKey.NanoAddress)
         if (err != nil) {
            fmt.Println("error: ", err.Error())
         }

      case "3":
         var newSeed keyMan.Key

         conn, err := pgx.Connect(context.Background(), databaseUrl)
         if (err != nil) {
            fmt.Println("main: ", err)
            return
         }
         defer conn.Close(context.Background())

         err = keyMan.GenerateSeed(&newSeed)
         if (err != nil) {
            fmt.Println("main: ", err)
            break
         }

         hexString := hex.EncodeToString(newSeed.Seed)
         fmt.Println("seed: ", hexString)

         _, err = insertSeed(conn, newSeed.Seed)

         if (err != nil) {
            fmt.Println("main: ", err)
            break //menu
         }

      case "4":
         verbosity = 5
         adhocAddress := "nano_1hiqiw6j9wo33moia3scoajhheweysiq5w1xjqeqt8m6jx6so6gj39pae5ea"
         blarg, _, err := getNewAddress(adhocAddress)
         if (err != nil) {
            fmt.Println(err)
         }
         fmt.Println("New address: ", blarg.NanoAddress)
      case "5":
         verbosity = 5
         _, err := findTotalBalance()
         if (err != nil) {
            fmt.Println("err: ", err.Error())
         }
      case "6":
         seed, _ := getSeedFromIndex(1, 7)
         err := receivedNano(seed.NanoAddress)
         if (err != nil) {
            fmt.Println("err: ", err.Error())
         }

      case "7":
         keyMan.WalletVerbose(true)
         verbosity = 5
         fmt.Scan(&usr)
         seed, _ := strconv.Atoi(usr)
         fmt.Scan(&usr)
         index, _ := strconv.Atoi(usr)

         _, err := getSeedFromIndex(seed, index)
         if (err != nil) {
            fmt.Println(err.Error())
         }

      case "8":
         fmt.Print("Seed: ")
         fmt.Scan(&usr)
         seed, _ := strconv.Atoi(usr)
         fmt.Print("Index: ")
         fmt.Scan(&usr)
         index, _ := strconv.Atoi(usr)
         fmt.Print("Nano: ")
         fmt.Scan(&usr)
         nano, _ := strconv.Atoi(usr)

         manualWalletUpdate(seed, index, int64(nano))

      case "9":
         verbosity = 5
         getBlockCount()
      case "A":
         verbosity = 5
         fmt.Print("Seed: ")
         fmt.Scan(&usr)
         seed, _ := strconv.Atoi(usr)
         fmt.Print("Index: ")
         fmt.Scan(&usr)
         index, _ := strconv.Atoi(usr)
         seedKey, _ := getSeedFromIndex(seed, index)
         err := clearPoW(seedKey.NanoAddress)
         if (err != nil) {
            fmt.Println("error: ", err.Error())
         }
      case "B":
         verbosity = 5
         err := telemetry()
         if (err != nil) {
            fmt.Println("err: ", err.Error())
         }
      case "C":
         verbosity = 5
         getAccountInfo("nano_1afhc54bhcqkrdwz7rwmwcexfssnwsbbkzwfhj3a7wxa17k93zbh3k4cknkb")
      case "E":
         verbosity = 5
         h, _ := hex.DecodeString("BCEFF64B4B62B20CAC352D77F518128FB2E46A2717CE7F40C3E04A8570211BFC")
         //block, _ := getBlockInfo(h)
         getBlockInfo(h)

      case "H":
         received, _, _,  err := Receive("nano_17iperkf8wx68akk66t4zynhuep7oek397nghh75oenahauch41g6pfgtrnu")
         if (err != nil) {
            fmt.Println("Error: ", err.Error())
         }
         fmt.Println("Received:", received)

      case "I":
         var seedSend *keyMan.Key
         var seedReceive *keyMan.Key
         seedSend, _ = getSeedFromIndex(1, 0)
         seedReceive, _ = getSeedFromIndex(1, 12)
         SendEasy(seedSend.NanoAddress,
                  seedReceive.NanoAddress,
                  nt.NewRawFromNano(0.50),
                  false)
      case "J":
         verbosity = 5
         var seedSend *keyMan.Key
         var seedReceive *keyMan.Key
         seedSend, _ = getSeedFromIndex(1, 0)
         seedReceive, _ = getSeedFromIndex(1, 7)
         SendEasy(seedSend.NanoAddress,
                  seedReceive.NanoAddress,
                  nt.NewRawFromNano(1.00),
                  false)
         //seedSend, _ = getSeedFromIndex(1, 0)
         //seedReceive, _ = getSeedFromIndex(1, 2)
         //SendEasy(seedSend.NanoAddress,
                  //seedReceive.NanoAddress,
                  //nt.NewRawFromNano(0.50),
                  //false)
         //seedSend, _ = getSeedFromIndex(1, 0)
         //seedReceive, _ = getSeedFromIndex(1, 3)
         //SendEasy(seedSend.NanoAddress,
                  //seedReceive.NanoAddress,
                  //nt.NewRawFromNano(0.5),
                  //false)
         //time.Sleep(5 * time.Second)
         //seedSend, _ = getSeedFromIndex(1, 10)
         //seedReceive, _ = getSeedFromIndex(1, 3)
         //SendEasy(seedSend.NanoAddress,
                  //seedReceive.NanoAddress,
                  //nt.NewRawFromNano(0.5),
                  //false)
         //seedSend, _ = getSeedFromIndex(1, 0)
         //seedReceive, _ = getSeedFromIndex(1, 6)
         //SendEasy(seedSend.NanoAddress,
                  //seedReceive.NanoAddress,
                  //nt.NewRawFromNano(1.0),
                  //false)
         //seedSend, _ = getSeedFromIndex(1, 11)
         //seedReceive, _ = getSeedFromIndex(1, 0)
         //SendEasy(seedSend.NanoAddress,
                  //seedReceive.NanoAddress,
                  //nt.NewRawFromNano(0.5),
                  //false)
      case "K":
         verbosity = 5

         for i := 0; i <= 41; i++ {
            fmt.Println("--------------", i, "-------------")
            seedReceive, _ := getSeedFromIndex(1, i)
            err := ReceiveAll(seedReceive.NanoAddress)
            if (err != nil) {
               fmt.Println("Error:", err.Error())
            }
         }
      case "L":
         conn, err := pgx.Connect(context.Background(), databaseUrl)
         if (err != nil) {
            fmt.Println(err.Error())
         }
         defer conn.Close(context.Background())

         seedSend, _ := getSeedFromIndex(1, 0)
         seedReceive, _ := getSeedFromIndex(1, 8)
         blacklist(conn, seedSend.PublicKey, seedReceive.PublicKey)

      case "M":
         verbosity = 5
         seed, _ := getSeedFromIndex(1, 5)
         blarg, _ := getPendingHashes(seed.NanoAddress)
         fmt.Println(blarg[seed.NanoAddress][0])
      case "N":
         verbosity = 5
         //fmt.Print("Seed: ")
         //fmt.Scan(&usr)
         //seed, _ := strconv.Atoi(usr)
         //fmt.Print("Index: ")
         //fmt.Scan(&usr)
         //index, _ := strconv.Atoi(usr)
         for index := 0; index <= 41; index++ {
            seedkey, _ := getSeedFromIndex(1, index)
            err := checkBalance(seedkey.NanoAddress)
            if (err != nil) {
               fmt.Println(err.Error())
            }
         }
      case "O":
         verbosity = 5

         seedkey, _ := getSeedFromIndex(1, 0)
         go preCalculateNextPoW(seedkey.NanoAddress, true)
         //time.Sleep(5 * time.Second)
         work := calculateNextPoW(seedkey.NanoAddress, true)

         fmt.Println("work: ", work)
      case "P":
         verbosity = 10

         seed, _ := getSeedFromIndex(1, 0)
         var bigString string
         for i:=0; i < 5000; i++ {
            seed.Index = i
            keyMan.SeedToKeys(seed)

            bigString += seed.NanoAddress
            if (i % 200 == 0) {
               fmt.Println(i)
            }
         }
         fmt.Println(bigString)


      default:
         break menu
      }
   }



   wg.Wait()
}

// initNanoymousCore sets up our variables that need to preexist before other
// functions can be called.
func initNanoymousCore() error {
   // Init loggers
   _, err := os.ReadDir("./logs")
   if (err != nil ){
      if (strings.Contains(err.Error(), "no such file or directory")) {
         os.Mkdir("./logs", 0755)
      } else {
         return fmt.Errorf("initNanoymousCore: %w", err)
      }
   }

   logFile, err := os.OpenFile("./logs/logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
   if err != nil {
    return fmt.Errorf("initNanoymousCore: %w", err)
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

   resetInUse()

   feeDividend = int64(math.Trunc(100/FEE_PERCENT))
   minPayment = oneNano()

   activePoW = make(map[string]int)
   workChannel = make(map[string]chan string)

   // Seed randomness
   random = rand.New(rand.NewSource(time.Now().UnixNano()))

   ch := make(chan int)
   go websocketListener(ch)
   // Wait until websockets are initialized
   <-ch

   return nil
}

// getNewAddress finds the next availalbe address given the keys stored in the
// database and returns address B. If "receivingAddress" A is not an empty
// string, then it will also place A->B into the blacklist.
func getNewAddress(receivingAddress string) (*keyMan.Key, int, error) {
   var seed keyMan.Key

   Info.Println("New address requested")
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

   queryString :=
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
         keyMan.SeedToKeys(&seed)
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
      "wallets(parent_seed, index, balance, hash) " +
   "VALUES " +
      "($1, $2, 0, $3)"

   hash := blake2b.Sum256(seed.PublicKey)
   rowsAffected, err := tx.Exec(context.Background(), queryString, id, seed.Index, hash[:])
   if (err != nil) {
      return nil, 0, fmt.Errorf("getNewAddress: %w", err)
   }
   if (rowsAffected.RowsAffected() < 1) {
      return nil, 0, fmt.Errorf("getNewAddress: no rows affected in insert")
   }

   // Generate work for first use
   go preCalculateNextPoW(seed.NanoAddress, true)

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
      err = setClientAddress(id, seed.Index, receivingAddressByte)
      if (err != nil) {
         return nil, 0, fmt.Errorf("getNewAddress: %w", err)
      }
   }

   err = tx.Commit(context.Background())
   if (err != nil) {
      return nil, 0, fmt.Errorf("getNewAddress: %w", err)
   }

   // Track confirmations on websocket
   select {
      case addWebSocketSubscription <- seed.NanoAddress:
      case <-time.After(3 * time.Second):
         if (verbosity >= 5) {
            fmt.Println("Subscription add failed!")
         }
         Warning.Println("Add subscription timout")
   }


   return &seed, id, nil
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
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return fmt.Errorf("receivedNano: %w", err)
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

// receivedNano is a large function that does most of the work for nanonymous.
// Upon receiving nano it does 5 distinct things:
//    (1) Updates the database with the newly recived nano
//    (2) Checks if we were expecting the tranaction
//    (3) Calculates the fee
//    (4) Finds the wallet(s) with enough funds to support the transaction
//        (minus the blacklisted ones)
//    (5) Sends the funds to the client
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

   payment, receiveHash, _, err := Receive(nanoAddress)
   if (err != nil) {
      return fmt.Errorf("receivedNano: %w", err)
   }
   if (payment == nil || payment.Cmp(nt.NewRaw(0)) == 0) {
      return fmt.Errorf("receivedNano: No payment received")
   }

   // If anything goes wrong the transactionManager will make sure to clean up
   // the mess.
   var t Transaction
   go transactionManager(&t)
   t.commChannel = make(chan int)
   t.errChannel = make(chan error)
   t.paymentParentSeedId = parentSeedId
   t.paymentIndex = index
   t.receiveHash = receiveHash
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
   seed, _ := getSeedFromIndex(1, 8)
   clientPub, _ := keyMan.AddressToPubKey(seed.NanoAddress)
   setClientAddress(t.paymentParentSeedId, t.paymentIndex, clientPub)
   // TODO end of debugging code

   // Get client address for later use.
   t.clientAddress = getClientAddress(t.paymentParentSeedId, t.paymentIndex)
   if (t.clientAddress == nil) {
      err = fmt.Errorf("receivedNano: no active transaction available")
      return err
   }
       //keyMan.NewRaw(0).Div(payment, keyMan.NewRaw(feeDividend))
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

   err = sendNanoToClient(&t)
   if (err != nil) {
      err = fmt.Errorf("receivedNano: %w", err)
      return err
   }

   return nil
}

var lock sync.Mutex
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
      "in_use = FALSE " +
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
      foundEntry, tmpKey, err = checkBlackList(tmpSeed, tmpIndex, t.clientAddress)
      if (err != nil) {
         return fmt.Errorf("findSendingWallets: %w", err)
      }
      if (!foundEntry) {
         // Uset this address
         t.sendingKeys = append(t.sendingKeys, tmpKey)
         t.walletSeed = append(t.walletSeed, tmpSeed)
         t.walletBalance = append(t.walletBalance, tmpBalance)
         setAddressInUse(tmpKey.NanoAddress)
         foundAddress = true
         if (verbosity >= 5) {
            fmt.Println("sending from:", tmpSeed, tmpIndex, "to client")
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
         "index = $2)" +
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
         foundEntry, tmpKey, err = checkBlackList(tmpSeed, tmpIndex, t.clientAddress)
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
         return fmt.Errorf("findSendingWallets: not enough funds")
      }
   }

   return nil
}

func sendNanoToClient(t *Transaction) error {
   var err error

   // Send nano to client
   if (len(t.sendingKeys) == 1) {
      t.individualSendAmount = append(t.individualSendAmount, t.amountToSend)
      go Send(t.sendingKeys[0], t.clientAddress, t.amountToSend, t.commChannel, t.errChannel, 0)
      t.commChannel <- 1
   } else if (len(t.sendingKeys) > 1) {
      // Need to do a multi-send; Get a new wallet to combine all funds into
      t.transitionalKey, t.transitionSeedId, err = getNewAddress("")
      if (err != nil) {
         return fmt.Errorf("sendNanoToClient: %w", err)
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
         } else {
            currentSend = t.walletBalance[i]
         }
         t.individualSendAmount = append(t.individualSendAmount, currentSend)
         go Send(key, t.transitionalKey.PublicKey, currentSend, t.commChannel, t.errChannel, i)
         if (i == 0) {
            t.commChannel <- 1
         }
         totalSent.Add(totalSent, currentSend)
         if (verbosity >= 5) {
            fmt.Println("Sending", currentSend.Int, "from", t.walletSeed[i], key.Index, "to", t.transitionSeedId, t.transitionalKey.Index)
         }
      }

      // Now send to client
      if (verbosity >= 5) {
         fmt.Println("Sending", t.amountToSend, "from", t.transitionSeedId, t.transitionalKey.Index, "to client.")
      }
      go ReceiveAndSend(t.transitionalKey, t.clientAddress, t.amountToSend, t.commChannel, t.errChannel, &t.receiveWg, &t.abort)
   } else {
      return fmt.Errorf("sendNanoToClient: not enough funds(2)")
   }

   return nil
}

// rawToNANO is used to convert raw to NANO AKA Mnano (the communnity just calls
// this a nano). We don't have a conversion to go the other way as all
// operations should be done in raw to avoid rounding errors. We only want to
// convert when outputing for human readable format.
func rawToNANO(raw *nt.Raw) float64 {
   // 1 NANO is 10^30 raw
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

func getClientAddress(parentSeedId int, index int) []byte {
   key := strconv.Itoa(parentSeedId) + "-" + strconv.Itoa(index)
   return activeTransactionList[key]
}

func setClientAddress(parentSeedId int, index int, clientAddress []byte) error {
   key := strconv.Itoa(parentSeedId) + "-" + strconv.Itoa(index)
   if (activeTransactionList[key] != nil) {
      return fmt.Errorf("setClientAddress: address already exists in active trnasaction list")
   }

   if (len(clientAddress) != 0 ) {
      activeTransactionList[key] = clientAddress
   } else {
      delete(activeTransactionList, key)
   }

   return nil
}

func ReceiveAndSend(transitionalKey *keyMan.Key, toPublicKey []byte, amount *nt.Raw, commCh chan int, errCh chan error, transactionWg *sync.WaitGroup, abort *bool) {
   wg.Add(1)
   defer wg.Done()

   // Wait until all sends have processed and confirmed
   transactionWg.Wait()
   transactionWg.Add(1)
   if (*abort) {
      return
   }

   // All transactions are still pending. Receive the funds.
   err := ReceiveAll(transitionalKey.NanoAddress)
   if (err != nil) {
      err = fmt.Errorf("ReceiveAndSend: %w", err)
      errCh <- err
   } else {
      commCh <- 2
   }

   // Wait until all receives have processed and confirmed
   transactionWg.Wait()
   transactionWg.Add(1)
   if (*abort) {
      return
   }

   // Finally, send to client.
   err = Send(transitionalKey, toPublicKey, amount, nil, nil, -1)
   if (err != nil) {
      err = fmt.Errorf("ReceiveAndSend: %w", err)
      errCh <- err
   } else {
      commCh <- 3
   }
}

func Send(fromKey *keyMan.Key, toPublicKey []byte, amount *nt.Raw, commCh chan int, errCh chan error, i int) error {

   err := sendNano(fromKey, toPublicKey, amount)
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
      return fmt.Errorf("Send: %w", err)
   }

   setAddressNotInUse(fromKey.NanoAddress)

   if (commCh != nil) {
      commCh <- i
   }

   return nil
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
      err = Send(&fromKey, toKey.PublicKey, amount, nil, nil, -1)
      if (err != nil && verbosity >= 5) {
         fmt.Println("Error: ", err.Error())
      }
   }
}

func sendNano(fromKey *keyMan.Key, toPublicKey []byte, amountToSend *nt.Raw) error {
   var block keyMan.Block

   accountInfo, err := getAccountInfo(fromKey.NanoAddress)
   if (err != nil) {
      return fmt.Errorf("sendNano: %w", err)
   }

   // if (Balance < amountToSend)
   if (accountInfo.Balance.Cmp(amountToSend) < 0) {
      return fmt.Errorf("sendNano: not enough funds in account.\n have: %s\n need: %s", accountInfo.Balance, amountToSend)
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
      return fmt.Errorf("sendNano: %w", err)
   }

   if (verbosity >= 5) {
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
      return fmt.Errorf("sendNano: %w", err)
   }

   // Send RCP request
   newHash, err := publishSend(block, sig, PoW)
   if (err != nil){
      return fmt.Errorf("sendNano: %w", err)
   }
   if (len(newHash) != 32){
      return fmt.Errorf("sendNano: no block hash returned from node", err)
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
      return fmt.Errorf("sendNano: updatebalance error %w", databaseError)
   }

   return nil
}

func sendAllNano(fromKey *keyMan.Key, toPublicKey []byte) error {
   defer setAddressNotInUse(fromKey.NanoAddress)

   balance, _, err := getAccountBalance(fromKey.NanoAddress)
   if (err != nil) {
      return fmt.Errorf("sendAllNano: %w", err)
   }

   err = sendNano(fromKey, toPublicKey, balance)
   if (err != nil) {
      return fmt.Errorf("sendAllNano: %w", err)
   }

   return nil
}

func manualWalletUpdate(seed int, index int, nano int64) error {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return fmt.Errorf("manualWalletUpdatn: %w", err)
   }
   defer conn.Close(context.Background())

   amount := nt.NewRaw(0).Mul(nt.NewRaw(nano), nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil))

   queryString :=
   "UPDATE " +
      "wallets " +
   "SET " +
      "\"balance\" = \"balance\" + $1 " +
   "WHERE " +
      "\"parent_seed\" = $2 AND " +
      "\"index\" = $3;"

   rowsAffected, err := conn.Exec(context.Background(), queryString, amount, seed, index)
   if (err != nil) {
      return fmt.Errorf("manualWalletUpdate: Update: %w", err)
   }
   if (rowsAffected.RowsAffected() < 1) {
      return fmt.Errorf("manualWalletUpdate: no rows affected during index incrament")
   }

   return nil
}

func ReceiveAll(account string) error {

   for {
      _, _, numLeft, err := Receive(account)
      if (err != nil) {
         return fmt.Errorf("ReceiveAll: %w", err)
      }
      if (numLeft <= 0) {
         break
      }
   }

   return nil
}

func Receive(account string) (*nt.Raw, nt.BlockHash, int, error) {
   var block keyMan.Block
   var pendingInfo BlockInfo
   var newHash nt.BlockHash

   key, _, _, err := getSeedFromAddress(account)
   if (err != nil) {
      return nil, nil, 0, fmt.Errorf("receive: %w", err)
   }

   pendingHashes, _ := getPendingHashes(account)
   numOfPeningHashes := len(pendingHashes[account])

   if (numOfPeningHashes > 0) {
      pendingHash := pendingHashes[account][0]
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

      if (verbosity >= 5) {
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
         return nil, nil, 0, fmt.Errorf("receive: no block hash returned from node", err)
      }

      numOfPeningHashes--

      // We'ved used any stored PoW, clear it out for next use
      clearPoW(block.Account)
      go preCalculateNextPoW(block.Account, false)

      // Update database records
      err = updateBalance(block.Account, block.Balance)
      if (err != nil) {
         Error.Println("Balance update failed from receive:", err.Error())
         return pendingInfo.Amount, newHash, 0, fmt.Errorf("sendNano: updatebalance error %w", databaseError)
      }
   }

   return pendingInfo.Amount, newHash, numOfPeningHashes, err
}

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

   rand := random.Intn(len(hardcodedList))
   return hardcodedList[rand]
}

// preCalculateNextPoW finds the proper hash, calcluates the PoW and saves it to
// the database for future use.
func preCalculateNextPoW(nanoAddress string, isReceiveBlock bool) {
   wg.Add(1)
   defer wg.Done()
   if (testing) {
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

var activePoW map[string]int
var workChannel map[string]chan string

// calculateNextPoW finds the hash and calculates the PoW using any number of
// different work servers setup at runtime. It returns the work generated as a
// string. It also will make sure not to request work that is already being
// worked on, and so can safely be called multiple times on the same account.
func calculateNextPoW(nanoAddress string, isReceiveBlock bool) string {
   if (testing) {
      return ""
   }
   // TODO should receiveblocks request work from the node instead?

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
         Warning.Println("Failed to connect to work server")
         // Fall back server
         work, err = generateWorkOnNode(hash, difficulty)
         if (err != nil) {
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

func calculateFee(payment *nt.Raw) *nt.Raw {

   // Find base fee simply by taking the percentage
   fee := nt.NewRaw(0).Div(payment, nt.NewRaw(feeDividend))

   // Don't want the user to have to deal with dust so I'll round the fee down
   // to the nearest .001 * minimum
   minDust := oneNano().Div(minPayment, nt.NewRaw(1000))

   _, dust := oneNano().DivMod(fee, minDust)

   // Remove any dust from the fee
   fee.Sub(fee, dust)

   return fee
}

func oneNano() *nt.Raw {
   return nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil)
}
