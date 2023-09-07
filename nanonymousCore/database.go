package main

import (
   "fmt"
   "context"
   "strings"

   // Local packages
   keyMan "nanoKeyManager"
   nt "nanoTypes"

   // 3rd party packages
   pgx "github.com/jackc/pgx/v4"
   "golang.org/x/crypto/blake2b"
)

func updateBalance(nanoAddress string, balance *nt.Raw) error {
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

func getBalance(nanoAddress string) (*nt.Raw, error) {
   var balance = nt.NewRaw(0)
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return balance,fmt.Errorf("getBalance: %w", err)
   }
   defer conn.Close(context.Background())

   queryString :=
   "SELECT " +
      "balance " +
   "FROM " +
      "wallets " +
   "WHERE " +
      "\"hash\" = $1;"

   pubKey,  _ := keyMan.AddressToPubKey(nanoAddress)
   nanoAddressHash := blake2b.Sum256(pubKey)
   err = conn.QueryRow(context.Background(), queryString, nanoAddressHash[:]).Scan(balance)
   if (err != nil) {
      return  balance, fmt.Errorf("getBalance: %w", err)
   }

   return balance, nil
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
      return 0, 0, fmt.Errorf("getWalletFromAddress: %w", err)
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
      return nil, fmt.Errorf("getSeedFromDatabase: %w", err)
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

// WARNING: You are responsible for closing Conn when you're done with it!!
func getSeedRowsFromDatabase() (pgx.Rows, *pgx.Conn, error) {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return nil, conn, fmt.Errorf("getSeedFromDatabase: %w", err)
   }

   queryString :=
   "SELECT " +
      "pgp_sym_decrypt_bytea(seed, $1)," +
      "current_index, " +
      "id " +
   "FROM " +
      "seeds " +
   "WHERE " +
      "current_index <= $2 " +
   "ORDER BY " +
      "id DESC;"

   rows, err := conn.Query(context.Background(), queryString, databasePassword, MAX_INDEX)
   if (err != nil) {
      return nil, conn, fmt.Errorf("getSeedRowsFrom: %w", err)
   }

   return rows, conn, nil
}

// WARNING: You are responsible for closing Conn when you're done with it!!
func getEncryptedSeedRowsFromDatabase() (pgx.Rows, *pgx.Conn, error) {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return nil, conn, fmt.Errorf("getSeedFromDatabase: %w", err)
   }

   queryString :=
   "SELECT " +
      "id, " +
      "seed, " +
      "current_index " +
   "FROM " +
      "seeds " +
   "WHERE " +
      "current_index <= $1 " +
   "ORDER BY " +
      "id;"

   rows, err := conn.Query(context.Background(), queryString, MAX_INDEX)
   if (err != nil) {
      return nil, conn, fmt.Errorf("getSeedRowsFrom: %w", err)
   }

   return rows, conn, nil
}

// WARNING: You are responsible for closing Conn when you're done with it!!
func getWalletRowsFromDatabase() (pgx.Rows, *pgx.Conn, error) {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return nil, conn, fmt.Errorf("getSeedFromDatabase: %w", err)
   }

   queryString :=
   "SELECT " +
      "parent_seed, " +
      "index, " +
      "balance, " +
      "pow, " +
      "in_use, " +
      "receive_only, " +
      "mixer " +
   "FROM " +
      "wallets " +
   "WHERE " +
      "receive_only = FALSE " +
   "ORDER BY " +
      "parent_seed, " +
      "index;"

   rows, err := conn.Query(context.Background(), queryString)
   if (err != nil) {
      return nil, conn, fmt.Errorf("getSeedRowsFrom: %w", err)
   }

   return rows, conn, nil
}

// WARNING: You are responsible for closing Conn when you're done with it!!
func getManagedWalletsRowsFromDatabase(startingPoint int, seed int) (pgx.Rows, *pgx.Conn, error) {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return nil, conn, fmt.Errorf("getSeedFromDatabase: %w", err)
   }

   queryString :=
   "SELECT " +
      "index " +
   "FROM " +
      "wallets " +
   "WHERE " +
      "mixer = FALSE AND " +
      "parent_seed = $1 AND " +
      "index >= $2 " +
   "ORDER BY " +
      "index;"

   rows, err := conn.Query(context.Background(), queryString, seed, startingPoint)
   if (err != nil) {
      return nil, conn, fmt.Errorf("getSeedRowsFrom: %w", err)
   }

   return rows, conn, nil
}

// WARNING: You are responsible for closing Conn when you're done with it!!
func getBlacklistRowsFromDatabase() (pgx.Rows, *pgx.Conn, error) {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return nil, conn, fmt.Errorf("getSeedFromDatabase: %w", err)
   }

   queryString :=
   "SELECT " +
      "hash " +
   "FROM " +
      "blacklist;"

   rows, err := conn.Query(context.Background(), queryString)
   if (err != nil) {
      return nil, conn, fmt.Errorf("getSeedRowsFrom: %w", err)
   }

   return rows, conn, nil
}

// WARNING: You are responsible for closing Conn when you're done with it!!
func getProfitRowsFromDatabase() (pgx.Rows, *pgx.Conn, error) {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return nil, conn, fmt.Errorf("getSeedFromDatabase: %w", err)
   }

   queryString :=
   "SELECT " +
      "id, " +
      "trans_id, " +
      "time, " +
      "nano_gained, " +
      "nano_usd_value " +
   "FROM " +
      "profit_record " +
   "ORDER BY " +
      "trans_id;"

   rows, err := conn.Query(context.Background(), queryString)
   if (err != nil) {
      return nil, conn, fmt.Errorf("getSeedRowsFrom: %w", err)
   }

   return rows, conn, nil
}

// WARNING: You are responsible for closing Conn when you're done with it!!
func getAllWalletRowsFromDatabase() (pgx.Rows, *pgx.Conn, error) {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return nil, conn, fmt.Errorf("getSeedFromDatabase: %w", err)
   }

   queryString :=
   "SELECT " +
      "parent_seed, " +
      "index, " +
      "balance, " +
      "pow, " +
      "in_use, " +
      "receive_only, " +
      "mixer " +
   "FROM " +
      "wallets " +
   "ORDER BY " +
      "parent_seed, " +
      "index;"

   rows, err := conn.Query(context.Background(), queryString)
   if (err != nil) {
      return nil, conn, fmt.Errorf("getSeedRowsFrom: %w", err)
   }

   return rows, conn, nil
}

func getCurrentIndexFromDatabase(id int) (int, error) {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return 0, fmt.Errorf("getCurrentIndexFromDatabase: %w", err)
   }
   defer conn.Close(context.Background())

   queryString :=
   "SELECT " +
      "current_index " +
   "FROM " +
      "seeds " +
   "WHERE " +
      "id = $1;"

   var currentIndex int
   _ = conn.QueryRow(context.Background(), queryString, id).Scan(&currentIndex)

   return currentIndex, nil
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
      "\"in_use\" = TRUE " +
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
      "\"in_use\" = FALSE " +
   "WHERE " +
      "\"hash\" = $1;"

   pubKey,  _ := keyMan.AddressToPubKey(nanoAddress)
   nanoAddressHash := blake2b.Sum256(pubKey)
   conn.Exec(context.Background(), queryString, nanoAddressHash[:])
}

func resetInUse() {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return
   }
   defer conn.Close(context.Background())

   queryString :=
   "UPDATE " +
      "wallets " +
   "SET " +
      "\"in_use\" = FALSE;"

   conn.Exec(context.Background(), queryString)
}

func isAddressInUse(nanoAddress string) (bool, error) {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return false, fmt.Errorf("isAddressInUse: %w", err)
   }
   defer conn.Close(context.Background())

   queryString :=
   "SELECT " +
      "in_use " +
   "FROM " +
      "wallets " +
   "WHERE " +
      "hash = $1;"

   var inUse bool
   pubKey, err := keyMan.AddressToPubKey(nanoAddress)
   if (err != nil) {
      return false, fmt.Errorf("isAddressInUse: %w", err)
   }
   hash := blake2b.Sum256(pubKey)
   conn.QueryRow(context.Background(), queryString, hash[:]).Scan(&inUse)

   return inUse, nil
}

func addressExsistsInDB(nanoAddress string) bool {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return false
   }
   defer conn.Close(context.Background())

   queryString :=
   "SELECT " +
      "hash " +
   "FROM " +
      "wallets " +
   "WHERE " +
      "\"hash\" = $1;"

   pubKey, _ := keyMan.AddressToPubKey(nanoAddress)
   hash := blake2b.Sum256(pubKey)
   row, _ := conn.Query(context.Background(), queryString, hash[:])

   if (row.Next()) {
      return true
   } else {
      return false
   }
}

func addressIsReceiveOnly(nanoAddress string) bool {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return false
   }
   defer conn.Close(context.Background())

   queryString :=
   "SELECT " +
      "receive_only " +
   "FROM " +
      "wallets " +
   "WHERE " +
      "\"hash\" = $1;"

   pubKey, _ := keyMan.AddressToPubKey(nanoAddress)
   hash := blake2b.Sum256(pubKey)
   var receiveOnly bool
   err = conn.QueryRow(context.Background(), queryString, hash[:]).Scan(&receiveOnly)
   if (err != nil) {
      // TODO log err
      fmt.Println("addressIsReceiveOnly: QueryRow failed: %w", err)
      return false
   }

   return receiveOnly
}

func setAddressReceiveOnly(nanoAddress string) {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return
   }
   defer conn.Close(context.Background())

   queryString :=
   "UPDATE " +
      "wallets " +
   "SET " +
      "\"receive_only\" = TRUE " +
   "WHERE " +
      "\"hash\" = $1;"

   pubKey,  _ := keyMan.AddressToPubKey(nanoAddress)
   nanoAddressHash := blake2b.Sum256(pubKey)
   conn.Exec(context.Background(), queryString, nanoAddressHash[:])
}

func setAddressNotReceiveOnly(nanoAddress string) {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return
   }
   defer conn.Close(context.Background())

   queryString :=
   "UPDATE " +
      "wallets " +
   "SET " +
      "\"receive_only\" = FALSE " +
   "WHERE " +
      "\"hash\" = $1;"

   pubKey,  _ := keyMan.AddressToPubKey(nanoAddress)
   nanoAddressHash := blake2b.Sum256(pubKey)
   conn.Exec(context.Background(), queryString, nanoAddressHash[:])
}

func addressIsMixer(nanoAddress string) bool {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return false
   }
   defer conn.Close(context.Background())

   queryString :=
   "SELECT " +
      "mixer " +
   "FROM " +
      "wallets " +
   "WHERE " +
      "\"hash\" = $1;"

   pubKey, _ := keyMan.AddressToPubKey(nanoAddress)
   hash := blake2b.Sum256(pubKey)
   var mixer bool
   err = conn.QueryRow(context.Background(), queryString, hash[:]).Scan(&mixer)
   if (err != nil) {
      // TODO log err
      fmt.Println("addressIsMixer: QueryRow failed: %w", err)
      return false
   }

   return mixer
}

// insertSeed saves an encrytped version of the seed given into the database.
func insertSeed(conn psqlDB, seed []byte) (int, error) {
   var id int

   queryString :=
   "INSERT INTO " +
     "seeds (seed, current_index) " +
   "VALUES " +
     "(pgp_sym_encrypt_bytea($1, $2), -1) " +
   "RETURNING id;"

   rows, err := conn.Query(context.Background(), queryString, seed, databasePassword)
   if (err != nil) {
      return -1, fmt.Errorf("insertSeed: %w", err)
   }

   if (rows.Next()) {
      err = rows.Scan(&id)
      if (err != nil) {
         return -1, fmt.Errorf("insertSeed: %w ", err)
      }
   }

   rows.Close()

   return id, nil
}

// findTotalBalance is a simple function that adds up all the nano there is
// amongst all the wallets and returns the amount in Nano, the amount of nano
// there is in the managed wallets in Raw and the amount of nano there is in the
// mixer in Raw.
// TODO needs to not break if there are no mixer addresses
func findTotalBalance() (*nt.Raw, *nt.Raw, *nt.Raw, error) {
   zero := nt.NewRaw(0)
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return zero, zero, zero, fmt.Errorf("FindTotalBalance: %w", err)
   }
   defer conn.Close(context.Background())

   queryString :=
   "SELECT " +
      "SUM(balance) " +
   "FROM " +
      "wallets;"

   var rawBalance = nt.NewRaw(0)
   var nanoBalance *nt.Raw
   row, err := conn.Query(context.Background(), queryString)
   if (err != nil) {
      return zero, zero, zero, fmt.Errorf("QueryRow failed: %w", err)
   }

   if (row.Next()) {
      err = row.Scan(rawBalance)
      if (err != nil) {
         return zero, zero, zero, fmt.Errorf("findTotalBalance: Query faild on total: %w", err)
      }

      nanoBalance = nt.NewFromRaw(rawBalance)
   }

   row.Close()

   queryString =
   "SELECT " +
      "SUM(balance) " +
   "FROM " +
      "wallets " +
   "WHERE " +
      "mixer = false;"

   err = conn.QueryRow(context.Background(), queryString).Scan(rawBalance)
   if (err != nil) {
      if (strings.Contains(err.Error(), "(<nil>)")) {
         // Just Scan complaining about nil like little baby.
         rawBalance = nt.NewRaw(0)
      } else {
         return zero, zero, zero, fmt.Errorf("findTotalBalance: QueryRow failed on managed: %w", err)
      }
   }

   managed := nt.NewFromRaw(rawBalance)

   queryString =
   "SELECT " +
      "SUM(balance) " +
   "FROM " +
      "wallets " +
   "WHERE " +
      "mixer = TRUE;"

   err = conn.QueryRow(context.Background(), queryString).Scan(rawBalance)
   if (err != nil) {
      if (strings.Contains(err.Error(), "(<nil>)")) {
         // Just Scan complaining about nil like little baby.
         rawBalance = nt.NewRaw(0)
      } else {
         return zero, zero, zero, fmt.Errorf("findTotalBalance: QueryRow failed on mixer: %w", err)
      }
   }

   mixer := nt.NewFromRaw(rawBalance)

   return nanoBalance, managed, mixer, nil
}

func getNextTransactionId() (int, error) {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return -1, fmt.Errorf("getNextTransactionId: %w", err)
   }
   defer conn.Close(context.Background())

   queryString :=
   "UPDATE " +
      "transaction " +
   "SET " +
      "unique_id = unique_id + 1 " +
   "RETURNING " +
      "unique_id;"

   var id int
   err = conn.QueryRow(context.Background(), queryString).Scan(&id)
   if (err != nil) {
      return -1, fmt.Errorf("getNextTransactionId: QueryRow failed: %w", err)
   }

   return id, nil
}

func recordProfit(gross *nt.Raw, tid int) error {

   // Don't bother recording if there was no fee.
   if (gross.Cmp(nt.NewRaw(0)) == 0) {
      return nil
   }

   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return fmt.Errorf("recordProfit: %w", err)
   }
   defer conn.Close(context.Background())

   nanoUsdValue, err := getNanoUSDValue()
   if (err != nil) {
      return fmt.Errorf("recordProfit: %w", err)
   }

   queryString :=
   "INSERT INTO " +
      "profit_record (time, nano_gained, nano_usd_value, trans_id) " +
   "VALUES " +
      "(NOW(), $1, $2, $3);"

   rowsAffected, err := conn.Exec(context.Background(), queryString, gross, nanoUsdValue, tid)
   if (err != nil) {
      return fmt.Errorf("updateBalance: %w", err)
   }
   if (rowsAffected.RowsAffected() < 1) {
      return fmt.Errorf("updateBalance: no rows affected in update")
   }

   return nil
}

// WARNING: You are responsible for closing Conn when you're done with it!!
func getMixerRows() (pgx.Rows, *pgx.Conn, error) {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return nil, conn, fmt.Errorf("getMixerRows: %w", err)
   }

   queryString :=
   "SELECT " +
      "parent_seed, " +
      "index, " +
      "balance " +
   "FROM " +
      "wallets " +
   "WHERE " +
      "balance > 0 AND " +
      "in_use = FALSE AND " +
      "receive_only = FALSE AND " +
      "mixer = TRUE " +
   "ORDER BY " +
      "balance, " +
      "index;"

   rows, err := conn.Query(context.Background(), queryString)
   if (err != nil) {
      return nil, conn, fmt.Errorf("getMixerRows: %w", err)
   }

   return rows, conn, err
}
