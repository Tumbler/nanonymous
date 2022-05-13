package main

import (
   "fmt"
   "context"

   // Local packages
   keyMan "nanoKeyManager"

   // 3rd party packages
   pgx "github.com/jackc/pgx/v4"
   "golang.org/x/crypto/blake2b"
)

func updateBalance(nanoAddress string, balance *keyMan.Raw) error {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return fmt.Errorf("updateBalance: %w", err)
   }
   defer conn.Close(context.Background())

   queryString :=
   "UPDATE " +
      "wallets " +
   "SET " +
      "\"balance\" = $1" +
   "WHERE " +
      "\"hash\" = $2;"

   pubKey,  _ := keyMan.AddressToPubKey(nanoAddress)
   nanoAddressHash := blake2b.Sum256(pubKey)
   rowsAffected, err := conn.Exec(context.Background(), queryString, balance, nanoAddressHash[:])
   if (err != nil) {
      return fmt.Errorf("updateBalance: %w", err)
   }
   if (rowsAffected.RowsAffected() < 1) {
      return fmt.Errorf("updateBalance: no rows affected in update")
   }

   return nil
}

func clearPoW(nanoAddress string) error {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return fmt.Errorf("updateBalance: %w", err)
   }
   defer conn.Close(context.Background())

   queryString :=
   "UPDATE " +
      "wallets " +
   "SET " +
      "\"pow\" = ''" +
   "WHERE " +
      "\"hash\" = $1;"

   pubKey,  _ := keyMan.AddressToPubKey(nanoAddress)
   nanoAddressHash := blake2b.Sum256(pubKey)
   rowsAffected, err := conn.Exec(context.Background(), queryString, nanoAddressHash[:])
   if (err != nil) {
      return fmt.Errorf("clearPoW: %w", err)
   }
   if (rowsAffected.RowsAffected() < 1) {
      return fmt.Errorf("clearPoW: no rows affected in update")
   }

   return nil
}

func getWalletFromAddress(nanoAddress string) (int, int, error) {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return 0, 0, fmt.Errorf("getSeedFromAddress: %w", err)
   }
   defer conn.Close(context.Background())

   queryString :=
   "SELECT " +
      "parent_seed, " +
      "index " +
   "FROM " +
      "wallets " +
   "WHERE " +
      "hash = $1;"

   pubkey, err := keyMan.AddressToPubKey(nanoAddress)
   if (err != nil) {
      return 0, 0, fmt.Errorf("getWalletFromAddress: %w", err)
   }

   recivedHash := blake2b.Sum256(pubkey)

   var parentSeed int
   var index int
   err = conn.QueryRow(context.Background(), queryString, recivedHash[:]).Scan(&parentSeed, &index)
   if (err != nil) {
      return  0, 0, fmt.Errorf("getWalletFromAddress: %w", err)
   }
   if (parentSeed == 0) {
      return  0, 0, fmt.Errorf("getWalletFromAddress: address not found in database")
   }

   return parentSeed, index, nil
}

func getSeedFromAddress(nanoAddress string) (keyMan.Key, int, int, error) {
   var key keyMan.Key

   parentSeed, index, err := getWalletFromAddress(nanoAddress)
   if (err != nil) {
      return key, 0, 0, fmt.Errorf("getSeedFromAddress: %w", err)
   }
   key.Seed, _ = getSeedFromDatabase(parentSeed)
   key.Index = index
   keyMan.SeedToKeys(&key)

   return key, parentSeed, index, nil
}

func getSeedFromIndex(seed int, index int) (*keyMan.Key, error) {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return nil, fmt.Errorf("getSeedFromIndex: %w", err)
   }
   defer conn.Close(context.Background())

   queryString :=
   "SELECT " +
      "pgp_sym_decrypt_bytea(seed, $1)" +
   "FROM " +
      "seeds " +
   "WHERE " +
      "id = $2;"

   row, err := conn.Query(context.Background(), queryString, databasePassword, seed)
   if (err != nil) {
      return nil, fmt.Errorf("getSeedFromIndex: %w", err)
   }

   var key keyMan.Key
   if (row.Next()) {
      err = row.Scan(&key.Seed)
      if (err != nil) {
         return nil, fmt.Errorf("getSeedFromIndex: %w ", err)
      } else {
         row.Close()

         key.Index = index
         err = keyMan.SeedToKeys(&key)
         if (err != nil) {
            return nil, fmt.Errorf("getSeedFromIndex: %w", err)
         }
      }
   }

   if (key.NanoAddress == "") {
      return nil, fmt.Errorf("getSeedFromIndex: nil key: either bad address or password")
   }

   return &key, nil
}


func getSeedFromDatabase(id int) ([]byte, error) {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return nil, fmt.Errorf("getNewAddress: %w", err)
   }
   defer conn.Close(context.Background())

   queryString :=
   "SELECT " +
      "pgp_sym_decrypt_bytea(seed, $1) " +
   "FROM " +
      "seeds " +
   "WHERE " +
      "id = $2;"

   var seed []byte
   _ = conn.QueryRow(context.Background(), queryString, databasePassword, id).Scan(&seed)

   return seed, nil
}

func setAddressInUse(nanoAddress string) {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return
   }
   defer conn.Close(context.Background())

   queryString :=
   "UPDATE " +
      "wallets " +
   "SET " +
      "\"in_use\" = TRUE" +
   "WHERE " +
      "\"hash\" = $1;"

   pubKey,  _ := keyMan.AddressToPubKey(nanoAddress)
   nanoAddressHash := blake2b.Sum256(pubKey)
   conn.Exec(context.Background(), queryString, nanoAddressHash[:])
}

func setAddressNotInUse(nanoAddress string) {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return
   }
   defer conn.Close(context.Background())

   queryString :=
   "UPDATE " +
      "wallets " +
   "SET " +
      "\"in_use\" = FALSE" +
   "WHERE " +
      "\"hash\" = $1;"

   pubKey,  _ := keyMan.AddressToPubKey(nanoAddress)
   nanoAddressHash := blake2b.Sum256(pubKey)
   conn.Exec(context.Background(), queryString, nanoAddressHash[:])
}
