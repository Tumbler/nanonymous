package main

import (
   "fmt"
   "sync"
   "errors"
   "time"
   "strings"
   "regexp"
   "strconv"

   // Local packages
   keyMan "nanoKeyManager"
)

// TODO check all timeouts and make sure that they do all the refunding and aborting that they should

type Transaction struct {
   paymentAddress []byte
   receiveHash keyMan.BlockHash
   multiSend bool
   receiveWg sync.WaitGroup
   clientAddress []byte
   fee *keyMan.Raw
   amountToSend *keyMan.Raw
   sendingKeys []*keyMan.Key
   walletSeed []int
   walletBalance []*keyMan.Raw
   transitionalKey *keyMan.Key
   transitionSeedId int
   individualSendAmount []*keyMan.Raw
   abort bool
   commChannel chan int
   errChannel chan error
   confirmationChannel chan string
}

var databaseError = errors.New("database error")

const RetryNumber = 3

func transactionManager(t *Transaction) {
   t.receiveWg.Add(1)
   var numDone = 0
   var operation = 0

   wg.Add(1)
   defer wg.Done()

   address, _ := keyMan.PubKeyToAddress(t.paymentAddress)
   defer func() {
      setAddressNotInUse(address)
      for _, key := range t.sendingKeys {
         setAddressNotInUse(key.NanoAddress)
      }
      if (t.transitionalKey != nil) {
         setAddressNotInUse(t.transitionalKey.NanoAddress)
      }
   }()

   // Waiting until first send
   select {
      case <-t.commChannel:
         // Proceed to next step
         if (t.multiSend) {
            t.confirmationChannel = make(chan string)
            registerConfirmationListener(t.transitionalKey.NanoAddress, t.confirmationChannel, "send")
            defer unregisterConfirmationListener(t.transitionalKey.NanoAddress, "send")
         }
      case err := <-t.errChannel:
         // There was a problem. Refund the payment and abort the transaction.
         if (verbose) {
            fmt.Println("Error:", err.Error())
         }
         Refund(t.receiveHash)
         t.abort = true
         t.receiveWg.Done()
         return
      case <-time.After(5 * time.Minute):
         // Timeout. Refund the payment and abort the transaction.
         Refund(t.receiveHash)
         t.abort = true
         t.receiveWg.Done()
         return
   }

   // First manage all iniital sends
   numOfSends := len(t.sendingKeys)
   for {
      // All sends finished
      if (numDone >= numOfSends) {
         if (verbose) {
            fmt.Println("Done with Sends!")
         }
         break
      }
      select {
         case i := <-t.commChannel:
            // A send finished with no errors
            numDone++

            // This is known as the "reverse-blacklist." It makes sure that we
            // don't send funds from the address associated with address C to
            // address A. (see blacklist documentation)
            if (t.walletBalance[i].Cmp(t.individualSendAmount[i]) > 0) {
               go blacklistHash(t.sendingKeys[i].PublicKey, t.receiveHash)
            }
         case err := <-t.errChannel:
            // There was an error. Deal with it.
            if (verbose) {
               fmt.Println("Error with sends", err.Error())
            }
            if (t.multiSend) {
               findIndex, _ := regexp.Compile(">>([0-9]+)<<")
               regexResults := findIndex.FindSubmatch([]byte(err.Error()))
               var whichSend int
               if (len(regexResults) > 1) {
                  whichSend, _ = strconv.Atoi(string(findIndex.FindSubmatch([]byte(err.Error()))[1]))
               }

               if (handleMultiSendError(t, operation, whichSend, err)) {
                  // We recovered so go back to regular operation
                  numDone++
               } else {
                  // Abort rest of transaction
                  t.abort = true
                  t.receiveWg.Done()
                  return
               }
            } else {
               if (handleSingleSendError(t, err)) {
                  // We recovered so go back to regular operation
                  numDone++
               } else {
                  // Abort transaction
                  return
               }
            }
         case <-time.After(5 * time.Minute):
            // TODO log
            if (verbose) {
               if (t.multiSend) {
                  fmt.Println("Transaction error: timout during sends")
               } else {
                  fmt.Println("Transaction error: timout during single send")
               }
            }

            // Refund and reset
            Refund(t.receiveHash)
            reverseTransitionalAddress(t)

            t.abort = true
            t.receiveWg.Done()
            return
      }
   }


   // Sends are done, wait for them to be confirmed
   if (t.multiSend) {
      trackConfirms := make(map[string]bool)
      var numConfirmed int

      for (numConfirmed < numOfSends) {
         select {
            case hash := <-t.confirmationChannel:
               // Make sure we didn't receive the same block twice
               if (trackConfirms[hash] == false) {
                  trackConfirms[hash] = true
                  numConfirmed++
               }
               if (verbose) {
                  fmt.Println("[S]Confirmed: ", numConfirmed)
               }
            case <-time.After(5 * time.Minute):
               // TODO log
               t.abort = true
               t.receiveWg.Done()
               return
         }
      }
      if (verbose) {
         fmt.Println("All sends confirmed!")
      }
      // Signal receives to start in ReceiveAndSend()
      t.receiveWg.Done()
   }


   // Now if it's a multi send then we need monitor the last step of receiving
   // and sending to the client
   if (t.multiSend) {
      operation = 1

      registerConfirmationListener(t.transitionalKey.NanoAddress, t.confirmationChannel, "receive")
      defer unregisterConfirmationListener(t.transitionalKey.NanoAddress, "receive")

      for (operation < 3) {
         select {
            case operation = <-t.commChannel:
               if (operation == 2) {
                  // Done with receives
                  if (verbose) {
                     fmt.Println("Done with receives")
                  }

                  // Recives have been published, now wait for them to confirm
                  trackConfirms := make(map[string]bool)
                  var numConfirmed int
                  for numConfirmed < numOfSends {
                     select {
                        case hash := <-t.confirmationChannel:
                           // Make sure we didn't receive the same block twice
                           if (trackConfirms[hash] == false) {
                              trackConfirms[hash] = true
                              numConfirmed++
                           }
                           if (verbose) {
                              fmt.Println("[R]Confirmed: ", numConfirmed)
                           }
                        case <-time.After(5 * time.Minute):
                           // TODO log
                           t.abort = true
                           t.receiveWg.Done()
                           return
                     }
                  }
                  t.receiveWg.Done()
               } else if (operation == 3) {
                  // All finished
                  if (verbose) {
                     fmt.Println("Done with everything!")
                  }
               }
            case err := <-t.errChannel:
               // There was an error. Deal with it.
               if (verbose) {
                  fmt.Println("Error with receives or final send", err.Error())
               }
               if (handleMultiSendError(t, operation, -1, err)) {
                  operation++
                  t.receiveWg.Done()
               } else {
                  t.abort = true
                  t.receiveWg.Done()

                  // This is just to exit the loop
                  operation = 10

                  break
               }
            case <-time.After(5 * time.Minute):
               // TODO log
               if (verbose) {
                  if (operation == 1) {
                     fmt.Println("Transaction error: timout during receives")
                  } else {
                     fmt.Println("Transaction error: timout during final send")
                  }
               }

               // Refund and reset
               Refund(t.receiveHash)
               reverseTransitionalAddress(t)

               // This is just to exit the loop
               operation = 10

               break
         }
      }
   }

   if (verbose) {
      fmt.Println("Transaction Complete!")
   }
}

func handleSingleSendError(t *Transaction, prevError error) bool {
   var retryCount = 0
   var err error

   // If there was some problem with PoW then regenerate it.
   if (prevError != nil && strings.Contains(prevError.Error(), "work")){
      clearPoW(t.sendingKeys[0].NanoAddress)
   }
   if (errors.Is(prevError, databaseError)) {
      // Just a database error. Update internal database and move on
      checkBalance(t.sendingKeys[0].NanoAddress)
   }

   for (retryCount < RetryNumber) {
      err = Send(t.sendingKeys[0], t.clientAddress, t.amountToSend, nil, nil, -1)
      if (err != nil) {
         if (verbose) {
            fmt.Println("Error with resend: ", retryCount, err.Error())
         }
         retryCount++
      } else {
         return true
      }
   }

   // TODO log

   // Transaction failed... attempt refund
   err = Refund(t.receiveHash)
   if (err != nil && verbose) {
      fmt.Println("Refund failed: ", err.Error())
   }

   return false
}

func handleMultiSendError(t *Transaction, operation int, i int, err error) bool {

   if (errors.Is(err, databaseError)) {
      // Just a database error. Update internal database and move on
      if (operation == 0) {
         checkBalance(t.sendingKeys[i].NanoAddress)
      } else {
         checkBalance(t.transitionalKey.NanoAddress)
      }
      return true
   }

   switch (operation){
      case 0:
         // Problem with initial sends
         if (retryMultiSend(t, i, err)) {
            return true
         }
      case 1:
         // Problem with receives
         if (retryReceives(t, err)) {
            return true
         }
      case 2:
         // Problem with final send
         if (retryFinalSend(t, err)) {
            return true
         }
      case 3:
         // How did you get here?? The transaction is already complete!
         // TODO log
         return false
   }

   // Couldn't salvage the transaction, begin refunds
   Refund(t.receiveHash)
   reverseTransitionalAddress(t)

   return false
}

// Refund just takes a single receive block hash and reverses it.
func Refund(receiveHash keyMan.BlockHash)  error {
   // Find the address that send the payment so we can send it back
   if (verbose) {
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
      err = Send(&sendingKey, clientOriginalAddress, blockInfo.Amount, nil, nil, -1)
      if (err != nil) {
         if (verbose) {
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

// reverseTransitionalAddress takes all funds that were sent to one of our
// internal addresses and returns them to their original wallets. This is so
// that the wallets can continue using their own blacklist entries correctly.
func reverseTransitionalAddress(t *Transaction) {

   if !(t.multiSend) {
      return
   }

   nanoAddress := t.transitionalKey.NanoAddress

   if !(addressExsistsInDB(nanoAddress)) {
      return
   }

   // Give some time for any transactions that might be in progres (might not
   // even be pending yet) to finish before trying to find them all.
   time.Sleep(10 * time.Second)

   ReceiveAll(nanoAddress)

   // TODO wait for blocks to be confirmed
   time.Sleep(10 * time.Second)

   // -1 means full history
   history, _ := getAccountHistory(nanoAddress, -1)

   for _, block := range history.History {
      if (block.Type == "receive" && addressExsistsInDB(block.Account)) {
         pubKey, _ := keyMan.AddressToPubKey(block.Account)
         sendNano(t.transitionalKey, pubKey, block.Amount)
      }
   }

   setAddressNotInUse(nanoAddress)
   // TODO test reverse transitional
}

func retryMultiSend(t *Transaction, i int, prevError error) bool {
   retryCount := 0

   // If there was some problem with PoW then regenerate it.
   if (prevError != nil && strings.Contains(prevError.Error(), "work")){
      clearPoW(t.sendingKeys[i].NanoAddress)
   }

   for (retryCount < RetryNumber) {
      err := Send(t.sendingKeys[i], t.transitionalKey.PublicKey, t.individualSendAmount[i], nil, nil, -1)
      if (err != nil) {
         retryCount++
      } else {
         return true
      }
   }
   return false
}

func retryFinalSend(t *Transaction, prevError error) bool {
   retryCount := 0

   // If there was some problem with PoW then regenerate it.
   if (prevError != nil && strings.Contains(prevError.Error(), "work")){
      clearPoW(t.transitionalKey.NanoAddress)
   }
   // TODO put other common fixes to problems here as you find them

   for (retryCount < RetryNumber) {
      err := Send(t.transitionalKey, t.clientAddress, t.amountToSend, nil, nil, -1)
      if (err != nil) {
         retryCount++
      } else {
         return true
      }
   }
   return false
}

func retryReceives(t *Transaction, prevError error) bool {

   // If there was some problem with PoW then regenerate it.
   if (prevError != nil && strings.Contains(prevError.Error(), "work")){
      clearPoW(t.transitionalKey.NanoAddress)
   }

   err := ReceiveAll(t.transitionalKey.NanoAddress)
   if (err != nil) {
      return false
   }

   return true
}
