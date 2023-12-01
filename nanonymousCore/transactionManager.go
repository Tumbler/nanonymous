package main

import (
   "fmt"
   "sync"
   "errors"
   "time"
   "strings"
   "regexp"
   "strconv"
   "encoding/hex"

   // Local packages
   keyMan "nanoKeyManager"
   nt "nanoTypes"
)

type Transaction struct {
   id int
   paymentAddress []byte
   paymentParentSeedId int
   paymentIndex int
   payment *nt.Raw
   receiveHash nt.BlockHash
   receiveWg []sync.WaitGroup // One per subsend
   recipientAddress []byte
   fee *nt.Raw
   amountToSend []*nt.Raw // One per subsend.
   sendingKeys [][]*keyMan.Key // One per subsend.
   walletSeed [][]int // One per subsend.
   walletBalance [][]*nt.Raw // One per subsend.
   individualSendAmount [][]*nt.Raw // One per subsend.
   transitionalKey []*keyMan.Key // One per subsend.
   transitionSeedId []int // One per subsend.
   finalHash []nt.BlockHash // One per subsend.
   commChannel []chan transactionComm // One per subsend
   errChannel []chan error // One per subsend
   confirmationChannel []chan string // One per subsend
   percents []int
   delays []int
   bridge bool
   numSubSends int // This is how many individual sends the client has asked to split the transaction into.
   multiSend []bool // This is if we're using multiple wallets to satisfy a single send. One per subsend.
   dirtyAddress []int // The sendingKeys address that has been linked to other addresses but not blacklisted. One per subsend.
   transactionSuccessful []bool // One per subsend.
   abort bool
   abortchan chan int
}

func (t Transaction) String() string {
   var ret string

   pa, _ := keyMan.PubKeyToAddress(t.paymentAddress)
   ra, _ := keyMan.PubKeyToAddress(t.recipientAddress)

   ret =
   "\n  id: "+ strconv.Itoa(t.id) +
   "\n  paymentAddress: "+ pa +
   "\n  paymentID: "+ strconv.Itoa(t.paymentParentSeedId) +","+ strconv.Itoa(t.paymentIndex) +
   "\n  receiveHash: "+ t.receiveHash.String() +
   "\n  recipientAddress: "+ ra +
   "\n  fee: "+ t.fee.String() +
   "\n  percents: "+ fmt.Sprint(t.percents) +
   "\n  delays: "+ fmt.Sprint(t.delays) +
   "\n  amountToSend: "+ fmt.Sprint(t.amountToSend) +
   "\n  sendingKeys: "+ fmt.Sprint(t.sendingKeys) +
   "\n  walletSeed: "+ fmt.Sprint(t.walletSeed) +
   "\n  walletBalance: "+ fmt.Sprint(t.walletBalance) +
   "\n  individualSendAmount: "+ fmt.Sprint(t.individualSendAmount) +
   "\n  transitionalKey: "+ fmt.Sprint(t.transitionalKey) +
   "\n  finalHash: "+ fmt.Sprint(t.finalHash) +
   "\n  multiSend: "+ fmt.Sprint(t.multiSend) +
   "\n  dirtyAddress: "+ fmt.Sprint(t.dirtyAddress) +
   "\n  numOfSends: "+ fmt.Sprint(t.numSubSends) +
   "\n  transactionSuccessful: "+ fmt.Sprint(t.transactionSuccessful) +
   "\n  bridge: "+ fmt.Sprint(t.bridge) +
   "\n  abort: "+ strconv.FormatBool(t.abort)

   return ret
}

type transactionComm struct {
   i int
   hashes []nt.BlockHash
}

var registeredClientComunicationPipes map[string]chan string

var databaseError = errors.New("database error")

const RetryNumber = 3

// transactionManager tracks all on-chain actions spawned by receivedNano(),
// coordinates send/receives, and handles refunds if something goes wrong.
func transactionManager(t *Transaction) {

   // WARNING: use wg.Add(1) before calling
   defer wg.Done()

   address, _ := keyMan.PubKeyToAddress(t.paymentAddress)

   // We have a lot of clean up to do
   defer func() {
      recoverMessage := recover()
      if (recoverMessage != nil) {
         Error.Println("transactionManager panic: ", recoverMessage)
      }

      // If we're doing a safe exit then leave the transaction intact to be
      // picked up next time we start.
      if (!safeExit) {
         // Remove active transaction
         err := setRecipientAddress(t.paymentParentSeedId, t.paymentIndex, nil, false, []int{}, []int{})
         if (err != nil) {
            Warning.Println("defer transactionManager: ", err.Error())
         }

         // Cancel all the things
         if !(fullTransactionWasSuccessfull(t.transactionSuccessful)) {
            sendInfoToClient("info=There was an internal error. Your transaction has been refunded.", t.paymentAddress)
            err := checkPartialRefund(t)
            if (err != nil) {
               // VERY BAD! Just accepted money, failed to deliver it, and didn't
               //           refund the user!
               nanoAddress, _ := keyMan.PubKeyToAddress(t.paymentAddress)
               Error.Println("Refund failed!! Address:", nanoAddress, " error:", err.Error())
               sendEmail("IMMEDIATE ATTENTION REQUIRED", "Refund failed!! Address: "+ nanoAddress +" error: "+ err.Error() +
                  "\n\nPayment Hash: "+ t.receiveHash.String() +
                  "\nID: "+ strconv.Itoa(t.paymentParentSeedId) +","+ strconv.Itoa(t.paymentIndex) +
                  "\nAmount: "+ strconv.FormatFloat(rawToNANO(t.payment), 'f', -1, 64))
            }
            mixTransitionalAddresses(t)
            t.abort = true
            for i, _ := range t.receiveWg {
               t.receiveWg[i].Done()
            }
            if (t.numSubSends > 1) {
               Warning.Println("Sub send transaction failed (", t.id, ")")
            } else if (t.multiSend[0]) {
               Warning.Println("Multi transaction failed (", t.id, ")")
            } else {
               Warning.Println("Transaction failed (", t.id, ")")
            }
            if (verbosity >= 5) {
               fmt.Println("Transaction failed...")
            }
         } else {
            recordProfit(t.fee, t.id)
            Info.Println("Transaction", t.id, "Complete")

            for _, hash := range t.finalHash {
               // There was a bug that I saw occasionally where it would send a blank
               // hash to the client. I haven't been able to reproduce it recently,
               // but if it happens, log it and try to track it down.
               if (len(hash) < 32) {
                  Warning.Println("Final hash for transaction is blank:\n", t)
                  sendEmail("WARNInG", "Final hash for transaction is blank:\n"+ t.String())
               }
            }
            if (t.bridge) {
               // Final hash would leak recipients address to sender. Redact it.
               for i, _ := range t.finalHash {
                  t.finalHash[i] = []byte("COFFEE")
               }
            }
            sendFinalHash(t.finalHash, t.paymentAddress)

            // Send any dirty addresses to the mixer.
            for i, dirtyAddress := range t.dirtyAddress {
               if (dirtyAddress != -1) {
                  if (verbosity >= 8) {
                     fmt.Println("Sending dirty address to mixer")
                  }
                  err := sendToMixer(t.sendingKeys[i][dirtyAddress], 1)

                  if (err != nil) {
                     Error.Println("Mixer Error:", err.Error())
                     sendEmail("WARNING", "Mixer Error: "+ err.Error())
                  }
               }
            }
         }

         // Un-mark all addresses
         setAddressNotInUse(address)
         for _, sendingKeys := range t.sendingKeys {
            for _, key := range sendingKeys {
               setAddressNotInUse(key.NanoAddress)
            }
         }
         for _, transitionalKey := range t.transitionalKey {
            if (transitionalKey != nil) {
               setAddressNotInUse(transitionalKey.NanoAddress)
            }
         }

         err = deleteTransactionRecord(t.id)
         if (err != nil) {
            Warning.Println("Delayed transaction delete failed:", err)
         }
      } else {
         if (len(t.delays) > 0) {
            fmt.Println("Transaction stopped, but saved to db")
         }
      }

   }()

   var subWait sync.WaitGroup
   // Start up a mini manager for every sub-send
   t.receiveWg = make([]sync.WaitGroup, t.numSubSends)
   for i := 0; i < t.numSubSends; i++ {
      t.receiveWg[i].Add(1)

      // Check to see if prevous info has been loaded or we need to init new arrays
      if (len(t.transactionSuccessful) == t.numSubSends) {
         if (t.transactionSuccessful[i]) {
            // Transaction has already been completed. Skip this one.
            continue
         }
      } else {
         // Init new stuff
         t.transactionSuccessful = append(t.transactionSuccessful, false)
         t.finalHash = append(t.finalHash, make([]byte, 0))
      }

      subWait.Add(1)
      go monitorSubSend(t, &subWait, i)

   }

   subWait.Wait()

   if (verbosity >= 5 && fullTransactionWasSuccessfull(t.transactionSuccessful)) {
      fmt.Println("Transaction Complete!")
   }
}

func monitorSubSend(t *Transaction, tWait *sync.WaitGroup, subSend int) {
   defer func() {
      recoverMessage := recover()
      if (recoverMessage != nil) {
         Error.Println("monitorSubSend panic: ", recoverMessage)
      }
   }()

   defer tWait.Done()
   var numDone = 0
   var operation = 0

   if (t.abort) {
      return
   }

   defer updateDelayRecords(t)

   var initialWait time.Duration
   if (len(t.delays) > subSend) {
      // 5 minutes + the delay time.
      initialWait = (time.Duration(5 * 60 + t.delays[subSend]) * time.Second)
   } else {
      initialWait = (5 * time.Minute)
   }

   // Waiting until first send
   select {
      case <-t.commChannel[subSend]:
         // Proceed to next step
         if (t.multiSend[subSend]) {
            t.confirmationChannel[subSend] = make(chan string)
            registerConfirmationListener(t.transitionalKey[subSend].NanoAddress, t.confirmationChannel[subSend], "send")
            defer unregisterConfirmationListener(t.transitionalKey[subSend].NanoAddress, "send")
         }
      case err := <-t.errChannel[subSend]:
         // There was a problem.
         if (verbosity >= 5) {
            fmt.Println("Error:", err.Error())
         }
         return
      case <-time.After(initialWait):
         // Timeout.
         Info.Println("Transaction timeout(0)")
         return
      case <-t.abortchan:
         if (verbosity >= 5) {
            fmt.Println("1 Aborting early on subSend", subSend)
         }
         return
      case <-safeExitChan:
         // If we're waiting for the first send and we get a safeExit then exit.
         // Otherwise let the subsend finish before exiting.
         if (verbosity >= 5) {
            fmt.Println("Safe exit early on subSend", subSend)
         }
         return
   }

   // First manage all initial sends
   numOfSends := len(t.sendingKeys[subSend])
   for {
      // All sends finished
      if (numDone >= numOfSends) {
         if !(t.multiSend[subSend]) {
            t.transactionSuccessful[subSend] = true
         }
         if (verbosity >= 5) {
            fmt.Println("Done with Sends!")
         }

         break
      }
      select {
         case i := <-t.commChannel[subSend]:
            // A send finished with no errors
            numDone++

            if !(t.multiSend[subSend]) {
               t.finalHash[subSend] = i.hashes[0]
            }

            reverseBlacklist(t, subSend, i.i)
         case err := <-t.errChannel[subSend]:
            // There was an error. Deal with it.
            if (verbosity >= 5) {
               fmt.Println("Error with sends", err.Error())
            }
            if (t.multiSend[subSend]) {
               findIndex, _ := regexp.Compile(">>([0-9]+)<<")
               regexResults := findIndex.FindSubmatch([]byte(err.Error()))
               var whichSend int
               if (len(regexResults) > 1) {
                  whichSend, _ = strconv.Atoi(string(findIndex.FindSubmatch([]byte(err.Error()))[1]))
               }

               if (handleMultiSendError(t, operation, whichSend, subSend, err)) {
                  // We recovered so go back to regular operation
                  numDone++
               } else {
                  // Abort rest of transaction
                  t.abort = true
                  broadcastAbort(t)
                  return
               }
            } else {
               if (handleSingleSendError(t, subSend, err)) {
                  // We recovered so go back to regular operation
                  numDone++
               } else {
                  // Abort transaction
                  t.abort = true
                  broadcastAbort(t)
                  return
               }
            }
         case <-time.After(5 * time.Minute):
            Info.Println("Transaction timeout(1)")
            if (verbosity >= 5) {
               if (t.multiSend[subSend]) {
                  fmt.Println("Transaction error: timout during sends")
               } else {
                  fmt.Println("Transaction error: timout during single send")
               }
            }
            return
         case <-t.abortchan:
            // The send to the recipient needs to be able to finish to make sure
            // we don't send funds and then also refund them.
            fmt.Println("Received abort from someone")
            if (t.multiSend[subSend]) {
               if (verbosity >= 5) {
                  fmt.Println("3 Aborting early on subSend", subSend)
               }
               return
            }
      }
   }

   if (t.abort) {
      if (verbosity >= 5) {
         fmt.Println("4 Aborting early on subSend", subSend)
      }
      return
   }

   // Sends are done, wait for them to be confirmed
   if (t.multiSend[subSend]) {
      trackConfirms := make(map[string]bool)
      var numConfirmed int

      for (numConfirmed < numOfSends) {
         if (inTesting) {
            break
         }

         select {
            case hash := <-t.confirmationChannel[subSend]:
               // Make sure we didn't receive the same block twice
               if (trackConfirms[hash] == false) {
                  trackConfirms[hash] = true
                  numConfirmed++
               }
               if (verbosity >= 5) {
                  fmt.Println("[S]Confirmed: ", numConfirmed)
               }
            case <-time.After(5 * time.Minute):
               Info.Println("Transaction timeout(2)")
               t.abort = true
               broadcastAbort(t)
               t.receiveWg[subSend].Done()
               return
            case <-t.abortchan:
               if (verbosity >= 5) {
                  fmt.Println("5 Aborting early on subSend", subSend)
               }
               return
         }
      }
      if (verbosity >= 5) {
         fmt.Println("All sends confirmed!")
      }
      // Signal receives to start in ReceiveAndSend()
      t.receiveWg[subSend].Done()
   }

   if (t.abort) {
      if (verbosity >= 5) {
         fmt.Println("6 Aborting early on subSend", subSend)
      }
      return
   }

   // Now if it's a multi send then we need monitor the last step of receiving
   // and sending to the recipient
   if (t.multiSend[subSend]) {
      operation = 1

      registerConfirmationListener(t.transitionalKey[subSend].NanoAddress, t.confirmationChannel[subSend], "receive")
      defer unregisterConfirmationListener(t.transitionalKey[subSend].NanoAddress, "receive")

      for (operation < 3) {
         select {
            case tComm := <-t.commChannel[subSend]:
               operation = tComm.i
               if (operation == 2) {
                  // Done with receives
                  if (verbosity >= 5) {
                     fmt.Println("Done with receives")
                  }

                  // Receives have been published, now wait for them to confirm

                  trackConfirms := make(map[string]bool)
                  for _, hash := range tComm.hashes {
                     trackConfirms[hash.String()] = false
                  }

                  var timeLimit = time.Now().Add(5 * time.Minute)

                  var numConfirmed int
                  for numConfirmed < numOfSends {
                     if (inTesting) {
                        break
                     }

                     select {
                        case hash := <-t.confirmationChannel[subSend]:
                           // Make sure we didn't receive the same block twice
                           if (trackConfirms[hash] == false) {
                              if (verbosity >= 9) {
                                 fmt.Println("hash:", hash)
                              }
                              trackConfirms[hash] = true
                              numConfirmed++
                           }
                           if (verbosity >= 5) {
                              fmt.Println("[R]Confirmed: ", numConfirmed)
                           }

                           // Time limit gets reset if we're still getting confirmations
                           timeLimit = time.Now().Add(5 * time.Minute)
                        case <-time.After(5 * time.Second):
                           // It's been some time, let's poll the hashes manually.
                           for hash, seen := range trackConfirms {
                              if (!seen) {
                                 encodedHash, _ := hex.DecodeString(hash)
                                 blockInfo, err := getBlockInfo(encodedHash)
                                 if (err == nil) {
                                    if (blockInfo.Confirmed) {
                                       trackConfirms[hash] = true
                                       numConfirmed++
                                       if (verbosity >= 5) {
                                          fmt.Println("[R]Confirmed: ", numConfirmed)
                                       }
                                    }
                                 }
                              }
                           }
                        case <-time.After(timeLimit.Sub(time.Now())):
                           Info.Println("Transaction timeout(3)")
                           t.abort = true
                           broadcastAbort(t)
                           t.receiveWg[subSend].Done()
                           return
                     }
                  }
                  t.receiveWg[subSend].Done()
               } else if (operation == 3) {
                  // All finished
                  t.finalHash[subSend] = tComm.hashes[0]
                  t.transactionSuccessful[subSend] = true
                  if (verbosity >= 5) {
                     fmt.Println("Done with everything!")
                  }
               }
            case err := <-t.errChannel[subSend]:
               // There was an error. Deal with it.
               if (verbosity >= 5) {
                  fmt.Println("Error with receives or final send", err.Error())
               }
               if (handleMultiSendError(t, operation, -1, subSend, err)) {
                  operation++
                  t.receiveWg[subSend].Done()
               } else {
                  // This is just to exit the loop
                  operation = 10
                  t.abort = true
                  broadcastAbort(t)
                  return
               }
            case <-time.After(5 * time.Minute):
               Info.Println("Transaction timeout(4)")
               if (verbosity >= 5) {
                  if (operation == 1) {
                     Info.Println("Transaction error: timout during receives")
                  } else {
                     Info.Println("Transaction error: timout during final send")
                  }
               }
               t.abort = true
               broadcastAbort(t)
               return
            case <-t.abortchan:
               // The send to the recipient needs to be able to finish to make sure
               // we don't send funds and then also refund them.
               if (operation == 1) {
                  if (verbosity >= 5) {
                     fmt.Println("7 Aborting early on subSend", subSend)
                  }
                  return
               }
         }
      }
   }
}

func handleSingleSendError(t *Transaction, subSend int, prevError error) bool {
   var retryCount = 0
   var err error

   // If there was some problem with PoW then regenerate it.
   if (prevError != nil && strings.Contains(prevError.Error(), "work")){
      clearPoW(t.sendingKeys[subSend][0].NanoAddress)
   }
   if (errors.Is(prevError, databaseError)) {
      // Just a database error. Update internal database and move on
      checkBalance(t.sendingKeys[subSend][0].NanoAddress)
      return true
   }

   for (retryCount < RetryNumber) {
      blockHash, err := Send(t.sendingKeys[subSend][0], t.recipientAddress, t.amountToSend[subSend], nil, nil, -1)
      if (err != nil) {
         if (verbosity >= 5) {
            fmt.Println("Error with resend: ", retryCount, err.Error())
         }
         retryCount++
      } else {
         t.finalHash[subSend] = blockHash
         reverseBlacklist(t, subSend, 0)
         return true
      }
   }

   Error.Println("ID", t.id, "Single send error, orig:", prevError, "final:", err)

   // Transaction failed...
   return false
}

func handleMultiSendError(t *Transaction, operation int, i int, subSend int, err error) bool {

   if (errors.Is(err, databaseError)) {
      // Just a database error. Update internal database and move on
      if (operation == 0) {
         checkBalance(t.sendingKeys[subSend][i].NanoAddress)
      } else {
         checkBalance(t.transitionalKey[subSend].NanoAddress)
      }
      return true
   }

   switch (operation){
      case 0:
         // Problem with initial sends
         if (retryMultiSend(t, subSend, i, err)) {
            return true
         }
      case 1:
         // Problem with receives
         if (retryReceives(t, subSend, err)) {
            return true
         }
      case 2:
         // Problem with final send
         if (retryFinalSend(t, subSend, err)) {
            return true
         }
      case 3:
         // How did you get here?? The transaction is already complete!
         Warning.Println("Unreachable(1)")
         return true
   }

   // Couldn't salvage the transaction
   return false
}

// This function will find out how much of the transaction is unrecoverable and
// refund the rest. (Fee is fully refunded)
func checkPartialRefund(t *Transaction) error {
   var partialSuccess bool
   for _, success := range t.transactionSuccessful {
      if (success) {
         partialSuccess = true
      }
   }

   if (!partialSuccess) {
      // Just refund the whole thing
      return Refund(t.receiveHash, nt.NewRaw(0))
   } else {
      // Find out how much is unrecoverable.
      amountSent := nt.NewRaw(0)
      for i, success := range t.transactionSuccessful {
         if (success) {
            amountSent.Add(amountSent, t.amountToSend[i])
         }
      }

      // Return everything they paid that hasn't got sent already
      return Refund(t.receiveHash, nt.NewRaw(0).Sub(t.payment, amountSent))
   }
}

// Refund just takes a single receive block hash and reverses it.
// Will refund the whole receive if amount is 0.
func Refund(receiveHash nt.BlockHash, amount *nt.Raw)  error {
   if (inTesting) {
      fmt.Println("Refunding in testing")
      return nil
   }

   // Find the address that send the payment so we can send it back
   if (verbosity >= 5) {
      fmt.Println("Refunding!")
   }

   blockInfo, err := getBlockInfo(receiveHash)
   if (err != nil) {
      return fmt.Errorf("Refund: %w", err)
   }
   sendingKey, _, _, err := getSeedFromAddress(blockInfo.Contents.Account)
   if (err != nil) {
      return fmt.Errorf("Refund: %w", err)
   }

   if (blockInfo.Subtype == "receive") {
      blockInfo, err = getBlockInfo(blockInfo.Contents.Link)
   } else {
      return fmt.Errorf("Refund: Given hash was not a receive")
   }

   clientOriginalAddress, err := keyMan.AddressToPubKey(blockInfo.Contents.Account)
   if (err != nil) {
      return fmt.Errorf("Refund: %w", err)
   }

   retryCount := 0
   for (retryCount < RetryNumber) {
      // If the amount given is 0 or is larger than the actual amount, refund
      // the whole block.
      if (amount.Cmp(nt.NewRaw(0)) == 0 || amount.Cmp(blockInfo.Amount) > 0) {
         _, err = Send(&sendingKey, clientOriginalAddress, blockInfo.Amount, nil, nil, -1)
      } else {
         _, err = Send(&sendingKey, clientOriginalAddress, amount, nil, nil, -1)
      }
      if (err != nil) {
         if (verbosity >= 5) {
            fmt.Println("Refund send error: ", err.Error())
         }
         retryCount++
      } else {
         break
      }
   }
   if (retryCount >= RetryNumber) {
      return fmt.Errorf("Refund: refund failed: %w", err)
   }

   return nil
}

// mixTransitionalAddresses takes all funds that were sent to one of our
// internal addresses and sends it to the mixer because it's now been combined
// in a way that can't be undone.
func mixTransitionalAddresses(t *Transaction) {

   for i := 0; i < t.numSubSends; i++ {
      if !(t.multiSend[i]) {
         return
      }

      nanoAddress := t.transitionalKey[i].NanoAddress

      if !(addressExsistsInDB(nanoAddress)) {
         return
      }

      // Give some time for any transactions that might be in progres (might not
      // even be pending yet) to finish before trying to find them all.
      time.Sleep(10 * time.Second)

      hashList, _ := ReceiveAll(nanoAddress)
      err := waitForConfirmations(hashList)
      if (err != nil) {
         Error.Println("reverseTransitional: ", err)
      }

      // We're stuck with a bunch of funds that have been combined. We have no way
      // to forward blacklist entries, so the best thing to do is just mix it all.
      sendToMixer(t.transitionalKey[i], 1)

      setAddressNotInUse(nanoAddress)
   }
}

func retryMultiSend(t *Transaction, subSend int, i int, prevError error) bool {
   retryCount := 0

   // If there was some problem with PoW then regenerate it.
   if (prevError != nil && strings.Contains(prevError.Error(), "work")){
      clearPoW(t.sendingKeys[subSend][i].NanoAddress)
   }

   var err error
   for (retryCount < RetryNumber) {
      _, err = Send(t.sendingKeys[subSend][i], t.transitionalKey[subSend].PublicKey, t.individualSendAmount[subSend][i], nil, nil, -1)
      if (err != nil) {
         retryCount++
      } else {
         return true
      }
   }
   Error.Println("ID", t.id, "Problem with multi send ", i, ", orig:", prevError, "final:", err)
   return false
}

func retryFinalSend(t *Transaction, subSend int, prevError error) bool {
   retryCount := 0

   var err error
   for (retryCount < RetryNumber) {
      var newHash nt.BlockHash
      newHash, err = Send(t.transitionalKey[subSend], t.recipientAddress, t.amountToSend[subSend], nil, nil, -1)
      if (err != nil) {
         retryCount++
      } else {
         t.finalHash[subSend] = newHash
         return true
      }
   }
   Error.Println("ID", t.id, "Problem with final send, orig:", prevError, "final:", err)
   return false
}

func retryOrigReceive(nanoAddress string, prevError error) (*nt.Raw, nt.BlockHash, error) {

   var retryCount int

   var err = prevError
   var payment *nt.Raw
   var receiveHash nt.BlockHash

   for (retryCount < RetryNumber) {
      payment, receiveHash, _, _, err = Receive(nanoAddress)
      if (err != nil) {
         if (verbosity >= 5) {
            fmt.Println("Error with re-receive: ", retryCount, err.Error())
         }
         retryCount++
      } else {
         return payment, receiveHash, nil
      }

   }

   Error.Println("Problem with original receive, orig:", prevError, "final:", err)

   return nil, nil, err
}

func retryReceives(t *Transaction, subSend int, prevError error) bool {

   receiveHashes, err := ReceiveAll(t.transitionalKey[subSend].NanoAddress)
   if (err != nil) {
      Error.Println("ID", t.id, "Problem with multi receives orig:", prevError, "final:", err)
      return false
   }

   // Need to give the transaction Manager a bump or it will hang.
   go func() {
      var tComm transactionComm
      tComm.i = 2
      tComm.hashes = receiveHashes
      t.commChannel[subSend] <- tComm
   }()

   return true
}

func registerClientComunicationPipe(nanoAddress string, ch chan string) {
   if (inTesting) {
      return
   }

   registeredClientComunicationPipes[nanoAddress] = ch
}

func unregisterClientComunicationPipe(nanoAddress string) {
   if (inTesting) {
      return
   }

   delete(registeredClientComunicationPipes, nanoAddress)
}

func sendInfoToClient(info string, clientPubkey []byte) {
   if (inTesting) {
      return
   }

   clientAddress, _ := keyMan.PubKeyToAddress(clientPubkey)
   if (registeredClientComunicationPipes[clientAddress] != nil) {
      registeredClientComunicationPipes[clientAddress] <- info
   }
}

func sendFinalHash(hashes []nt.BlockHash, pubkey []byte) {
   if (inTesting) {
      return
   }

   var response = "hash="

   for _, hash := range hashes {
      response += hash.String() + ","
   }
   // remove final comma
   if (response[len(response)-1] == ',') {
      response = response[:len(response)-1]
   }

   nanoAddress, _ := keyMan.PubKeyToAddress(pubkey)
   if (registeredClientComunicationPipes[nanoAddress] != nil) {
      select {
         case registeredClientComunicationPipes[nanoAddress] <- response:
         case <-time.After(5 * time.Second):
      }
   }
}

func fullTransactionWasSuccessfull(subSuccess []bool) bool {
   for _, success := range subSuccess {
      if (!success) {
         return false
      }
   }

   return true
}

func broadcastAbort(t *Transaction) {

   // Send to all other subsends in no particular order. Some might actually
   // accept more than one so make sure we send more than the total amount to
   // ensure everybody hears it at least once.
   for i := 0; i < t.numSubSends * 2; i++ {
      select {
         case t.abortchan <- 1:
         case <-time.After(5 * time.Second):
            // Don't wait forever, some might have already finished.
      }
   }
}

// This function will update the database with the success or failure of
// subsends. (This only matters if there are delays as we don't store
// transactions in the database otherwise)
func updateDelayRecords(t *Transaction) {

   if (len(t.delays) > 0) {
      err := upsertTransactionRecord(t)
      if (err != nil) {
         fmt.Println("updateDelayRecords:", err)
         Warning.Println("updateDelayRecords:", err)
      }
   }
}

// This is known as the "reverse-blacklist." It makes sure that we don't send
// funds from the address associated with address C to address A. (see blacklist
// documentation)
func reverseBlacklist(t *Transaction, subSend int, multiSendIndex int) {
   if (t.walletBalance[subSend][multiSendIndex].Cmp(t.individualSendAmount[subSend][multiSendIndex]) > 0) {
      go func() {
         defer func() {
            recoverMessage := recover()
            if (recoverMessage != nil) {
               Error.Println("reverseBlacklist panic: ", recoverMessage)
            }
         }()

         err := blacklistHash(t.sendingKeys[subSend][multiSendIndex].PublicKey, t.receiveHash)
         if (err != nil) {
            _, seedID, index, _ := getSeedFromAddress(t.sendingKeys[subSend][multiSendIndex].NanoAddress)
            Error.Println("BlacklistHash failed to blacklist ", seedID, ",", index, ":", err.Error())
            sendEmail("WARNING", "Blacklist failed."+
            "\n\nID: "+ strconv.Itoa(seedID) +","+strconv.Itoa(index) +
              "\nHash: "+ t.receiveHash.String())

            // Wait for address to not be in use by the transaction.
            for {
               in_use, _ := isAddressInUse(t.sendingKeys[subSend][multiSendIndex].NanoAddress)
               if (in_use) {
                  time.Sleep(10 * time.Second)
               } else {
                  break
               }
            }

            // Keep funds locked until manual intervention.
            setAddressInUse(t.sendingKeys[subSend][multiSendIndex].NanoAddress)
         }
      }()
   }
}
