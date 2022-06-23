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

   myKey, err := getSeedFromIndex(1, 0)
   rawBalance, _ := getBalance(myKey.NanoAddress)
   NanoBalance := rawToNANO(rawBalance)
   prompt := fmt.Sprintf("(%.3f)%s> ", NanoBalance, myKey.NanoAddress)
   wallet, err := readline.New(prompt)
   wallet.Config.AutoComplete = walletCompleter
   if (err != nil) {
      fmt.Println(fmt.Errorf("CLI: %w", err))
   }
   defer wallet.Close()

   RPC, err := readline.New("RCP> ")
   RPC.Config.AutoComplete = RCPCompleter
   if (err != nil) {
      fmt.Println(fmt.Errorf("CLI: %w", err))
   }
   defer RPC.Close()

   DB, err := readline.New("DB> ")
   DB.Config.AutoComplete = DBCompleter
   if (err != nil) {
      fmt.Println(fmt.Errorf("CLI: %w", err))
   }
   defer DB.Close()


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
         walletMenu:
         for {
            fmt.Println()
            wallet.SetPrompt(prompt)
            line, err := wallet.Readline()
            if (err != nil) {
               break
            }
            array := strings.Split(strings.ToLower(line), " ")

            switch(array[0]) {
               case "send":
                  CLIsend(myKey, array, &prompt)
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
                  err = CLIselect(myKey, array, &prompt)
                  if (err != nil) {
                     fmt.Println(fmt.Errorf("CLI: %w", err))
                  }
               case "new":
                  err = CLInew(myKey, array, &prompt)
                  if (err != nil) {
                     fmt.Println(fmt.Errorf("CLI: %w", err))
                  }
               case "peek":
                  err = CLIpeek(array)
                  if (err != nil) {
                     fmt.Println(fmt.Errorf("CLI: %w", err))
                  }
               case "receive":
                  err = CLIreceive(myKey, array, &prompt)
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
               case "verbosity":
                  if (len(array) >= 2 && len(array[1]) > 0) {
                     verbosity, _ = strconv.Atoi(array[1])
                  } else {
                     fmt.Println(" Verbosity: ", verbosity)
                  }
               case "exit":
                  fallthrough
               case "q":
                  break walletMenu
               default:
                  println(array[0], `not recognized as a command. Try "help" or "-h"`)
            }
         }
      case "2":
         RCPMenu:
         for {
            fmt.Println()
            line, err := RPC.Readline()
            if (err != nil) {
               break
            }
            array := strings.Split(strings.ToLower(line), " ")
            fmt.Println()

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
                        continue
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
                        continue
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
               case "account_representative":
                  if (len(array) >= 2) {
                     // validate the address
                     _, err := keyMan.AddressToPubKey(array[1])
                     if (err != nil) {
                        fmt.Println("Invalid Address!")
                        continue
                     }
                     accountInfo, err := getAccountRep(array[1])
                     if (err != nil) {
                        fmt.Println("RCP error: ", err.Error())
                        continue
                     }
                     fmt.Println()
                     fmt.Println("Account rep for", array[1], ":")
                     fmt.Println("  representative: ", accountInfo.Representative)
                     fmt.Println()

                  } else {
                     fmt.Println("Not enough arguments... Expecting a nano address.")
                  }
               case "account_weight":
                  if (len(array) >= 2) {
                     // validate the address
                     _, err := keyMan.AddressToPubKey(array[1])
                     if (err != nil) {
                        fmt.Println("Invalid Address!")
                        continue
                     }
                     weightRaw, err := getAccountWeight(array[1])
                     if (err != nil) {
                        fmt.Println("RCP error: ", err.Error())
                        continue
                     }
                     fmt.Println()
                     fmt.Println("Weight delegated to", array[1], ":")
                     fmt.Print  ("  weight: ", weightRaw)
                     fmt.Print  (" (", rawToNANO(weightRaw), ")")
                     fmt.Println()

                  } else {
                     fmt.Println("Not enough arguments... Expecting a nano address.")
                  }
               case "available_supply":
                  supply, err := getAvailableSupply()
                  if (err != nil) {
                     fmt.Println("RCP error: ", err.Error())
                  }
                  fmt.Println()
                  fmt.Print("Supply: ", supply, " (", rawToNANO(supply), ")")
                  fmt.Println()
               case "block_account":
                  if (len(array) >= 2) {
                     address, err := getOwnerOfBlock(array[1])
                     if (err != nil) {
                        fmt.Println("RCP error: ", err.Error())
                        continue
                     }
                     fmt.Println()
                     fmt.Println("Owner of hash", array[1], ":")
                     fmt.Println("  account: ", address)
                     fmt.Println()

                  } else {
                     fmt.Println("Not enough arguments... Expecting a block hash")
                  }
               case "block_count":
                  count, unchecked, cemented, err := getBlockCount()
                  if (err != nil) {
                     fmt.Println("RCP error: ", err.Error())
                  }
                  fmt.Println("  count: ", count)
                  fmt.Println("  unchecked: ", unchecked)
                  fmt.Println("  cemented:  ", cemented)
               case "block_info":
                  if (len(array) >= 2) {
                     var hash nt.BlockHash
                     hash, err = hex.DecodeString(array[1])
                     if (err != nil) {
                        fmt.Println("Invalid hash!")
                        continue
                     }
                     blockInfo, err := getBlockInfo(hash)
                     if (err != nil) {
                        fmt.Println("RCP error: ", err.Error())
                        continue
                     }
                     fmt.Println()
                     fmt.Println("Block info for", array[1], ":")
                     fmt.Print  ("  amount: ", blockInfo.Amount)
                     fmt.Print  (" (", rawToNANO(blockInfo.Amount), ")\n")
                     fmt.Println("  contents: ")
                     fmt.Println("    type: ", blockInfo.Contents.Type)
                     fmt.Println("    account: ", blockInfo.Contents.Account)
                     fmt.Println("    previous: ", blockInfo.Contents.Previous)
                     fmt.Print  ("    balance: ", blockInfo.Contents.Balance)
                     fmt.Print  (" (", rawToNANO(blockInfo.Contents.Balance), ")\n")
                     fmt.Println("    link: ", blockInfo.Contents.Link)
                     fmt.Println("    link_as_account: ", blockInfo.Contents.LinkAsAccount)
                     fmt.Println("    signature: ", blockInfo.Contents.Signature)
                     fmt.Println("    work: ", blockInfo.Contents.Work)
                     fmt.Println("  height:  ", blockInfo.Height)
                     fmt.Println("  local_timestamp: ", blockInfo.LocalTimestamp)
                     fmt.Println("  successor: ", blockInfo.Successor)
                     fmt.Println("  confirmed: ", blockInfo.Confirmed)
                     fmt.Println("  sybtype: ", blockInfo.Subtype)
                     fmt.Println()
                  }
               case "bootstrap_status":
                  err := printBootstrapStatus()
                  if (err != nil) {
                     fmt.Println("RCP error:", err.Error())
                  }

               case "chain":
                  if (len(array) >= 2) {
                     var count = -1
                     if (len(array) >= 3) {
                        count, _ = strconv.Atoi(array[2])
                     }
                     hash, err := hex.DecodeString(array[1])
                     if (err != nil) {
                        fmt.Println("Invalid hash!")
                        continue
                     }
                     blocks, err := getBlocksInChain(hash, count)
                     if (err != nil) {
                        fmt.Println("RCP error: ", err.Error())
                        continue
                     }
                     fmt.Println()
                     fmt.Println("Blocks in chain:", array[1], ":")
                     for _, block := range blocks {
                        fmt.Println("   ", block)
                     }
                     fmt.Println()
                  } else {
                     fmt.Println("Not enough arguments... Expecting a block hash")
                  }

               case "confirmation_active":
                  //var showBlocks bool
                  //if (contains(array, "showblocks")) {
                     //showBlocks = true
                  //}

                  confirmations, unconfirmed, confirmed, err := getActiveConfirmations()
                  if (err != nil) {
                     fmt.Println("RCP error: ", err.Error())
                     continue
                  }

                  fmt.Println()
                  //if (showBlocks) {
                     //for _, block := range confirmations {
                        //fmt.Println(" ", block)
                     //}
                  //} else {
                     fmt.Println("  confirmations: ", confirmations)
                  //}
                  fmt.Println("  unconfirmed: ", unconfirmed)
                  fmt.Println("  confirmed: ", confirmed)
                  fmt.Println()
               case "confirmation_history":
                  err := printConfirmationHistory()
                  if (err != nil) {
                     fmt.Println("RCP error: ", err.Error())
                  }
               case "confirmation_quorum":
                  err := printConfirmationQuorum()
                  if (err != nil) {
                     fmt.Println("RCP error: ", err.Error())
                  }
               case "delgators_count":
                  if (len(array) >= 2) {
                     // validate the address
                     _, err := keyMan.AddressToPubKey(array[1])
                     if (err != nil) {
                        fmt.Println("Invalid Address!")
                        continue
                     }
                     delegators, err := getNumberOfDelegators(array[1])
                     if (err != nil) {
                        fmt.Println("RCP error: ", err.Error())
                        continue
                     }
                     fmt.Println()
                     fmt.Println("Number of delegators for", array[1], ":")
                     fmt.Print  ("  count: ", delegators)
                     fmt.Println()

                  } else {
                     fmt.Println("Not enough arguments... Expecting a nano address.")
                  }
               case "frontier_count":
                  count, err := getFrontierCount()
                  if (err != nil) {
                     fmt.Println("RCP error: ", err.Error())
                     continue
                  }
                  fmt.Println()
                  fmt.Println("Frontiers: ", count)
                  fmt.Println()
               case "peers":
                  err := printPeers()
                  if (err != nil) {
                     fmt.Println("RCP error: ", err.Error())
                     continue
                  }
               case "receivable":
                  if (len(array) >= 2) {
                     // validate the address
                     _, err := keyMan.AddressToPubKey(array[1])
                     if (err != nil) {
                        fmt.Println("Invalid Address!")
                        continue
                     }
                     var count = -1
                     if (len(array) >= 3) {
                        count, _ = strconv.Atoi(array[2])
                     }
                     blocks, err := getReceivable(array[1], count)
                     if (err != nil) {
                        fmt.Println("RCP error: ", err.Error())
                        continue
                     }
                     fmt.Println()
                     for _, block := range blocks {
                        fmt.Println("  ", block)
                     }
                     fmt.Println()

                  } else {
                     fmt.Println("Not enough arguments... Expecting a nano address.")
                  }
               case "republish":
                  if (len(array) >= 2) {
                     hash, err := hex.DecodeString(array[1])
                     if (err != nil) {
                        fmt.Println("Invalid hash!")
                        continue
                     }
                     err = republish(hash)
                     if (err != nil) {
                        fmt.Println("RCP error: ", err.Error())
                        continue
                     }
                     fmt.Println()
                  } else {
                     fmt.Println("Not enough arguments... Expecting a block hash")
                  }
               case "stats":
                  if (len(array) >= 2) {
                     if (array[1] == "counters" ||
                         array[1] == "samples" ||
                         array[1] == "objects" ||
                         array[1] == "database") {
                        err := printStats(array[1])
                        if (err != nil) {
                           fmt.Println("RCP error: ", err.Error())
                           continue
                        }
                     } else {
                        fmt.Println("Bad argument... Expecting \"counters\", \"samples\", \"objects\", or \"database.\"")
                     }
                  } else {
                     fmt.Println("Not enough arguments... Expecting \"counters\", \"samples\", \"objects\", or \"database.\"")
                  }
               case "successors":
                  if (len(array) >= 2) {
                     hash, err := hex.DecodeString(array[1])
                     if (err != nil) {
                        fmt.Println("Invalid hash!")
                        continue
                     }
                     var count = -1
                     if (len(array) >= 3) {
                        count, _ = strconv.Atoi(array[2])
                     }
                     blocks, err := getSuccessors(hash, count)
                     if (err != nil) {
                        fmt.Println("RCP error: ", err.Error())
                        continue
                     }
                     for i, block := range blocks {
                        fmt.Println("  successor", i, ":", block)
                     }
                     fmt.Println()
                  } else {
                     fmt.Println("Not enough arguments... Expecting a block hash")
                  }
               case "telemetry":
                  err := printTelemetry()
                  if (err != nil) {
                     fmt.Println("RCP error: ", err.Error())
                     continue
                  }
               case "version":
                  err := printVersion()
                  if (err != nil) {
                     fmt.Println("RCP error: ", err.Error())
                     continue
                  }
               case "unchecked":
                  var count = 1
                  if (len(array) >= 2) {
                     count, _ = strconv.Atoi(array[1])
                  }
                  verboseSave := verbosity
                  verbosity = 9
                  _, err := getUncheckedBlocks(count)
                  verbosity = verboseSave
                  if (err != nil) {
                     fmt.Println("RCP error: ", err.Error())
                     continue
                  }
               case "uptime":
                  secs, err := getUptime()
                  if (err != nil) {
                     fmt.Println("RCP error: ", err.Error())
                     continue
                  }
                  fmt.Println()
                  fmt.Println("Uptime is", secs, "seconds.")
                  if (secs > 3600) {
                     hours := secs / 3600
                     mins := secs / 60 % 60
                     fmt.Println("That's", hours, "hours and", mins, "minutes.")
                  } else if (secs > 180) {
                     mins := secs / 60
                     fmt.Println("That's", mins, "minutes.")
                  }
                  fmt.Println()
               case "verbosity":
                  if (len(array) >= 2 && len(array[1]) > 0) {
                     verbosity, _ = strconv.Atoi(array[1])
                  } else {
                     fmt.Println(" Verbosity: ", verbosity)
                  }
               case "q":
                  fallthrough
               case "exit":
                  break RCPMenu
               default:
                  fmt.Println(array[0], "is not a recognized command.")
            }

         }
      case "3":
         DBMenu:
         for {
            fmt.Println()
            line, err := DB.Readline()
            if (err != nil) {
               break
            }
            array := strings.Split(strings.ToLower(line), " ")
            fmt.Println()

            switch(array[0]) {
               case "select":
                  if (len(array) >= 2) {
                     switch(array[1]) {
                        case "seeds":
                        case "wallets":
                           CLIprintWallets()
                           rows, err := getAllWalletRowsFromDatabase()
                           if (err != nil) {
                              fmt.Println(fmt.Errorf("CLI: %w", err))
                              continue
                           }
                           printPsqlRows(rows)
                        default:
                           fmt.Println("Table not recognized")
                     }
                  } else {
                     fmt.Println("Not enough arguments... Expecting a table.")
                  }
               case "verbosity":
                  if (len(array) >= 2 && len(array[1]) > 0) {
                     verbosity, _ = strconv.Atoi(array[1])
                  } else {
                     fmt.Println(" Verbosity: ", verbosity)
                  }
               case "q":
                  fallthrough
               case "exit":
                  break DBMenu
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

func CLIsend(myKey *keyMan.Key, args []string, prompt *string) {
   toPubKey, err := keyMan.AddressToPubKey(args[2])
   if (err != nil) {
      fmt.Println(args[2], "is not a valid address")
      return
   }

   amountNano, err := strconv.ParseFloat(args[1], 64)
   if (err != nil) {
      fmt.Println(args[1], "is not a valid nano amount")
      return
   }

   amountRaw := nt.NewRawFromNano(amountNano)
   _, err = Send(myKey, toPubKey, amountRaw, nil, nil, -1)
   if (err != nil) {
      fmt.Println(fmt.Errorf("CLIsend: %w", err))
   }

   rawBalance, _ := getBalance(myKey.NanoAddress)
   NanoBalance := rawToNANO(rawBalance)
   format := fmt.Sprintf("(%.3f)%s> ", NanoBalance, myKey.NanoAddress)

   *prompt = format
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
      var dummy1 string
      var dummy2 bool
      var dummy3 bool

      rows.Scan(&seed, &index, balance, &dummy1, &dummy2, &dummy3)

      if (ignoreZeroBalance && balance.Cmp(nt.NewRaw(0)) == 0) {
         continue
      }

      balanceInNano := rawToNANO(balance)

      fmt.Print(seed, ",", index, ":  Ӿ ", balanceInNano, "\n")
   }

   return nil
}

func CLIselect(myKey *keyMan.Key, args []string, prompt *string) error {

   seed, err := strconv.Atoi(args[1])
   if (err != nil) {
      return fmt.Errorf("CLIselect: %w", err)
   }

   index, err := strconv.Atoi(args[2])
   if (err != nil) {
      return fmt.Errorf("CLIselect: %w", err)
   }

   newKey, err := getSeedFromIndex(seed, index)
   if (err != nil) {
      return fmt.Errorf("CLIselect: %w", err)
   }

   *myKey = *newKey

   rawBalance, _ := getBalance(myKey.NanoAddress)
   NanoBalance := rawToNANO(rawBalance)
   format := fmt.Sprintf("(%.3f)%s> ", NanoBalance, myKey.NanoAddress)

   *prompt = format

   return nil
}

func CLIreceive(myKey *keyMan.Key, args []string, prompt *string) error {

   raw, block, numLeft, err := Receive(myKey.NanoAddress)
   if (err != nil) {
      return fmt.Errorf("CLIreceive: %w", err)
   }
   if (len(block) == 0) {
      fmt.Println("Nothing to receive!")
      return nil
   }
   fmt.Print  ("   Received ", raw, " (", rawToNANO(raw), ")\n")
   fmt.Println("   block:", block)
   fmt.Println("   receives remaining:", numLeft)
   fmt.Println()

   rawBalance, _ := getBalance(myKey.NanoAddress)
   NanoBalance := rawToNANO(rawBalance)
   format := fmt.Sprintf("(%.3f)%s> ", NanoBalance, myKey.NanoAddress)

   *prompt = format

   return nil
}

func CLInew(myKey *keyMan.Key, args []string, prompt *string) error {

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

      newKey, err := getSeedFromIndex(seed, key.Index)
      if (err != nil) {
         return fmt.Errorf("CLInew: %w", err)
      }

      *myKey = *newKey
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

   *prompt = format

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

func CLIprintWallets() {
   rows, err := getWalletRowsFromDatabase()
   if (err != nil) {
      fmt.Println(fmt.Errorf("CLI: %w", err))
      return
   }
   var lengths [2]int
   // Go through once to count, and then again to print.
   for rows.Next() {
      var seed int
      var index int
      var balance = nt.NewRaw(0)
      var pow string
      var inUse bool
      var receiveOnly bool

      rows.Scan(&seed, &index, balance, &pow, &inUse, &receiveOnly)
      idlen := len(strconv.Itoa(seed) +"-"+ strconv.Itoa(index))
      if (lengths[0] < idlen + 1) {
         lengths[0] = idlen + 1
      }
      ballen := len(balance.String() +" ("+ fmt.Sprintf("%f", rawToNANO(balance)) +")")
      if (lengths[1] < ballen + 1) {
         lengths[1] = ballen + 1
      }
   }
   rows, err = getAllWalletRowsFromDatabase()
   if (err != nil) {
      fmt.Println(fmt.Errorf("CLI: %w", err))
      return
   }
   afterPad1 := strings.Repeat(" ", (lengths[0]-1)/2)
   beforePad1 := strings.Repeat(" ", (lengths[0]-1) - (lengths[0]-1)/2)
   afterPad2 := strings.Repeat(" ", (lengths[1]-6)/2)
   beforePad2 := strings.Repeat(" ", (lengths[1]-6) - (lengths[1]-6)/2)
   fmt.Print(beforePad1 +"ID"+ afterPad1 +
   "|"+ beforePad2 +"balance"+ afterPad2 +
   "|       pow        | use | receive \n")
   fmt.Print(strings.Repeat("-", lengths[0]+1) +"+"+
   strings.Repeat("-", lengths[1]+1) +
   "+------------------+-----+---------\n")

   for rows.Next() {
      var seed int
      var index int
      var balance = nt.NewRaw(0)
      var pow string
      var inUse bool
      var receiveOnly bool

      rows.Scan(&seed, &index, balance, &pow, &inUse, &receiveOnly)
      id := strconv.Itoa(seed) +"-"+ strconv.Itoa(index)
      bal := balance.String() +" ("+ fmt.Sprintf("%f", rawToNANO(balance)) +")"
      idPadding := strings.Repeat(" ", lengths[0] - len(id))
      balPadding := strings.Repeat(" ", lengths[1] - len(bal))
      var useString string
      if (inUse) {
         useString = "t"
      } else {
         useString = "f"
      }
      var receiveString string
      if (inUse) {
         receiveString = "t"
      } else {
         receiveString = "f"
      }
      if (len(pow) == 0) {
         pow = "                "
      }

      fmt.Print(idPadding + id + " |")
      fmt.Print(balPadding + bal + " |")
      fmt.Print(" "+ pow +" |")
      fmt.Print("  "+ useString +"  |")
      fmt.Print("   "+ receiveString +"\n")
   }
   fmt.Println()
}

func printPsqlRows(rows pgx.Rows) {

   blarg := rows.FieldDescriptions()
   fmt.Println(len(blarg))
   fmt.Println(string(blarg[0].Name))
   //fmt.Println(rows.FieldDescriptions())
   //fmt.Println(rows.Values())
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
   readline.PcItem("receive"),
   readline.PcItem("receiveonly",
      readline.PcItem("off"),
   ),
   readline.PcItem("help"),
   readline.PcItem("exit"),
   readline.PcItem("verbosity"),
)

var RCPCompleter = readline.NewPrefixCompleter(
   readline.PcItem("account_balance"),
   readline.PcItem("account_block_count"),
   readline.PcItem("account_history"),
   readline.PcItem("account_info"),
   readline.PcItem("account_representative"),
   readline.PcItem("account_weight"),
   readline.PcItem("available_supply"),
   readline.PcItem("block_account"),
   readline.PcItem("block_count"),
   readline.PcItem("block_info"),
   readline.PcItem("bootstrap_status"),
   readline.PcItem("chain"),
   readline.PcItem("confirmation_active",
      readline.PcItem("showblocks"),
   ),
   readline.PcItem("verbosity"),
   readline.PcItem("confirmation_history"),
   readline.PcItem("confirmation_quorum"),
   readline.PcItem("delgators_count"),
   readline.PcItem("frontier_count"),
   readline.PcItem("peers"),
   readline.PcItem("receivable"),
   readline.PcItem("republish"),
   readline.PcItem("stats",
      readline.PcItem("counters"),
      readline.PcItem("samples"),
      readline.PcItem("objects"),
      readline.PcItem("database"),
   ),
   readline.PcItem("successors"),
   readline.PcItem("telemetry"),
   readline.PcItem("version"),
   readline.PcItem("unchecked"),
   readline.PcItem("uptime"),
)

var DBCompleter = readline.NewPrefixCompleter(
   readline.PcItem("select",
      readline.PcItem("seeds"),
      readline.PcItem("wallets"),
      readline.PcItem("blacklist"),
      readline.PcItem("profitrecord"),
   ),
   readline.PcItem("verbosity"),
)
