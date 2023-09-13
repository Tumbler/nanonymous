package main

import (
   "fmt"
   "context"

   // Local packages
   keyMan "nanoKeyManager"
   nt "nanoTypes"
)

// sendToMixer takes an address and shuffles it all into the mixer. shufflesLeft
// is how many times to shuffle before being considered clean. (Should be 1 in
// almost all cases)
func sendToMixer(key *keyMan.Key, shufflesLeft int) error {
   var err error

   origSeed, origIndex, err := getWalletFromAddress(key.NanoAddress)
   if (err != nil) {
      return fmt.Errorf("sendToMixer: Failed to get wallet: %w", err)
   }

   // Get new mixer addresses
   mix1, _, err := getNewAddress("", false, true, 0)
   if (err != nil) {
      return fmt.Errorf("sendToMixer: Orig: %d, %d, Failed to get new address1: %w:", origSeed, origIndex, err)
   }
   mix2, _, err := getNewAddress("", false, true, 0)
   if (err != nil) {
      return fmt.Errorf("sendToMixer: Orig: %d, %d, Failed to get new address2: %w:", origSeed, origIndex, err)
   }

   if (!inTesting) {
      addWebSocketSubscription <- mix1.NanoAddress
      addWebSocketSubscription <- mix2.NanoAddress
   }

   // Randomize amounts

   bal, err := getBalance(key.NanoAddress)
   if (err != nil) {
      return fmt.Errorf("sendToMixer: Orig: %d, %d - %w", origSeed, origIndex, err)
   }

   percent := int64(random.Intn(98) + 1) // 1 - 99 %
   onePercent := nt.NewRaw(0).Div(bal, nt.NewRaw(100))
   amount1 := nt.NewRaw(0).Mul(onePercent, nt.NewRaw(percent))
   amount2 := nt.NewRaw(0).Sub(bal, amount1)

   tries := 0
   var sendHash nt.BlockHash
   var hashList []nt.BlockHash
   sendHash, err = sendNano(key, mix1.PublicKey, amount1)
   for (err != nil) {
      if (tries < RetryNumber) {
         sendHash, err = sendNano(key, mix1.PublicKey, amount1)
         tries++
         continue
      }
      return fmt.Errorf("sendToMixer: Send 1 error: Orig: %d, %d - %w", origSeed, origIndex, err)
   }
   hashList = append(hashList, sendHash)
   tries = 0
   send1Seed, send1Index, _ := getWalletFromAddress(mix1.NanoAddress)

   sendHash, err = sendNano(key, mix2.PublicKey, amount2)
   for (err != nil) {
      if (tries < RetryNumber) {
         sendHash, err = sendNano(key, mix2.PublicKey, amount2)
         tries++
         continue
      }
      return fmt.Errorf("sendToMixer: Send 2 error: Orig: %d, %d, Send 1: %d, %d - %w", origSeed, origIndex, send1Seed, send1Index, err)
   }
   hashList = append(hashList, sendHash)
   tries = 0
   send2Seed, send2Index, _ := getWalletFromAddress(mix1.NanoAddress)

   // Make sure they're confirmed.
   waitForConfirmations(hashList)

   _, _, _, err = Receive(mix1.NanoAddress)
   for (err != nil) {
      if (tries < RetryNumber) {
         _, _, _, err = Receive(mix1.NanoAddress)
         tries++
         continue
      }
      return fmt.Errorf("sendToMixer: Orig: %d, %d, Send1: %d, %d, Send2: %d, %d - Receive 1 error: %w", origSeed, origIndex, send1Seed, send1Index, send2Seed, send2Index, err)
   }
   tries = 0
   _, _, _, err = Receive(mix2.NanoAddress)
   for (err != nil) {
      if (tries < RetryNumber) {
         _, _, _, err = Receive(mix2.NanoAddress)
         tries++
         continue
      }
      return fmt.Errorf("sendToMixer: Orig: %d, %d, Send1: %d, %d, Send2: %d, %d - Receive 2 error: %w", origSeed, origIndex, send1Seed, send1Index, send2Seed, send2Index, err)
   }
   tries = 0

   if (shufflesLeft > 0) {
      err = sendToMixer(mix1, shufflesLeft-1)
      if (err != nil) {
         return fmt.Errorf("sendToMixer: Orig: %d, %d, Send1: %d, %d, Send2: %d, %d - sendToMixer 2 error: %w", origSeed, origIndex, send1Seed, send1Index, send2Seed, send2Index, err)
      }
      err = sendToMixer(mix2, shufflesLeft-1)
      if (err != nil) {
         return fmt.Errorf("sendToMixer: Orig: %d, %d, Send1: %d, %d, Send2: %d, %d - sendToMixer 3 error: %w", origSeed, origIndex, send1Seed, send1Index, send2Seed, send2Index, err)
      }
   }

   if (!inTesting) {
      removeWebSocketSubscription <- mix1.NanoAddress
      removeWebSocketSubscription <- mix2.NanoAddress
   }
   setAddressNotInUse(mix1.NanoAddress)
   setAddressNotInUse(mix2.NanoAddress)

   return err
}

// getKeysFromMixer the transaction manager is very good at making things get
// done properly, so just hand it the keys it needs to get enough funds and let
// it handle the actual sends.
func getKeysFromMixer(amountNeeded *nt.Raw) ([]*keyMan.Key, []int, []*nt.Raw, error) {
   var err error

   mixerBalance, err := getReadyMixerFunds()

   // Mixer Balance < amount to send
   if (err != nil) {
      return []*keyMan.Key{}, []int{}, []*nt.Raw{}, fmt.Errorf("getFromMixer: problem getting funds from mixer: %w", err)
   }  else if (mixerBalance.Cmp(amountNeeded) < 0) {
      return []*keyMan.Key{}, []int{}, []*nt.Raw{}, fmt.Errorf("getFromMixer: not enough funds in mixer")
   }

   rows, conn, err := getMixerRows()
   if (err != nil) {
      return []*keyMan.Key{}, []int{}, []*nt.Raw{}, fmt.Errorf("getFromMixer: get rows: %w", err)
   }
   defer conn.Close(context.Background())

   var totalBalance = nt.NewRaw(0)
   var keys []*keyMan.Key
   var seeds []int
   var balances []*nt.Raw

   var seed int
   var index int
   var balance = nt.NewRaw(0)
   for (rows.Next()) {
      rows.Scan(&seed, &index, balance)

      key, err := getSeedFromIndex(seed, index)
      if (err != nil) {
         return []*keyMan.Key{}, []int{}, []*nt.Raw{}, fmt.Errorf("getFromMixer: %w", err)
      }
      keys = append(keys, key)
      seeds = append(seeds, seed)
      balances = append(balances, nt.NewFromRaw(balance))

      totalBalance.Add(totalBalance, balance)

      if (totalBalance.Cmp(amountNeeded) >= 0) {
         break
      }
   }

   return keys, seeds, balances, nil
}


func extractFromMixer(amountToSend *nt.Raw, publicKey []byte) (nt.BlockHash, error) {

   var addressList []string
   // Make sure anything we touched is back to not in use
   defer func() {
      for _, address := range addressList {
         setAddressNotInUse(address)
      }
   }()

   var err error
   var finalHash nt.BlockHash
   var dirtyAddress *keyMan.Key

   _, _, mixerBalance, err := findTotalBalance()

   // Mixer Balance < amount to send
   if (err != nil) {
      return finalHash, fmt.Errorf("extractFromMixer: problem getting funds from mixer: %w", err)
   }  else if (mixerBalance.Cmp(amountToSend) < 0) {
      return finalHash, fmt.Errorf("extractFromMixer: not enough funds in mixer")
   }

   rows, conn, err := getMixerRows()
   if (err != nil) {
      return finalHash, fmt.Errorf("extractFromMixer: get rows: %w", err)
   }

   defer conn.Close(context.Background())

   transitionalAddress, _, err := getNewAddress("", false, true, 0)
   if (err != nil) {
      return finalHash, fmt.Errorf("extractFromMixer: Can't get transitionaladdress: %w", err)
   }
   addressList = append(addressList, transitionalAddress.NanoAddress)

   var totalSent = nt.NewRaw(0)

   var seed int
   var index int
   var balance = nt.NewRaw(0)
   var hashList []nt.BlockHash
   for (rows.Next()) {
      rows.Scan(&seed, &index, balance)

      key, err := getSeedFromIndex(seed, index)
      if (err != nil) {
         return finalHash, fmt.Errorf("extractFromMixer: %w", err)
      }
      setAddressInUse(key.NanoAddress)
      addressList = append(addressList, key.NanoAddress)

      var currentSend = nt.NewRaw(0)
      if (nt.NewRaw(0).Add(totalSent, balance).Cmp(amountToSend) > 0) {
         currentSend = nt.NewRaw(0).Sub(amountToSend, totalSent)

         dirtyAddress, _ = getSeedFromIndex(seed, index)
      } else {
         currentSend = balance
      }

      hash, err := sendNano(key, transitionalAddress.PublicKey, currentSend)
      if (err != nil) {
         return finalHash, fmt.Errorf("extractFromMixer: %w", err)
      }
      hashList = append(hashList, hash)

      totalSent.Add(totalSent, currentSend)

      if (totalSent.Cmp(amountToSend) >= 0) {
         break
      }

   }

   // Make sure they're confirmed.
   waitForConfirmations(hashList)

   hashList, err = ReceiveAll(transitionalAddress.NanoAddress)
   if (err != nil) {
      return finalHash, fmt.Errorf("extractFromMixer: %w", err)
   }

   // enable websockets??
   // Make sure they're confirmed.
   waitForConfirmations(hashList)

   finalHash, err = sendNano(transitionalAddress, publicKey, amountToSend)
   if (err != nil) {
      return finalHash, fmt.Errorf("extractFromMixer: %w", err)
   }

   if (dirtyAddress != nil) {
      sendToMixer(dirtyAddress, 1)
   }

   return finalHash, nil
}
