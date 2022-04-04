package main

import (
   "fmt"
   "bufio"
   "os"
   "strings"
   "encoding/hex"
   "filippo.io/edwards25519"
   "encoding/base32"
   "errors"
   "golang.org/x/crypto/blake2b"
   "crypto/sha256"
   "crypto/rand"
)

// ED25519
const NANO_ADDRESS_ENCODING = "13456789abcdefghijkmnopqrstuwxyz"

type key struct {
   initialized bool
   keyType     int    // 0 - Full seed; 1 - private key; 2 - public key
   seed        []byte
   index       int
   mnemonic    string
   privateKey  []byte
   publicKey   []byte
   nanoAddress string
}

var activeSeed key

func main() {

   var usr string



   menu:
   for {
      fmt.Print("1. Generate Seed\n",
                "2. Input Mnomonic\n",
                "3. Get next address\n",
                "4. Delete stored seed\n")
      fmt.Scan(&usr)

      switch (usr) {
      case "1":
         GenerateSeed(&activeSeed)
      case "2":
         inputReader := bufio.NewReader(os.Stdin)
         fmt.Print("Mnemonic: ")
         input1, _ := inputReader.ReadString('\n')
         input1 += input1
         input, _ := inputReader.ReadString('\n')
         GenerateSeedFromMnemonic(input, &activeSeed)
      case "3":
         if (activeSeed.initialized) {
            if (activeSeed.keyType < 2) {
               activeSeed.index++
               SeedToKeys(&activeSeed)

               fmt.Print("Index ", activeSeed.index, ":\n")
               fmt.Print("Private key:  0x", strings.ToUpper(hex.EncodeToString(activeSeed.privateKey[:])), "\n")
               fmt.Print("Public  key:  0x", strings.ToUpper(hex.EncodeToString(activeSeed.publicKey[:])), "\n")
               fmt.Print("Nano Address: ", activeSeed.nanoAddress, "\n")
            } else {
               fmt.Println("ERROR: Key doesn't support this operatoin!")
            }
         } else {
            fmt.Println("ERROR: No active seed!")
         }
      case "4":
         if (activeSeed.initialized) {
            fmt.Println("Are you sure??? y/n")
            fmt.Scan(&usr)

            if (usr == "y") {

               activeSeed.initialized = false
               activeSeed.keyType = 0
               activeSeed.seed = make([]byte, 0)
               activeSeed.index = 0
               activeSeed.mnemonic = ""
               activeSeed.privateKey = make([]byte, 0)
               activeSeed.publicKey = make([]byte, 0)
               activeSeed.nanoAddress = ""
            }
         } else {
            fmt.Println("ERROR: No active seed!")
         }
      default:
         break menu
      }
   }

   fmt.Println("Peace!")
}

// GenerateSeed gets entropy and generates a new mnemonic/seed pair along with
// their public keys. Takes a key struct to fill in.
func GenerateSeed(newKey *key) {
   var entropy, entropySquirt = GetBipEntropy()
   var seed = make([]byte, 32)

   if (newKey.initialized) {
      fmt.Println("Key already initialized. Delete key first!")
      return
   }

   newKey.initialized = true

   for i := 0; i < len(seed); i++ {
      seed[i] = entropy[i]
   }
   newKey.seed = append(newKey.seed, seed...)

   newKey.mnemonic = SeedToMnemonic(entropySquirt)

   SeedToKeys(newKey);


   fmt.Println("mnemonic is:\"", newKey.mnemonic, "\"")
   fmt.Print("Seed: 0x", strings.ToUpper(hex.EncodeToString(newKey.seed)), "\n")
   fmt.Print("Index 0\n")
   fmt.Print("Private key:  0x", strings.ToUpper(hex.EncodeToString(newKey.privateKey[:])), "\n")
   fmt.Print("Public  key:  0x", strings.ToUpper(hex.EncodeToString(newKey.publicKey[:])), "\n")
   fmt.Print("Nano Address: ", newKey.nanoAddress, "\n")
}

// SeedToMnemonic takes a nano seed and returns the corresponding BIP39
// compliant mnemonic.
func SeedToMnemonic(seed bitSquirt) string {
   var file, err = os.Open("bip39-English.txt")
   var wordlist [2048]string
   var mnemonic = make([]string, 0)

   if (err != nil) {
      fmt.Println("bip39-English.txt not found")
      return ""
   }

   defer file.Close();
   var scanner = bufio.NewScanner(file)

   // Read wordlist from file called "bip39-English.txt"
   for i := 0; i < 2048 && scanner.Scan(); i++ {
      wordlist[i] = scanner.Text()
   }

   // TODO caculate checksum instead of letting entropy function do it.
   // get the mnemoic
   for i := 0; i < 24; i++ {
      index := seed.squirtBits(11)
      mnemonic = append(mnemonic, wordlist[index])
   }

   return strings.Join(mnemonic, " ")
}

// GenerateSeedFromMnemonic converts an exisiting BIP39 mnemonic to a nano seed.
func GenerateSeedFromMnemonic(mnemonic string, newKey *key) {
   var seed []byte
   var squirt bitSquirt

   if (newKey.initialized) {
      fmt.Println("Error! Delete current active seed first")
      return
   }

   if (len(strings.Split(mnemonic, " ")) != 24) {
      fmt.Println("Invalid Mnemonic, 24 entries required!")
      return
   } else {
      mnemonic = strings.Trim(mnemonic, "\n\r")
      mnemonicArray := strings.Split(mnemonic, " ")
      var file, err = os.Open("bip39-English.txt")
      var wordlist = make(map[string]int)

      if (err != nil) {
         fmt.Println("bip39-English.txt not found")
         return
      }

      defer file.Close();
      var scanner = bufio.NewScanner(file)

      // Read wordlist from file called "bip39-English.txt"
      for i := 0; i < 2048 && scanner.Scan(); i++ {
         wordlist[scanner.Text()] = i
      }

      for i := 0; i < 24; i++ {
         squirt.slurpBits(int64(wordlist[mnemonicArray[i]]), 11)
      }
      squirt.restartBitSquirt()

      for i :=0; i < squirt.getBitSquirtLength(); i++ {
         seed = append(seed, (byte)(squirt.squirtBits(8)))
      }

      // Check checksum
      var checksum byte
      if (len(seed) == 33) {
         // Remove checksum from seed imputed from mnemonic
         checksum = seed[len(seed)-1]
         seed = seed[:len(seed)-1]

         // Recalculate checksum
         hash := sha256.New()
         hash.Write(seed)
         calculatedChecksum := hash.Sum(nil)

         if (checksum != calculatedChecksum[0]) {
            fmt.Println("ERROR! Checksum mismatch!")
         } else {
            // Everything checks out, proceed to save key
            newKey.seed = seed
            newKey.mnemonic = mnemonic
            newKey.initialized = true
            SeedToKeys(newKey)

            fmt.Println("mnemonic is:\"", newKey.mnemonic, "\"")
            fmt.Print("Seed: 0x", strings.ToUpper(hex.EncodeToString(newKey.seed)), "\n")
            fmt.Print("Index 0\n")
            fmt.Print("Private key:  0x", strings.ToUpper(hex.EncodeToString(newKey.privateKey[:])), "\n")
            fmt.Print("Public  key:  0x", strings.ToUpper(hex.EncodeToString(newKey.publicKey[:])), "\n")
            fmt.Print("Nano Address: ", newKey.nanoAddress, "\n")
         }
      }
   }
}

// SeedToKeys takes a prepopulated seed from a key struct and gernerates the
// corresponding private key, public key, and public nano address.
func SeedToKeys(seed *key) {

   var index = seed.index

   // Temporary spot to store the seed + index
   var seedIndex []byte

   // Append 32 bit integer form of index to the seed
   seedIndex = append(seed.seed, (byte)((index & 0xFF000000) >> 24))
   seedIndex = append(seedIndex, (byte)((index & 0x00FF0000) >> 16))
   seedIndex = append(seedIndex, (byte)((index & 0x0000FF00) >> 8))
   seedIndex = append(seedIndex, (byte)(index & 0x000000FF))

   // blake2b hash the seed + index
   var address = blake2b.Sum256(seedIndex)
   seed.privateKey = make([]byte, 0)
   seed.privateKey = append(seed.privateKey, address[:]...)

   h := blake2b.Sum512(address[:])
   s, _ := edwards25519.NewScalar().SetBytesWithClamping(h[:32])
   A := (&edwards25519.Point{}).ScalarBaseMult(s)
   // TODO Error check variable "_"

   seed.publicKey = A.Bytes()
   var pubCopy = make([]byte, len(seed.publicKey))
   _ = copy(pubCopy, seed.publicKey)

   checksum := Checksum(pubCopy)
   pubCopy = append([]byte{0, 0, 0}, pubCopy...)
   b32 := base32.NewEncoding(NANO_ADDRESS_ENCODING)

   seed.nanoAddress = "nano_" + b32.EncodeToString(pubCopy)[4:] + b32.EncodeToString(checksum)

   return
}


func Checksum(pubkey []byte) (checksum []byte) {
   hash, err := blake2b.New(5, nil)
   if err != nil {
      return
   }
   hash.Write(pubkey)
   for _, b := range hash.Sum(nil) {
      checksum = append([]byte{b}, checksum...)
   }
   return
}

// AddressToPubKey takes a nano address and converts it to the corresponding
// public key.
func AddressToPubKey(nanoAddress string) (pubKey []byte, err error) {
   var address string
   err = errors.New("invalid address")

   if (len(nanoAddress) == 64) {
      if (nanoAddress[:4] != "xrb_") {
         return
      } else {
         address = nanoAddress[4:]
      }
   } else if (len(nanoAddress) == 65) {
      if (nanoAddress[:5] != "nano_" ) {
         return
      } else {
         address = nanoAddress[5:]
      }
   } else {
      return
   }

   b32 := base32.NewEncoding(NANO_ADDRESS_ENCODING)

   pubKey, err = b32.DecodeString("1111" + address[:52])
   if (err != nil) {
      return
   }

   pubKey = pubKey[3:]

   checksum := Checksum(pubKey)

   if (b32.EncodeToString(checksum) != address[52:]) {
      err = errors.New("checksum mismatch")
   }

   return
}

//TODO look into import crypto/rand for entropy generation
// Generate array of 11 bit ints to be used for bip39 word generation
// TODO Don't generate checksum in entropy
func GetBipEntropy() ([]byte, bitSquirt) {

   // Seeds are 32 bytes long
   var entropy = make([]byte, 32)
   rand.Read(entropy)

   hash := sha256.New()
   hash.Write(entropy)
   checksum := hash.Sum(nil)

   entropy = append(entropy, checksum[0])

   //structContainer := newBitSquirt(entropy)
   //storedData = *structContainer

   var squirt bitSquirt

   for _, bite := range entropy {
      squirt.slurpBits(int64(bite), 8)
   }
   squirt.restartBitSquirt()

   return entropy, squirt
}

func ReturnActiveSeed() *key {
   return &activeSeed
}
