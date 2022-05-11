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
                "H. OpenAccount 0, 1\n",
                "I. GenerateWork\n",
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
         _, err := findTotalBalance()
         if (err != nil) {
            fmt.Println("err: ", err.Error())
         }
      case "6":
         adhocAddress := []string{
            "nano_3agqwxuotzbojiyjskxr8ix59f4sx6fryt8eaptk3awgi3ktyrz7gqq1kc7p",
            "lkjlkj",
         }
         nanoRecieved := keyMan.NewRaw(0).Exp(keyMan.NewRaw(10), keyMan.NewRaw(30), nil)
         err := receivedNano(adhocAddress[0], nanoRecieved)
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

         _, err := getWalletInfo(seed, index)
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
         getAccountInfo("nano_1s3dw5dn1m74hm73wxj96i5eouigp1w7nesw83tjo8kchrx8t6ekaymp6dgs")
      case "E":
         h, _ := hex.DecodeString("A77C8B3CAE8DAEEDC2A2C9A52778175A1621AAB4EA6C3766F440C9712D45B4BA")
         block, _ := getBlockInfo(h)

         fmt.Println("balance:", block.Balance)
         fmt.Println("linke account:", block.LinkAsAccount)

         block.Balance.Sub(block.Balance, keyMan.NewRaw(72))
         fmt.Println("New balance:", block.Balance)
         block.Previous = h

         hash, _ := block.Hash()
         fmt.Println("New balance:", hash)

         s, _ := hex.DecodeString("6698E42FFDD05E83212CDEBED6FE0B021DEA7C865C436DC566D4E43214F2A89E")
         block.Seed.Seed = s
         keyMan.SeedToKeys(&block.Seed)

         fmt.Println("private key len:", len(block.Seed.PrivateKey))
         fmt.Println("private key:", strings.ToUpper(hex.EncodeToString(block.Seed.PrivateKey)))
         fmt.Println("hash:", strings.ToUpper(hex.EncodeToString(hash)))

         sig, err := block.Sign()
         if (err != nil) {
            fmt.Println("Error: %s", err.Error())
         }
         fmt.Println("Signed: ", strings.ToUpper(hex.EncodeToString(sig)))
         fmt.Println("3rd   :  5794D0B497EF171F036A43661ECC08FB65184F418BAA8C1F70B4AF35582F355E8BF9BA7C13BE2C47163F6CE9C62C974FC99E197EF21C116C8060143073255C07")

      case "H":
         if (verbose) {
            fmt.Println(" verbose == true")
         }
         pending, err := OpenAccount("nano_3j7pzwsg999gpx99fm4xcwndfdpcsyxu91ed5nq6dmpdots6nfj1hkp8ahsk")
         if (err != nil) {
            fmt.Println("Error: ", err.Error())
         }
         fmt.Println("Pending after transaction: ", pending)

      case "I":
         verbose = true
         go preCalculateNextPoW("nano_3j7pzwsg999gpx99fm4xcwndfdpcsyxu91ed5nq6dmpdots6nfj1hkp8ahsk")

      default:
         break //menu
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
func receivedNano(nanoAddress string, payment *keyMan.Raw) error {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return fmt.Errorf("receivedNano: %w", err)
   }
   defer conn.Close(context.Background())

   tx, _ := conn.BeginTx(context.Background(), pgx.TxOptions{})
   defer tx.Rollback(context.Background())

   // Find which address just received funds based on the hash
   queryString :=
   "SELECT " +
      "parent_seed, " +
      "index " +
   "FROM " +
      "wallets " +
   "WHERE " +
      "hash = $1;"

   pubkey, err := keyMan.AddressToPubKey(nanoAddress)
   if (err != nil) {
      return fmt.Errorf("receivedNano: %w", err)
   }

   recivedHash := blake2b.Sum256(pubkey)

   row, err := tx.Query(context.Background(), queryString, recivedHash[:])
   if (err != nil) {
      return fmt.Errorf("receiviedNano: %w", err)
   }

   var parentSeed int
   var index int
   if (row.Next()) {
      err = row.Scan(&parentSeed, &index)
      row.Close()
      if (err != nil) {
         return fmt.Errorf("receivedNano: %w", err)
      }
   } else {
      return fmt.Errorf("receivedNano: address not found in active wallets")
   }

   // TODO This is just for debugging
   //clientPub, _ := keyMan.AddressToPubKey("nano_36uqf39z8nydejhehihtkopyd8hjouqi7su9ccxw85dwft3mtm15myzgz3mx")
   //setClientAddress(parentSeed, index, clientPub)
   // TODO end of debugging code

   // Get client address for later use. TODO check for nil
   clientAddress := getClientAddress(parentSeed, index)
   if (clientAddress == nil) {
      // No active transaction, send the funds back to owner
      // TODO sendNano()
      return fmt.Errorf("receivedNano: no active transaction available")
   }

   // Add funds we got into our database of wallets
   queryString =
   "UPDATE " +
      "wallets "+
   "SET " +
      "\"balance\" = \"balance\" + $1 " +
   "WHERE " +
      "\"parent_seed\" = $2 AND " +
      "\"index\" = $3;"

   rowsAffected, err := tx.Exec(context.Background(), queryString, payment, parentSeed, index)
   if (err != nil) {
      return fmt.Errorf("receivedNano: Update: %w", err)
   }
   if (rowsAffected.RowsAffected() < 1) {
      return fmt.Errorf("receivedNano: no rows affected during index incrament")
   }

   fee := keyMan.NewRaw(0).Div(payment, keyMan.NewRaw(feeDividend))
   amountToSend := keyMan.NewRaw(0).Sub(payment, fee)
   if (verbose) {
      fmt.Println("payment:        ", payment, "\r\nfee:            ", fee, "\r\namount to send: ", amountToSend)
   }

   // Find all wallets that have enough funds to send out the payment that
   // aren't the wallet we just received in.
   queryString =
   "SELECT " +
      "parent_seed, " +
      "index, " +
      "balance " +
   "FROM " +
      "wallets " +
   "WHERE " +
      "balance >= $1 AND NOT" +
      "( parent_seed = $2 AND " +
      "  index = $3 )" +
   "ORDER BY " +
      "balance, " +
      "index;"

   rows, err := tx.Query(context.Background(), queryString, amountToSend, parentSeed, index)
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

      rows, err := tx.Query(context.Background(), queryString, parentSeed, index)
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
      sendNano(sendingKeys[0].PrivateKey, clientAddress, keyMan.NewRaw(4100))
      sendInDatabase(walletSeed[0], sendingKeys[0].Index, amountToSend, 0, 0, tx)
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
         sendNano(key.PrivateKey, transitionalAddress.PublicKey, currentSend)
         sendInDatabase(walletSeed[i], key.Index, currentSend, transitionSeedId, transitionalAddress.Index, tx)
         totalSent.Add(totalSent, currentSend)
         if (verbose) {
            fmt.Println("Sending", currentSend.Int, "from", walletSeed[i], key.Index, "to", transitionSeedId, transitionalAddress.Index)
         }
      }
      // Now send to client
      if (verbose) {
         fmt.Println("Sending", amountToSend, "from", transitionSeedId, transitionalAddress.Index, "to client.")
      }
      sendNano(transitionalAddress.PrivateKey, clientAddress, keyMan.NewRaw(4100))
      sendInDatabase(transitionSeedId, transitionalAddress.Index, amountToSend, 0, 0, tx)
   } else {
      return fmt.Errorf("receivedNano: not enough funds(2)")
   }

   tx.Commit(context.Background())

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

func sendNano(fromPrivateKey []byte, toPublicKey []byte, amount *keyMan.Raw) bool {
   // TODO TODO TODO
   return true
}

// sendInDatabase does the same work that sendNano does, but just in our local database instead.
func sendInDatabase(fromSeed int, fromIndex int, amount *keyMan.Raw, toSeed int, toIndex int, conn psqlDB) error {

   queryString :=
   "UPDATE " +
      "wallets "+
   "SET " +
      "\"balance\" = \"balance\" - $1 " +
   "WHERE " +
      "\"parent_seed\" = $2 AND " +
      "\"index\" = $3;"

   rowsAffected, err := conn.Exec(context.Background(), queryString, amount, fromSeed, fromIndex)
   if (err != nil) {
      return fmt.Errorf("sendInDatabase: Update: %w", err)
   }
   if (rowsAffected.RowsAffected() < 1) {
      return fmt.Errorf("sendInDatabase: no rows affected during index incrament")
   }

   if (toSeed != 0) {
      queryString =
      "UPDATE " +
         "wallets "+
      "SET " +
         "\"balance\" = \"balance\" + $1 " +
      "WHERE " +
         "\"parent_seed\" = $2 AND " +
         "\"index\" = $3;"
      rowsAffected, err := conn.Exec(context.Background(), queryString, amount, toSeed, toIndex)
      if (err != nil) {
         return fmt.Errorf("sendInDatabase: Update: %w", err)
      }
      if (rowsAffected.RowsAffected() < 1) {
         return fmt.Errorf("sendInDatabase: no rows affected during index incrament")
      }
   }

   return nil
}

func getWalletInfo(seed int, index int) (*keyMan.Key, error) {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return nil, fmt.Errorf("getWalletInfo: %w", err)
   }
   defer conn.Close(context.Background())

   queryString :=
   "SELECT " +
      "pgp_sym_decrypt_bytea(seed, $1)" +
   "FROM " +
      "seeds " +
   "WHERE " +
      "id = $2;"

   row, err := conn.Query(context.Background(), queryString, databasePassword, seed)
   if (err != nil) {
      return nil, fmt.Errorf("getWalletInfo: %w", err)
   }

   var key keyMan.Key
   if (row.Next()) {
      err = row.Scan(&key.Seed)
      if (err != nil) {
         return nil, fmt.Errorf("getWalletInfo: %w ", err)
      } else {
         row.Close()

         key.Index = index
         err = keyMan.SeedToKeys(&key)
         if (err != nil) {
            return nil, fmt.Errorf("getWalletInfo: %w", err)
         }
      }
   }

   if (key.NanoAddress == "") {
      return nil, fmt.Errorf("getWalletInfo: nil key: either bad address or password")
   }

   return &key, nil
}

func getSeedFromAddress(nanoAddress string) (keyMan.Key, error) {
   var key keyMan.Key

   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return key, fmt.Errorf("getSeedFromAddress: %w", err)
   }
   defer conn.Close(context.Background())

   queryString :=
   "SELECT " +
      "parent_seed, " +
      "index " +
   "FROM " +
      "wallets " +
   "WHERE " +
      "hash = $1;"

   pubkey, err := keyMan.AddressToPubKey(nanoAddress)
   if (err != nil) {
      return key, fmt.Errorf("getSeedFromAddress: %w", err)
   }

   recivedHash := blake2b.Sum256(pubkey)

   var parentSeed int
   var index int
   err = conn.QueryRow(context.Background(), queryString, recivedHash[:]).Scan(&parentSeed, &index)
   if (err != nil) {
      return key, fmt.Errorf("getSeedFromAddress: %w", err)
   }
   if (parentSeed == 0) {
      return key, fmt.Errorf("getSeedFromAddress: address not found in database")
   }

   key.Seed, _ = getSeedFromDatabase(parentSeed)
   key.Index = index
   keyMan.SeedToKeys(&key)

   return key, nil
}

func getSeedFromDatabase(id int) ([]byte, error) {

   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return nil, fmt.Errorf("getNewAddress: %w", err)
   }
   defer conn.Close(context.Background())

   queryString :=
   "SELECT " +
      "pgp_sym_decrypt_bytea(seed, $1) " +
   "FROM " +
      "seeds " +
   "WHERE " +
      "id = $2;"

   var seed []byte
   _ = conn.QueryRow(context.Background(), queryString, databasePassword, id).Scan(&seed)

   return seed, nil
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

func OpenAccount(account string) (int, error) {
   var block keyMan.Block
   var err error

   _, receivable, err := getAccountBalance(account)
   if (err != nil) {
      return -1, fmt.Errorf("OpenAccount: %w", err)
   }
   // if (receivable == 0)
   if (receivable.Cmp(keyMan.NewRaw(0)) == 0) {
      return -1, fmt.Errorf("OpenAccount: no receiveable nano in account %s", account)
   }

   key, err := getSeedFromAddress(account)
   if (err != nil) {
      return -1, fmt.Errorf("OpenAccount: %w", err)
   }
   pendingHashes := getPendingHash(account)
   pendingTransactions := len(pendingHashes[account])
   if (pendingTransactions > 0 ) {

      // Fill block with relavent information
      block.Previous = make([]byte, 32)
      block.Account = account
      block.Representative = getRepresentative()
      block.Balance = receivable
      // Link is not an addres!!!!!
      block.Link = pendingHashes[account][0]
      block.Seed = key
      if (err != nil) {
         return -1, fmt.Errorf("OpenAccount: %w", err)
      }

      sig, err := block.Sign()
      if (err != nil) {
         return -1, fmt.Errorf("OpenAccount: %w", err)
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

      PoW, err := getPoW(block.Seed.NanoAddress)
      if (err != nil) {
         return -1, fmt.Errorf("OpenAccount: %w", err)
      }

      // Send RCP request
      newHash, err := openAccount(block, sig, PoW)
      if (err != nil){
         return -1, fmt.Errorf("OpenAccount: %w", err)
      }
      if (len(newHash) != 32){
         return -1, fmt.Errorf("OpenAccount: no block hash returned from node", err)
      }

      // We'ved used any stored PoW, clear it out for next use
      clearPoW(block.Account)
      go preCalculateNextPoW(block.Account)
   }

   // Update database records
   err = updateBalance(block.Account, block.Balance)

   return pendingTransactions - 1, err
}
func getRepresentative() string {
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

   work, _ := generateWorkOnNode(hash)
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
      hash, _ = keyMan.AddressToPubKey(nanoAddress)
   }

   work, _ := generateWorkOnNode(hash)
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
      return "", fmt.Errorf("getPoW: %w", err)
   }

   if (PoW == "") {
      // none stored; need to generate it
      PoW = calculateNextPoW(nanoAddress, nanoAddressHash[:])
   }

   return PoW, nil
}
