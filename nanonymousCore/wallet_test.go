package main

import (
   "testing"
   "strings"
   "encoding/hex"
   "errors"
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
         generatedSeed := strings.ToUpper(hex.EncodeToString(key.seed))

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

   if        (!key.initialized) {
      t.Errorf("Key not initialized")
   } else if (key.keyType != 0) {
      t.Errorf("Keytype not expected type, want: 0, got: %d", key.keyType)
   } else if (len(key.seed) != 32) {
      t.Errorf("Malformed seed, len: %d", len(key.seed))
   } else if (!(len(key.mnemonic) > 0)) {
      t.Errorf("Empty mnemonic")
   } else if (len(key.privateKey) != 32) {
      t.Errorf("Malformed privateKey, len: %d", len(key.privateKey))
   } else if (len(key.publicKey) != 32) {
      t.Errorf("Malformed mnemonic, len: %d", len(key.publicKey))
   } else if (len(key.nanoAddress) != 65) {
      t.Errorf("Malformed nano Address, len: %d", len(key.nanoAddress))
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
      key := Key{seed: seedBytes}
      err := SeedToKeys(&key)

      if (err != nil) {
         t.Errorf("Error in execution: %s", err.Error())
      } else {
         var privateKeyString = strings.ToUpper(hex.EncodeToString(key.privateKey))
         if (test.privateKey != privateKeyString) {
            t.Errorf("Invalid private key\r\n want: %s\r\n got:  %s", test.privateKey, privateKeyString)
         }
         var publicKeyString = strings.ToUpper(hex.EncodeToString(key.publicKey))
         if (test.publicKey != publicKeyString) {
            t.Errorf("Invalid public key\r\n want: %s\r\n got:  %s", test.publicKey, publicKeyString)
         }
         if (test.nanoAddress != key.nanoAddress) {
            t.Errorf("Invalid nano address\r\n want: %s\r\n got:  %s", test.nanoAddress, key.nanoAddress)
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

// TODO test checksums
//func TestChecksumSha(t *testing.T) {
   //test1 := []struct {
      //{}
   //}
//}

//func TestNanoAddressChecksum(t *testing.T) {
//}

// TODO get new input from keys tools
func TestNanoED25519PublicKey(t *testing.T) {
   test1 := []struct {
      privateKey string
      publicKey string
   }{
      {"2835308381D823CE083C4DF12140C8A8855EF642BDCA4ED36EF7FC438A3728D8",
       "2B2876C9DC03B09D8B5BDFABF2560CAB6565BB6C20E20D05169EB635659D2F80"},
      {"F89BF8E93811CC70A52219F0CCE058622975F8AE3371D3BAB9DD63B7B5A266D6",
       "7AFB3AF46F7E6F32EE05F0FC81FF0F584AED4A02FE29412B0ACEBDC12F13A346"},
      {"F2EE208CB377C38544F07FFEF2D3667B1D23CA7836AAC1522763791646ADC221",
       "F4DC80C806BCC9FE641F65199F4FC226906250F218461C51AD59EAA317CA48B5"},
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
