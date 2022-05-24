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
   nt "nanoTypes"
)

type Transaction struct {
   id int
   paymentAddress []byte
   paymentParentSeedId int
   paymentIndex int
   receiveHash nt.BlockHash
   receiveWg sync.WaitGroup
   clientAddress []byte
   fee *nt.Raw
   amountToSend *nt.Raw
   sendingKeys []*keyMan.Key
   walletSeed []int
   walletBalance []*nt.Raw
   individualSendAmount []*nt.Raw
   transitionalKey *keyMan.Key
   transitionSeedId int
   commChannel chan int
   errChannel chan error
   confirmationChannel chan string
   multiSend bool
   abort bool
}

var databaseError = errors.New("database error")

const RetryNumber = 3

func transactionManager(t *Transaction) {
   t.receiveWg.Add(1)
   var numDone = 0
   var operation = 0
   var transcationSucessfull bool


   wg.Add(1)
   defer wg.Done()

   address, _ := keyMan.PubKeyToAddress(t.paymentAddress)

   // We have a lot of clean up to do
   defer func() {
      // Un-mark all address
      setAddressNotInUse(address)
      for _, key := range t.sendingKeys {
         setAddressNotInUse(key.NanoAddress)
      }
      if (t.transitionalKey != nil) {
         setAddressNotInUse(t.transitionalKey.NanoAddress)
      }

      // Remove active transaction
      setClientAddress(t.paymentParentSeedId, t.paymentIndex, nil)

      // Cancel all the things
      if !(transcationSucessfull) {
         err := Refund(t.receiveHash)
         if (err != nil) {
            // VERY BAD! Just accepted money, failed to deliver it, and didn't
            //           refund the user!
            Error.Println("Refund failed!! Address:", t.paymentAddress, " error:", err.Error())
            // TODO email myself??
         }
         reverseTransitionalAddress(t)
         t.abort = true
         t.receiveWg.Done()
         if (t.multiSend) {
            Warning.Println("Multi transaction failed (", t.id, ")")
         } else {
            Warning.Println("Transaction failed (", t.id, ")")
         }
         if (verbosity >= 5) {
            fmt.Println("Transaction failed...")
         }
      } else {
         Info.Println("Transaction", t.id, "Complete")
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
         // There was a problem.
         if (verbosity >= 5) {
            fmt.Println("Error:", err.Error())
         }
         return
      case <-time.After(5 * time.Minute):
         // Timeout.
         return
   }

   // First manage all iniital sends
   numOfSends := len(t.sendingKeys)
   for {
      // All sends finished
      if (numDone >= numOfSends) {
         if !(t.multiSend) {
            transcationSucessfull = true
         }
         if (verbosity >= 5) {
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
            if (verbosity >= 5) {
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
            Info.Println("Transaction timeout(1)")
            if (verbosity >= 5) {
               if (t.multiSend) {
                  fmt.Println("Transaction error: timout during sends")
               } else {
                  fmt.Println("Transaction error: timout during single send")
               }
            }
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
               if (verbosity >= 5) {
                  fmt.Println("[S]Confirmed: ", numConfirmed)
               }
            case <-time.After(5 * time.Minute):
               Info.Println("Transaction timeout(2)")
               t.abort = true
               t.receiveWg.Done()
               return
         }
      }
      if (verbosity >= 5) {
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
                  if (verbosity >= 5) {
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
                           if (verbosity >= 5) {
                              fmt.Println("[R]Confirmed: ", numConfirmed)
                           }
                        case <-time.After(5 * time.Minute):
                           Info.Println("Transaction timeout(3)")
                           t.abort = true
                           t.receiveWg.Done()
                           return
                     }
                  }
                  t.receiveWg.Done()
               } else if (operation == 3) {
                  // All finished
                  transcationSucessfull = true
                  if (verbosity >= 5) {
                     fmt.Println("Done with everything!")
                  }
               }
            case err := <-t.errChannel:
               // There was an error. Deal with it.
               if (verbosity >= 5) {
                  fmt.Println("Error with receives or final send", err.Error())
               }
               if (handleMultiSendError(t, operation, -1, err)) {
                  operation++
                  t.receiveWg.Done()
               } else {
                  // This is just to exit the loop
                  operation = 10
                  return
               }
            case <-time.After(5 * time.Minute):
               Info.Println("Transaction timeout(4)")
               if (verbosity >= 5) {
                  if (operation == 1) {
                     fmt.Println("Transaction error: timout during receives")
                  } else {
                     fmt.Println("Transaction error: timout during final send")
                  }
               }
               return
         }
      }
   }

   if (verbosity >= 5) {
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
         if (verbosity >= 5) {
            fmt.Println("Error with resend: ", retryCount, err.Error())
         }
         retryCount++
      } else {
         return true
      }
   }

   Warning.Println("Single send error:", err.Error())

   // Transaction failed...
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
         Warning.Println("Unreachable(1)")
         return true
   }

   // Couldn't salvage the transaction
   return false
}

// Refund just takes a single receive block hash and reverses it.
func Refund(receiveHash nt.BlockHash)  error {
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
      err = Send(&sendingKey, clientOriginalAddress, blockInfo.Amount, nil, nil, -1)
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
