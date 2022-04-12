package nanoKeyManager

import (
   "testing"
   "strings"
   "encoding/hex"
   "errors"
   "reflect"
)


func TestGenerateSeedFromMnemonic(t *testing.T) {
   test1 := []struct {
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
   test2 := []struct {
      mnemonic string
   }{
      {"jeans ecology family expand prosper sauce above quality hire hole already joke lady guard such must debris club meadow vendor express agree still vacuut"},
      {"light rare feature yard matrix surprise cluster stable payment fatal wealth glimpse obey behind erase room laugh burger original next south tide tree obscure"},
      {"mail hobby prefer must between hip basket victory embrace enter museum clinic stick vault depart reform early hamster entire fault patrol silver brass quality"},
   }

   // Test input/output
   for _, test := range test1 {
      var key Key
      err := GenerateSeedFromMnemonic(test.mnemonic, &key)
      if (err != nil) {
         t.Errorf("Error in execution: %s", err.Error())
      } else {
         generatedSeed := strings.ToUpper(hex.EncodeToString(key.Seed))

         if (generatedSeed != test.seed) {
            t.Errorf("Generation from \"%s\" was incorrect, \n\rgot:  %s, \n\rwant: %s", test.mnemonic, generatedSeed, test.seed)
         }
      }
   }

   // Test checksum detection
   for _, test := range test2 {
      var key Key
      // Inputs have malformed mnemonics. Should fail here.
      err := GenerateSeedFromMnemonic(test.mnemonic, &key)
      if (!errors.Is(err, ChecksumMismatch)) {
         t.Errorf("Bad checksum detection failed; expected checksum error on \"%s\"", test.mnemonic)
      }
   }
}

func TestGenerateSeed(t *testing.T) {
   var key Key
   err := GenerateSeed(&key)

   if (err != nil) {
      t.Errorf("Unexpected Error: %s", err.Error())
      return
   }

   if        (!key.Initialized) {
      t.Errorf("Key not initialized")
   } else if (key.KeyType != 0) {
      t.Errorf("Keytype not expected type, want: 0, got: %d", key.KeyType)
   } else if (len(key.Seed) != 32) {
      t.Errorf("Malformed seed, len: %d", len(key.Seed))
   } else if (!(len(key.Mnemonic) > 0)) {
      t.Errorf("Empty mnemonic")
   } else if (len(key.PrivateKey) != 32) {
      t.Errorf("Malformed privateKey, len: %d", len(key.PrivateKey))
   } else if (len(key.PublicKey) != 32) {
      t.Errorf("Malformed mnemonic, len: %d", len(key.PublicKey))
   } else if (len(key.NanoAddress) != 65) {
      t.Errorf("Malformed nano Address, len: %d", len(key.NanoAddress))
   }
}

func TestSeedToMnemonic(t *testing.T) {
   test1 := []struct {
      seed string
      mnemonic string
   }{
      {"77A8C14A28142B7F0025796C0D941CBC27C8CEB61C9038658626F9150C0A3577",
       "jeans ecology family expand dress sauce above quality hire hole already joke lady guard such must debris club meadow vendor express agree still vacuum"},
      {"1BC399A38CB80E6F87EAE8B100CE5BF0602AADE0635A66BD434B09127EDF6CCD",
       "bridge broom happy bonus liberty daughter cabbage fringe rain all notice this actor finger light hero grit tube spot banana child hurry rebel romance"},
      {"7D9FE5DFE54B66C717606E4F2F6D7C19411EB893892824C9FED0658E86003F9C",
       "laugh you jewel skin reopen glory frown assist execute kiwi gallery crater ball fox evolve celery cave exit reduce slam trick abandon wrist man"},
   }

   // Test input/output
   for _, test := range test1 {
      seedBytes, _ := hex.DecodeString(test.seed)
      generatedMnemonic, err := SeedToMnemonic(seedBytes)
      if (err != nil) {
         t.Errorf("Error in execution: %s", err.Error())
      } else {
         if (generatedMnemonic != test.mnemonic) {
            t.Errorf("Generation from \"%s\" was incorrect, \n\rgot:  %s, \n\rwant: %s", test.seed, generatedMnemonic, test.mnemonic)
         }
      }
   }
}

func TestSeedToKeys(t *testing.T) {
   test1 := []struct {
      seed string
      privateKey string
      publicKey string
      nanoAddress string
   }{
      {"BF28C0C840027166F3B9FD9372DAD1BA78AFB86029DADD88A76022A368641A6A",
       "2835308381D823CE083C4DF12140C8A8855EF642BDCA4ED36EF7FC438A3728D8",
       "2B2876C9DC03B09D8B5BDFABF2560CAB6565BB6C20E20D05169EB635659D2F80",
       "nano_1csagu6xr1ximp7oqqxdybd1scu7epxpra943n4jf9op8okstdw1rquwusrc"},
      {"11DC328F5DAE2F49E3AA833FA29F217C58C5E8B5F77C14E46FB8630741602DCE",
       "F89BF8E93811CC70A52219F0CCE058622975F8AE3371D3BAB9DD63B7B5A266D6",
       "7AFB3AF46F7E6F32EE05F0FC81FF0F584AED4A02FE29412B0ACEBDC12F13A346",
       "nano_1yqu9dt8yzmh8dq1dw9wi9ziyp4cxo717zjba6oiomoxr6qj9at8sze7zzt3"},
      {"CCE5C6FC0B720516B6CBA0D9976D4AA73BC76B2DF8309A92A2D3A86CA975AD49",
       "F2EE208CB377C38544F07FFEF2D3667B1D23CA7836AAC1522763791646ADC221",
       "F4DC80C806BCC9FE641F65199F4FC226906250F218461C51AD59EAA317CA48B5",
       "nano_3x8wi561fh8bzsk3ysasmx9w6bniebah68485jattphcnedwnk7or7py9iho"},
   }


   for _, test := range test1 {
      seedBytes, _ := hex.DecodeString(test.seed)
      key := Key{Seed: seedBytes}
      err := SeedToKeys(&key)

      if (err != nil) {
         t.Errorf("Error in execution: %s", err.Error())
      } else {
         var privateKeyString = strings.ToUpper(hex.EncodeToString(key.PrivateKey))
         if (test.privateKey != privateKeyString) {
            t.Errorf("Invalid private key\r\n want: %s\r\n got:  %s", test.privateKey, privateKeyString)
         }
         var publicKeyString = strings.ToUpper(hex.EncodeToString(key.PublicKey))
         if (test.publicKey != publicKeyString) {
            t.Errorf("Invalid public key\r\n want: %s\r\n got:  %s", test.publicKey, publicKeyString)
         }
         if (test.nanoAddress != key.NanoAddress) {
            t.Errorf("Invalid nano address\r\n want: %s\r\n got:  %s", test.nanoAddress, key.NanoAddress)
         }
      }
   }
}

// TODO get new input from keys tools
func TestAddressToPubKey(t *testing.T) {
   test1 := []struct {
      nanoAddress string
      pubKey string
   }{
      {"nano_1csagu6xr1ximp7oqqxdybd1scu7epxpra943n4jf9op8okstdw1rquwusrc",
       "2B2876C9DC03B09D8B5BDFABF2560CAB6565BB6C20E20D05169EB635659D2F80"},
      {"nano_1yqu9dt8yzmh8dq1dw9wi9ziyp4cxo717zjba6oiomoxr6qj9at8sze7zzt3",
       "7AFB3AF46F7E6F32EE05F0FC81FF0F584AED4A02FE29412B0ACEBDC12F13A346"},
      {"nano_3x8wi561fh8bzsk3ysasmx9w6bniebah68485jattphcnedwnk7or7py9iho",
       "F4DC80C806BCC9FE641F65199F4FC226906250F218461C51AD59EAA317CA48B5"},
   }
   test2 := []struct {
      nanoAddress string
   }{
      {"nano_33abgezpdmxy9ey1twkcifooc4ow4nznqsq9h4zoqhy4n91hfrtidt857ray"},
      {"nano_3u1q4k31trmfkq3his63nnrzhs8k4bx1yheb3miu9ftrwm6815ker9td8sd8"},
      {"nano_3fzt5u4opy84jzuhusgx6yos817r3jjhdcyfwrgwyp1x91y7d9ocjdb1syxx"},
   }

   for _, test := range test1 {
      pubkey, err := AddressToPubKey(test.nanoAddress)

      if (err != nil) {
         t.Errorf("Error in execution: %s", err.Error())
      } else {
         if (test.pubKey != strings.ToUpper(hex.EncodeToString(pubkey))) {
            t.Errorf("Invalid public key\r\n want: %s\r\n got:  %s", test.pubKey, strings.ToUpper(hex.EncodeToString(pubkey)))
         }
      }
   }

   // Test checksum detection
   for _, test := range test2 {
      // Inputs have malformed addresses. Should fail here.
      _, err := AddressToPubKey(test.nanoAddress)

      if !(errors.Is(err, ChecksumMismatch)) {
         t.Errorf("Bad checksum detection failed; expected checksum error on \"%s\"", test.nanoAddress)
      }
   }
}

func TestChecksumSha(t *testing.T) {
   test1 := []struct {
      input string
      output string
   }{
      {"CE104A38563B568251CA3AA38FC7ED10FFB6E32BF79F28E3D0517C5912427C3B",
       "B0"},
      {"C6013A11B1144820974611BEB7200E9790ADB587D8A909B9AB5CDDB4E6642396",
       "BA"},
      {"82EBADD6C7FB2EA14EBE",
       "02"},
   }

   for _, test := range test1 {
      inputbytes, _ := hex.DecodeString(test.input)
      bite := ChecksumSha(inputbytes)

      expectedBite, _ := hex.DecodeString(test.output)
      if (bite != expectedBite[0]) {
         t.Errorf("Bad SHA checksum\r\n want: %x\r\n got:  %x", expectedBite, bite)
      }
   }
}

func TestNanoAddressChecksum(t *testing.T) {
   test1 := []struct {
      publicKey string
      checksum []byte
   }{
      {"D40AD7C3617FBBA92BE6B528DD5D47527502E8DAB9774BA7C5B1C96D46BF06B7",
      []byte{0xA1, 0xE0, 0xF3, 0xBA, 0xC1}},
      {"021D3EB06E65538F324E4985A15A16E511ACD7BAA90226DFFBAE2CC55B5D97DC",
      []byte{0x57, 0xDD, 0xB5, 0x79, 0xEF}},
      {"5DE44E0C67F7208C8DB7F37EE439D8BA5F8A6DC9B17180B4EE6AB18999FE281A",
      []byte{0x22, 0xEF, 0xDE, 0x20, 0x16}},
   }

   for _, test := range test1 {
      publicKeyByte, _ := hex.DecodeString(test.publicKey)
      newChecksum, err := NanoAddressChecksum(publicKeyByte)

      if (err != nil) {
         t.Errorf("Error in execution: %s", err.Error())
      } else {
         if (!reflect.DeepEqual(newChecksum, test.checksum)) {
            t.Errorf("Bad BLAKE checksum\r\n want: %x\r\n got:  %x", test.checksum, newChecksum)
         }
      }
   }

}

func TestNanoED25519PublicKey(t *testing.T) {
   test1 := []struct {
      privateKey string
      publicKey string
   }{
      {"7FB019CF97ABF40A9CAE666C75FEBDC63E07451FE0F7A7D4BF881E1A244D673D",
       "27BB04B055B73D8C3D1C2746C9B2D5983E68DAEB3BDAAC3604FA9FD1791A8D60"},
      {"97E8CC6BBB2E4637291F5B7EA34F58367403F577526BDE4C9866114BFEAB100F",
       "26B537E1BF707E2CF4969583F100EEB6F778FF2C795DDC9CCAD3C0EEEE85AC4C"},
      {"3DCA4EBC021E655EDD6A28BD61E0F7568A793B7612E13CF3FC8ADAA62D8C66CB",
       "6C94AA7D9B36EA6885F7BCD3E6B11F7E66A0E8D7ECC50A6ABB7869589AE19031"},
    }

   for _, test := range test1 {
      privateKeyBytes, _ := hex.DecodeString(test.privateKey)
      generatedpublicKey, err := NanoED25519PublicKey(privateKeyBytes)

      generatedpublicKeyString := strings.ToUpper(hex.EncodeToString(generatedpublicKey))

      if (err != nil) {
         t.Errorf("Error in execution: %s", err.Error())
      } else {
         if (test.publicKey != generatedpublicKeyString) {
            t.Errorf("Public key incorrect:\r\n want: %s \r\n got:  %s", test.publicKey, generatedpublicKeyString)
         }
      }
   }
}

func TestBitBucket(t *testing.T) {
   test1 := []int64 {
       3242339803082304,
       4750080550102598,
       8362602201806792,
       41,
       9223372036854775807,
   }

   for _, test := range test1 {
      var bucket bitBucket
      err := bucket.slurpBits(test, 64)

      if (err != nil) {
         t.Errorf("Error in execution: %s", err.Error())
         return
      }

      output, _ := bucket.squirtBits(64)

      if (output != test) {
         t.Errorf("BitBucket produced wrong int\r\n want: %d\r\n got:  %d", test, output)
      }
   }
}
