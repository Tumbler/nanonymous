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

const NANO_ADDRESS_ENCODING = "13456789abcdefghijkmnopqrstuwxyz"
const BYTES_IN_KEY = 32
const NUMBER_OF_MNEMONICS = 2048
const MNEMNOIC_WORDS = 24
const BITS_IN_ONE_WORD = 11

type Key struct {
   initialized bool
   keyType     int    // 0 - Full seed; 1 - private key; 2 - public key
   seed        []byte
   index       int
   mnemonic    string
   privateKey  []byte
   publicKey   []byte
   nanoAddress string
}

var activeSeed Key

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
      case "5":
         var blarg bitBucket
         blarg.slurpBits(0b101001, 6)
         blarg.slurpBits(0b101001, 6)
         blarg.slurpBits(0b101001, 6)
         blarg.slurpBits(0b10, 2)

         bits, num := blarg.squirtBits(23)
         fmt.Println("number is:", bits, "read ", num, "bits")
      default:
         break menu
      }
   }

   fmt.Println("Peace!")
}

// GenerateSeed gets entropy and generates a new mnemonic/seed pair along with
// their public keys. Takes a key struct to fill in.
func GenerateSeed(newKey *Key) {
   var entropy = GetBipEntropy()
   var seed = make([]byte, BYTES_IN_KEY)

   if (newKey.initialized) {
      fmt.Println("Key already initialized. Delete key first!")
      return
   }

   newKey.initialized = true

   for i := 0; i < len(seed); i++ {
      seed[i] = entropy[i]
   }
   newKey.seed = append(newKey.seed, seed...)

   newKey.mnemonic = SeedToMnemonic(seed)

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
func SeedToMnemonic(seed []byte) string {
   var file, err = os.Open("bip39-English.txt")
   var wordlist [NUMBER_OF_MNEMONICS]string
   var mnemonic = make([]string, 0)

   if (err != nil) {
      fmt.Println("bip39-English.txt not found")
      return ""
   }

   defer file.Close();
   var scanner = bufio.NewScanner(file)

   // Read wordlist from file called "bip39-English.txt"
   for i := 0; i < NUMBER_OF_MNEMONICS && scanner.Scan(); i++ {
      wordlist[i] = scanner.Text()
   }

   // Caclulate checksum and append it onto seed
   seed = append(seed, ChecksumSha(seed))

   var bucket bitBucket
   for _, bite := range seed {
      bucket.slurpBits(int64(bite), 8)
   }

   // Transcribe to mnemonic
   for {
      index, numRead := bucket.squirtBits(BITS_IN_ONE_WORD)
      if (numRead == BITS_IN_ONE_WORD) {
         mnemonic = append(mnemonic, wordlist[index])
      } else if (numRead > 0) {
         mnemonic = append(mnemonic, wordlist[index])
         break
      } else {
         break
      }
   }

   return strings.Join(mnemonic, " ")
}

// GenerateSeedFromMnemonic converts an exisiting BIP39 mnemonic to a nano seed.
func GenerateSeedFromMnemonic(mnemonic string, newKey *Key) {
   var seed []byte
   var bucket bitBucket

   if (newKey.initialized) {
      fmt.Println("Error! Delete current active seed first")
      return
   }

   if (len(strings.Split(mnemonic, " ")) != MNEMNOIC_WORDS) {
      fmt.Println("Invalid Mnemonic,", MNEMNOIC_WORDS, "entries required!")
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
      for i := 0; i < NUMBER_OF_MNEMONICS && scanner.Scan(); i++ {
         wordlist[scanner.Text()] = i
      }

      for i := 0; i < MNEMNOIC_WORDS; i++ {
         bucket.slurpBits(int64(wordlist[mnemonicArray[i]]), BITS_IN_ONE_WORD)
      }

      for {
         bits, numRead := bucket.squirtBits(8)
         if (numRead == 8) {
            seed = append(seed, byte(bits))
         } else if (numRead > 0) {
            seed = append(seed, byte(bits))
            break
         } else {
            break
         }
      }

      // Check checksum
      var checksum byte
      if (len(seed) == BYTES_IN_KEY + 1) {
         // Remove checksum from seed obtained from mnemonic
         checksum = seed[len(seed)-1]
         seed = seed[:len(seed)-1]

         // Recalculate checksum
         calculatedChecksum := ChecksumSha(seed)

         if (checksum != calculatedChecksum) {
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
func SeedToKeys(seed *Key) {

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

   seed.publicKey = NanoED25519PublicKey(address[:])

   var pubCopy = make([]byte, len(seed.publicKey))
   _ = copy(pubCopy, seed.publicKey)

   checksum := NanoAddressChecksum(pubCopy)
   pubCopy = append([]byte{0, 0, 0}, pubCopy...)
   b32 := base32.NewEncoding(NANO_ADDRESS_ENCODING)

   seed.nanoAddress = "nano_" + b32.EncodeToString(pubCopy)[4:] + b32.EncodeToString(checksum)

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

   checksum := NanoAddressChecksum(pubKey)

   if (b32.EncodeToString(checksum) != address[52:]) {
      err = errors.New("checksum mismatch")
   }

   return
}

// Generate array of 11 bit ints to be used for bip39 word generation
func GetBipEntropy() []byte {
   // Seeds are 32 bytes long
   var entropy = make([]byte, BYTES_IN_KEY)
   rand.Read(entropy)

   return entropy
}

// ChecksumSha uses the sha256 algorithm to hash the data and returns one byte
// to use as a checksum
func ChecksumSha(data []byte) byte{
   hash := sha256.New()
   hash.Write(data)
   return hash.Sum(nil)[0]
}

// NanoAddressChecksum finds the checksum for the public nano address from a
// public key. It derives 5 bytes using the blake2b algorithm and reverses the
// order.
func NanoAddressChecksum(pubkey []byte) (checksum []byte) {
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

// NanoED25519PublicKey uses the ED25519 curve to derive public key. Nano uses
// blake2b instead of sha.
func NanoED25519PublicKey(privateKey []byte) []byte {
   h := blake2b.Sum512(privateKey)
   s, _ := edwards25519.NewScalar().SetBytesWithClamping(h[:BYTES_IN_KEY])
   A := (&edwards25519.Point{}).ScalarBaseMult(s)
   // TODO Error check variable "_"

   return A.Bytes()
}

func ReturnActiveSeed() *Key {
   return &activeSeed
}
