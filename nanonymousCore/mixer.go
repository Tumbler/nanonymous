package main

import (
   "fmt"
   "time"

   // Local packages
   keyMan "nanoKeyManager"
   nt "nanoTypes"
)

// TODO sometimes errors?? Might have fixed with confirmation check.... monitoring
func sendToMixer(key *keyMan.Key, shufflesLeft int) error {
   var err error

   origSeed, origIndex, err := getWalletFromAddress(key.NanoAddress)
   if (err != nil) {
      return fmt.Errorf("sendToMixer: Failed to get wallet: %w", err)
   }

   fmt.Println(key.NanoAddress)

   // Get new mixer addresses
   mix1, _, err := getNewAddress("", false, true, 0)
   if (err != nil) {
      return fmt.Errorf("sendToMixer: Orig: %d, %d, Failed to get new address1: %w:", origSeed, origIndex, err)
   }
   mix2, _, err := getNewAddress("", false, true, 0)
   if (err != nil) {
      return fmt.Errorf("sendToMixer: Orig: %d, %d, Failed to get new address2: %w:", origSeed, origIndex, err)
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

   fmt.Println("amount1:", amount1)
   fmt.Println("amount2:", amount2)

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
   var confirms int
   for (confirms < 2) {
      for i := len(hashList)-1; i >= 0; i-- {
         blockInfo, err := getBlockInfo(hashList[i])
         if (err != nil) {
            if (verbosity >= 5) {
               fmt.Println(fmt.Errorf("sendToMixer warning: %w", err))
            }
         }
         if (blockInfo.Confirmed) {
            confirms++
            hashList[i] = hashList[len(hashList)-1]
            hashList = hashList[:len(hashList)-1]
         }
      }
      if (confirms < 2) {
         time.Sleep(5 * time.Second)
      }
   }

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

   setAddressNotInUse(mix1.NanoAddress)
   setAddressNotInUse(mix2.NanoAddress)

   return err
}

// getKeysFromMixer the transaction manager is very good at making things get
// done properly, so just hand it the keys it needs to get enough funds and let
// it handle the actual sends.
func getKeysFromMixer(amountNeeded *nt.Raw) ([]*keyMan.Key, []int, []*nt.Raw, error) {
   var err error

   _, _, mixerBalance, err := findTotalBalance()

   // Mixer Balance < amount to send
   if (err != nil) {
      return []*keyMan.Key{}, []int{}, []*nt.Raw{}, fmt.Errorf("getFromMixer: problem getting funds from mixer: %w", err)
   }  else if (mixerBalance.Cmp(amountNeeded) < 0) {
      return []*keyMan.Key{}, []int{}, []*nt.Raw{}, fmt.Errorf("getFromMixer: not enough funds in mixer")
   }

   rows, err := getMixerRows()
   if (err != nil) {
      return []*keyMan.Key{}, []int{}, []*nt.Raw{}, fmt.Errorf("getFromMixer: get rows: %w", err)
   }

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
      fmt.Println("Mixer Total Balance:", rawToNANO(totalBalance), "/", rawToNANO(amountNeeded))

      if (totalBalance.Cmp(amountNeeded) >= 0) {
         break
      }
   }

   return keys, seeds, balances, nil
}
