package main

import (
   "fmt"
   "math/rand"

   // Local packages
   keyMan "nanoKeyManager"
   nt "nanoTypes"
)

// TODO sometimes errors??
func sendToMixer(key *keyMan.Key, depth int) error {
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

   amount1 := nt.NewRaw(0).Div(bal, nt.NewRaw(int64(rand.Intn(98) + 1)))
   amount2 := nt.NewRaw(0).Sub(bal, amount1)

   fmt.Println("amount1:", amount1)
   fmt.Println("amount2:", amount2)

   tries := 0
   _, err = sendNano(key, mix1.PublicKey, amount1)
   for (err != nil) {
      if (tries < RetryNumber) {
         _, err = sendNano(key, mix1.PublicKey, amount1)
         tries++
         continue
      }
      return fmt.Errorf("sendToMixer: Send 1 error: Orig: %d, %d - %w", origSeed, origIndex, err)
   }
   tries = 0
   send1Seed, send1Index, _ := getWalletFromAddress(mix1.NanoAddress)

   _, err = sendNano(key, mix2.PublicKey, amount2)
   for (err != nil) {
      if (tries < RetryNumber) {
         _, err = sendNano(key, mix2.PublicKey, amount2)
         tries++
         continue
      }
      return fmt.Errorf("sendToMixer: Send 2 error: Orig: %d, %d, Send 1: %d, %d - %w", origSeed, origIndex, send1Seed, send1Index, err)
   }
   tries = 0
   send2Seed, send2Index, _ := getWalletFromAddress(mix1.NanoAddress)

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

   if (depth < 1) {
      err = sendToMixer(mix1, depth+1)
      if (err != nil) {
         return fmt.Errorf("sendToMixer: Orig: %d, %d, Send1: %d, %d, Send2: %d, %d - sendToMixer 2 error: %w", origSeed, origIndex, send1Seed, send1Index, send2Seed, send2Index, err)
      }
      err = sendToMixer(mix2, depth+1)
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
   if (mixerBalance.Cmp(amountNeeded) < 0 || err != nil) {
      return []*keyMan.Key{}, []int{}, []*nt.Raw{}, fmt.Errorf("getFromMixer: not enough funds in mixer")
   } else if (err != nil) {
      return []*keyMan.Key{}, []int{}, []*nt.Raw{}, fmt.Errorf("getFromMixer: problem getting funds from mixer: %w", err)
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
      balances = append(balances, balance)

      totalBalance.Add(totalBalance, balance)
      fmt.Println("Mixer Total Balance:", totalBalance, "/", amountNeeded)

      if (totalBalance.Cmp(amountNeeded) >= 0) {
         break
      }
   }

   return keys, seeds, balances, nil
}
