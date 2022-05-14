package main

import (
   "fmt"
   "sync"
   "errors"

   // Local packages
   keyMan "nanoKeyManager"
)

type Transaction struct {
   paymentAddress []byte
   payment *keyMan.Raw
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
   started bool
   commChannel chan int
   errChannel chan error
}

var sendError = errors.New("send error")
var receiveError = errors.New("receive error")

func transactionManager(transaction *Transaction) {
   transaction.receiveWg.Add(1)
   var numDone = 0

   // Wait to organize unil the first send has started
   for !(transaction.started) {
   }

   for {
      select {
         case <-transaction.commChannel:
            // All sends finished with no errors

            numDone++
            if (numDone >= len(transaction.sendingKeys)) {
               // Signal receives to start in ReceiveAndSend() if necessary
               transaction.receiveWg.Done()

               if (verbose) {
                  fmt.Println("Done with Sends!")
                  break
               }
            }
         case err := <-transaction.errChannel:
            // There was an error. Deal with it.
            if (verbose) {
               fmt.Println("Error with sends %s", err.Error())
            }
            if (transaction.multiSend) {
               if (handleMultiSendError(transaction)) {
                  // We recovered go back to regular operation
                  numDone++
               }
            } else {
               if (handleSingleSendError(transaction)) {
                  // We recovered go back to regular operation
                  numDone++
               }
            }
      }
   }

   // If it's a multi send then we need monitor the last step of receiving
   // and sending to the client
   if (transaction.multiSend) {

      select {
         case <-transaction.commChannel:
            // All finished
            if (verbose) {
               fmt.Println("Done with everything!")
            }
            break
         case err := <-transaction.errChannel:
            // There was an error. Deal with it.
            if (verbose) {
               fmt.Println("Error with receives or final send %s", err.Error())
            }
            handleMultiSendError(transaction)
      }
   }
}

func handleSingleSendError(t *Transaction) bool {
   var retryCount = 0
   var err error

   for (retryCount < 3) {
      err = Send(t.sendingKeys[0], t.clientAddress, t.amountToSend, nil, nil)
      if (err != nil) {
         retryCount++
      } else {
         return true
      }
   }

   fmt.Println("Transaction error: %s", err.Error())

   // Transaction failed... attempt refund
   retryCount = 0
   for (retryCount < 3) {
      err := Send(t.sendingKeys[0], t.paymentAddress, t.payment, nil, nil)
      if (err != nil) {
         retryCount++
      } else {
         // TODO log error and probably email because we just took someones money and it didn't work
         break
      }
   }

   return false
}

func handleMultiSendError(t *Transaction) bool{

   return false
}
