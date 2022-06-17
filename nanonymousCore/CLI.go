package main

import (
   "fmt"
   "context"
   "encoding/hex"
   "strings"
   "strconv"
   _"embed"

   // Local packages
   keyMan "nanoKeyManager"
   nt "nanoTypes"

   // 3rd party packages
   pgx "github.com/jackc/pgx/v4"
)

func CLI() {
   var usr string
   var seed keyMan.Key

   menu:
   for {
      fmt.Print(
         //"1. Wallet\n"
         //"2. Database\n"
         //"3. RCP\n"
         "1. Generate Seed\n",
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
         h, _ := hex.DecodeString("EECE7188B8557634EBCFCCA9ABB5A640556FA67AD9AC13E9BE4D767A7230C040")
         block, _ := getBlockInfo(h)
         fmt.Println("Block", block)

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
         // Send commnad
         verbosity = 5
         fmt.Print("Seed: ")
         fmt.Scan(&usr)
         seed, _ := strconv.Atoi(usr)
         fmt.Print("Index: ")
         fmt.Scan(&usr)
         index, _ := strconv.Atoi(usr)
         sendKey, _ := getSeedFromIndex(seed, index)
         fmt.Print("Nano address: ")
         fmt.Scan(&usr)
         nanoAddress := usr
         toPubKey, err := keyMan.AddressToPubKey(nanoAddress)
         if (err != nil) {
            fmt.Println("Error: ", err.Error())
            continue
         }
         fmt.Print("Amount in Nano: ")
         fmt.Scan(&usr)
         amountNano, _ := strconv.ParseFloat(usr, 64)
         amountRaw := nt.NewRawFromNano(amountNano)
         Send(sendKey, toPubKey, amountRaw, nil, nil, -1)
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

         err := returnAllReceiveable()
         if (err != nil) {
            fmt.Println(fmt.Errorf("main: %w", err))
         }
      default:
         break menu
      }
   }
}
