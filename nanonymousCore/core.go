package main

import (
   "fmt"
   "time"
   "context"
   "encoding/hex"

   // Local packages
   keyMan "nanoKeyManager"

   // 3rd party packages
   pgx "github.com/jackc/pgx/v4"
)

func main() {
   fmt.Println("Starting nanonymous Core on ", time.Now())

   var usr string
   var seed keyMan.Key

   //menu:
   //for {
      fmt.Print("1. Generate Seed\n",
                "2. Database test\n",
                "3. Insert into database\n")
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
         var databaseUrl = "postgres://postgres:folliCle99@localhost:5432/nanonymousdb"
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

         err = insertSeed(conn, newSeed.Seed)

         if (err != nil) {
            fmt.Println("main: ", err)
            break //menu
         }

      default:
         break //menu
      }
   //}


}


func insertSeed(conn *pgx.Conn, seed []byte) error {
   queryString :=
   "INSERT INTO " +
     "seeds (seed, currentIndex) " +
   "VALUES " +
     "($1, 0);"

   rowsAffected, err := conn.Exec(context.Background(), queryString, seed)
   if (err != nil) {
      return fmt.Errorf("insertSeed: %w", err)
   }

   if (rowsAffected.RowsAffected() < 1) {
      return fmt.Errorf("insertSeed: no rows affected")
   }

   return nil
}
