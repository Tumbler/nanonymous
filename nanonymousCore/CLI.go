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
   "github.com/chzyer/readline"
)

func CLI() {
   var usr string
   var seed keyMan.Key
   if (verbosity == 0) {
      verbosity = 5
   }

   menu:
   for {
      fmt.Print(
         "1. Wallet\n",
         "2. RCP\n",
         "3. Database\n",
         //"1. Generate Seed\n",
         //"2. Get Account info\n",
         //"3. Insert into database\n",
         //"4. Send pretend request for new address\n",
         //"5. Find total balance\n",
         //"6. Pretend nano receive\n",
         //"7. Get Wallet Info\n",
         //"8. Add nano to wallet\n",
         //"9. Get block count\n",
         //"A. Clear PoW\n",
         //"B. Telemetry\n",
         //"C. Get Account Info\n",
         //"D. Sign Block\n",
         //"E. Block Info\n",
         //"H. OpenAccount\n",
         //"I. GenerateWork\n",
         //"J. Send\n",
         //"K. Recive All\n",
         //"L. Black list\n",
         //"M. Get Pending\n",
         //"N. Check Balance\n",
         //"O. Channel Test\n",
         //"P. Test\n",
      )
      fmt.Scan(&usr)

      switch (strings.ToUpper(usr)) {
      case "1":
         myKey, err := getSeedFromIndex(1, 0)
         rawBalance, _ := getBalance(myKey.NanoAddress)
         NanoBalance := rawToNANO(rawBalance)
         format := fmt.Sprintf("(%.3f)%s> ", NanoBalance, myKey.NanoAddress)
         wallet, err := readline.New(format)
         wallet.Config.AutoComplete = walletCompleter
         if (err != nil) {
            fmt.Println(fmt.Errorf("CLI: %w", err))
         }

         defer wallet.Close()

         walletMenu:
         for {
            println()
            line, err := wallet.Readline()
            if (err != nil) {
               break
            }
            array := strings.Split(strings.ToLower(line), " ")

            switch(array[0]) {
               case "send":
                  CLIsend(myKey, array)
               case "ls":
                  fallthrough
               case "list":
                  err = CLIlist(array)
                  if (err != nil) {
                     fmt.Println(fmt.Errorf("CLI: %w", err))
                  }
               case "select":
                  fallthrough
               case "set":
                  err = CLIselect(wallet, myKey, array)
                  if (err != nil) {
                     fmt.Println(fmt.Errorf("CLI: %w", err))
                  }
               case "new":
                  err = CLInew(wallet, myKey, array)
                  if (err != nil) {
                     fmt.Println(fmt.Errorf("CLI: %w", err))
                  }
               case "peek":
                  err = CLIpeek(array)
                  if (err != nil) {
                     fmt.Println(fmt.Errorf("CLI: %w", err))
                  }
               case "receiveonly":
                  err = CLIreceiveOnly(array)
                  if (err != nil) {
                     fmt.Println(fmt.Errorf("CLI: %w", err))
                  }
               case "-h":
                  fallthrough
               case "help":
                  CLIhelp(array)
               case "exit":
                  fallthrough
               case "q":
                  break walletMenu
               default:
                  println(array[0], `not recognized as a command. Try "help" or "-h"`)
            }
         }
      case "2":
         RPC, err := readline.New("RCP> ")
         RPC.Config.AutoComplete = RCPCompleter
         if (err != nil) {
            fmt.Println(fmt.Errorf("CLI: %w", err))
         }

         defer RPC.Close()

         RCPMenu:
         for {
            println()
            line, err := RPC.Readline()
            if (err != nil) {
               break
            }
            array := strings.Split(strings.ToLower(line), " ")
            println()

            switch(array[0]) {
               case "account_balance":
                  if (len(array) >= 2) {
                     // validate the address
                     _, err := keyMan.AddressToPubKey(array[1])
                     if (err != nil) {
                        fmt.Println("Invalid Address!")
                        continue
                     }
                     balance, receiveable, _ := getAccountBalance(array[1])
                     fmt.Print(array[1], ":\n   balance:    ", balance, "\n   receivable: ", receiveable)
                     fmt.Println()

                  } else {
                     fmt.Println("Not enough arguments... Expecting a nano address.")
                  }
               case "account_block_count":
                  if (len(array) >= 2) {
                     // validate the address
                     _, err := keyMan.AddressToPubKey(array[1])
                     if (err != nil) {
                        fmt.Println("Invalid Address!")
                        continue
                     }
                     blocks, err := getAccountBlockCount(array[1])
                     if (err != nil) {
                        fmt.Println("RCP error: ", err.Error())
                     }
                     fmt.Print(array[1], ":\n   block_count: ", blocks)
                     fmt.Println()

                  } else {
                     fmt.Println("Not enough arguments... Expecting a nano address.")
                  }
               case "account_history":
                  if (len(array) >= 2) {
                     // validate the address
                     _, err := keyMan.AddressToPubKey(array[1])
                     if (err != nil) {
                        fmt.Println("Invalid Address!")
                        continue
                     }
                     var histNum = -1
                     if (len(array) >= 3) {
                        histNum, _ = strconv.Atoi(array[2])
                     }
                     accountHistory, err := getAccountHistory(array[1], histNum)
                     if (err != nil) {
                        fmt.Println("RCP error: ", err.Error())
                     }
                     for i, block := range accountHistory.History {
                        fmt.Println()
                        fmt.Println("Block", i, "from frontier")
                        fmt.Println("  type: ", block.Type)
                        fmt.Println("  account: ", block.Account)
                        fmt.Print  ("  amount:  ", block.Amount)
                        fmt.Print  (" (", fmt.Sprintf("%f", rawToNANO(block.Amount)), ")\n")
                        fmt.Println("  local_timestamp: ", block.LocalTimestamp)
                        fmt.Println("  height: ", block.Height)
                        fmt.Println("  hash: ", block.Hash)
                        fmt.Println("  confirmed: ", block.Confirmed)
                     }
                     fmt.Println()

                  } else {
                     fmt.Println("Not enough arguments... Expecting a nano address.")
                  }
               case "account_info":
                  if (len(array) >= 2) {
                     // validate the address
                     _, err := keyMan.AddressToPubKey(array[1])
                     if (err != nil) {
                        fmt.Println("Invalid Address!")
                        continue
                     }
                     accountInfo, err := getAccountInfo(array[1])
                     if (err != nil) {
                        fmt.Println("RCP error: ", err.Error())
                     }
                     fmt.Println()
                     fmt.Println("Account info for", array[1], ":")
                     fmt.Println("  frontier: ", accountInfo.Frontier)
                     fmt.Println("  open_block: ", accountInfo.OpenBlock)
                     fmt.Println("  representative_block:  ", accountInfo.RepresentativeBlock)
                     fmt.Println("  balance: ", accountInfo.Balance)
                     fmt.Println("  modified_timestamp: ", accountInfo.ModifiedTimestamp)
                     fmt.Println("  block_count: ", accountInfo.BlockCount)
                     fmt.Println("  account_version: ", accountInfo.AccountVersion)
                     fmt.Println("  confirmation_height: ", accountInfo.ConfirmationHeight)
                     fmt.Println("  confirmation_height_frontier: ", accountInfo.ConfirmationHeightFrontier)
                     fmt.Println()

                  } else {
                     fmt.Println("Not enough arguments... Expecting a nano address.")
                  }
               case "q":
                  fallthrough
               case "exit":
                  break RCPMenu
               default:
            }

         }
      case "one":
         keyMan.WalletVerbose(true)

         err := keyMan.GenerateSeed(&seed)
         if (err != nil){
            fmt.Println(err.Error())
         }

         keyMan.WalletVerbose(false)
      case "two":
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

      case "three":
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
         blarg, _, err := getNewAddress(adhocAddress, false, 0)
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

func CLIsend(myKey *keyMan.Key, args []string) {
   toPubKey, err := keyMan.AddressToPubKey(args[2])
   if (err != nil) {
      println(args[2], "is not a valid address")
      return
   }

   amountNano, err := strconv.ParseFloat(args[1], 64)
   if (err != nil) {
      println(args[1], "is not a valid nano amount")
      return
   }

   amountRaw := nt.NewRawFromNano(amountNano)
   _, err = Send(myKey, toPubKey, amountRaw, nil, nil, -1)
   if (err != nil) {
      println(fmt.Errorf("CLIsend: %w", err))
   }
}

func CLIlist(args []string) error {
   var err error
   var rows pgx.Rows

   var ignoreZeroBalance bool
   if !(contains(args, "-z")) {
      ignoreZeroBalance = true
   }

   if (contains(args, "-a")) {
      rows, err = getAllWalletRowsFromDatabase()
   } else {
      rows, err = getWalletRowsFromDatabase()
   }

   if (err != nil) {
      return fmt.Errorf("CLIlist: %w", err)
   }

   for (rows.Next()) {
      var seed int
      var index int
      var balance = nt.NewRaw(0)

      rows.Scan(&seed, &index, balance)

      if (ignoreZeroBalance && balance.Cmp(nt.NewRaw(0)) == 0) {
         continue
      }

      balanceInNano := rawToNANO(balance)

      fmt.Print(seed, ",", index, ":  Ӿ ", balanceInNano, "\n")
   }

   return nil
}

func CLIselect(r *readline.Instance, myKey *keyMan.Key, args []string) error {

   seed, err := strconv.Atoi(args[1])
   if (err != nil) {
      return fmt.Errorf("CLIselect: %w", err)
   }

   index, err := strconv.Atoi(args[2])
   if (err != nil) {
      return fmt.Errorf("CLIselect: %w", err)
   }

   myKey, err = getSeedFromIndex(seed, index)
   if (err != nil) {
      return fmt.Errorf("CLIselect: %w", err)
   }

   rawBalance, _ := getBalance(myKey.NanoAddress)
   NanoBalance := rawToNANO(rawBalance)
   format := fmt.Sprintf("(%.3f)%s> ", NanoBalance, myKey.NanoAddress)

   r.SetPrompt(format)

   return nil
}

func CLInew(r *readline.Instance, myKey *keyMan.Key, args []string) error {

   var receiveOnly bool
   var newSeed bool
   if (contains(args, "receiveonly")) {
      receiveOnly = true
   } else if (contains(args, "seed")) {
      newSeed = true
   }

   if !(newSeed) {
      // new wallet
      var seedId int
      if (len(args) >= 2) {
         seedId, _ = strconv.Atoi(args[1])
      }

      key, seed, err := getNewAddress("", receiveOnly, seedId)
      if (err != nil) {
         return fmt.Errorf("CLInew: %w", err)
      }

      myKey, err = getSeedFromIndex(seed, key.Index)
      if (err != nil) {
         return fmt.Errorf("CLIselect: %w", err)
      }
   } else {
      // new seed
      conn, err := pgx.Connect(context.Background(), databaseUrl)
      if (err != nil) {
         return fmt.Errorf("CLInew: %w", err)
      }
      defer conn.Close(context.Background())

      keyMan.ReinitSeed(myKey)
      err = keyMan.GenerateSeed(myKey)
      if (err != nil) {
         return fmt.Errorf("CLInew: %w", err)
      }

      hexString := hex.EncodeToString(myKey.Seed)
      fmt.Println("seed: ", hexString)

      _, err = insertSeed(conn, myKey.Seed)
      if (err != nil) {
         return fmt.Errorf("CLInew: %w", err)
      }
   }

      rawBalance, _ := getBalance(myKey.NanoAddress)
      NanoBalance := rawToNANO(rawBalance)
      format := fmt.Sprintf("(%.3f)%s> ", NanoBalance, myKey.NanoAddress)

   r.SetPrompt(format)

   return nil
}

func CLIpeek(args []string) error {

   seed, err := strconv.Atoi(args[1])
   if (err != nil) {
      return fmt.Errorf("CLIpeek: %w", err)
   }

   index, err := strconv.Atoi(args[2])
   if (err != nil) {
      return fmt.Errorf("CLIpeek: %w", err)
   }

   tmpKey, err := getSeedFromIndex(seed, index)
   if (err != nil) {
      return fmt.Errorf("CLIpeek: %w", err)
   }

   rawBalance, _ := getBalance(tmpKey.NanoAddress)
   NanoBalance := rawToNANO(rawBalance)
   fmt.Printf("%s(%d,%d): Ӿ %f\n", tmpKey.NanoAddress, seed, tmpKey.Index, NanoBalance)

   return nil
}

func contains(a []string, search string) bool {
   for _, s := range a {
      if s == search {
         return true
      }
   }
   return false
}

func CLIreceiveOnly(args []string) error {

   seed, err := strconv.Atoi(args[1])
   if (err != nil) {
      return fmt.Errorf("CLIreceiveOnly: %w", err)
   }

   index, err := strconv.Atoi(args[2])
   if (err != nil) {
      return fmt.Errorf("CLIreceiveOnly: %w", err)
   }

   key, err := getSeedFromIndex(seed, index)
   if (err != nil) {
      return fmt.Errorf("CLIreceiveOnly: %w", err)
   }

   if (contains(args, "off")) {
      setAddressNotReceiveOnly(key.NanoAddress)
   } else {
      setAddressReceiveOnly(key.NanoAddress)
   }

   return nil
}

func CLIhelp(args []string) {
   var term string
   if (len(args) >= 2) {
      term = args[1]
   }
   switch(term) {
      case "send":
         CLIhelpsend()
      case "ls":
         fallthrough
      case "list":
         CLIhelplist()
      case "select":
         fallthrough
      case "set":
         CLIhelpselect()
      case "new":
         CLIhelpnew()
      case "peek":
         CLIhelppeek()
      case "receiveonly":
         CLIhelpreceiveOnly()
      case "-h":
         fallthrough
      case "help":
         CLIhelphelp()
      case "exit":
         fallthrough
      case "q":
         CLIhelpexit()
      default:
         fmt.Print("\nThis is the wallet CLI. The supported commands are:\n")
         CLIhelpexit()
         fmt.Print("\n   - help [command] | This help message\n",
                   "          Flags: [command] | the individual help for the indicated command\n")
         CLIhelplist()
         CLIhelpnew()
         CLIhelppeek()
         CLIhelpreceiveOnly()
         CLIhelpselect()
         CLIhelpsend()
   }
}

func CLIhelplist() {
   fmt.Println()
   fmt.Print("   - list/ls [-a] [-z] | Shows a list of wallet balances\n",
             "             Flags: -a | By default does not show wallets listed as \"receieveOnly\" This\n",
             "                         flag will show wallets marked as such\n",
             "                    -z | By default does not show wallets with zero balance. This flag\n",
             "                         will show wallets regardless of their balance\n",
   )
}

func CLIhelpnew() {
   fmt.Println()
   fmt.Print("   - new [seedID] [\"receiveonly\"] [\"seed\"] | Adds a new wallet to the database\n",
             "            Flags: receiveonly | When adding the wallet also sets the receive_only flag\n",
             "                   seed        | Instead of adding a new wallet, it adds a new seed (not\n",
             "                                 compatible with the \"receiveonly\" option)\n",
             "                   seedID      | Opens the next wallet in seedID\n",
   )
}

func CLIhelppeek() {
   fmt.Println()
   fmt.Print("   - peek {seed} {index} | Opens the wallet with the specified seed and index and displays\n",
             "                           its nano address and balance\n",
   )
}

func CLIhelphelp() {
   fmt.Println()
   fmt.Print("What do you want from me? It's a help function, you know what to do with it!\n")
}

func CLIhelpexit() {
   fmt.Println()
   fmt.Print("   - exit/q | Takes you back to the previous menu\n")
}

func CLIhelpsend() {
   fmt.Println()
   fmt.Print("   - send {Nano} {address} | Sends \"Nano\" amount from the currently slected address\n",
             "                             to \"address\"",
   )
}

func CLIhelpselect() {
   fmt.Println()
   fmt.Print("   - set/select {seed} {index} | Sets the current wallet to indicated seed and index\n")
}

func CLIhelpreceiveOnly() {
   fmt.Println()
   fmt.Print("   - receiveonly {seed} {index} [\"off\"] | Sets the wallet indicated by seed and index to\n",
             "                                        be receive only\n",
             "                        Flags: off    | When set it will set the indicated wallet to NOT\n",
             "                                        be receive only\n",
   )
}

var walletCompleter = readline.NewPrefixCompleter(
   readline.PcItem("new",
      readline.PcItem("receiveonly"),
      readline.PcItem("seed"),
   ),
   readline.PcItem("send"),
   readline.PcItem("ls"),
   readline.PcItem("list"),
   readline.PcItem("select"),
   readline.PcItem("set"),
   readline.PcItem("peek"),
   readline.PcItem("receiveonly",
      readline.PcItem("off"),
   ),
   readline.PcItem("help"),
   readline.PcItem("exit"),
)

var RCPCompleter = readline.NewPrefixCompleter(
   readline.PcItem("account_balance"),
   readline.PcItem("account_block_count"),
   readline.PcItem("account_history"),
   readline.PcItem("account_info"),
)
