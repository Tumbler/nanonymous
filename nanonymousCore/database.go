package main

import (
   "fmt"
   "context"
   "strings"
   "strconv"
   "encoding/hex"
   "time"

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
      "\"balance\" = $1 " +
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
      return nil, conn, fmt.Errorf("getSeedRowsFromDatabase: %w", err)
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
      return nil, conn, fmt.Errorf("getSeedRowsFromDatabase: %w", err)
   }

   return rows, conn, nil
}

// WARNING: You are responsible for closing Conn when you're done with it!!
func getEncryptedSeedRowsFromDatabase() (pgx.Rows, *pgx.Conn, error) {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return nil, conn, fmt.Errorf("getEncryptedSeedRowsFromDatabase: %w", err)
   }

   queryString :=
   "SELECT " +
      "id, " +
      "seed, " +
      "current_index, " +
      "active " +
   "FROM " +
      "seeds " +
   "WHERE " +
      "current_index <= $1 " +
   "ORDER BY " +
      "id;"

   rows, err := conn.Query(context.Background(), queryString, MAX_INDEX)
   if (err != nil) {
      return nil, conn, fmt.Errorf("getEncryptedSeedRowsFromDatabase: %w", err)
   }

   return rows, conn, nil
}

// WARNING: You are responsible for closing Conn when you're done with it!!
func getWalletRowsFromDatabase() (pgx.Rows, *pgx.Conn, error) {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return nil, conn, fmt.Errorf("getWalletRowsFromDatabase: %w", err)
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
      return nil, conn, fmt.Errorf("getWalletRowsFromDatabase: %w", err)
   }

   return rows, conn, nil
}

// WARNING: You are responsible for closing Conn when you're done with it!!
func getWalletRowsFromDatabaseFromSeed(seedID int) (pgx.Rows, *pgx.Conn, error) {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return nil, conn, fmt.Errorf("getWalletRowsFromDatabaseFromSeed: %w", err)
   }

   queryString :=
   "SELECT " +
      "index " +
   "FROM " +
      "wallets " +
   "WHERE " +
      "in_use = FALSE AND " +
      "parent_seed = $1 " +
   "ORDER BY " +
      "index;"

   rows, err := conn.Query(context.Background(), queryString, seedID)
   if (err != nil) {
      return nil, conn, fmt.Errorf("getWalletRowsFromDatabaseFromSeed: %w", err)
   }

   return rows, conn, nil
}

// WARNING: You are responsible for closing Conn when you're done with it!!
func getWalletRowsForRetirement(seedID int) (pgx.Rows, *pgx.Conn, error) {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return nil, conn, fmt.Errorf("getWalletRowsForRetirement: %w", err)
   }

   queryString :=
   "SELECT " +
      "index " +
   "FROM " +
      "wallets " +
   "WHERE " +
      "balance > 0 AND " +
      "in_use = FALSE AND " +
      "mixer = FALSE AND " +
      "parent_seed = $1 " +
   "ORDER BY " +
      "index;"

   rows, err := conn.Query(context.Background(), queryString, seedID)
   if (err != nil) {
      return nil, conn, fmt.Errorf("getWalletRowsForRetirement: %w", err)
   }

   return rows, conn, nil
}

// WARNING: You are responsible for closing Conn when you're done with it!!
func getManagedWalletsRowsFromDatabase(seed int) (pgx.Rows, *pgx.Conn, error) {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return nil, conn, fmt.Errorf("getManagedWalletsRowsFromDatabase: %w", err)
   }

   queryString :=
   "SELECT " +
      "index " +
   "FROM " +
      "wallets " +
   "WHERE " +
      "mixer = FALSE AND " +
      "parent_seed = $1 " +
   "ORDER BY " +
      "index DESC;"

   rows, err := conn.Query(context.Background(), queryString, seed)
   if (err != nil) {
      return nil, conn, fmt.Errorf("getManagedWalletsRowsFromDatabase: %w", err)
   }

   return rows, conn, nil
}

// WARNING: You are responsible for closing Conn when you're done with it!!
func getBlacklistRowsFromDatabase() (pgx.Rows, *pgx.Conn, error) {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return nil, conn, fmt.Errorf("getBlacklistRowsFromDatabase: %w", err)
   }

   queryString :=
   "SELECT " +
      "hash, " +
      "seed_id " +
   "FROM " +
      "blacklist;"

   rows, err := conn.Query(context.Background(), queryString)
   if (err != nil) {
      return nil, conn, fmt.Errorf("getBlacklistRowsFromDatabase: %w", err)
   }

   return rows, conn, nil
}

// WARNING: You are responsible for closing Conn when you're done with it!!
func getProfitRowsFromDatabase() (pgx.Rows, *pgx.Conn, error) {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return nil, conn, fmt.Errorf("getProfitRowsFromDatabase: %w", err)
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
      return nil, conn, fmt.Errorf("getProfitRowsFromDatabase: %w", err)
   }

   return rows, conn, nil
}

// WARNING: You are responsible for closing Conn when you're done with it!!
func getAllWalletRowsFromDatabase() (pgx.Rows, *pgx.Conn, error) {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return nil, conn, fmt.Errorf("getAllWalletRowsFromDatabase: %w", err)
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
      return nil, conn, fmt.Errorf("getAllWalletRowsFromDatabase: %w", err)
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
   err = conn.QueryRow(context.Background(), queryString, id).Scan(&currentIndex)
   if (err != nil) {
      return 0, fmt.Errorf("getCurrentIndexFromDatabase: %w", err)
   }

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

func resetInUse() error {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return fmt.Errorf("resetInUse: %w", err)
   }
   defer conn.Close(context.Background())

   queryString :=
   "UPDATE " +
      "wallets " +
   "SET " +
      "\"in_use\" = FALSE;"

   conn.Exec(context.Background(), queryString)

   return nil
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
      Warning.Println("addressIsReceiveOnly: QueryRow failed: ", err)
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
      Warning.Println("addressIsMixer: QueryRow failed: ", err)
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
         if (strings.Contains(err.Error(), "(<nil>)")) {
            // Just Scan complaining about nil like little baby.
            rawBalance = nt.NewRaw(0)
         } else {
            return zero, zero, zero, fmt.Errorf("findTotalBalance: Query faild on total: %w", err)
         }
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

func peekAtNextTransactionId() (int, error) {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return -1, fmt.Errorf("peekAtNextTransactionId: %w", err)
   }
   defer conn.Close(context.Background())

   queryString :=
   "SELECT " +
      "unique_id " +
   "FROM " +
      "transaction;"

   var id int
   err = conn.QueryRow(context.Background(), queryString).Scan(&id)
   if (err != nil) {
      return -1, fmt.Errorf("peekAtNextTransactionId: QueryRow failed: %w", err)
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

func getReadyMixerFunds() (*nt.Raw, error) {
   zero := nt.NewRaw(0)
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return zero, fmt.Errorf("getReadyMixerFunds: %w", err)
   }

   queryString :=
   "SELECT " +
      "SUM(balance) " +
   "FROM " +
      "wallets " +
   "WHERE " +
      "in_use = FALSE AND " +
      "mixer = TRUE;"

   var rawBalance = nt.NewRaw(0)
   err = conn.QueryRow(context.Background(), queryString).Scan(rawBalance)
   if (err != nil) {
      return zero, fmt.Errorf("getReadyMixerFunds: %w", err)
   }

   return rawBalance, err
}

func setSeedInactive() (int, error) {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return 0, fmt.Errorf("setSeedInactive: %w", err)
   }

   queryString :=
   "UPDATE " +
      "seeds " +
   "SET " +
      "active = FALSE " +
   "WHERE " +
      "active = TRUE AND " +
      "id = (SELECT MAX(id) FROM seeds WHERE active = TRUE) " +
   "RETURNING " +
      "id;"

   var id int
   err = conn.QueryRow(context.Background(), queryString).Scan(&id)
   if (err != nil) {
      return 0, fmt.Errorf("setSeedInactive: %w", err)
   }

   return id, nil
}

// This is defined as 1 - the highest seed.
func getOldestActiveSeedIndex() (int, error) {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return 0, fmt.Errorf("getOldestActiveSeedIndex: %w", err)
   }

   queryString :=
   "SELECT " +
      "MAX(id) " +
   "FROM " +
      "seeds;"

   var seedID int
   err = conn.QueryRow(context.Background(), queryString).Scan(&seedID)
   if (err != nil) {
      return 0, fmt.Errorf("getOldestActiveSeedIndex: %w", err)
   }

   return seedID - 1, nil
}

func isSeedActive(seedID int) (bool, error) {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return true, fmt.Errorf("isSeedActive: %w", err)
   }

   queryString :=
   "SELECT " +
      "active " +
   "FROM " +
      "seeds " +
   "WHERE " +
      "id = $1;"

   var active bool
   err = conn.QueryRow(context.Background(), queryString, seedID).Scan(&active)
   if (err != nil) {
      return true, fmt.Errorf("isSeedActive: %w", err)
   }

   return active, nil
}

// This is a scary fundtion. Probably don't call this. Use pruneBlacklist()
// instead.
func deleteBlacklist(seedID int) error {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return fmt.Errorf("deleteBlacklist: %w", err)
   }

   queryString :=
   "DELETE FROM " +
      "blacklist " +
   "WHERE " +
      "seed_id = $1;"

   rowsAffected, err := conn.Exec(context.Background(), queryString, seedID)
   if (err != nil) {
      return fmt.Errorf("deleteBlacklist: %w", err)
   }
   if (rowsAffected.RowsAffected() < 1) {
      Warning.Println("deleteBlacklist: no rows affected in update")
   }

   return nil
}

func balanceInSeed(seedID int) (*nt.Raw, error) {
   zero := nt.NewRaw(0)
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return zero, fmt.Errorf("balanceInSeed: %w", err)
   }

   queryString :=
   "SELECT " +
      "SUM(balance) " +
   "FROM " +
      "wallets "+
   "WHERE " +
      "parent_seed = $1;"

   var balance = nt.NewRaw(0)
   err = conn.QueryRow(context.Background(), queryString, seedID).Scan(balance)
   if (err != nil) {
      return zero, fmt.Errorf("balanceInSeed: %w", err)
   }

   return balance, nil
}

func getProfitSince(date time.Time) (*nt.Raw, error) {
   zero := nt.NewRaw(0)
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return zero, fmt.Errorf("getProfitSince: %w", err)
   }

   queryString :=
   "SELECT " +
      "SUM(nano_gained) " +
   "FROM " +
      "profit_record "+
   "WHERE " +
      "time > $1;"

   if (date.IsZero()) {
      // Zero time (Make sure it's inititialized)
      date = time.Time{}
   }

   var gross = nt.NewRaw(0)
   err = conn.QueryRow(context.Background(), queryString, date).Scan(gross)
   if (err != nil) {
      return zero, fmt.Errorf("getProfitSince: %w", err)
   }

   return gross, nil
}

func getUSDSinceAtTimeOfTransaction(date time.Time) (float64, error) {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return 0.0, fmt.Errorf("getUSDSinceAtTimeOfTransaction: %w", err)
   }

   queryString :=
   "SELECT " +
      "SUM((nano_gained / 10^30) * nano_usd_value) " +
   "FROM " +
      "profit_record "+
   "WHERE " +
      "time > $1;"

   if (date.IsZero()) {
      // Zero time (Make sure it's inititialized)
      date = time.Time{}
   }

   var USD float64
   err = conn.QueryRow(context.Background(), queryString, date).Scan(&USD)
   if (err != nil) {
      return 0.0, fmt.Errorf("getUSDSinceAtTimeOfTransaction: %w", err)
   }


   return USD, nil
}

func getNumOfTransactionsSince(date time.Time) (int, error) {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return 0, fmt.Errorf("getNumOfTransactionsSince: %w", err)
   }

   queryString :=
   "SELECT " +
      "COUNT(nano_gained) " +
   "FROM " +
      "profit_record "+
   "WHERE " +
      "time > $1;"

   var num int
   err = conn.QueryRow(context.Background(), queryString, date).Scan(&num)
   if (err != nil) {
      return 0, fmt.Errorf("getNumOfTransactionsSince: %w", err)
   }

   return num, nil
}

// WARNING: You are responsible for closing Conn when you're done with it!!
func getRowsOfWalletsWithAnyBalance() (pgx.Rows, *pgx.Conn, error) {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return nil, conn, fmt.Errorf("getRowsOfWalletsWithAnyBalance: %w", err)
   }

   queryString :=
   "SELECT " +
      "parent_seed, " +
      "index " +
   "FROM " +
      "wallets " +
   "WHERE " +
      "balance > 0;"

   rows, err := conn.Query(context.Background(), queryString)
   if (err != nil) {
      return nil, conn, fmt.Errorf("getRowsOfWalletsWithAnyBalance: %w", err)
   }

   return rows, conn, nil
}

func upsertTransactionRecord(t *Transaction) error {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return fmt.Errorf("updateTransactionRecord: %w", err)
   }

   var stamps []time.Time
   for _, delay := range t.delays {
      stamps = append(stamps, time.Now().Add(time.Duration(delay) * time.Second))
   }

   var sendingKeys [][]string
   var longestArray int
   for i, keys := range t.sendingKeys {
      sendingKeys = append(sendingKeys, make([]string, 0))

      if (len(keys) == 0) {
         sendingKeys[i] = append(sendingKeys[i], "0,0")
      }
      for j, key := range keys {
         if (key != nil) {
            seedId, index, err := getWalletFromAddress(key.NanoAddress)
            if (err != nil) {
               Warning.Println("upsertTransactionRecord: getting sending keys: ", err)
               fmt.Println("upsertTransactionRecord: getting sending keys: ", err)
            }
            sendingKeys[i] = append(sendingKeys[i], strconv.Itoa(seedId) +","+ strconv.Itoa(index))
         } else {
            sendingKeys[i] = append(sendingKeys[i], "0,0")
         }
         if (j+1 > longestArray) {
            longestArray = j+1
         }
      }
   }
   // Append dummy values because postgres's multi-dimensional arrays support is lacking.
   for i, keys := range sendingKeys {
      for j := len(keys); j < longestArray; j++ {
         sendingKeys[i] = append(sendingKeys[i], "-1,-1")
      }
   }

   var transitionalKey []string
   for i, key := range t.transitionalKey {
      if (t.multiSend[i]) {
         seedId, index, err := getWalletFromAddress(key.NanoAddress)
         if (err != nil) {
            fmt.Println("upsertTransactionRecord: getting transitional keys: ", err)
         }
         transitionalKey = append(transitionalKey, strconv.Itoa(seedId) +","+ strconv.Itoa(index))
      } else {
         transitionalKey = append(transitionalKey, "0,0")
      }
   }

   var finalHash string
   // Turn into one long string so we can encrypt it.
   for _, hash := range t.finalHash {
      hashString := hash.String()

      finalHash += hashString +","
   }
   if (finalHash[len(finalHash)-1] == ',') {
      finalHash = finalHash[:len(finalHash)-1]
   }

   queryString :=
   "INSERT INTO " +
      "delayed_transactions ( " +
         "id, " +
         "timestamps, " +
         "paymentAddress, " +
         "paymentParentSeedId, " +
         "paymentIndex, " +
         "payment, " +
         "receiveHash, " +
         "recipientAddress, " +
         "fee, " +
         "amountToSend, " +
         "sendingKeys, " +
         "transitionalKey, " +
         "finalHash, " +
         "percents, " +
         "bridge, " +
         "numSubSends, " +
         "dirtyAddress, " +
         "multisend, " +
         "transactionSuccessful " +
      ") " +
   "VALUES " +
      "($1, $2, pgp_sym_encrypt_bytea($3, $20), $4, $5, $6, pgp_sym_encrypt_bytea($7, $20), pgp_sym_encrypt_bytea($8, $20), $9, $10, $11, $12, pgp_sym_encrypt_bytea($13, $20), $14, $15, $16, $17, $18, $19) " +
   "ON CONFLICT " +
      "(id) " +
   "DO UPDATE " +
      "SET " +
         "sendingKeys = $11, " +
         "transitionalKey = $12, " +
         "finalHash = pgp_sym_encrypt_bytea($13, $20), " +
         "dirtyAddress = $17, " +
         "transactionSuccessful = $19 " +
      "WHERE " +
         "delayed_transactions.id = $1;"

   rowsAffected, err := conn.Exec(context.Background(), queryString,
      t.id,
      stamps,
      t.paymentAddress,
      t.paymentParentSeedId,
      t.paymentIndex,
      t.payment,
      t.receiveHash,
      t.recipientAddress,
      t.fee,
      nt.RawArrayToPostgres(t.amountToSend),
      sendingKeys,
      transitionalKey,
      finalHash,
      t.percents,
      t.bridge,
      t.numSubSends,
      t.dirtyAddress,
      t.multiSend,
      t.transactionSuccessful,
      databasePassword,
   )
   if (err != nil) {
      return fmt.Errorf("upsertTransactionRecord INSERT: %w", err)
   }
   if (rowsAffected.RowsAffected() < 1) {
      return fmt.Errorf("upsertTransactionRecord: no rows affected in update")
   }


   return nil
}

func getTranscationRecord(id int, t *Transaction) error {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return fmt.Errorf("getTranscationRecord: %w", err)
   }

   queryString :=
   "SELECT " +
      "numSubSends " +
   "FROM " +
      "delayed_transactions " +
   "WHERE " +
      "id = $1";

   var numSubSends int
   err = conn.QueryRow(context.Background(), queryString, id).Scan(&numSubSends)
   if (err != nil) {
      return fmt.Errorf("getTranscationRecord first query: %w", err)
   }

   fmt.Println(" subsends:", numSubSends)

   // Init what needs to be initted
   t.payment = nt.NewRaw(0)
   t.fee = nt.NewRaw(0)
   for (len(t.amountToSend) < numSubSends) {
      t.amountToSend = append(t.amountToSend, nt.NewRaw(0))
   }

   queryString =
   "SELECT " +
      "id, " +
      "timestamps, " +
      "pgp_sym_decrypt_bytea(paymentAddress, $2), " +
      "paymentParentSeedId, " +
      "paymentIndex, " +
      "payment, " +
      "pgp_sym_decrypt_bytea(receiveHash, $2), " +
      "pgp_sym_decrypt_bytea(recipientAddress, $2),  " +
      "fee, " +
      "amountToSend, " +
      "sendingKeys, " +
      "transitionalKey, " +
      "pgp_sym_decrypt_bytea(finalHash, $2), " +
      "percents, " +
      "bridge, " +
      "numSubSends, " +
      "dirtyAddress, " +
      "multisend, " +
      "transactionSuccessful " +
   "FROM " +
      "delayed_transactions " +
   "WHERE " +
      "id = $1;"

   var dates []time.Time
   var sendingKeys [][]string
   var transitionalKey []string
   var finalHash []byte
   err = conn.QueryRow(context.Background(), queryString, id, databasePassword).Scan(
         &t.id,
         &dates,
         &t.paymentAddress,
         &t.paymentParentSeedId,
         &t.paymentIndex,
          t.payment,
         &t.receiveHash,
         &t.recipientAddress,
          t.fee,
         (*nt.RawArray)(&t.amountToSend),
         &sendingKeys,
         &transitionalKey,
         &finalHash,
         &t.percents,
         &t.bridge,
         &t.numSubSends,
         &t.dirtyAddress,
         &t.multiSend,
         &t.transactionSuccessful,
      )

   if (err != nil) {
      return fmt.Errorf("getTranscationRecord second query: %w", err)
   }

   // Populate delays
   now := time.Now()
   for i, date := range dates {
      if (len(t.delays) <= i) {
         t.delays = append(t.delays, int(date.Sub(now).Seconds()))
      } else {
         t.delays[i] = int(date.Sub(now).Seconds())
      }
   }

   // Init arrays
   for i := 0; i < t.numSubSends; i++ {
      if (len(t.sendingKeys) <= i) {
         t.sendingKeys = append(t.sendingKeys, make([]*keyMan.Key, 0))
      }
      if (len(t.transitionalKey) <= i) {
         t.transitionalKey = append(t.transitionalKey, new(keyMan.Key))
      }
      if (len(t.transitionSeedId) <= i) {
         t.transitionSeedId = append(t.transitionSeedId, 0)
      }
      if (len(t.walletSeed) <= i) {
         t.walletSeed = append(t.walletSeed, make([]int, 0))
      }
      if (len(t.walletBalance) <= i) {
         t.walletBalance = append(t.walletBalance, make([]*nt.Raw, 0))
      }
      if (len(t.individualSendAmount) <= i) {
         t.individualSendAmount = append(t.individualSendAmount, make([]*nt.Raw, 0))
      }
   }

   // Populate sending keys
   for j, array := range sendingKeys {
      for i, IDindex := range array {
         nums := strings.Split(IDindex, ",")
         if (len(nums) < 2) {
            return fmt.Errorf("getTranscationRecord: Not enough integers in sendingKeys.")
         }

         seed, err := strconv.Atoi(nums[0])
         if (err != nil) {
            return fmt.Errorf("getTranscationRecord: Bad data in seed ID: %w", err)
         }
         id, err := strconv.Atoi(nums[1])
         if (err != nil) {
            return fmt.Errorf("getTranscationRecord: Bad data in index ID: %w", err)
         }

         if (seed > 0) {
            key, err := getSeedFromIndex(seed, id)
            if (err != nil) {
               return fmt.Errorf("getTranscationRecord: Coudn't get seed: %w", err)
            }

            if (len(t.sendingKeys[j]) <= i) {
               t.sendingKeys[j] = append(t.sendingKeys[j], key)
            } else {
               t.sendingKeys[j][i] = key
            }
         }
      }
   }

   // Populate transitional keys
   for i, IDindex := range transitionalKey {
      nums := strings.Split(IDindex, ",")
      if (len(nums) < 2) {
         return fmt.Errorf("getTranscationRecord: Not enough integers in transitionalKey.")
      }

      seed, err := strconv.Atoi(nums[0])
      if (err != nil) {
         return fmt.Errorf("getTranscationRecord: Bad data in seed ID: %w", err)
      }
      id, err := strconv.Atoi(nums[1])
      if (err != nil) {
         return fmt.Errorf("getTranscationRecord: Bad data in index ID: %w", err)
      }

      if (seed != 0) {
         key, err := getSeedFromIndex(seed, id)
         if (err != nil) {
            return fmt.Errorf("getTranscationRecord: Coudn't get seed: %w", err)
         }

         if (len(t.transitionalKey) <= i) {
            t.transitionalKey = append(t.transitionalKey, key)
         } else {
            t.transitionalKey[i] = key
         }
      }
   }

   // Populate final Hashes
   // This as an array but stored as plaintext so that we can encrypt it.
   for _, hash := range strings.Split(string(finalHash), ",") {
      hexString, err := hex.DecodeString(hash)
      if (err != nil) {
         Error.Println("getTranscationRecord: Decode String: ", err)
      }
      t.finalHash = append(t.finalHash, hexString)
   }

   return nil
}

func deleteTransactionRecord (id int) error {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return fmt.Errorf("deleteTransactionRecord: %w", err)
   }

   queryString :=
   "DELETE " +
   "FROM " +
      "delayed_transactions " +
   "WHERE " +
      "id = $1";

   _, err = conn.Exec(context.Background(), queryString, id)
   if (err != nil) {
      return fmt.Errorf("deleteTransactionRecord: %w", err)
   }

   return nil
}

func getDelayedIds() ([]int, []string, error) {
   conn, err := pgx.Connect(context.Background(), databaseUrl)
   if (err != nil) {
      return []int{}, []string{}, fmt.Errorf("getDelayedIds: %w", err)
   }

   queryString :=
   "SELECT " +
      "id, " +
      "pgp_sym_decrypt_bytea(paymentAddress, $1) " +
   "FROM " +
      "delayed_transactions " +
   "ORDER BY " +
      "id;"

   rows, err := conn.Query(context.Background(), queryString, databasePassword)
   if (err != nil) {
      return []int{}, []string{}, fmt.Errorf("getDelayedIds: %w", err)
   }

   var id int
   var ids []int
   var address []byte
   var addresses []string
   for (rows.Next()) {
      rows.Scan(&id, &address)
      ids = append(ids, id)
      pa, _ := keyMan.PubKeyToAddress(address)
      addresses = append(addresses, pa)
   }

   return ids, addresses, nil
}
