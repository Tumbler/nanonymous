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
   finalHash nt.BlockHash
   commChannel chan transactionComm
   errChannel chan error
   confirmationChannel chan string
   multiSend bool
   dirtyAddress int // The sendingKeys address that has been linked to other addresses but not blacklisted
   abort bool
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
   t.receiveWg.Add(1)
   var numDone = 0
   var operation = 0
   var transcationSucessfull bool


   wg.Add(1)
   defer wg.Done()

   address, _ := keyMan.PubKeyToAddress(t.paymentAddress)

   // We have a lot of clean up to do
   defer func() {
      // Remove active transaction
      // TODO I don't think this is working
      setClientAddress(t.paymentParentSeedId, t.paymentIndex, nil)

      // Cancel all the things
      if !(transcationSucessfull) {
         err := Refund(t.receiveHash)
         if (err != nil) {
            // VERY BAD! Just accepted money, failed to deliver it, and didn't
            //           refund the user!
            nanoAddress, _ := keyMan.PubKeyToAddress(t.paymentAddress)
            Error.Println("Refund failed!! Address:", nanoAddress, " error:", err.Error())
            sendEmail("IMMEDIATE ATTENTION REQUIRED", "Refund failed!! Address: "+ nanoAddress +" error: "+ err.Error() +
               "\n\nPayment Hash: "+ t.receiveHash.String() +
               "\nID: "+ strconv.Itoa(t.paymentParentSeedId) +","+ strconv.Itoa(t.paymentIndex) +
               "\nAmount: "+ strconv.FormatFloat(rawToNANO(nt.NewRaw(0).Add(t.amountToSend, t.fee)), 'f', -1, 64))
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
         recordProfit(t.fee, t.id)
         Info.Println("Transaction", t.id, "Complete")
         // TODO debug statements
         //fmt.Println("finalHash", t.finalHash)
         //fmt.Println("multi?:", t.multiSend)
         sendFinalHash(t.finalHash, t.clientAddress)

         // Send any dirty addresses to the mixer.
         if (t.dirtyAddress != -1) {
            if (verbosity >= 8) {
               fmt.Println("Sending dirty address to mixer")
            }
            err := sendToMixer(t.sendingKeys[t.dirtyAddress], 1)

            if (err != nil) {
               Error.Println("Mixer Error:", err.Error())
               sendEmail("WARNING", "Mixer Error: "+ err.Error())
            }
         }
      }

      // Un-mark all addresses
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
         // There was a problem.
         if (verbosity >= 5) {
            fmt.Println("Error:", err.Error())
         }
         return
      case <-time.After(5 * time.Minute):
         // Timeout.
         Info.Println("Transaction timeout(0)")
         return
   }

   // First manage all initial sends
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

            if !(t.multiSend) {
               t.finalHash = i.hashes[0]
            }

            // This is known as the "reverse-blacklist." It makes sure that we
            // don't send funds from the address associated with address C to
            // address A. (see blacklist documentation)
            if (t.walletBalance[i.i].Cmp(t.individualSendAmount[i.i]) > 0) {
               go blacklistHash(t.sendingKeys[i.i].PublicKey, t.receiveHash)
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
         if (inTesting) {
            break
         }

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
            case tComm := <-t.commChannel:
               operation = tComm.i
               if (operation == 2) {
                  // Done with receives
                  if (verbosity >= 5) {
                     fmt.Println("Done with receives")
                  }

                  // Recives have been published, now wait for them to confirm

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
                        case hash := <-t.confirmationChannel:
                           // Make sure we didn't receive the same block twice
                           if (trackConfirms[hash] == false) {
                              fmt.Println("hash:", hash)
                              trackConfirms[hash] = true
                              numConfirmed++
                           }
                           if (verbosity >= 5) {
                              fmt.Println("[R]Confirmed: ", numConfirmed)
                           }

                           // Time limit gets reset if we're still getting confirmations
                           timeLimit = time.Now().Add(5 * time.Minute)
                        case <-time.After(10 * time.Second):
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
                           t.receiveWg.Done()
                           return
                     }
                  }
                  t.receiveWg.Done()
               } else if (operation == 3) {
                  // All finished
                  t.finalHash = tComm.hashes[0]
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
                     Info.Println("Transaction error: timout during receives")
                  } else {
                     Info.Println("Transaction error: timout during final send")
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
      return true
   }

   for (retryCount < RetryNumber) {
      _, err = Send(t.sendingKeys[0], t.clientAddress, t.amountToSend, nil, nil, -1)
      if (err != nil) {
         if (verbosity >= 5) {
            fmt.Println("Error with resend: ", retryCount, err.Error())
         }
         retryCount++
      } else {
         return true
      }
   }

   Error.Println("ID", t.id, "Single send error, orig:", prevError, "final:", err)

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
      _, err = Send(&sendingKey, clientOriginalAddress, blockInfo.Amount, nil, nil, -1)
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

   // TODO how to make this more robust???
   // Give some time for any transactions that might be in progres (might not
   // even be pending yet) to finish before trying to find them all.
   time.Sleep(10 * time.Second)

   // Get ready to track how many receives we get
   ch := make(chan string)
   registerConfirmationListener(nanoAddress, ch, "receive")
   defer unregisterConfirmationListener(nanoAddress, "receive")
   hashes, _ := getReceivable(nanoAddress, -1)
   numToReceive := len(hashes)

   ReceiveAll(nanoAddress)

   // Wait for all to be confirmed
   trackConfirms := make(map[string]bool)
   var numConfirmed int
   confirmLoop:
   for (numConfirmed < numToReceive) {
      select {
         case hash := <-ch:
            if (trackConfirms[hash] == false) {
               trackConfirms[hash] = true
               numConfirmed++
            }
         case <-time.After(5 * time.Minute):
            Error.Println("reverseTransitional timeout")
            break confirmLoop
      }
   }

   // TODO reverseing this is actually bad because now all reversed addresses are dirty. Probably better to just mix it.
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

   var err error
   for (retryCount < RetryNumber) {
      _, err = Send(t.sendingKeys[i], t.transitionalKey.PublicKey, t.individualSendAmount[i], nil, nil, -1)
      if (err != nil) {
         retryCount++
      } else {
         return true
      }
   }
   Error.Println("ID", t.id, "Problem with multi send ", i, ", orig:", prevError, "final:", err)
   return false
}

func retryFinalSend(t *Transaction, prevError error) bool {
   retryCount := 0

   // If there was some problem with PoW then regenerate it.
   if (prevError != nil && strings.Contains(prevError.Error(), "work")){
      clearPoW(t.transitionalKey.NanoAddress)
   }
   // TODO put other common fixes to problems here as you find them

   var err error
   for (retryCount < RetryNumber) {
      newHash, err := Send(t.transitionalKey, t.clientAddress, t.amountToSend, nil, nil, -1)
      if (err != nil) {
         retryCount++
      } else {
         t.finalHash = newHash
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
      // If there was some problem with PoW then regenerate it.
      if (strings.Contains(err.Error(), "work")){
         clearPoW(nanoAddress)
      }

      payment, receiveHash, _, err = Receive(nanoAddress)
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

func retryReceives(t *Transaction, prevError error) bool {

   // If there was some problem with PoW then regenerate it.
   if (prevError != nil && strings.Contains(prevError.Error(), "work")){
      clearPoW(t.transitionalKey.NanoAddress)
   }

   // TODO track hashes
   receiveHashes, err := ReceiveAll(t.transitionalKey.NanoAddress)
   if (err != nil) {
      Error.Println("ID", t.id, "Problem with multi receives orig:", prevError, "final:", err)
      return false
   }

   // Need to give the transaction Manager a bump or it will hang.
   go func() {
      var tComm transactionComm
      tComm.i = 2
      tComm.hashes = receiveHashes
      t.commChannel <- tComm
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

// TODO confusion with the name client. DOes it refer to the sender or the receiver??
func sendInfoToClient(info string, clientPubkey []byte) {
   clientAddress, _ := keyMan.PubKeyToAddress(clientPubkey)
   if (registeredClientComunicationPipes[clientAddress] != nil) {
      registeredClientComunicationPipes[clientAddress] <- info
   }
}

func sendFinalHash(hash nt.BlockHash, pubkey []byte) {
   nanoAddress, _ := keyMan.PubKeyToAddress(pubkey)
   if (registeredClientComunicationPipes[nanoAddress] != nil) {
      registeredClientComunicationPipes[nanoAddress] <- "hash="+ hash.String()
   }
}
