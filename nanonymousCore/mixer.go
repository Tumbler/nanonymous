package main

import (
   "fmt"
   "math/rand"

   // Local packages
   keyMan "nanoKeyManager"
   nt "nanoTypes"
)

func sendToMixer(key *keyMan.Key, depth int) error {
   var err error

   fmt.Println(key.NanoAddress)

   // Get new mixer addresses
   mix1, _, err := getNewAddress("", false, true, 0)
   if (err != nil) {
      return fmt.Errorf("sendToMixer: Failed to get new address1: %w:", err)
   }
   mix2, _, err := getNewAddress("", false, true, 0)
   if (err != nil) {
      return fmt.Errorf("sendToMixer: Failed to get new address2: %w:", err)
   }

   // Randomize amounts

   bal, err := getBalance(key.NanoAddress)
   if (err != nil) {
      return fmt.Errorf("sendToMixer: %w", err)
   }

   amount1 := nt.NewRaw(0).Div(bal, nt.NewRaw(int64(rand.Intn(98) + 1)))
   amount2 := nt.NewRaw(0).Sub(bal, amount1)

   fmt.Println("amount1:", amount1)
   fmt.Println("amount2:", amount2)

   _, err = sendNano(key, mix1.PublicKey, amount1)
   if (err != nil) {
      fmt.Println("Send 1 error: %w", err)
      return err
   }
   _, err = sendNano(key, mix2.PublicKey, amount2)
   if (err != nil) {
      fmt.Println("Send 2 error: %w", err)
      return err
   }

   Receive(mix1.NanoAddress)
   Receive(mix2.NanoAddress)

   if (depth < 1) {
      sendToMixer(mix1, depth+1)
      sendToMixer(mix2, depth+1)
   }

   setAddressNotInUse(mix1.NanoAddress)
   setAddressNotInUse(mix2.NanoAddress)

   return err
}
