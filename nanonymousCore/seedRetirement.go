package main

import (
   "fmt"
   "context"
)

// retireCurrentSeed sets the most recent seed to inactive, mixes the previous
// seed, and prunes the blacklist. Can be called with launch option -r.
func retireCurrentSeed() error {

   retireSeedID, err := setSeedInactive()
   if (err != nil) {
      Warning.Println("retireCurrentSeed: ", err)
      return fmt.Errorf("retireCurrentSeed: %w", err)
   }
   if (retireSeedID == 0) {
      // nothing to retire.
      return nil
   }
   Info.Println("Retiring seed:", retireSeedID)
   if (verbosity >= 5) {
      fmt.Println("Retiring seed:", retireSeedID)
   }

   oldSeedID, err := getOldestActiveSeedIndex()
   if (err != nil) {
      Warning.Println("retireCurrentSeed: ", err)
      return fmt.Errorf("retireCurrentSeed: %w", err)
   }

   if (oldSeedID > 0) {
      Info.Println("Mixing seed:", oldSeedID)
      if (verbosity >= 5) {
         fmt.Println("Mixing seed:", oldSeedID)
      }

      rows, conn, err := getWalletRowsForRetirement(oldSeedID)
      if (err != nil) {
         Warning.Println("retireCurrentSeed: ", err)
         return fmt.Errorf("retireCurrentSeed: %w", err)
      }
      defer conn.Close(context.Background())

      var index int
      for (rows.Next()) {
         err := rows.Scan(&index)
         if (err != nil) {
            return fmt.Errorf("retireCurrentSeed: %w", err)
         }

         key, err := getSeedFromIndex(oldSeedID, index)
         if (err != nil) {
            return fmt.Errorf("retireCurrentSeed: %w", err)
         }

         err = sendToMixer(key, 1)
         if (err != nil) {
            return fmt.Errorf("retireCurrentSeed: %w", err)
         }
      }

      if (verbosity >= 5) {
         fmt.Println("Pruning blacklist")
      }
      err = pruneBlacklist(oldSeedID)
      if (err != nil) {
         Warning.Println("retireCurrentSeed: ", err)
         return fmt.Errorf("retireCurrentSeed: %w", err)
      }
   }

   return nil
}

func pruneBlacklist(seedID int) error {
   seedActive, err := isSeedActive(seedID)
   if (err != nil) {
      return fmt.Errorf("pruneBlacklist: %w", err)
   }

   if (seedActive) {
      return fmt.Errorf("pruneBlacklist: Cannot prune active seed")
   }

   err = deleteBlacklist(seedID)
   if (err != nil) {
      return fmt.Errorf("pruneBlacklist: %w", err)
   }

   return nil
}
