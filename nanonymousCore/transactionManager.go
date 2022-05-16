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

type Transaction struct {
   paymentAddress []byte
   payment *keyMan.Raw
   receiveHash keyMan.BlockHash
   multiSend bool
   receiveWg sync.WaitGroup
   clientAddress []byte
   fee *keyMan.Raw
   amountToSend *keyMan.Raw
   sendingKeys []*keyMan.Key
   walletSeed []int
   walletBalance []*keyMan.Raw
   transitionalAddress *keyMan.Key
   transitionSeedId int
   multiSendAmount []*keyMan.Raw
   abort bool
   commChannel chan int
   errChannel chan error
}

var sendError = errors.New("send error")
var receiveError = errors.New("receive error")

func transactionManager(transaction *Transaction) {
   transaction.receiveWg.Add(1)
   var numDone = 0
   var operation = 0

   // Waiting until first send
   select {
      case <-transaction.commChannel:
         // Proceed to next step
      case <-transaction.errChannel:
         // There was a problem. Refund the payment and abort the transaction.
         simpleRefund(transaction.receiveHash)
         transaction.abort = true
         transaction.receiveWg.Done()
         return
      case <-time.After(5 * time.Minute):
         // Timeout. Refund the payment and abort the transaction.
         simpleRefund(transaction.receiveHash)
         transaction.abort = true
         transaction.receiveWg.Done()
         return
   }

   for {
      finishedSends := make([]bool, len(transaction.sendingKeys))

      // All sends finished
      if (numDone >= len(transaction.sendingKeys)) {
         // Signal receives to start in ReceiveAndSend() if necessary
         transaction.receiveWg.Done()

         if (verbose) {
            fmt.Println("Done with Sends!")
         }
         break
      }
      select {
         case i := <-transaction.commChannel:
            // A send finished with no errors
            numDone++
            finishedSends[i] = true

            // This is known as the "reverse-blacklist." It makes sure that we
            // don't send funds from the address associated with address C to
            // address A. (see blacklist documentation)
            if (transaction.walletBalance[i].Cmp(transaction.multiSendAmount[i]) > 0) {
               go blacklistHash(transaction.sendingKeys[i].PublicKey, transaction.receiveHash)
            }
         case err := <-transaction.errChannel:
            // There was an error. Deal with it.
            if (verbose) {
               fmt.Println("Error with sends %s", err.Error())
            }
            findIndex, _ := regexp.Compile(">>([0-9]+)<<")
            whichSend, _ := strconv.Atoi(string(findIndex.FindSubmatch([]byte(err.Error()))[1]))
            if (transaction.multiSend) {
               if (handleMultiSendError(transaction, operation, whichSend)) {
                  // We recovered so go back to regular operation
                  numDone++
               } else {
                  // Abort rest of transaction
                  transaction.abort = true
                  transaction.receiveWg.Done()
                  return
               }
            } else {
               if (handleSingleSendError(transaction, err)) {
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
               if (transaction.multiSend) {
                  fmt.Println("Transaction error: timout during sends")
               } else {
                  fmt.Println("Transaction error: timout during single send")
               }
            }
            transaction.abort = true
            transaction.receiveWg.Done()
            return
      }
   }


   // If it's a multi send then we need monitor the last step of receiving
   // and sending to the client
   if (transaction.multiSend) {
      operation = 1

      for {
         select {
            case operation = <-transaction.commChannel:
               if (operation == 2) {
                  // Done with receives
                  if (verbose) {
                     fmt.Println("Done with receives")
                  }
               } else if (operation == 3) {
                  // All finished
                  if (verbose) {
                     fmt.Println("Done with everything!")
                  }
                  break
               }
            case err := <-transaction.errChannel:
               // There was an error. Deal with it.
               if (verbose) {
                  fmt.Println("Error with receives or final send %s", err.Error())
               }
               handleMultiSendError(transaction, operation, -1)
            case <-time.After(5 * time.Minute):
               // TODO log
               if (verbose) {
                  if (operation == 1) {
                     fmt.Println("Transaction error: timout during receives")
                  } else {
                     fmt.Println("Transaction error: timout during final send")
                  }
               }
               break
         }
      }
   }
}

func handleSingleSendError(t *Transaction, prevError error) bool {
   var retryCount = 0
   var err error

   // If there was some problem with PoW then regenerate it.
   if (prevError != nil && strings.Contains(prevError.Error(), "work")){
      clearPoW(t.sendingKeys[0].NanoAddress)
   }

   for (retryCount < 3) {
      err = Send(t.sendingKeys[0], t.clientAddress, t.amountToSend, nil, nil, -1)
      if (err != nil) {
         retryCount++
      } else {
         return true
      }
   }

   // TODO log
   fmt.Println("Transaction error: %s", err.Error())

   // Transaction failed... attempt refund
   simpleRefund(t.receiveHash)

   return false
}

func handleMultiSendError(t *Transaction, operation int, i int) bool{

   switch (operation){
      case 0:
      case 1:
      case 2:
      case 3:
   }

   return false
}

// simpleRefund just takes a receive block hash and reverses it.
func simpleRefund(receiveHash keyMan.BlockHash)  error {
   // Find the address that send the payment so we can send it back

   blockInfo, err := getBlockInfo(receiveHash)
   if (err != nil) {
      return fmt.Errorf("simpleRefund: %w", err)
   }
   sendingKey, _, _, err := getSeedFromAddress(blockInfo.Contents.Account)
   if (err != nil) {
      return fmt.Errorf("simpleRefund: %w", err)
   }

   if (blockInfo.Subtype == "receive") {
      blockInfo, err = getBlockInfo(blockInfo.Contents.Link)
   } else {
      return fmt.Errorf("simpleRefund: Given hash was not a receive")
   }

   clientOriginalAddress, err := keyMan.AddressToPubKey(blockInfo.Contents.Account)
   if (err != nil) {
      return fmt.Errorf("simpleRefund: %w", err)
   }

   retryCount := 0
   for (retryCount < 3) {
      err = Send(&sendingKey, clientOriginalAddress, blockInfo.Amount, nil, nil, -1)
      if (err != nil) {
         retryCount++
      } else {
         break
      }
   }
   if (retryCount >= 3) {
      return fmt.Errorf("simpleRefund: refund failed: %w", err)
   }

   return nil
}
