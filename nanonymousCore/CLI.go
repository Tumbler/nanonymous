package main

import (
   "fmt"
   "context"
   "encoding/hex"
   "strings"
   "strconv"
   "time"
   _"embed"

   // Local packages
   keyMan "nanoKeyManager"
   nt "nanoTypes"

   // 3rd party packages
   pgx "github.com/jackc/pgx/v4"
   "github.com/jackc/pgtype"
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
               case "refresh":
                  rawBalance, _ := getBalance(myKey.NanoAddress)
                  NanoBalance := rawToNANO(rawBalance)
                  format := fmt.Sprintf("(%.3f)%s> ", NanoBalance, myKey.NanoAddress)

                  prompt = format
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
               case "clearpow":
                  err = CLIclearPoW(myKey, array)
                  if (err != nil) {
                     fmt.Println(fmt.Errorf("CLI: %w", err))
                  }
               case "count":
                  Nano, managed, mixer, err := findTotalBalance()
                  if (err != nil) {
                     fmt.Println(fmt.Errorf("CLI: %w", err))
                  }
                  fmt.Println("Nano: Ӿ", rawToNANO(Nano), "\nManaged: Ӿ", rawToNANO(managed), "\nMixer: Ӿ", rawToNANO(mixer))
               case "mix":
                  err := sendToMixer(myKey , 1)
                  if (err != nil) {
                     fmt.Println(fmt.Errorf("CLI: %w", err))
                  }
               case "extract":
                  CLIextract(array)
               case "update":
                  err = CLIupdate(myKey, array, &prompt)
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
               case "":
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
                           // TODO copy without accessing DB???
                           rows, conn, err := getEncryptedSeedRowsFromDatabase()
                           rowsCopy, conn2, err := getEncryptedSeedRowsFromDatabase()
                           if (err != nil) {
                              fmt.Println(fmt.Errorf("CLI: %w", err))
                              conn.Close(context.Background())
                              conn2.Close(context.Background())
                              continue
                           }
                           printPsqlRows(rows, rowsCopy)
                           rows.Close()
                           rowsCopy.Close()
                           conn.Close(context.Background())
                           conn2.Close(context.Background())
                        case "wallets":
                           rows, conn, err := getAllWalletRowsFromDatabase()
                           rowsCopy, conn2, err := getAllWalletRowsFromDatabase()
                           if (err != nil) {
                              fmt.Println(fmt.Errorf("CLI: %w", err))
                              conn.Close(context.Background())
                              conn2.Close(context.Background())
                              continue
                           }
                           printPsqlRows(rows, rowsCopy)
                           rows.Close()
                           rowsCopy.Close()
                           conn.Close(context.Background())
                           conn2.Close(context.Background())
                        case "blacklist":
                           rows, conn, err := getBlacklistRowsFromDatabase()
                           rowsCopy, conn2, err := getBlacklistRowsFromDatabase()
                           if (err != nil) {
                              fmt.Println(fmt.Errorf("CLI: %w", err))
                              conn.Close(context.Background())
                              conn2.Close(context.Background())
                              continue
                           }
                           printPsqlRows(rows, rowsCopy)
                           rows.Close()
                           rowsCopy.Close()
                           conn.Close(context.Background())
                           conn2.Close(context.Background())
                        case "profitrecord":
                           rows, conn, err := getProfitRowsFromDatabase()
                           rowsCopy, conn2, err := getProfitRowsFromDatabase()
                           if (err != nil) {
                              fmt.Println(fmt.Errorf("CLI: %w", err))
                              conn.Close(context.Background())
                              conn2.Close(context.Background())
                              continue
                           }
                           printPsqlRows(rows, rowsCopy)
                           rows.Close()
                           rowsCopy.Close()
                           conn.Close(context.Background())
                           conn2.Close(context.Background())
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
         var dirtyAddress *keyMan.Key
         if (dirtyAddress != nil) {
            fmt.Println("len:", len(dirtyAddress.NanoAddress))
         }

      case "6":

         _, blarg, _ := getNewAddress("", false, true, 0)
         fmt.Println("index:", blarg)
         _, blarg, _  = getNewAddress("", false, true, 0)
         fmt.Println("index:", blarg)
         _, blarg, _  = getNewAddress("", false, true, 0)
         fmt.Println("index:", blarg)
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
         //verbosity = 5
//
         //for i := 0; i <= 41; i++ {
            //fmt.Println("--------------", i, "-------------")
            //seedReceive, _ := getSeedFromIndex(1, i)
            //err := ReceiveAll(seedReceive.NanoAddress)
            //if (err != nil) {
               //fmt.Println("Error:", err.Error())
            //}
         //}
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
      if (strings.Contains(err.Error(), "work")) {
         // Problem with work, try again.
         if (verbosity >= 5) {
            fmt.Println("Problem with work, trying again...")
         }
         clearPoW(myKey.NanoAddress)
         _, err = Send(myKey, toPubKey, amountRaw, nil, nil, -1)
         if (err != nil) {
            fmt.Println(fmt.Errorf("CLIsend: %w", err))
         }
      } else {
         fmt.Println(fmt.Errorf("CLIsend: %w", err))
      }
   }

   rawBalance, _ := getBalance(myKey.NanoAddress)
   NanoBalance := rawToNANO(rawBalance)
   format := fmt.Sprintf("(%.3f)%s> ", NanoBalance, myKey.NanoAddress)

   *prompt = format
}

// TODO This function randomly doesn't show all the rows. I don't know why.
func CLIlist(args []string) error {
   var err error
   var rows pgx.Rows
   var conn *pgx.Conn

   var ignoreZeroBalance bool
   if !(contains(args, "-z")) {
      ignoreZeroBalance = true
   }

   if (contains(args, "-a")) {
      rows, conn, err = getAllWalletRowsFromDatabase()
   } else {
      rows, conn, err = getWalletRowsFromDatabase()
   }

   if (err != nil) {
      return fmt.Errorf("CLIlist: %w", err)
   }

   for (rows.Next()) {
      var seed int
      var index int
      var balance = nt.NewRaw(0)
      var receiveOnly bool
      var mixer bool

      err = rows.Scan(&seed, &index, balance, nil, nil, &receiveOnly, &mixer)
      if (err != nil) {
         fmt.Println("CLILIST:", err)
         return fmt.Errorf("CLIList: %w", err)
      }

      if (ignoreZeroBalance && balance.Cmp(nt.NewRaw(0)) == 0) {
         continue
      }

      balanceInNano := rawToNANO(balance)

      var text = ""
      if (receiveOnly) {
         text = ", receive_only"
      }
      if (mixer) {
         text += ", mixer"
      }

      fmt.Print(seed, ",", index, ":  Ӿ ", balanceInNano, text, "\n")
   }

   if (conn != nil) {
      conn.Close(context.Background())
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

      key, seed, err := getNewAddress("", receiveOnly, false, seedId)
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

// TODO make program not crash if you don't have enough arguments
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

func CLIextract(args []string) {
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

   _, err = extractFromMixer(amountRaw, toPubKey)
   if (err != nil) {
      fmt.Println(fmt.Errorf("Problem with extraction: %w", err))
   }
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

func CLIclearPoW(myKey *keyMan.Key, args []string) error {

   err := clearPoW(myKey.NanoAddress)
   if (err != nil) {
      return fmt.Errorf("CLIclearPoW: %w", err)
   }

   fmt.Println("PoW cleared")

   return nil
}

func CLIupdate(myKey *keyMan.Key, args []string, prompt *string) error {
   checkBalance(myKey.NanoAddress)

   rawBalance, _ := getBalance(myKey.NanoAddress)
   NanoBalance := rawToNANO(rawBalance)
   format := fmt.Sprintf("(%.3f)%s> ", NanoBalance, myKey.NanoAddress)

   *prompt = format

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
      case "extract":
         CLIhelpextract()
      case "mix":
         CLIhelpmix()
      case "count":
         CLIhelpcount()
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
      case "receive":
         CLIhelpreceive()
      case "receiveonly":
         CLIhelpreceiveOnly()
      case "refresh":
         CLIhelprefresh()
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
         CLIhelpreceive()
         CLIhelpreceiveOnly()
         CLIhelpselect()
         CLIhelpsend()
         CLIhelpextract()
         CLIhelpmix()
         CLIhelpcount()
         CLIhelprefresh()
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

func CLIhelpextract() {
   fmt.Println()
   fmt.Print("   - extract {Nano} {address} | Sends \"Nano\" amount from the any number of mixer\n",
             "                                addresses to \"address\"",
   )
}

func CLIhelpmix() {
   fmt.Println()
   fmt.Print("   - mix | Splits up any funds in the current wallet and sends it to mixer addresses\n",)
}

func CLIhelpcount() {
   fmt.Println()
   fmt.Print("   - count | Displays all stored nano on all wallets\n",)
}

func CLIhelpselect() {
   fmt.Println()
   fmt.Print("   - set/select {seed} {index} | Sets the current wallet to indicated seed and index\n")
}

func CLIhelpreceive() {
   fmt.Println()
   fmt.Print("   - receive | Finds the next receivable hash for the active address and receives it\n",)
}

func CLIhelpreceiveOnly() {
   fmt.Println()
   fmt.Print("   - receiveonly {seed} {index} [\"off\"] | Sets the wallet indicated by seed and index to\n",
             "                                        be receive only\n",
             "                        Flags: off    | When set it will set the indicated wallet to NOT\n",
             "                                        be receive only\n",
   )
}

func CLIhelprefresh() {
   fmt.Println()
   fmt.Print("   - refresh | Syncs the current wallet with the database\n")
}

func printPsqlRows(rows pgx.Rows, rowsCopy pgx.Rows) {

   const INT = 23
   const BIGINT = 20
   const NUMERIC = 1700
   const STRING = 25
   const BOOL = 16
   const HASH = 17
   const TIMESTAMP = 1184
   const FLOAT = 701

   const TIME_FORMAT = "2006-01-02_15:04:05.000"

   titles := rows.FieldDescriptions()
   var lengths = make([]int, len(titles))

   // Find all lengths
   for i, header := range titles {
      lengths[i] = len(string(header.Name))
   }
   for (rows.Next()) {
      values, err := rows.Values()
      if (err != nil) {
         fmt.Println("Error :", err.Error())
      }
      for i, value := range values {
         switch(titles[i].DataTypeOID) {
            case BIGINT:
               if (lengths[i] < len(strconv.Itoa(int(value.(int64))))) {
                  lengths[i] = len(strconv.Itoa(int(value.(int64))))
               }
            case INT:
               if (lengths[i] < len(strconv.Itoa(int(value.(int32))))) {
                  lengths[i] = len(strconv.Itoa(int(value.(int32))))
               }
            case FLOAT:
               if (lengths[i] < len(fmt.Sprintf("%.2f", (value.(float64))))) {
                  lengths[i] = len(fmt.Sprintf("%.2f", (value.(float64))))
               }
            case NUMERIC:
               numberString := value.(pgtype.Numeric).Int.String() + strings.Repeat("0", int(value.(pgtype.Numeric).Exp))
               raw, _ := nt.OneNano().SetString(numberString, 10)
               numberString += " ("+ fmt.Sprintf("%f", rawToNANO(raw)) + ")"
               if (lengths[i] < len(numberString)) {
                  lengths[i] = len(numberString)
               }
            case STRING:
               if (lengths[i] < len(value.(string))) {
                  lengths[i] = len(value.(string))
               }
            case HASH:
               if (string(titles[i].Name) == "seed") {
                  if (lengths[i] < len("lol no")) {
                     lengths[i] = len("lol no")
                  }
               } else {
                  stringRep := "\\x"+ hex.EncodeToString(value.([]uint8))
                  if (lengths[i] < len(stringRep)) {
                     lengths[i] = len(stringRep)
                  }
               }
            case TIMESTAMP:
               if (lengths[i] < len(value.(time.Time).Format(TIME_FORMAT))) {
                  lengths[i] = len(value.(time.Time).Format(TIME_FORMAT))
               }
            case BOOL:
               // Max length of one so title will always be the correct length
         }
      }

   }

   // Print header
   for i, header := range titles {
      if (i > 0) {
         fmt.Print("|")
      }
      fmt.Print(headerSpacing(string(header.Name), lengths[i]))
   }
   fmt.Println()
   for i := 0; i < len(titles); i++ {
      if (i > 0) {
         fmt.Print("+")
      }
      fmt.Print(strings.Repeat("-", lengths[i]+2))
   }
   fmt.Println()

   // Print table
   for (rowsCopy.Next()) {
      values, _ := rowsCopy.Values()
      for i, value := range values {
         if (i > 0) {
            fmt.Print("|")
         }
         switch(titles[i].DataTypeOID) {
            case BIGINT:
               fmt.Print(" ", spacing(strconv.Itoa(int(value.(int64))), lengths[i]), " ")
            case INT:
               fmt.Print(" ", spacing(strconv.Itoa(int(value.(int32))), lengths[i]), " ")
            case FLOAT:
               fmt.Print(" ", spacing(fmt.Sprintf("%.2f", (value.(float64))), lengths[i]), " ")
            case NUMERIC:
               // Convert Numeric to string
               numberString := value.(pgtype.Numeric).Int.String() + strings.Repeat("0", int(value.(pgtype.Numeric).Exp))
               raw, _ := nt.OneNano().SetString(numberString, 10)
               numberString += " ("+ fmt.Sprintf("%f", rawToNANO(raw)) + ")"
               fmt.Print(" ", spacing(numberString, lengths[i]), " ")
            case STRING:
               fmt.Print(" ", spacing(value.(string), lengths[i]), " ")
            case HASH:
               if (string(titles[i].Name) == "seed") {
                  fmt.Print(" lol no ")
               } else {
                  stringRep := hex.EncodeToString(value.([]uint8))
                  fmt.Print(" \\x", stringRep, " ")
               }
            case BOOL:
               if (value.(bool)) {
                  fmt.Print(" ", spacing("t", lengths[i]), " ")
               } else {
                  fmt.Print(" ", spacing("f", lengths[i]), " ")
               }
            case TIMESTAMP:
               fmt.Print(" ", spacing(value.(time.Time).Format(TIME_FORMAT), lengths[i]), " ")
         }
      }
      fmt.Println()
   }
}

func spacing(input string, desiredLen int) string {
   output := strings.Repeat(" ", desiredLen - len(input))
   output += input
   return output
}

func headerSpacing(input string, desiredLen int) string {
   // Account for spacing
   desiredLen += 2

   output := strings.Repeat(" ", (desiredLen - len(input)) - (desiredLen - len(input))/2)
   output += input
   output += strings.Repeat(" ", (desiredLen - len(input))/2)
   return output
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
   readline.PcItem("update"),
   readline.PcItem("clearpow"),
   readline.PcItem("help"),
   readline.PcItem("exit"),
   readline.PcItem("verbosity"),
   readline.PcItem("count"),
   readline.PcItem("mix"),
   readline.PcItem("extract"),
   readline.PcItem("refresh"),
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
