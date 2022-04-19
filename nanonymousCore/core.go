package main

import (
   "fmt"
   "time"
   "context"
   "encoding/hex"
   "strings"
   _"embed"

   // Local packages
   keyMan "nanoKeyManager"

   // 3rd party packages
   pgx "github.com/jackc/pgx/v4"
   pgxErr "github.com/jackc/pgerrcode"
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
                "4. Send pretend request for new address\n")
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
         var databaseUrl = "postgres://postgres:folliCle99@localhost:5432/postgres"

         conn, err := pgx.Connect(context.Background(), databaseUrl)

         if (err != nil) {
            fmt.Println("main: ", err)
            return
         }
         defer conn.Close(context.Background())

         var id int
         var name string
         var number string

         rows, err := conn.Query(context.Background(), "SELECT * FROM gotest1")

         if (err != nil) {
            fmt.Println("QueryRow failed: ", err)
            return
         }

         for rows.Next() {
            err = rows.Scan(&id, &name, &number)
            if (err != nil) {
               fmt.Println("Scan failed: ", err)
               return
            }
            fmt.Println("ID: ", id, "Name: ", name, "Number: ", number)
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
      "wallets(parent_seed, index, balence) " +
   "VALUES " +
      "($1, $2, 0)"

   rowsAffected, err := conn.Exec(context.Background(), queryString, id, seed.Index)
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
