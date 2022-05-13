package main

import (
   "fmt"
   "time"
   "context"
   "encoding/hex"
   "encoding/json"
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

   //menu:
   //for {
      fmt.Print("1. Generate Seed\n",
                "2. Database test\n",
                "3. Insert into database\n",
                "4. Send pretend request for new address\n",
                "5. Find total balance\n",
                "6. Pretend nano receive\n",
                "7. Get Wallet Info\n",
                "8. Add nano to wallet\n",
                "9. Get block count\n",
                "A. Peers\n",
                "B. Telemetry\n",
                "C. Get Account Info\n",
                "D. Sign Block\n",
                "E. Block Info\n",
                "H. OpenAccount\n",
                "I. GenerateWork\n",
                "J. Send\n",
                "K. Recive\n",
                "L. ?\n",
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
         conn, err := pgx.Connect(context.Background(), databaseUrl)
         if (err != nil) {
            fmt.Println("main: ", err)
            return
         }
         defer conn.Close(context.Background())

         var parent_seed int
         var index int
         var balance = keyMan.NewRaw(0)
         var hash keyMan.HexData

         rows, err := conn.Query(context.Background(), "SELECT * FROM wallets")

         if (err != nil) {
            fmt.Println("QueryRow failed: ", err)
            return
         }

         for rows.Next() {
            err = rows.Scan(&parent_seed, &index, balance, &hash)
            if (err != nil) {
               fmt.Println("Scan failed: ", err)
               return
            }
            fmt.Println("parent: ", parent_seed, "index: ", index, "balance: ", balance.String())
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
         seed, _ := getSeedFromIndex(1, 1)
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
         printPeers()
      case "B":
         telemetry()
      case "C":
         verbose = true
         getAccountInfo("nano_1afhc54bhcqkrdwz7rwmwcexfssnwsbbkzwfhj3a7wxa17k93zbh3k4cknkb")
      case "E":
         verbose = true
         h, _ := hex.DecodeString("BCEFF64B4B62B20CAC352D77F518128FB2E46A2717CE7F40C3E04A8570211BFC")
         //block, _ := getBlockInfo(h)
         getBlockInfo(h)

      case "H":
         pending, received, err := Receive("nano_17iperkf8wx68akk66t4zynhuep7oek397nghh75oenahauch41g6pfgtrnu")
         if (err != nil) {
            fmt.Println("Error: ", err.Error())
         }
         fmt.Println("Received:", received, ". Pending after transaction: ", pending)

      case "I":
         verbose = true
         go preCalculateNextPoW("nano_3agqwxuotzbojiyjskxr8ix59f4sx6fryt8eaptk3awgi3ktyrz7gqq1kc7p")
      default:
         break //menu
      case "J":
         seedSend, _ := getSeedFromIndex(1, 3)
         seedReceive, _ := getSeedFromIndex(1, 1)
         verbose = true
         SendEasy(seedSend.NanoAddress,
                  seedReceive.NanoAddress,
                  keyMan.NewRawFromNano(1.0))
      case "K":
         //var b keyMan.BlockHash
         //b = make([]byte, 32)
//
         //fmt.Println("prev: ", hex.EncodeToString(b))

         seedReceive, _ := getSeedFromIndex(1, 5)
         verbose = true
         _, _, err := Receive(seedReceive.NanoAddress)
         if (err != nil) {
            fmt.Println("Error:", err.Error())
         }
      case "L":
         jsonText := `{"timestamp": "1652378161"}`
         response := struct {
            TimeStamp keyMan.JInt
         }{}

         err := json.Unmarshal([]byte(jsonText), &response)
         if (err != nil) {
            fmt.Println("Error: ", err.Error())
         }

         fmt.Println("timestamp", response.TimeStamp)
      }
   //}
   // TODO work only needs frontier to generate next PoW, or if it's an open block, the pubKey

   // work for next nano_3pwsg1enmhf77ai6d87fppu9rfjua5qgshya4muduoo879b87fjwyy8oxr33: 7d00f6d81793ccdd

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

   feeDividend = int64(math.Trunc(100/FEE_PERCENT))

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
   go preCalculateNextPoW(seed.NanoAddress)

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

// receivedNano is a large function that does most of the work for nanonymous.
// Upon receiving nano it does 5 distinct things:
//    (1) Updates the database with the newly recived nano
//    (2) Checks if we were expecting the tranaction
//    (3) Calculates the fee
//    (4) Finds the wallet(s) with enough funds to support the transaction
//        (minus the blacklisted ones)
//    (5) Sends the funds to the client
func receivedNano(nanoAddress string) error {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return fmt.Errorf("receivedNano: %w", err)
   }
   defer conn.Close(context.Background())

   parentSeed, index, err := getWalletFromAddress(nanoAddress)
   if (err != nil) {
      return fmt.Errorf("receivedNano: %w", err)
   }

   _, payment, err := Receive(nanoAddress)
   if (err != nil) {
      return fmt.Errorf("receivedNano: %w", err)
   }
   if (payment == nil || payment.Cmp(keyMan.NewRaw(0)) == 0) {
      return fmt.Errorf("receivedNano: No payment received")
   }

   // TODO This is just for debugging
   clientPub, _ := keyMan.AddressToPubKey("nano_3sani18ec6jj1z4p74436r9h6odxwh8fu115xgxaf3ujksi8tzyuqbk3ebgo")
   setClientAddress(parentSeed, index, clientPub)
   // TODO end of debugging code

   // Get client address for later use. TODO check for nil
   clientAddress := getClientAddress(parentSeed, index)
   if (clientAddress == nil) {
      // No active transaction, send the funds back to owner
      // TODO sendNano()
      return fmt.Errorf("receivedNano: no active transaction available")
   }

   fee := keyMan.NewRaw(0).Div(payment, keyMan.NewRaw(feeDividend))
   amountToSend := keyMan.NewRaw(0).Sub(payment, fee)
   if (verbose) {
      fmt.Println("payment:        ", payment,
                  "\r\nfee:            ", fee,
                  "\r\namount to send: ", amountToSend)
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

   rows, err := conn.Query(context.Background(), queryString, amountToSend, parentSeed, index)
   if (err != nil) {
      return fmt.Errorf("receiviedNano: Query: %w", err)
   }

   var foundAddress bool
   var sendingKeys []*keyMan.Key
   var walletSeed []int
   var walletBalance []*keyMan.Raw
   var tmpSeed int
   var tmpIndex int
   var tmpBalance = keyMan.NewRaw(0)
   for rows.Next() {
      err = rows.Scan(&tmpSeed, &tmpIndex, tmpBalance)
      if (err != nil) {
         return fmt.Errorf("receivedNano: Scan: %w", err)
      }

      // Check the blacklist before accepting
      foundEntry, tmpKey, err := checkBlackList(tmpSeed, tmpIndex, clientAddress)
      if (err != nil) {
         return fmt.Errorf("receivedNano: %w", err)
      }
      if (!foundEntry) {
         // Uset this address
         sendingKeys = append(sendingKeys, tmpKey)
         walletSeed = append(walletSeed, tmpSeed)
         walletBalance = append(walletBalance, tmpBalance)
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

      rows, err := conn.Query(context.Background(), queryString, parentSeed, index)
      if (err != nil) {
         return fmt.Errorf("receiviedNano: Query: %w", err)
      }

      var enough bool
      var totalBalance = keyMan.NewRaw(0)
      for rows.Next() {
         err = rows.Scan(&tmpSeed, &tmpIndex, tmpBalance)
         if (err != nil) {
            return fmt.Errorf("receivedNano: Scan(2): %w", err)
         }

         // Check the blacklist before adding to the list
         foundEntry, tmpKey, err := checkBlackList(tmpSeed, tmpIndex, clientAddress)
         if (err != nil) {
            return fmt.Errorf("receivedNano: %w", err)
         }
         if (!foundEntry) {
            sendingKeys = append(sendingKeys, tmpKey)
            walletSeed = append(walletSeed, tmpSeed)
            // tmpBalance contains a pointer, so we need a new address to add to the list
            newAddress := keyMan.NewFromRaw(tmpBalance)
            walletBalance = append(walletBalance, newAddress)
            totalBalance.Add(totalBalance, tmpBalance)
            if (totalBalance.Cmp(amountToSend) >= 0) {
               // We've found enough
               enough = true
               break
            }
         }
      }
      rows.Close()
      if (!enough) {
         return fmt.Errorf("receivedNano: not enough funds")
      }
   }

   // Send nano to client
   if (len(sendingKeys) == 1) {
      Send(sendingKeys[0], clientAddress, amountToSend)
   } else if (len(sendingKeys) > 1) {
      // Need to do a multi-send; Get a new wallet to combine all funds into
      transitionalAddress, transitionSeedId, err := getNewAddress("")
      if (err != nil) {
         return fmt.Errorf("receivedNano: %w", err)
      }

      // Go through list of wallets and send to interim address
      var totalSent = keyMan.NewRaw(0)
      var currentSend = keyMan.NewRaw(0)
      for i, key := range sendingKeys {

         // if (total + balance) > payment
         var arithmaticResult = keyMan.NewRaw(0)
         if (arithmaticResult.Add(totalSent, walletBalance[i]).Cmp(amountToSend) > 0) {
            currentSend = arithmaticResult.Sub(amountToSend, totalSent)
         } else {
            currentSend = walletBalance[i]
         }
         Send(key, transitionalAddress.PublicKey, currentSend)
         totalSent.Add(totalSent, currentSend)
         if (verbose) {
            fmt.Println("Sending", currentSend.Int, "from", walletSeed[i], key.Index, "to", transitionSeedId, transitionalAddress.Index)
         }
      }

      // Now send to client
      if (verbose) {
         fmt.Println("Sending", amountToSend, "from", transitionSeedId, transitionalAddress.Index, "to client.")
      }
      go ReceiveAndSend(transitionalAddress, clientAddress, amountToSend, transitionSeedId, transitionalAddress.Index, 0, 0, conn)
   } else {
      return fmt.Errorf("receivedNano: not enough funds(2)")
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
func checkBlackList(parentSeed int, index int, clientAddress []byte) (bool, *keyMan.Key, error) {
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
   row, err := conn.Query(context.Background(), queryString, databasePassword, parentSeed)
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
      return true, nil, fmt.Errorf("checkBlacklist: No such seed found: %d", parentSeed)
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

func getClientAddress(parentSeed int, index int) []byte {
   key := strconv.Itoa(parentSeed) + strconv.Itoa(index)
   return activeTransactionList[key]
}

func setClientAddress(parentSeed int, index int, clientAddress []byte) {
   key := strconv.Itoa(parentSeed) + strconv.Itoa(index)
   activeTransactionList[key] = clientAddress
}

func ReceiveAndSend(transitionalKey *keyMan.Key, toPublicKey []byte, amount *keyMan.Raw, fromSeed int, fromIndex int, toSeed int, toIndex int, conn psqlDB) {
   wg.Add(1)
   defer wg.Done()

   // All transactions are still pending. Receive the funds.
   ReceiveAll(transitionalKey.NanoAddress)

   // Finally, send to client.
   Send(transitionalKey, toPublicKey, amount)
}

func Send(fromKey *keyMan.Key, toPublicKey []byte, amount *keyMan.Raw) error {

   err := sendNano(fromKey, toPublicKey, amount)
   if (err != nil) {
      return fmt.Errorf("Send: %w", err)
   }

   setAddressNotInUse(fromKey.NanoAddress)

   return nil
}

// SendEasy is just a wrapper for Send(). If you have the info already Send()
// is more efficient so don't overuse this function.
func SendEasy(from string, to string, amount *keyMan.Raw) {

   fromKey, _, _, err := getSeedFromAddress(from)
   toKey, _, _, err := getSeedFromAddress(to)
   if (err != nil) {
      toKey.PublicKey, _ = keyMan.AddressToPubKey(to)
   }

   err = Send(&fromKey, toKey.PublicKey, amount)
   if (err != nil) {
      fmt.Println("Error: ", err.Error())
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

   PoW, err := getPoW(block.Account)
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
   go preCalculateNextPoW(block.Account)

   // Update database records
   err = updateBalance(block.Account, block.Balance)
   if (err != nil) {
      return fmt.Errorf("sendNano: %w", err)
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
      if num, _, err := Receive(account); num <= 0 {
         if (err != nil) {
            return fmt.Errorf("ReceiveAll: %w", err)
         } else {
            break
         }
      }
   }

   return nil
}

func Receive(account string) (int, *keyMan.Raw, error) {
   var block keyMan.Block
   var pendingInfo BlockInfo

   key, _, _, err := getSeedFromAddress(account)
   if (err != nil) {
      return -1, nil, fmt.Errorf("receive: %w", err)
   }

   pendingHashes := getPendingHash(account)
   numPendingTransactions := len(pendingHashes[account])

   if (numPendingTransactions > 0 ) {
      pendingHash := pendingHashes[account][0]
      pendingInfo, _ = getBlockInfo(pendingHash)
      accountInfo, err := getAccountInfo(account)
      if (err != nil) {
         return -1, nil, fmt.Errorf("Receive: %w", err)
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
         return -1, nil, fmt.Errorf("receive: %w", err)
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

      PoW, err := getPoW(block.Account)
      if (err != nil) {
         return -1, nil, fmt.Errorf("receive: %w", err)
      }

      // Send RCP request
      newHash, err := publishReceive(block, sig, PoW)
      if (err != nil){
         return -1, nil, fmt.Errorf("receive: %w", err)
      }
      if (len(newHash) != 32){
         return -1, nil, fmt.Errorf("receive: no block hash returned from node", err)
      }

      // We'ved used any stored PoW, clear it out for next use
      clearPoW(block.Account)
      go preCalculateNextPoW(block.Account)

      // Update database records
      err = updateBalance(block.Account, block.Balance)
   }

   return numPendingTransactions - 1, pendingInfo.Amount, err
}

func getNewRepresentative() string {
   return  "nano_1s3dw5dn1m74hm73wxj96i5eouigp1w7nesw83tjo8kchrx8t6ekaymp6dgs"
}

func preCalculateNextPoW(nanoAddress string) {
   wg.Add(1)
   defer wg.Done()
   if (testing) {
      return
   }

   var hash keyMan.BlockHash
   accountInfo, _ := getAccountInfo(nanoAddress)

   if (len(accountInfo.Frontier) == 32) {
      hash = accountInfo.Frontier
   } else {
      // New account, just use pubKey
      hash, _ = keyMan.AddressToPubKey(nanoAddress)
   }

   work, err := generateWorkOnWorkServer(hash)
   if (err != nil) {
      // TODO log error
      // Fall back server
      work, err = generateWorkOnNode(hash)
      if (err != nil) {
         return
      }
   }


   if (verbose) {
      fmt.Println("Work generated for address", nanoAddress, ":", work)
   }

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

func calculateNextPoW(nanoAddress string, addressHash []byte) string {
   if (testing) {
      return ""
   }

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

   work, err := generateWorkOnWorkServer(hash)
   if (err != nil) {
      // TODO log error
      // Fall back server
      work, err = generateWorkOnNode(hash)
      if (err != nil) {
         return ""
      }
   }
   if (verbose) {
      fmt.Println("Work generated for address", nanoAddress, ":", work)
   }

   return work
}

func getPoW(nanoAddress string) (string, error) {
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
      PoW = calculateNextPoW(nanoAddress, nanoAddressHash[:])
   }

   return PoW, nil
}
