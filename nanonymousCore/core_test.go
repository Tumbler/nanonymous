package main

import (
   "testing"
   "golang.org/x/crypto/blake2b"
   "encoding/hex"
)


func Test_Password(t *testing.T) {
   const SALT = "E29A8053962DB8E76A7"
   // This is my hash, GET YOUR OWN!
   const P_HASH = "a852f13dbf9f820d492a9afb7dff43bccca5d777cf32f4b4a92a7166d5c434c986f8ec64b1f9859f3b0b48c575e147481090f3c85ad87013ebf42be200658648"

   err := initNanoymousCore()
   if (err != nil) {
      t.Errorf(err.Error())
      return
   }

   salted := databasePassword + SALT
   passwordHash := blake2b.Sum512([]byte(salted))
   hashString := hex.EncodeToString(passwordHash[:])
   if (hashString != P_HASH) {
      t.Errorf("password incorrect!!")
   }
}

