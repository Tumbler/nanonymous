package main

import (
   "testing"
   "golang.org/x/crypto/blake2b"
   "encoding/hex"
   "context"
   "os/exec"
   "strings"

   // 3rd party packages
   pgx "github.com/jackc/pgx/v4"
)


func Test_Password(t *testing.T) {
   const SALT = "E29A8053962DB8E76A7"
   // This is my hash, GET YOUR OWN!
   const P_HASH = "bd7ba7e0de3500733f4066499255ffb3fa155def272c87f259d10148ee5d2bf8613fe492d833795e3350166443cec75e6fb861029f03ff85f25aa0447b1d7b93"

   err := initNanoymousCore()
   if (err != nil) {
      t.Errorf(err.Error())
      return
   }

   salted := databasePassword + SALT
   passwordHash := blake2b.Sum512([]byte(salted))
   hashString := hex.EncodeToString(passwordHash[:])
   if (hashString != P_HASH) {
      t.Errorf("password incorrect!! : %s", hashString)
   }
}

func Test_insertSeed(t *testing.T) {
   databaseUrl = "postgres://test:testing@localhost:5432/gotests"
   databasePassword = "testing"

   // Reset database to known state
   script := exec.Command("psql", "-f", "../scripts/resetTestDatabase.sql", "-U", "test", "-d", "gotests")
   script.Run()

   test1 := []struct {
      inputSeed string
      outputId int
   }{
      {"7DDCC4B0092452DC3BEBC58CA21E9B5591F87D0B6F5F9EC509AAB64EC130660D",
        2},
      {"44334B9E9918954C679C50A57B1E91728FD3CF9EDE769ABD7C8751D353CFDF60",
        3},
      {"2EEC5740213B93B687F4B3C768E5E6597CEEB37793D6F8A55A5F4ED0C7B4003D",
        4},
   }

   conn, _ := pgx.Connect(context.Background(), databaseUrl)
   defer conn.Close(context.Background())

   for _, test := range test1 {
      seedByte, _ := hex.DecodeString(test.inputSeed)
      id, err := insertSeed(conn, seedByte)
      if (err != nil) {
         t.Errorf(err.Error())
         return
      }

      if (id != test.outputId) {
         t.Errorf("Unexpected ID, want: %d, got: %d", test.outputId, id)
         return
      }

      // Check that encryption/decryption works
      queryString :=
      "SELECT " +
         "pgp_sym_decrypt_bytea(seed, $1) " +
      "FROM " +
         "seeds " +
      "WHERE " +
       "\"id\" = $2;"

       var seed []byte
       _ = conn.QueryRow(context.Background(), queryString, databasePassword, id).Scan(&seed)

       seedString := strings.ToUpper(hex.EncodeToString(seed))
       if (seedString != test.inputSeed) {
          t.Errorf("Bad decryption:\r\n want: %s\r\n got:  %s", test.inputSeed, seedString)
       }
   }
}

func Test_getNewAddress(t *testing.T) {
   databaseUrl = "postgres://test:testing@localhost:5432/gotests"
   databasePassword = "testing"

   // Reset database to known state
   script := exec.Command("psql", "-f", "../scripts/resetTestDatabase.sql", "-U", "test", "-d", "gotests")
   script.Run()

   test1 := []struct {
      inputAddress string
      outputBlacklist string
      outputAddress string
   }{
      {"nano_397i3mizjznjinko3b6rz5szigc18jrogzqd9pxz7meo4mqsuwjyshoxm3xq",
       "23776c9dc9ebad79157db84d28f15fe542c40b9a900382d25e1e8268ff3258d1",
       "nano_3f4pznen4utfxmeu7jmucnhg6ut4rd9fk87s7xnnrkr4okph65158j4xciqf" },
      {"nano_3qis7ubfx8ebmybeoks4f3cied1pzfykgp8ejj8gxk3pkdpmjo74gmhzjeba",
       "012c5241ea1d55a3a4d01e4689ffd86cb81ae995be222c1e412f41f6a5970988",
       "nano_1osskweb73zsqnjj638st4cjmf9s56hdnb7bh941iwkb9qszamxg5seeadhw" },
      {"nano_1b9opg96jquha58fueg8g8zjofauw5ua3pqfer6bdnjjj34x51s6j8dq83hh",
       "64d89bab0dca839602b7194cdc8620fff3db11353c71f5fc3393f374ce6ab30b",
       "nano_3whzcsaf9xq56dftxxnc1z554s7ii6gdp8r1jti5sdarh73qcfaj6xxpd5ui" },
   }

   for _, test := range test1 {
      key, _, err := getNewAddress(test.inputAddress)
      if (err != nil) {
         t.Errorf("Error during execution: %s", err.Error())
      }

      // Check that returned address is correct
      if (key.NanoAddress != test.outputAddress) {
         t.Errorf("Bad address\r\n want: %s\r\n got:  %s", test.outputAddress, key.NanoAddress)
      }

      // Check that blacklist was correct
      conn, _ := pgx.Connect(context.Background(), databaseUrl)
      defer conn.Close(context.Background())

      queryString :=
      "SELECT " +
         "hash " +
      "FROM " +
         "blacklist " +
      "WHERE " +
         "\"hash\" = $1;"

      blacklistByte, _ := hex.DecodeString(test.outputBlacklist)
      row, err := conn.Query(context.Background(), queryString, blacklistByte)
      if (err != nil) {
         t.Errorf("Error during execution: %s", err.Error())
      }

      if !(row.Next()) {
         t.Errorf("Blacklist not found\r\n %s", test.outputBlacklist)
      }

   }
}
