package main

import (
   "fmt"
   "time"
   keyMan "nanoKeyManager"
)

func main() {
   fmt.Println("Starting nanonymous Core on ", time.Now())

   var usr string
   var seed keyMan.Key

   menu:
   for {
      fmt.Print("1. Generate Seed\n",
                "2. Input Mnomonic\n",
                "3. Get next address\n",
                "4. Delete stored seed\n",
                "5. Test stuff\n",
                "6. Nano address to pubkey\n")
      fmt.Scan(&usr)

      keyMan.WalletVerbose(true)
      switch (usr) {
      case "1":
         err := keyMan.GenerateSeed(&seed)
         if (err != nil){
            fmt.Println(err.Error())
         }

      default:
         break menu
      }
      keyMan.WalletVerbose(false)
   }
}
