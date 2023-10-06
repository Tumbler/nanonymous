package main

import (
   "fmt"
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
   const P_HASH = "270fabb2f590b0baf1824383b2ecfd4452e2ebcfee9a60744862bfe1620fac941abe7aa9522cc0f91c8e375097c27953de2dbdf571a15bb1a0f13b5ce23b3393"

   inTesting = true
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
       "b5ee1b6e99f56bb7fc8149d8bb6fe3831756ab654e20067a491b58c29113f0fb",
       "nano_3usi45tj6oybeapseej9sgjzdyewz4gybrtpa88a8ttynifpc1w3nakxr8k5" },
      {"nano_1b9opg96jquha58fueg8g8zjofauw5ua3pqfer6bdnjjj34x51s6j8dq83hh",
       "bdd31576235494cf8283b9104cbb83c5b501c78f8c92574d075025efb6345752",
       "nano_3147zdiazd9kf6rnd4mdon19jemxbd89csatu8sxgnrtrdfuuxodm8hst4hw" },
      {"nano_34jz5qi36gkemhncpu3hbnzfjg1pam4b49fhg8oo9g96c795q9dz3e3n19bb",
       "2519a0e35da4f996f4aa0caa7c669d5990bc32efe0973c7b5a664430a341130d",
       "nano_3pa8fsb7ttmu4yw3jee69omcjhyampbz6xsnpxcxymrtfey8za5hhy1mmos1" },
   }

   for _, test := range test1 {
      key, _, err := getNewAddress(test.inputAddress, false, false, false, 0)
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

   balance, managed, mixer, err := findTotalBalance()
   if (err != nil) {
      t.Errorf("Error during execution: %s", err.Error())
   }

   if (balance.Cmp(nt.NewRawFromNano("64.8")) != 0) {
      t.Errorf("Bad balance, want: %.1f, got %.1f", 64.8, rawToNANO(balance))
   }
   if (managed.Cmp(nt.NewRawFromNano("44.8")) != 0) {
      t.Errorf("Bad balance, want: %.1f, got %.1f", 44.8, rawToNANO(managed))
   }
   if (mixer.Cmp(nt.NewRawFromNano("20")) != 0) {
      t.Errorf("Bad balance, want: %.1f, got %.1f", 20.0, rawToNANO(mixer))
   }
}

// TODO add test for recive_only
func Test_receivedNano(t *testing.T) {
   databaseUrl = "postgres://test:testing@localhost:5432/gotests"
   databasePassword = "testing"

   // Reset database to known state
   script := exec.Command("psql", "-f", resetScript, "-U", "test", "-d", "gotests")
   script.Run()
   inTesting = true

   test1 := []struct {
      nanoAddress string
      recipientAddress string
      seedId int
      index int
      nanoReceived *nt.Raw
      balances []*nt.Raw
      numOfPendingTxs []int
   }{
      {"nano_3f4pznen4utfxmeu7jmucnhg6ut4rd9fk87s7xnnrkr4okph65158j4xciqf",
       "nano_1bgho34hpofn4sxencbr8916sbbyyoosr5mmepewyguo8te15qkq8hefnrdn",
       1,
       3,
       nt.NewRaw(0).Mul(nt.NewRaw(10), nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil)),
 []*nt.Raw{nt.NewRaw(0).Mul(nt.NewRaw(3102), nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(28), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(6),    nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(29), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(32),   nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(29), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(10),   nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(5),    nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(5),    nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(10),   nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil))},
       []int{1},
      },
      {"nano_1ts8ejswbndstgp6r4wgi7yr593rg7ryab4wuzburmay3pxbrgu3i5f1fz3n",
       "nano_14gfu8wkz48o3xf869ehp7rd9oah1993d1deguqknkksidp5s4b46czn86sw",
       1,
       0,
       nt.NewRaw(0).Mul(nt.NewRaw(9), nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil)),
 []*nt.Raw{nt.NewRaw(0).Mul(nt.NewRaw(4002), nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(28), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(6),    nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(29), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(32),   nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(29), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(1018), nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(27), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(5),    nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(5),    nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(10),   nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil))},
       []int{1},
      },
      {"nano_1ie4own1s5qmmyd33u9a64169ox54kdb3khs1yt84gfgd7n7dshgcjkegxei",
       "nano_14gfu8wkz48o3xf869ehp7rd9oah1993d1deguqknkksidp5s4b46czn86sw",
       1,
       1,
       nt.NewRaw(0).Mul(nt.NewRaw(53), nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(29), nil)),
 []*nt.Raw{nt.NewRaw(0).Mul(nt.NewRaw(3473),   nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(28), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(59),     nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(29), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(32),     nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(29), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(1018),   nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(27), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(5),      nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(5),      nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(10),     nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil))},
       []int{1},
      },
      {"nano_1c73mmx64sxpudp1d46w56ct8kynnzt5bdufocfkspn8beknbb3mngj3a6br",
       // This address is blacklisted from address 1,0. That's why it won't take from there
       "nano_3gickb6kgex966fs9666jghehh7bwrpcmqmdbyqa1441i83dwufrr9uojn81",
       1,
       2,
       nt.NewRaw(0).Mul(nt.NewRaw(6), nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil)),
 []*nt.Raw{nt.NewRaw(0).Mul(nt.NewRaw(3473),   nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(28), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(0),     nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(28), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(92),     nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(29), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(0),      nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(28), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(5),      nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(5),      nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(10),     nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(0),      nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(28), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(0),      nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(28), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(0),      nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(28), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(21483),  nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(25), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(9207),   nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(25), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(224316), nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(24), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(398784), nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(24), nil))},
       []int{3,7},
      },
      {"nano_3f4pznen4utfxmeu7jmucnhg6ut4rd9fk87s7xnnrkr4okph65158j4xciqf",
       "nano_3gxo1dh5x6bai7dngpiy5sngnehx1qodr4acw8s1xowag6im7dba1iyswk58",
       1,
       3,
       nt.NewRaw(0).Mul(nt.NewRaw(52), nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil)),
 []*nt.Raw{nt.NewRaw(0).Mul(nt.NewRaw(0),      nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(28), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(0),      nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(28), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(0),      nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(29), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(52),     nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(0),      nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(28), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(0),      nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(28), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(10),     nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(0),      nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(0),      nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(0),      nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(0),      nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(0),      nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(0),      nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(0),      nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(0),      nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(0),      nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(0),      nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(17766216), nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(23), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(6242184),  nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(23), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(3885804),  nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(23), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(1745796),  nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(23), nil))},
       []int{9,7},
      },
   }

   for i, test := range test1 {
      resetInUse()
      // Add address to recipient list
      recipientPub, err := keyMan.AddressToPubKey(test.recipientAddress)
      if (err != nil) {
         t.Errorf("Error during execution: %s", err.Error())
      }
      setRecipientAddress(test.seedId, test.index, recipientPub, false)

      if (i == 5) {
         // Kinda weird to hardcode this, but whatever...
         // Adding a receive only address to test.
         getNewAddress("", true, false, false, 0)
      }

      testingPayment = append(make([]*nt.Raw, 0), test.nanoReceived)
      testingPaymentExternal = true
      testingReceiveAlls = 0
      testingSends = make(map[string][]*nt.Raw)
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

func Test_receiveOnly(t *testing.T) {
   databaseUrl = "postgres://test:testing@localhost:5432/gotests"
   databasePassword = "testing"

   // Reset database to known state
   script := exec.Command("psql", "-f", resetScript, "-U", "test", "-d", "gotests")
   script.Run()
   inTesting = true

   test1 := []struct {
      account string
      amount *nt.Raw
      balances []*nt.Raw
   }{
      {"nano_3usi45tj6oybeapseej9sgjzdyewz4gybrtpa88a8ttynifpc1w3nakxr8k5",
       nt.NewRaw(0).Mul(nt.NewRaw(3), nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil)),
 []*nt.Raw{nt.NewRaw(0).Mul(nt.NewRaw(41), nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(6),    nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(29), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(32),   nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(29), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(0) ,   nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(5),    nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(5),    nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(10),   nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(3),    nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil))},
      },
   }

   // Get new receive_only address
   getNewAddress("", true, false, false, 0)

   for _, test := range test1 {
      resetInUse()

      // Build the ConfirmationBlock
      cBlock := ConfirmationBlock{
         Message: struct{
               Block keyMan.Block
               Account string
               Amount *nt.Raw
               Hash nt.BlockHash
               ConfirmationType string `json:"confirmation_type"`
         }{
            Amount: test.amount,
            Block: keyMan.Block{
               Subtype: "send",
               LinkAsAccount: test.account,
            },
         },
      }

      testingPayment = append(make([]*nt.Raw, 0), test.amount)
      testingPaymentExternal = true
      handleNotification(cBlock)

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

   verbosity = 0
}

func Test_extractFromMixer(t *testing.T) {
   databaseUrl = "postgres://test:testing@localhost:5432/gotests"
   databasePassword = "testing"

   // Reset database to known state
   script := exec.Command("psql", "-f", resetScript, "-U", "test", "-d", "gotests")
   script.Run()
   inTesting = true

   test1 := []struct {
      recipientAddress string
      seedId int
      index int
      nanoToSend *nt.Raw
      balances []*nt.Raw
      numOfPendingTxs []int
   }{
      {"nano_1bgho34hpofn4sxencbr8916sbbyyoosr5mmepewyguo8te15qkq8hefnrdn",
       1,
       3,
       nt.NewRaw(0).Mul(nt.NewRaw(64),    nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(29), nil)),
 []*nt.Raw{nt.NewRaw(0).Mul(nt.NewRaw(41), nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(6),  nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(29), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(32), nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(29), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(0),  nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(0),  nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(0),  nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(10), nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(0),  nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(0),  nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(0),  nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(30), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(115632), nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(25), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(42768),  nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(25), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(197568), nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(25), nil)),
           nt.NewRaw(0).Mul(nt.NewRaw(4032),   nt.NewRaw(0).Exp(nt.NewRaw(10), nt.NewRaw(25), nil))},
       []int{2,6},
      },
   }

   for _, test := range test1 {
      resetInUse()

      recipientPub, err := keyMan.AddressToPubKey(test.recipientAddress)
      if (err != nil) {
         t.Errorf("Error during execution: %s", err.Error())
      }
      setRecipientAddress(test.seedId, test.index, recipientPub, false)

      testingPaymentExternal = false
      testingReceiveAlls = 0
      testingPendingHashesNum = test.numOfPendingTxs

      pubKey, err := keyMan.AddressToPubKey("nano_1tipnanogsu7q59pnie3qfc4w378wm43fg4ksqc8wmnnfnizrq1xrpt5geho")
      if (err != nil) {
         t.Errorf("Error during execution: %s", err.Error())
      }

      extractFromMixer(test.nanoToSend, pubKey)
      if (err != nil) {
         t.Errorf("Error during execution: %s", err.Error())
      }

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

func Test_seedRetire(t *testing.T){
   databaseUrl = "postgres://test:testing@localhost:5432/gotests"
   databasePassword = "testing"

   // Reset database to known state
   script := exec.Command("psql", "-f", resetScript, "-U", "test", "-d", "gotests")
   script.Run()
   inTesting = true

   test1 := []struct {
      finalSeed int
      numOfPendingTxs []int
   }{
      {
         3,
         []int{100},
      },
   }

   for _, test := range test1 {
      testingReceiveAlls = 0
      testingPendingHashesNum = test.numOfPendingTxs

      beforeBalance, err := balanceInSeed(1)
      if (err != nil) {
         t.Errorf("Error during execution: %s", err.Error())
      }

      retireCurrentSeed()
      getNewAddress("", false, false, false, 0)
      retireCurrentSeed()

      afterBalance, err := balanceInSeed(3)
      if (err != nil) {
         t.Errorf("Error during execution: %s", err.Error())
      }

      if (beforeBalance.Cmp(afterBalance) != 0) {
         if (err != nil) {
            t.Errorf("Error: Balances didn't match.\nBefore: %f\nAfter: %f", rawToNANO(beforeBalance), rawToNANO(afterBalance))
         }
      } else {
            fmt.Println("Not error: Balances didn't match.\nBefore:", rawToNANO(beforeBalance), "\nAfter: ",  rawToNANO(afterBalance))
      }

      rows, conn, err := getSeedRowsFromDatabase()
      if (err != nil) {
         t.Errorf("Error during execution: %s", err.Error())
      }
      defer rows.Close()
      defer conn.Close(context.Background())

      var finalSeed int
      rows.Next()
      err = rows.Scan(nil, nil, &finalSeed)
      if (err != nil) {
         t.Errorf("Error during execution: %s", err.Error())
      }

      if (finalSeed != test.finalSeed) {
         t.Errorf("Error: final seed does not match.\nwant: %d\nfound: %d", test.finalSeed, finalSeed)
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
      recipientAddress string
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
      recipientAddress string
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

      recipientPub, err := keyMan.AddressToPubKey(test.recipientAddress)
      if (err != nil) {
         t.Errorf("Error during execution: %s", err.Error())
      }
      check, _, err := checkBlackList(test.seedId, test.index, recipientPub)
      if (err != nil) {
         t.Errorf("Error during execution: %s", err.Error())
      }
      if !(check) {
         t.Errorf("Expected blacklist, but could not find. %d,%d", test.seedId, test.index)
      }
   }

   for _, test := range test2 {

      recipientPub, err := keyMan.AddressToPubKey(test.recipientAddress)
      if (err != nil) {
         t.Errorf("Error during execution: %s", err.Error())
      }
      check, _, err := checkBlackList(test.seedId, test.index, recipientPub)
      if (err != nil) {
         t.Errorf("Error during execution: %s", err.Error())
      }
      if (check) {
         t.Errorf("No blacklist exists but function returned true. %d,%d", test.seedId, test.index)
      }
   }

}

