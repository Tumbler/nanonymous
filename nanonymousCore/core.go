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
   "sync"
   _"embed"

   // Local packages
   keyMan "nanoKeyManager"

   // 3rd party packages
   pgx "github.com/jackc/pgx/v4"
   pgxErr "github.com/jackc/pgerrcode"
   "github.com/jackc/pgconn"
   //"github.com/shopspring/decimal"
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
// "work = [work_server_address]" in embed.txt to set this value
var workServer string

// Should only be set to true in test functions
var testing = false

const MAX_INDEX = 4294967295

// Fee in %
const FEE_PERCENT = float64(0.2)
var feeDividend int64

var wg sync.WaitGroup

// TODO what happens when there is a collision? (I.E 2 keys are identical)
var activeTransactionList = make(map[string][]byte)

type psqlDB interface {
   Begin(ctx context.Context) (pgx.Tx, error)
   Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
   Query(ctx context.Context, sql string, optionsAndArgs ...interface{}) (pgx.Rows, error)
   QueryRow(ctx context.Context, sql string, optionsAndArgs ...interface{}) pgx.Row
}

var verbose bool
func main() {
   fmt.Println("Starting nanonymous Core on ", time.Now())

   err := initNanoymousCore()
   if (err != nil) {
      fmt.Println(err.Error())
      return
   }

   var usr string
   var seed keyMan.Key

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
                "P. PoW test\n",
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
         verbose = true
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
         verbose = true
         adhocAddress := "nano_1hiqiw6j9wo33moia3scoajhheweysiq5w1xjqeqt8m6jx6so6gj39pae5ea"
         blarg, _, err := getNewAddress(adhocAddress)
         if (err != nil) {
            fmt.Println(err)
         }
         fmt.Println("New address: ", blarg.NanoAddress)
      case "5":
         verbose = true
         _, err := findTotalBalance()
         if (err != nil) {
            fmt.Println("err: ", err.Error())
         }
      case "6":
         verbose = true
         seed, _ := getSeedFromIndex(1, 6)
         err := receivedNano(seed.NanoAddress)
         if (err != nil) {
            fmt.Println("err: ", err.Error())
         }

      case "7":
         keyMan.WalletVerbose(true)
         verbose = true
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
         verbose = true
         getBlockCount()
      case "A":
         verbose = true
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
         verbose = true
         err := telemetry()
         if (err != nil) {
            fmt.Println("err: ", err.Error())
         }
      case "C":
         verbose = true
         getAccountInfo("nano_1afhc54bhcqkrdwz7rwmwcexfssnwsbbkzwfhj3a7wxa17k93zbh3k4cknkb")
      case "E":
         verbose = true
         h, _ := hex.DecodeString("BCEFF64B4B62B20CAC352D77F518128FB2E46A2717CE7F40C3E04A8570211BFC")
         //block, _ := getBlockInfo(h)
         getBlockInfo(h)

      case "H":
         received, _,  err := Receive("nano_17iperkf8wx68akk66t4zynhuep7oek397nghh75oenahauch41g6pfgtrnu")
         if (err != nil) {
            fmt.Println("Error: ", err.Error())
         }
         fmt.Println("Received:", received)

      case "I":
         verbose = true
         go preCalculateNextPoW("nano_3agqwxuotzbojiyjskxr8ix59f4sx6fryt8eaptk3awgi3ktyrz7gqq1kc7p", false)
      default:
         break //menu
      case "J":
         verbose = true
         var seedSend *keyMan.Key
         var seedReceive *keyMan.Key
         seedSend, _ = getSeedFromIndex(1, 0)
         seedReceive, _ = getSeedFromIndex(1, 6)
         SendEasy(seedSend.NanoAddress,
                  seedReceive.NanoAddress,
                  keyMan.NewRawFromNano(1.5),
                  false)
         //seedSend, _ = getSeedFromIndex(1, 10)
         //seedReceive, _ = getSeedFromIndex(1, 1)
         //SendEasy(seedSend.NanoAddress,
                  //seedReceive.NanoAddress,
                  //keyMan.NewRawFromNano(0.5),
                  //false)
         //seedSend, _ = getSeedFromIndex(1, 10)
         //seedReceive, _ = getSeedFromIndex(1, 2)
         //SendEasy(seedSend.NanoAddress,
                  //seedReceive.NanoAddress,
                  //keyMan.NewRawFromNano(0.5),
                  //false)
         //time.Sleep(5 * time.Second)
         //seedSend, _ = getSeedFromIndex(1, 10)
         //seedReceive, _ = getSeedFromIndex(1, 3)
         //SendEasy(seedSend.NanoAddress,
                  //seedReceive.NanoAddress,
                  //keyMan.NewRawFromNano(0.5),
                  //true)
         //seedSend, _ = getSeedFromIndex(1, 0)
         //seedReceive, _ = getSeedFromIndex(1, 6)
         //SendEasy(seedSend.NanoAddress,
                  //seedReceive.NanoAddress,
                  //keyMan.NewRawFromNano(1.0),
                  //false)
         //seedSend, _ = getSeedFromIndex(1, 11)
         //seedReceive, _ = getSeedFromIndex(1, 0)
         //SendEasy(seedSend.NanoAddress,
                  //seedReceive.NanoAddress,
                  //keyMan.NewRawFromNano(0.5),
                  //false)
      case "K":
         verbose = true

         for i := 0; i <= 14; i++ {
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
         seedReceive, _ := getSeedFromIndex(1, 10)
         blacklist(conn, seedSend.PublicKey, seedReceive.PublicKey)

      case "M":
         verbose = true
         seed, _ := getSeedFromIndex(1, 5)
         blarg := getPendingHash(seed.NanoAddress)
         fmt.Println(blarg[seed.NanoAddress][0])
      case "N":
         verbose = true
         //fmt.Print("Seed: ")
         //fmt.Scan(&usr)
         //seed, _ := strconv.Atoi(usr)
         //fmt.Print("Index: ")
         //fmt.Scan(&usr)
         //index, _ := strconv.Atoi(usr)
         for index := 0; index <= 12; index++ {
            seedkey, _ := getSeedFromIndex(1, index)
            err := checkBalance(seedkey.NanoAddress)
            if (err != nil) {
               fmt.Println(err.Error())
            }
         }
      case "O":
         verbose = true

         seedkey, _ := getSeedFromIndex(1, 0)
         go preCalculateNextPoW(seedkey.NanoAddress, true)
         //time.Sleep(5 * time.Second)
         work := calculateNextPoW(seedkey.NanoAddress, true)

         fmt.Println("work: ", work)
      case "P":
         verbose = true

         resetInUse()
      }



   wg.Wait()
}

// initNanoymousCore sets up our variables that need to preexist before other
// functions can be called.
func initNanoymousCore() error {
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

   // TODO
   //resetInUse()
   //seedSend, _ := getSeedFromIndex(1, 0)
   //setAddressInUse(seedSend.NanoAddress)

   feeDividend = int64(math.Trunc(100/FEE_PERCENT))

   activePoW = make(map[string]int)
   workChannel = make(map[string]chan string)

   return nil
}

// inserSeed saves an encrytped version of the seed given into the database.
func insertSeed(conn *pgx.Conn, seed []byte) (int, error) {
   var id int

   queryString :=
   "INSERT INTO " +
     "seeds (seed, current_index) " +
   "VALUES " +
     "(pgp_sym_encrypt_bytea($1, $2), -1) " +
   "RETURNING id;"

   rows, err := conn.Query(context.Background(), queryString, seed, databasePassword)
   if (err != nil) {
      return -1, fmt.Errorf("insertSeed: %w", err)
   }

   if (rows.Next()) {
      err = rows.Scan(&id)
      if (err != nil) {
         return -1, fmt.Errorf("insertSeed: %w ", err)
      }
   }

   rows.Close()

   return id, nil
}

// getNewAddress finds the next availalbe address given the keys stored in the
// database and returns address B. If "receivingAddress" A is not an empty
// string, then it will also place A->B into the blacklist.
func getNewAddress(receivingAddress string) (*keyMan.Key, int, error) {
   var seed keyMan.Key

   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return nil, 0, fmt.Errorf("getNewAddress: %w", err)
   }
   defer conn.Close(context.Background())

   queryString :=
   "SELECT " +
      "id, " +
      "pgp_sym_decrypt_bytea(seed, $1), " +
      "current_index " +
   "FROM " +
      "seeds " +
   "WHERE " +
      "current_index < $2" +
   "ORDER BY " +
      "id;"

   // TODO start a tranasction and increment current_index. Only commit after everthing checks out
   rows, err := conn.Query(context.Background(), queryString, databasePassword, MAX_INDEX)
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

      id, err = insertSeed(conn, seed.Seed)
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
   rowsAffected, err := conn.Exec(context.Background(), queryString, id, seed.Index, hash[:])
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

   rowsAffected, err = conn.Exec(context.Background(), queryString, seed.Index, id)
   if (err != nil) {
      return nil, 0, fmt.Errorf("getNewAddress: %w", err)
   }
   if (rowsAffected.RowsAffected() < 1) {
      return nil, 0, fmt.Errorf("getNewAddress: no rows affected during index incrament")
   }

   // blacklist new addres with the receiving address
   if (receivingAddress != "") {
      receivingAddressByte, err := keyMan.AddressToPubKey(receivingAddress)
      if (err != nil) {
         return nil, 0, fmt.Errorf("getNewAddress: %w", err)
      }

      err = blacklist(conn, seed.PublicKey, receivingAddressByte)
      if (err != nil) {
         return nil, 0, fmt.Errorf("getNewAddress: Blacklist falied: %w", err)
      }

      setClientAddress(id, seed.Index, receivingAddressByte)
   }

   return &seed, id, nil
}

// blacklist takes two public addresses, hashes them and stores them in the
// database. The purpose of the blacklist is to securely store a transaction
// pair that should never happen. If, for example, Alice sends nano from address
// A to address B in order to receive it ananomously at another address C, then
// Alice wants to be sure that her address A and address C are never associated.
// However, if she later orders another transaction to address C, nanonymous
// would run the risk of using address B to send to C. Before doing such a
// transaction, nanonymous regenerates the blacklist hash and checks the
// blacklist. If it doesn't exist, then we can be sure that there will be no
// unintentional associations.
func blacklist(conn *pgx.Conn, sendingAddress []byte, receivingAddress []byte) error {

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
func blacklistHash(sendingAddress []byte, receivingHash keyMan.BlockHash) error {
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
   var foundEntry bool
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return fmt.Errorf("receivedNano: %w", err)
   }
   defer conn.Close(context.Background())

   parentSeedId, index, err := getWalletFromAddress(nanoAddress)
   if (err != nil) {
      return fmt.Errorf("receivedNano: %w", err)
   }

   payment, receiveHash, err := Receive(nanoAddress)
   if (err != nil) {
      return fmt.Errorf("receivedNano: %w", err)
   }
   if (payment == nil || payment.Cmp(keyMan.NewRaw(0)) == 0) {
      return fmt.Errorf("receivedNano: No payment received")
   }
   fmt.Println(" receivehash: ", receiveHash)

   // If anything goes wrong the transactionManager will make sure to clean up
   // the mess.
   var t Transaction
   go transactionManager(&t)
   t.commChannel = make(chan int)
   t.errChannel = make(chan error)
   defer func () {
      if (err != nil) {
         t.errChannel <- err
      }
   }()

   t.paymentAddress, err = keyMan.AddressToPubKey(nanoAddress)
   if (err != nil) {
      err = fmt.Errorf("receivedNano: %w", err)
      return err
   }
   t.receiveHash = receiveHash
   fmt.Println(" receivehash: ", t.receiveHash.String())

   // TODO This is just for debugging
   seed, _ := getSeedFromIndex(1, 10)
   clientPub, _ := keyMan.AddressToPubKey(seed.NanoAddress)
   setClientAddress(parentSeedId, index, clientPub)
   // TODO end of debugging code

   // Get client address for later use.
   t.clientAddress = getClientAddress(parentSeedId, index)
   if (t.clientAddress == nil) {
      err = fmt.Errorf("receivedNano: no active transaction available")
      return err
   }

   t.fee = keyMan.NewRaw(0).Div(payment, keyMan.NewRaw(feeDividend))
   t.amountToSend = keyMan.NewRaw(0).Sub(payment, t.fee)
   if (verbose) {
      fmt.Println("payment:        ", payment,
                  "\r\nfee:            ", t.fee,
                  "\r\namount to send: ", t.amountToSend)
   }

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
   rows, err = conn.Query(context.Background(), queryString, t.amountToSend, parentSeedId, index)
   if (err != nil) {
      err = fmt.Errorf("receiviedNano: Query: %w", err)
      return err
   }

   var foundAddress bool
   var tmpSeed int
   var tmpIndex int
   var tmpBalance = keyMan.NewRaw(0)
   for rows.Next() {
      err = rows.Scan(&tmpSeed, &tmpIndex, tmpBalance)
      if (err != nil) {
         err = fmt.Errorf("receivedNano: Scan: %w", err)
         return err
      }

      // Check the blacklist before accepting
      var tmpKey *keyMan.Key
      foundEntry, tmpKey, err = checkBlackList(tmpSeed, tmpIndex, t.clientAddress)
      if (err != nil) {
         err = fmt.Errorf("receivedNano: %w", err)
         return err
      }
      if (!foundEntry) {
         // Uset this address
         t.sendingKeys = append(t.sendingKeys, tmpKey)
         t.walletSeed = append(t.walletSeed, tmpSeed)
         t.walletBalance = append(t.walletBalance, tmpBalance)
         setAddressInUse(tmpKey.NanoAddress)
         foundAddress = true
         if (verbose) {
            fmt.Println("sending from:", tmpSeed, tmpIndex)
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

      rows, err := conn.Query(context.Background(), queryString, parentSeedId, index)
      if (err != nil) {
         err = fmt.Errorf("receiviedNano: Query: %w", err)
         return err
      }

      var enough bool
      var totalBalance = keyMan.NewRaw(0)
      for rows.Next() {
         err = rows.Scan(&tmpSeed, &tmpIndex, tmpBalance)
         if (err != nil) {
            err = fmt.Errorf("receivedNano: Scan(2): %w", err)
            return err
         }

         // Check the blacklist before adding to the list
         var tmpKey *keyMan.Key
         foundEntry, tmpKey, err = checkBlackList(tmpSeed, tmpIndex, t.clientAddress)
         if (err != nil) {
            err = fmt.Errorf("receivedNano: %w", err)
            return err
         }
         if (!foundEntry) {
            t.sendingKeys = append(t.sendingKeys, tmpKey)
            t.walletSeed = append(t.walletSeed, tmpSeed)
            // tmpBalance contains a pointer, so we need a new address to add to the list
            newAddress := keyMan.NewFromRaw(tmpBalance)
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
         err = fmt.Errorf("receivedNano: not enough funds")
         return err
      }
   }

   // Send nano to client
   if (len(t.sendingKeys) == 1) {
      go Send(t.sendingKeys[0], t.clientAddress, t.amountToSend, t.commChannel, t.errChannel, 0)
      t.commChannel <- 1
   } else if (len(t.sendingKeys) > 1) {
      // Need to do a multi-send; Get a new wallet to combine all funds into
      t.transitionalKey, t.transitionSeedId, err = getNewAddress("")
      if (err != nil) {
         err = fmt.Errorf("receivedNano: %w", err)
         return err
      }
      t.multiSend = true

      // Go through list of wallets and send to interim address
      var totalSent = keyMan.NewRaw(0)
      var currentSend = keyMan.NewRaw(0)
      for i, key := range t.sendingKeys {

         // if (total + balance) > payment
         var arithmaticResult = keyMan.NewRaw(0)
         if (arithmaticResult.Add(totalSent, t.walletBalance[i]).Cmp(t.amountToSend) > 0) {
            currentSend = arithmaticResult.Sub(t.amountToSend, totalSent)
         } else {
            currentSend = t.walletBalance[i]
         }
         t.multiSendAmount = append(t.multiSendAmount, currentSend)
         go Send(key, t.transitionalKey.PublicKey, currentSend, t.commChannel, t.errChannel, i)
         if (i == 0) {
            t.commChannel <- 1
         }
         totalSent.Add(totalSent, currentSend)
         if (verbose) {
            fmt.Println("Sending", currentSend.Int, "from", t.walletSeed[i], key.Index, "to", t.transitionSeedId, t.transitionalKey.Index)
         }
      }

      // Now send to client
      if (verbose) {
         fmt.Println("Sending", t.amountToSend, "from", t.transitionSeedId, t.transitionalKey.Index, "to client.")
      }
      go ReceiveAndSend(t.transitionalKey, t.clientAddress, t.amountToSend, t.commChannel, t.errChannel, &t.receiveWg, &t.abort)
   } else {
      err = fmt.Errorf("receivedNano: not enough funds(2)")
      return err
   }

   return nil
}

// findTotalBalace is a simple function that adds up all the nano there is
// amongst all the wallets and returns the amount in Nano.
func findTotalBalance() (float64, error) {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return -1.0, fmt.Errorf("FindTotalBalance: %w", err)
   }
   defer conn.Close(context.Background())

   queryString :=
   "SELECT " +
      "SUM(balance) " +
   "FROM " +
      "wallets;"

   var rawBalance = keyMan.NewRaw(0)
   var nanoBalance float64
   row, err := conn.Query(context.Background(), queryString)
   if (err != nil) {
      return -1.0, fmt.Errorf("QueryRow failed: %w", err)
   }

   if (row.Next()) {
      err = row.Scan(rawBalance)
      if (err != nil) {
         return -1.0, fmt.Errorf("findTotalBalance: %w", err)
      }

      nanoBalance = rawToNANO(rawBalance)

      if (verbose) {
         fmt.Println("Total Balance is: Ӿ", nanoBalance)
      }
   }

   return nanoBalance, nil
}

// rawToNANO is used to convert raw to NANO AKA Mnano (the communnity just calls
// this a nano). We don't have a conversion to go the other way as all
// operations should be done in raw to avoid rounding errors. We only want to
// convert when outputing for human readable format.
func rawToNANO(raw *keyMan.Raw) float64 {
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

   if (verbose) {
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

func setClientAddress(parentSeedId int, index int, clientAddress []byte) {
   key := strconv.Itoa(parentSeedId) + "-" + strconv.Itoa(index)
   activeTransactionList[key] = clientAddress
}

func ReceiveAndSend(transitionalKey *keyMan.Key, toPublicKey []byte, amount *keyMan.Raw, commCh chan int, errCh chan error, transactionWg *sync.WaitGroup, abort *bool) {
   wg.Add(1)
   defer wg.Done()

   fmt.Println(" --------- 1 ---------")

   // Wait until all sends have processed
   transactionWg.Wait()
   if (*abort) {
      return
   }
   transactionWg.Add(1)

   fmt.Println(" --------- 2 ---------")

   // TODO wait for blocks to be confirmed
   fmt.Println("Waiting for blocks...")
   time.Sleep(10 * time.Second)
   fmt.Println(" --------- 3 ---------")

   // All transactions are still pending. Receive the funds.
   err := ReceiveAll(transitionalKey.NanoAddress)
   if (err != nil) {
      err = fmt.Errorf("ReceiveAndSend: %w", err)
      errCh <- err
   } else {
      commCh <- 2
   }
   fmt.Println(" --------- 4 ---------")

   // Wait until all receives have processed
   transactionWg.Wait()
   if (*abort) {
      return
   }
   transactionWg.Add(1)

   // TODO wait for blocks to be confirmed
   fmt.Println("Waiting for blocks(2)...")
   time.Sleep(10 * time.Second)
   fmt.Println(" --------- 5 ---------")

   // Finally, send to client.
   err = Send(transitionalKey, toPublicKey, amount, nil, nil, -1)
   if (err != nil) {
      err = fmt.Errorf("ReceiveAndSend: %w", err)
      errCh <- err
   } else {
      commCh <- 3
   }
   fmt.Println(" --------- 6 ---------")
}

func Send(fromKey *keyMan.Key, toPublicKey []byte, amount *keyMan.Raw, commCh chan int, errCh chan error, i int) error {

   err := sendNano(fromKey, toPublicKey, amount)
   if (err != nil) {
      if (errCh != nil) {
         if (verbose) {
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
func SendEasy(from string, to string, amount *keyMan.Raw, all bool) {

   fromKey, _, _, err := getSeedFromAddress(from)
   toKey, _, _, err := getSeedFromAddress(to)
   if (err != nil) {
      toKey.PublicKey, _ = keyMan.AddressToPubKey(to)
   }

   if (all) {
      err = sendAllNano(&fromKey, toKey.PublicKey)
   } else {
      err = Send(&fromKey, toKey.PublicKey, amount, nil, nil, -1)
      if (err != nil) {
         fmt.Println("Error: ", err.Error())
      }
   }
}

func sendNano(fromKey *keyMan.Key, toPublicKey []byte, amountToSend *keyMan.Raw) error {
   var block keyMan.Block

   accountInfo, err := getAccountInfo(fromKey.NanoAddress)
   if (err != nil) {
      return fmt.Errorf("sendNano: %w", err)
   }

   // if (Balance < amountToSend)
   if (accountInfo.Balance.Cmp(amountToSend) < 0) {
      return fmt.Errorf("sendNano: not enough funds in account.\r\n have: %s\r\n need: %s", accountInfo.Balance, amountToSend)
   }

   // Create send block
   block.Previous = accountInfo.Frontier
   block.Seed = *fromKey
   block.Account = block.Seed.NanoAddress
   block.Representative = accountInfo.Representative
   block.Balance = keyMan.NewRaw(0).Sub(accountInfo.Balance, amountToSend)
   block.Link = toPublicKey

   sig, err := block.Sign()
   if (err != nil) {
      return fmt.Errorf("sendNano: %w", err)
   }

   if (verbose) {
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
   if (block.Balance.Cmp(keyMan.NewRaw(0)) != 0) {
      go preCalculateNextPoW(block.Account, false)
   }

   // Update database records
   err = updateBalance(block.Account, block.Balance)
   if (err != nil) {
      // TODO log
      return fmt.Errorf("sendNano: updatebalance error %w", databaseError)
      // TODO we don't want to throw an error here because that would make us
      //      refund the client after we already sent his funds
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

   amount := keyMan.NewRaw(0).Mul(keyMan.NewRaw(nano), keyMan.NewRaw(0).Exp(keyMan.NewRaw(10), keyMan.NewRaw(30), nil))

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
      if amt, _, err := Receive(account); amt != nil {
         if (err != nil) {
            return fmt.Errorf("ReceiveAll: %w", err)
            break
         }
         if (amt.Cmp(keyMan.NewRaw(0)) == 0) {
            break
         }
      } else {
         break
      }
   }

   return nil
}

func Receive(account string) (*keyMan.Raw, keyMan.BlockHash, error) {
   var block keyMan.Block
   var pendingInfo BlockInfo
   var newHash keyMan.BlockHash

   key, _, _, err := getSeedFromAddress(account)
   if (err != nil) {
      return nil, nil, fmt.Errorf("receive: %w", err)
   }

   pendingHashes := getPendingHash(account)
   doesPendingExist := len(pendingHashes[account])

   fmt.Println("hashes: ", pendingHashes, "\r\nNum: ", doesPendingExist)

   if (doesPendingExist > 0 ) {
      pendingHash := pendingHashes[account][0]
      pendingInfo, _ = getBlockInfo(pendingHash)
      accountInfo, err := getAccountInfo(account)
      if (err != nil) {
         return nil, nil, fmt.Errorf("Receive: %w", err)
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
         block.Balance = keyMan.NewRaw(0).Add(accountInfo.Balance, pendingInfo.Amount)
      }
      block.Account = account
      block.Link = pendingHash
      block.Seed = key

      sig, err := block.Sign()
      if (err != nil) {
         return nil, nil, fmt.Errorf("receive: %w", err)
      }

      if (verbose) {
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
         return nil, nil, fmt.Errorf("receive: %w", err)
      }

      // Send RCP request
      newHash, err = publishReceive(block, sig, PoW)
      if (err != nil){
         return nil, nil, fmt.Errorf("receive: %w", err)
      }
      if (len(newHash) != 32){
         return nil, nil, fmt.Errorf("receive: no block hash returned from node", err)
      }

   fmt.Println("newHash: ", newHash)

      // We'ved used any stored PoW, clear it out for next use
      clearPoW(block.Account)
      go preCalculateNextPoW(block.Account, false)

      // Update database records
      err = updateBalance(block.Account, block.Balance)
      if (err != nil) {
         // TODO log
         return pendingInfo.Amount, newHash, fmt.Errorf("sendNano: updatebalance error %w", databaseError)
         // TODO we don't want to throw an error here because that would make us
         //      refund the client after we already sent his funds
      }
   }

   return pendingInfo.Amount, newHash, err
}

func getNewRepresentative() string {
   return  "nano_1s3dw5dn1m74hm73wxj96i5eouigp1w7nesw83tjo8kchrx8t6ekaymp6dgs"
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
   // TODO receiveblocks go to node?

   // Check if PoW is already being calculated
   if (activePoW[nanoAddress] == 0) {
      // PoW is not being calculated, so do it!
      activePoW[nanoAddress] = 1

      var hash keyMan.BlockHash
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
         // TODO log error
         // Fall back server
         work, err = generateWorkOnNode(hash, difficulty)
         if (err != nil) {
            return ""
         }
      }
      if (verbose) {
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
            if (verbose) {
               fmt.Println("Got", work, "from channel")
            }
            return work
         case <-time.After(5 * time.Minute):
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

   if (receiveable.Cmp(keyMan.NewRaw(0)) != 0) {
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
