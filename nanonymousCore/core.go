package main

import (
   "fmt"
   "time"
   "context"
   "encoding/hex"
   "strings"
   "math/big"
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

//go:embed databaseUrl.txt
var databaseUrl string
// Need to trim later

const MAX_INDEX = 4294967295

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
                "6. Pretend nano receive\n")
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

         if (err != nil) {
            fmt.Println("main: ", err)
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
         adhocAddress := "nano_183t7xkm6is3ge3dedfuxepyhd36i9qmehc9yenjzd8ahytu8xjw5pt7eec3"
         nanoRecieved := new(big.Int).Exp(big.NewInt(10), big.NewInt(30), nil)
         err := receivedNano(adhocAddress, nanoRecieved)
         if (err != nil) {
            fmt.Println("err: ", err.Error())
         }

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
     "($1, -1) " +
   "RETURNING id;"

   rows, err := conn.Query(context.Background(), queryString, seed)
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
      "seed, " +
      "current_index " +
   "FROM " +
      "seeds " +
   "WHERE " +
      "current_index < $1" +
   "ORDER BY " +
      "id;"

   // TODO start a tranasction and increment current_index. Only commit after everthing checks out
   rows, err := conn.Query(context.Background(), queryString, MAX_INDEX)
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

   row, err := conn.Query(context.Background(), queryString, recivedHash[:])
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

   queryString =
   "UPDATE " +
      "wallets "+
   "SET " +
      "\"balance\" = $1 " +
   "WHERE " +
      "\"parent_seed\" = $2 AND " +
      "\"index\" = $3;"

   newBalance := balance.Add(decimal.NewFromBigInt(payment, 0))
   rowsAffected, err := conn.Exec(context.Background(), queryString, newBalance, parentSeed, index)
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

   queryString =
   "SELECT " +
      "parent_seed, " +
      "index, " +
      "balance " +
   "FROM " +
      "wallets " +
   "WHERE " +
      "balance >= $1 " +
   "ORDER BY " +
      "balance;"

   rows, err := conn.Query(context.Background(), queryString, amountToSendDecimal)
   if (err != nil) {
      return fmt.Errorf("receiviedNano: Query: %w", err)
   }

   for rows.Next() {
      err = rows.Scan(&parentSeed, &index, &balance)
      if (err != nil) {
         return fmt.Errorf("receivedNano: Scan: %w", err)
      }

      fmt.Println(parentSeed, index, ":", balance.BigInt())
   }
   row.Close()


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
