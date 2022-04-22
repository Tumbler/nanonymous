package main

import (
   "fmt"
   "time"
   "context"
   "encoding/hex"
   "strings"
   "math/big"
   "strconv"
   _"embed"

   // Local packages
   keyMan "nanoKeyManager"

   // 3rd party packages
   pgx "github.com/jackc/pgx/v4"
   //"github.com/jackc/pgtype"
   pgxErr "github.com/jackc/pgerrcode"
   "github.com/shopspring/decimal"
   "golang.org/x/crypto/blake2b"
)

// TODO IP lock transactions 1 per 30 seconds??

//go:embed databaseUrl.txt
var databaseUrl string
// Need to trim later

const MAX_INDEX = 4294967295

//type activeTransaction struct {
   //parentSeed int,
   //index int,
   //publicAddress []byte,
//}

var activeTransactionList = make(map[string][]byte)

var password = "aweqoiuoiasdfho"

func main() {
   fmt.Println("Starting nanonymous Core on ", time.Now())
   databaseUrl = strings.Trim(databaseUrl, "\r\n")

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
                "7. Rando\n")
      fmt.Scan(&usr)

      switch (usr) {
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

         var id int
         var seed []byte
         var current_index int

         rows, err := conn.Query(context.Background(), "SELECT * FROM seeds")

         if (err != nil) {
            fmt.Println("QueryRow failed: ", err)
            return
         }

         for rows.Next() {
            err = rows.Scan(&id, &seed, &current_index)
            if (err != nil) {
               fmt.Println("Scan failed: ", err)
               return
            }
            fmt.Println("ID: ", id, "Name: ", seed, "Number: ", current_index)
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

         if (err != nil) { fmt.Println("main: ", err)
            break //menu
         }

      case "4":
         adhocAddress := "nano_157gdx49th7w6yrtgi6ciys9kfcs49iew5fa64j8584178p5zjaum33jfut6"
         err := userRequestsNewAddress(adhocAddress)
         if (err != nil) {
            fmt.Println(err)
         }
      case "5":
         _, err := findTotalBalance()
         if (err != nil) {
            fmt.Println("err: ", err.Error())
         }
      case "6":
         adhocAddress := []string{
            "nano_183t7xkm6is3ge3dedfuxepyhd36i9qmehc9yenjzd8ahytu8xjw5pt7eec3",
            "lkjlkj",
         }
         nanoRecieved := new(big.Int).Exp(big.NewInt(10), big.NewInt(30), nil)
         err := receivedNano(adhocAddress[0], nanoRecieved)
         if (err != nil) {
            fmt.Println("err: ", err.Error())
         }

      case "7":

         keyMan.WalletVerbose(true)
         var seed keyMan.Key
         keyMan.GenerateSeed(&seed)
         conn, _ := pgx.Connect(context.Background(), databaseUrl)

         id, _ := insertSeed(conn, seed.Seed)

         fmt.Println("ID: ", id)

         queryString :=
         "SELECT " +
            "pgp_sym_decrypt_bytea(seed, $1)" +
         "FROM " +
            "seeds " +
         "WHERE " +
            "id = $2;"

         var seed2 keyMan.Key
         row, _ := conn.Query(context.Background(), queryString, password, id)
         //if (err != nil) {
            //return true, fmt.Errorf("checkBlackList: seed query: %w", err)
         //}
         if (row.Next()) {
            row.Scan(&seed2.Seed)
            row.Close()
            //if (err != nil) {
               //return true, fmt.Errorf("checkBlacklist: %w", err)
            //}
         } else {
            //return true, fmt.Errorf("checkBlacklist: No such seed found: %d", parentSeed)
         }
         row.Close()
         keyMan.SeedToKeys(&seed2)
         fmt.Println("wallet: ", seed2.NanoAddress)


      default:
         break //menu
      }
   //}


}


func insertSeed(conn *pgx.Conn, seed []byte) (int, error) {
   var id int

   queryString :=
   "INSERT INTO " +
     "seeds (seed, current_index) " +
   "VALUES " +
     "(pgp_sym_encrypt_bytea($1, $2), -1) " +
   "RETURNING id;"

   rows, err := conn.Query(context.Background(), queryString, seed, password)
   if (err != nil) {
      return -1, fmt.Errorf("insertSeed: %w", err)
   }

   if (rows.Next()) {
      err = rows.Scan(&id)
      if (err != nil) {
         return -1, fmt.Errorf("inserSeed: &w ", err)
      }
   }

   rows.Close()

   return id, nil
}

func userRequestsNewAddress(receivingAddress string) error {
   var seed keyMan.Key

   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return fmt.Errorf("userRequestsNewAddress: %w", err)
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
   rows, err := conn.Query(context.Background(), queryString, password, MAX_INDEX)
   if (err != nil) {
      return fmt.Errorf("userRequestsNewAddress: Query seed: ", err)
   }

   // Get a current seed. If it fails, generate a new one.
   var id int
   if (rows.Next()) {
      err = rows.Scan(&id, &seed.Seed, &seed.Index)
      if (err != nil) {
         return fmt.Errorf("userRequestsNewAddress: &w ", err)
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
         return fmt.Errorf("userRequestsNewAddress: &w ", err)
      }

      id, err = insertSeed(conn, seed.Seed)
      if (err != nil) {
         return fmt.Errorf("userRequestsNewAddress: &w ", err)
      }
   }
   fmt.Println("seed ID: ", id)


   // Add to list of managed wallets
   queryString =
   "INSERT INTO "+
      "wallets(parent_seed, index, balance, hash) " +
   "VALUES " +
      "($1, $2, 0, $3)"

   hash := blake2b.Sum256(seed.PublicKey)
   rowsAffected, err := conn.Exec(context.Background(), queryString, id, seed.Index, hash[:])
   if (err != nil) {
      return fmt.Errorf("userRequestsNewAddress: %w", err)
   }
   if (rowsAffected.RowsAffected() < 1) {
      return fmt.Errorf("userRequestsNewAddress: no rows affected in insert")
   }

   fmt.Println("Next address retrieved: ", seed.NanoAddress)

   queryString =
   "UPDATE " +
      "\"seeds\"" +
   "SET " +
      "\"current_index\" = $1 " +
   "WHERE " +
      "\"id\" = $2;"

   rowsAffected, err = conn.Exec(context.Background(), queryString, seed.Index, id)
   if (err != nil) {
      return fmt.Errorf("userRequestsNewAddress: Update: %w", err)
   }
   if (rowsAffected.RowsAffected() < 1) {
      return fmt.Errorf("userRequestsNewAddress: no rows affected during index incrament")
   }

   // Blacklist new addres with the receiving address
   receivingAddressByte, err := keyMan.AddressToPubKey(receivingAddress)
   if (err != nil) {
      return fmt.Errorf("userRequestsNewAddress: %w", err)
   }

   err = blacklist(conn, seed.PublicKey, receivingAddressByte)
   if (err != nil) {
      return fmt.Errorf("userRequestsNewAddress: Blacklist falied: %w", err)
   }

   setClientAddress(id, seed.Index, receivingAddressByte)

   return nil
}

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

func receivedNano(nanoAddress string, payment *big.Int) error {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return fmt.Errorf("receivedNano: %w", err)
   }
   defer conn.Close(context.Background())

   tx, _ := conn.BeginTx(context.Background(), pgx.TxOptions{})
   defer tx.Rollback(context.Background())

   // TODO Don't accept if we don't have an active transaction available

   queryString :=
   "SELECT " +
      "* " +
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
   var balance decimal.Decimal
   var hash []byte
   if (row.Next()) {
      err = row.Scan(&parentSeed, &index, &balance, &hash)
      row.Close()
      if (err != nil) {
         return fmt.Errorf("receivedNano: %w", err)
      }
   } else {
      return fmt.Errorf("receivedNano: address not found in active wallets")
   }

   // TODO This is just for testing
   clientPub, _ := keyMan.AddressToPubKey("nano_157gdx49th7w6yrtgi6ciys9kfcs49iew5fa64j8584178p5zjaum33jfut6")
   setClientAddress(parentSeed, index, clientPub)
   // TODO end of test code

   // Get client address for later use. TODO check for nil
   clientAddress := getClientAddress(parentSeed, index)

   // Add funds we got into our database of wallets
   queryString =
   "UPDATE " +
      "wallets "+
   "SET " +
      "\"balance\" = \"balance\" + $1 " +
   "WHERE " +
      "\"parent_seed\" = $2 AND " +
      "\"index\" = $3;"

   paymentDecimal := decimal.NewFromBigInt(payment, 0)
   rowsAffected, err := tx.Exec(context.Background(), queryString, paymentDecimal, parentSeed, index)
   if (err != nil) {
      return fmt.Errorf("receivedNano: Update: %w", err)
   }
   if (rowsAffected.RowsAffected() < 1) {
      return fmt.Errorf("receivedNano: no rows affected during index incrament")
   }


   var feePercent int64
   feePercent = 10
   fee := new(big.Int).Div(payment, big.NewInt(feePercent))
   amountToSend := new(big.Int).Sub(payment, fee)
   amountToSendDecimal := decimal.NewFromBigInt(amountToSend, 0)

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

   rows, err := tx.Query(context.Background(), queryString, amountToSendDecimal, parentSeed, index)
   if (err != nil) {
      return fmt.Errorf("receiviedNano: Query: %w", err)
   }

   var foundAddress bool
   for rows.Next() {
      err = rows.Scan(&parentSeed, &index, &balance)
      if (err != nil) {
         return fmt.Errorf("receivedNano: Scan: %w", err)
      }

      // Check the blacklist before accepting
      foundEntry, err := checkBlackList(parentSeed, index, clientAddress)
      if (err != nil) {
         return fmt.Errorf("receivedNano: %w", err)
      }
      if (!foundEntry) {
         // Uset this address
         foundAddress = true
         break
      }
   }
   rows.Close()

   if (!foundAddress) {
      // TODO add support for multi wallet sending
      return fmt.Errorf("receivedNano: Not enough funds in a single wallet")
   }

   // TODO do a send trasaction to specified wallet stored activeTransactionList

   queryString =
   "UPDATE " +
      "wallets "+
   "SET " +
      "\"balance\" = \"balance\" - $1 " +
   "WHERE " +
      "\"parent_seed\" = $2 AND " +
      "\"index\" = $3;"

   rowsAffected, err = tx.Exec(context.Background(), queryString, amountToSendDecimal, parentSeed, index)
   if (err != nil) {
      return fmt.Errorf("receivedNano: Update: %w", err)
   }
   if (rowsAffected.RowsAffected() < 1) {
      return fmt.Errorf("receivedNano: no rows affected during index incrament")
   }

   tx.Commit(context.Background())

   return nil
}

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

   var rawBalance decimal.Decimal
   var nanoBalance float64
   row, err := conn.Query(context.Background(), queryString)
   if (err != nil) {
      return -1.0, fmt.Errorf("QueryRow failed: %w", err)
   }

   if (row.Next()) {
      err = row.Scan(&rawBalance)
      if (err != nil) {
         return -1.0, fmt.Errorf("findTotalBalance: %w", err)
      }

      nanoBalance = rawToNANO(rawBalance.BigInt())

      fmt.Println("Total Balance is: Ӿ", nanoBalance)
   }

   return nanoBalance, nil
}

// rawToNANO is used to convert raw to NANO AKA Mnano (the communnity just calls
// this a nano). We don't have a conversion to go the other way as all
// operations should be done in raw to avoid rounding errors. We only want to
// convert when outputing for human readable format.
func rawToNANO(raw *big.Int) float64{
   // 1 NANO is 10^30 raw
   rawConv := new(big.Int).Exp(big.NewInt(10), big.NewInt(30), nil)
   rawConvFloat := new(big.Float).SetInt(rawConv)
   rawFloat := new(big.Float).SetInt(raw)

   NanoFloat := new(big.Float).Quo(rawFloat, rawConvFloat)

   NanoFloat64, _ := NanoFloat.Float64()

   return NanoFloat64
}

func checkBlackList(parentSeed int, index int, clientAddress []byte) (bool, error) {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return true, fmt.Errorf("checkBlackList: %w", err)
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
   row, err := conn.Query(context.Background(), queryString, password, parentSeed)
   if (err != nil) {
      return true, fmt.Errorf("checkBlackList: seed query: %w", err)
   }
   if (row.Next()) {
      err = row.Scan(&seed.Seed)
      row.Close()
      if (err != nil) {
         return true, fmt.Errorf("checkBlacklist: %w", err)
      }
   } else {
      return true, fmt.Errorf("checkBlacklist: No such seed found: %d", parentSeed)
   }
   row.Close()

   seed.Index = index
   err = keyMan.SeedToKeys(&seed)
   if (err != nil) {
      return true, fmt.Errorf("checkBlackList: %w", err)
   }

   concat := append(seed.PublicKey, clientAddress[:]...)
   blackListHash := blake2b.Sum256(concat)

   fmt.Println("check blacklist for:", hex.EncodeToString(blackListHash[:]))

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
      return true, fmt.Errorf("checkBlackList: blacklist query: %w", err)
   }

   if (blacklistRows.Next()) {
      // Found entry in blacklist
      return true, nil
   } else {
      return false, nil
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

