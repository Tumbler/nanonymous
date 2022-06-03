package main

import (
   "testing"
   "golang.org/x/crypto/blake2b"
   "encoding/hex"
   "context"
   "os/exec"
   "strings"

   // Local packages
   keyMan "nanoKeyManager"
   nt "nanoTypes"

   // 3rd party packages
   pgx "github.com/jackc/pgx/v4"
)

const resetScript = "../scripts/resetTestDatabase.sql"

func Test_Password(t *testing.T) {
   const SALT = "E29A8053962DB8E76A7"
   // This is my hash, GET YOUR OWN!
   const P_HASH = "bd7ba7e0de3500733f4066499255ffb3fa155def272c87f259d10148ee5d2bf8613fe492d833795e3350166443cec75e6fb861029f03ff85f25aa0447b1d7b93"

   err := initNanoymousCore(false)
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
   script := exec.Command("psql", "-f", resetScript, "-U", "test", "-d", "gotests")
   script.Run()
   inTesting = true

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
   script := exec.Command("psql", "-f", resetScript, "-U", "test", "-d", "gotests")
   script.Run()
   inTesting = true

   test1 := []struct {
      inputAddress string
      outputBlacklist string
      outputAddress string
   }{
      {"nano_3qis7ubfx8ebmybeoks4f3cied1pzfykgp8ejj8gxk3pkdpmjo74gmhzjeba",
       "012c5241ea1d55a3a4d01e4689ffd86cb81ae995be222c1e412f41f6a5970988",
       "nano_1osskweb73zsqnjj638st4cjmf9s56hdnb7bh941iwkb9qszamxg5seeadhw" },
      {"nano_1b9opg96jquha58fueg8g8zjofauw5ua3pqfer6bdnjjj34x51s6j8dq83hh",
       "64d89bab0dca839602b7194cdc8620fff3db11353c71f5fc3393f374ce6ab30b",
       "nano_3whzcsaf9xq56dftxxnc1z554s7ii6gdp8r1jti5sdarh73qcfaj6xxpd5ui" },
      {"nano_34jz5qi36gkemhncpu3hbnzfjg1pam4b49fhg8oo9g96c795q9dz3e3n19bb",
       "25b16d3f62c5f93d2a4ffff8da3cc7c28fe714d374329973b569c7e08e7475a5",
       "nano_3gwfb61goagc5pftnghkpy85rf4qszkcp5e1pczo9qhqgxqwzoiby75ybwuj" },
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

func Test_findTotalBalance(t *testing.T) {
   databaseUrl = "postgres://test:testing@localhost:5432/gotests"
   databasePassword = "testing"

   // Reset database to known state
   script := exec.Command("psql", "-f", resetScript, "-U", "test", "-d", "gotests")
   script.Run()
   inTesting = true

   balance, err := findTotalBalance()
   if (err != nil) {
      t.Errorf("Error during execution: %s", err.Error())
   }

   if (balance != 44.8) {
      t.Errorf("Bad balance, want: %.1f, got %.1f", 45.8, balance)
   }
}

func Test_receivedNano(t *testing.T) {
   databaseUrl = "postgres://test:testing@localhost:5432/gotests"
   databasePassword = "testing"

   // Reset database to known state
   script := exec.Command("psql", "-f", resetScript, "-U", "test", "-d", "gotests")
   script.Run()
   inTesting = true

   test1 := []struct {
      nanoAddress string
      clientAddress string
      seedId int
      index int
      nanoReceived *nt.Raw
      balances []*nt.Raw
      numOfPendingTxs int
      intermediaryTx []*nt.Raw
   }{
      {"nano_3f4pznen4utfxmeu7jmucnhg6ut4rd9fk87s7xnnrkr4okph65158j4xciqf",
       "nano_1bgho34hpofn4sxencbr8916sbbyyoosr5mmepewyguo8te15qkq8hefnrdn",
       1,
       3,
       nt.NewRaw(0).Mul(nt.NewRaw(10), nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil)),
 []*nt.Raw{nt.NewRaw(0).Mul(nt.NewRaw(3102), nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(28), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(6),    nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(29), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(32),   nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(29), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(10),   nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil))},
       1,
       []*nt.Raw{},
      },
      {"nano_1ts8ejswbndstgp6r4wgi7yr593rg7ryab4wuzburmay3pxbrgu3i5f1fz3n",
       "nano_14gfu8wkz48o3xf869ehp7rd9oah1993d1deguqknkksidp5s4b46czn86sw",
       1,
       0,
       nt.NewRaw(0).Mul(nt.NewRaw(9), nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil)),
 []*nt.Raw{nt.NewRaw(0).Mul(nt.NewRaw(4002), nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(28), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(6),    nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(29), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(32),   nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(29), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(1018), nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(27), nil))},
       1,
       []*nt.Raw{},
      },
      {"nano_1ie4own1s5qmmyd33u9a64169ox54kdb3khs1yt84gfgd7n7dshgcjkegxei",
       "nano_14gfu8wkz48o3xf869ehp7rd9oah1993d1deguqknkksidp5s4b46czn86sw",
       1,
       1,
       nt.NewRaw(0).Mul(nt.NewRaw(53), nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(29), nil)),
 []*nt.Raw{nt.NewRaw(0).Mul(nt.NewRaw(3473),   nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(28), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(59),     nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(29), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(32),     nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(29), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(1018),   nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(27), nil))},
       1,
       []*nt.Raw{},
      },
      {"nano_1c73mmx64sxpudp1d46w56ct8kynnzt5bdufocfkspn8beknbb3mngj3a6br",
       // This address is blacklisted from address 1,0. That's why it won't take from there
       "nano_3gickb6kgex966fs9666jghehh7bwrpcmqmdbyqa1441i83dwufrr9uojn81",
       1,
       2,
       nt.NewRaw(0).Mul(nt.NewRaw(6), nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil)),
 []*nt.Raw{nt.NewRaw(0).Mul(nt.NewRaw(3473),   nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(28), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(93),     nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(28), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(92),     nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(29), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(0),      nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(28), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(0),      nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(28), nil))},
       3,
       []*nt.Raw{nt.NewRaw(0).Mul(nt.NewRaw(1018), nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(27), nil)),
                 nt.NewRaw(0).Mul(nt.NewRaw(4970), nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(27), nil))},
      },
   }

   for _, test := range test1 {
      resetInUse()
      // Add address to client list
      clientPub, err := keyMan.AddressToPubKey(test.clientAddress)
      if (err != nil) {
         t.Errorf("Error during execution: %s", err.Error())
      }
      setClientAddress(test.seedId, test.index, clientPub)

      testingPayment = append(make([]*nt.Raw, 0), test.nanoReceived)
      testingPayment = append(testingPayment, test.intermediaryTx...)
      testingPaymentIndex = 0
      testingPendingHashesNum = test.numOfPendingTxs
      err = receivedNano(test.nanoAddress)
      if (err != nil) {
         t.Errorf("Error during execution: %s", err.Error())
      }

      // Wait for transaction to complete
      wg.Wait()

      // Now check that the database is as we expect
      conn, err := pgx.Connect(context.Background(), databaseUrl)
      if (err != nil) {
         t.Errorf("Error during execution: %s", err.Error())
      }
      queryString :=
      "SELECT " +
         "balance, " +
         "parent_seed, " +
         "index " +
      "FROM " +
         "wallets " +
      "ORDER BY " +
         "index;"

      rows, err := conn.Query(context.Background(), queryString)

      var balance = nt.NewRaw(0)
      var seedId int
      var index int
      for i := 0; rows.Next(); i++ {
         err = rows.Scan(balance, &seedId, &index)
         if (err != nil) {
            t.Errorf("Error during execution: %s", err.Error())
         }
         if (i >= len(test.balances)) {
            t.Errorf("Too many wallets in database")
            return
         }

         if (balance.Cmp(test.balances[i]) != 0) {
            t.Errorf("Wrong balance at %d,%d\r\n want: %d\r\n got:  %d", seedId, index, test.balances[i], balance.Int)
         }
      }
   }

}

func Test_RawToNano(t *testing.T) {

   test1 := []struct {
      input *nt.Raw
      output float64
   }{
      {nt.NewRaw(0).Mul(nt.NewRaw(41), nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil)),
       41},
      {nt.NewRaw(0).Mul(nt.NewRaw(917), nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(28), nil)),
       9.17},
      {nt.NewRaw(0).Mul(nt.NewRaw(148), nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(27), nil)),
       0.148},
      {nt.NewRaw(0).Mul(nt.NewRaw(314), nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(28), nil)),
       3.14},
      {nt.NewRaw(0).Mul(nt.NewRaw(4857), nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(29), nil)),
       485.7},
   }

   for _, test := range test1 {

      check := rawToNANO(test.input)
      if (check != test.output) {
         t.Errorf("Conversion incorrect, want: %f, got, %f", test.output, check)
      }
   }
}

// Blacklist in test database:
// nano_1ts8ejswbndstgp6r4wgi7yr593rg7ryab4wuzburmay3pxbrgu3i5f1fz3n
// nano_3gickb6kgex966fs9666jghehh7bwrpcmqmdbyqa1441i83dwufrr9uojn81
//
// nano_1ie4own1s5qmmyd33u9a64169ox54kdb3khs1yt84gfgd7n7dshgcjkegxei
// nano_1x6r8jefor8z4sswemkqqcnognf3m7wu7u53jbe5auqpg8rrakbjbrit3xg9
//
// nano_1c73mmx64sxpudp1d46w56ct8kynnzt5bdufocfkspn8beknbb3mngj3a6br
// nano_1i7jmxg4t4jdqha8rnmoep6gst36d57r7c97p3amazyjdb45wkizp161fs6f
//
// nano_3f4pznen4utfxmeu7jmucnhg6ut4rd9fk87s7xnnrkr4okph65158j4xciqf
// nano_3gq8fhso1ukbegs63tu5nk7p8mum7z58qgi9iibkh8tc3jqptt8ds7k9jj9h

func Test_checkBlackList(t *testing.T) {
   databaseUrl = "postgres://test:testing@localhost:5432/gotests"
   databasePassword = "testing"

   // Reset database to known state
   script := exec.Command("psql", "-f", resetScript, "-U", "test", "-d", "gotests")
   script.Run()
   inTesting = true

   // Check to make sure they're there
   test1 := []struct {
      seedId int
      index int
      clientAddress string
   }{
      {1,
       0,
      "nano_3gickb6kgex966fs9666jghehh7bwrpcmqmdbyqa1441i83dwufrr9uojn81"},
      {1,
       1,
      "nano_1x6r8jefor8z4sswemkqqcnognf3m7wu7u53jbe5auqpg8rrakbjbrit3xg9"},
      {1,
       2,
      "nano_1i7jmxg4t4jdqha8rnmoep6gst36d57r7c97p3amazyjdb45wkizp161fs6f"},
      {1,
       3,
      "nano_3gq8fhso1ukbegs63tu5nk7p8mum7z58qgi9iibkh8tc3jqptt8ds7k9jj9h"},
   }

   // Check for false positives
   test2 := []struct {
      seedId int
      index int
      clientAddress string
   }{
      {1,
       0,
      "nano_35iadfbqk7r7purmds56fmqkeffo6se4zpncyy6ftun7wr7p9rg4r8k9wr44"},
      {1,
       1,
      "nano_3m91gkpno5s46gquu6upxrknk3zamrwiibd35ryioz6zwuq5kyx4m36t35u6"},
      {1,
       2,
      "nano_1rscbkzpacdz49ujso1eh7xm5hkui8ee1hgienhk8z64r9byzdeqwko3tio1"},
      {1,
       3,
      "nano_3jcq5bk5i3idpyw9hozrnuksy96jn9jyy3f8auhdr7ozbudc43rrnrdcy5er"},
   }

   for _, test := range test1 {

      clientPub, err := keyMan.AddressToPubKey(test.clientAddress)
      if (err != nil) {
         t.Errorf("Error during execution: %s", err.Error())
      }
      check, _, err := checkBlackList(test.seedId, test.index, clientPub)
      if (err != nil) {
         t.Errorf("Error during execution: %s", err.Error())
      }
      if !(check) {
         t.Errorf("Expected blacklist, but could not find. %d,%d", test.seedId, test.index)
      }
   }

   for _, test := range test2 {

      clientPub, err := keyMan.AddressToPubKey(test.clientAddress)
      if (err != nil) {
         t.Errorf("Error during execution: %s", err.Error())
      }
      check, _, err := checkBlackList(test.seedId, test.index, clientPub)
      if (err != nil) {
         t.Errorf("Error during execution: %s", err.Error())
      }
      if (check) {
         t.Errorf("No blacklist exists but function returned true. %d,%d", test.seedId, test.index)
      }
   }

}

