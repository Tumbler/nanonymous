package main

import (
   "testing"
   "strings"
   "encoding/hex"
)


func TestGenerateSeedFromMnemonic(t *testing.T) {
   tables := []struct {
      mnemonic string
      seed string
   }{
      {"jeans ecology family expand dress sauce above quality hire hole already joke lady guard such must debris club meadow vendor express agree still vacuum",
       "77A8C14A28142B7F0025796C0D941CBC27C8CEB61C9038658626F9150C0A3577"},
      {"bridge broom happy bonus liberty daughter cabbage fringe rain all notice this actor finger light hero grit tube spot banana child hurry rebel romance",
       "1BC399A38CB80E6F87EAE8B100CE5BF0602AADE0635A66BD434B09127EDF6CCD"},
      {"laugh you jewel skin reopen glory frown assist execute kiwi gallery crater ball fox evolve celery cave exit reduce slam trick abandon wrist man",
       "7D9FE5DFE54B66C717606E4F2F6D7C19411EB893892824C9FED0658E86003F9C"},
   }

   for _, table := range tables {
      var key Key
      err := GenerateSeedFromMnemonic(table.mnemonic, &key)
      if (err != nil) {
         t.Errorf("Error in execution: %s", err.Error())
      } else {
         generatedSeed := strings.ToUpper(hex.EncodeToString(key.seed))

         if (generatedSeed != table.seed) {
            t.Errorf("Generation from \"%s\" was incorrect, \n\rgot:  %s, \n\rwant: %s", table.mnemonic, generatedSeed, table.seed)
         }
      }
   }
}
