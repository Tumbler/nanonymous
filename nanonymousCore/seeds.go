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

var ChecksumMismatch = errors.New("checksum mismatch")

var activeSeed Key
var verbose bool

func main() {

   var usr string

   menu:
   for {
      fmt.Print("1. Generate Seed\n",
                "2. Input Mnomonic\n",
                "3. Get next address\n",
                "4. Delete stored seed\n",
                "5. Test stuff\n",
                "6. Nano address to pubkey\n")
      fmt.Scan(&usr)

      switch (usr) {
      case "1":
         err := GenerateSeed(&activeSeed)
         if (err != nil){
            fmt.Println(err.Error())
         }
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
               err := SeedToKeys(&activeSeed)
               if (err != nil) {
                  fmt.Println(err.Error())
                  break
               }

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
               ReinitSeed(&activeSeed)
            }
         } else {
            fmt.Println("ERROR: No active seed!")
         }
      case "5":
         var blarg bitBucket
         blarg.slurpBits(0b101001, 6)
         blarg.slurpBits(0b101001, 6)
         blarg.slurpBits(0b101001, 6)
         err := blarg.slurpBits(0b10, 2)
         if (err != nil) {
            fmt.Println(err.Error())
         }

         bits, num := blarg.squirtBits(23)
         fmt.Println("number is:", bits, "read ", num, "bits")
      case "6":
         inputReader := bufio.NewReader(os.Stdin)
         fmt.Print("nano Address: ")
         input1, _ := inputReader.ReadString('\n')
         input1 += input1
         input, _ := inputReader.ReadString('\n')
         var blarg, _ = AddressToPubKey(input)
         fmt.Print("Public  key:  0x", strings.ToUpper(hex.EncodeToString(blarg)), "\n")
      default:
         break menu
      }
   }

   fmt.Println("Peace!")
}

// GenerateSeed gets entropy and generates a new mnemonic/seed pair along with
// their public keys. Takes a key struct to fill in.
func GenerateSeed(newKey *Key) error {
   var entropy, err = GetBipEntropy()
   var seed = make([]byte, BYTES_IN_KEY)

   if (err != nil) {
      return fmt.Errorf("GenerateSeed: %w", err)
   }

   if (newKey.initialized) {
      err = errors.New("Key already initialized. Delete key first!")
      return fmt.Errorf("GenerateSeed: %w", err)
   } else {
      // Make sure we start with a clean slate
      ReinitSeed(newKey)
   }

   for i := 0; i < len(seed); i++ {
      seed[i] = entropy[i]
   }
   newKey.seed = append(newKey.seed, seed...)

   newKey.mnemonic, err = SeedToMnemonic(seed)
   if (err != nil) {
      return fmt.Errorf("GenerateSeed: %w", err)
   }

   err = SeedToKeys(newKey);
   if (err != nil) {
      return fmt.Errorf("GenerateSeed: %w", err)
   }

   if (verbose) {
      fmt.Println("mnemonic is:\"", newKey.mnemonic, "\"")
      fmt.Print("Seed: 0x", strings.ToUpper(hex.EncodeToString(newKey.seed)), "\n")
      fmt.Print("Index 0\n")
      fmt.Print("Private key:  0x", strings.ToUpper(hex.EncodeToString(newKey.privateKey[:])), "\n")
      fmt.Print("Public  key:  0x", strings.ToUpper(hex.EncodeToString(newKey.publicKey[:])), "\n")
      fmt.Print("Nano Address: ", newKey.nanoAddress, "\n")
   }

   newKey.initialized = true
   return nil
}

// SeedToMnemonic takes a nano seed and returns the corresponding BIP39
// compliant mnemonic.
func SeedToMnemonic(seed []byte) (string, error) {
   var file, err = os.Open("bip39-English.txt")
   var wordlist [NUMBER_OF_MNEMONICS]string
   var mnemonic = make([]string, 0)

   if (err != nil) {
      return "", fmt.Errorf("SeedToMnemonic: %w", err)
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
      err = bucket.slurpBits(int64(bite), 8)
   }
   if (err != nil) {
      return "", fmt.Errorf("SeedToMnemonic: %w", err)
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

   return strings.Join(mnemonic, " "), nil
}

// GenerateSeedFromMnemonic converts an exisiting BIP39 mnemonic to a nano seed.
func GenerateSeedFromMnemonic(mnemonic string, newKey *Key) error {
   var seed []byte
   var bucket bitBucket

   if (newKey.initialized) {
      err := errors.New("cannot overwrite active seed; delete current active seed first")
      return fmt.Errorf("GenerateSeedFromMnemonic: %w", err);
   }

   if (len(strings.Split(mnemonic, " ")) != MNEMNOIC_WORDS) {
      err := errors.New(fmt.Sprint("Invalid Mnemonic,", MNEMNOIC_WORDS, " entries required!"))
      return fmt.Errorf("GenerateSeedFromMnemonic: %w", err);
   }

   mnemonic = strings.Trim(mnemonic, "\n\r")
   mnemonicArray := strings.Split(mnemonic, " ")
   var file, err = os.Open("bip39-English.txt")
   var wordlist = make(map[string]int)

   if (err != nil) {
      return fmt.Errorf("GenerateSeedFromMnemonic: %w", err);
   }

   defer file.Close();
   var scanner = bufio.NewScanner(file)

   // Read wordlist from file
   for i := 0; i < NUMBER_OF_MNEMONICS && scanner.Scan(); i++ {
      wordlist[scanner.Text()] = i
   }

   // Convert mnemonic into bit string
   for i := 0; i < MNEMNOIC_WORDS; i++ {
      err = bucket.slurpBits(int64(wordlist[mnemonicArray[i]]), BITS_IN_ONE_WORD)
   }
   if (err != nil) {
      return fmt.Errorf("GenerateSeedFromMnemonic: %w", err)
   }

   // Read bit string into seed
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
   if (len(seed) != BYTES_IN_KEY + 1) {
      err = errors.New(fmt.Sprint("invalid seed length: ", len(seed)))
      return fmt.Errorf("GenerateSeedFromMnemonic: %w", err);
   }

   // Remove checksum from seed obtained from mnemonic
   checksum = seed[len(seed)-1]
   seed = seed[:len(seed)-1]

   // Recalculate checksum
   calculatedChecksum := ChecksumSha(seed)

   if (checksum != calculatedChecksum) {
      return fmt.Errorf("GenerateSeedFromMnemonic: %w", ChecksumMismatch);
   }

   // Everything checks out, proceed to save key
   newKey.seed = seed
   newKey.mnemonic = mnemonic
   err = SeedToKeys(newKey)
   if (err != nil) {
      return fmt.Errorf("GenerateSeedFromMnemonic: %w", err)
   }

   if (verbose) {
      fmt.Println("mnemonic is:\"", newKey.mnemonic, "\"")
      fmt.Print("Seed: 0x", strings.ToUpper(hex.EncodeToString(newKey.seed)), "\n")
      fmt.Print("Index 0\n")
      fmt.Print("Private key:  0x", strings.ToUpper(hex.EncodeToString(newKey.privateKey[:])), "\n")
      fmt.Print("Public  key:  0x", strings.ToUpper(hex.EncodeToString(newKey.publicKey[:])), "\n")
      fmt.Print("Nano Address: ", newKey.nanoAddress, "\n")
   }

   newKey.initialized = true
   return nil
}

// SeedToKeys takes a prepopulated seed from a key struct and gernerates the
// corresponding private key, public key, and public nano address.
func SeedToKeys(seed *Key) error {

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
   seed.privateKey = make([]byte, len(address))
   copy(seed.privateKey, address[:])

   var err error
   seed.publicKey, err = NanoED25519PublicKey(seed.privateKey)
   if (err != nil) {
      return fmt.Errorf("SeedToKeys: %w", err)
   }

   // Make a copy that we can manuipulate
   var pubCopy = make([]byte, len(seed.publicKey))
   copy(pubCopy, seed.publicKey)

   checksum, err := NanoAddressChecksum(pubCopy)
   pubCopy = append([]byte{0, 0, 0}, pubCopy...)
   b32 := base32.NewEncoding(NANO_ADDRESS_ENCODING)

   if (err != nil) {
      return fmt.Errorf("SeedToKeys: %w", err)
   }

   seed.nanoAddress = "nano_" + b32.EncodeToString(pubCopy)[4:] + b32.EncodeToString(checksum)

   return nil
}


// AddressToPubKey takes a nano address and converts it to the corresponding
// public key.
func AddressToPubKey(nanoAddress string) ([]byte, error) {
   var address string
   err := errors.New("invalid address")

   nanoAddress = strings.Trim(nanoAddress, "\n\r")
   if (len(nanoAddress) == 64) {
      if (nanoAddress[:4] != "xrb_") {
         return nil, fmt.Errorf("AddressToPubKey: %w", err)
      } else {
         address = nanoAddress[4:]
      }
   } else if (len(nanoAddress) == 65) {
      if (nanoAddress[:5] != "nano_" ) {
         return nil, fmt.Errorf("AddressToPubKey: %w", err)
      } else {
         address = nanoAddress[5:]
      }
   } else {
      return nil, fmt.Errorf("AddressToPubKey: %w", err)
   }

   err = errors.New("encode issue")
   b32 := base32.NewEncoding(NANO_ADDRESS_ENCODING)

   pubKey, err := b32.DecodeString("1111" + address[:52])
   if (err != nil) {
      return nil, fmt.Errorf("AddressToPubKey: %w", err)
   }

   pubKey = pubKey[3:]

   checksum, err := NanoAddressChecksum(pubKey)
   if (err != nil) {
      return nil, fmt.Errorf("AddressToPubKey: %w", err)
   }

   if (b32.EncodeToString(checksum) != address[52:]) {
      return nil, fmt.Errorf("AddressToPubKey: %w", ChecksumMismatch)
   }

   return pubKey, nil
}

// Generate array of 11 bit ints to be used for bip39 word generation
func GetBipEntropy() ([]byte, error) {
   // Seeds are 32 bytes long
   var entropy = make([]byte, BYTES_IN_KEY)
   _, err := rand.Read(entropy)

   if (err != nil) {
      return nil, fmt.Errorf("GetBipEntropy: %w", err)
   }

   return entropy, nil
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
func NanoAddressChecksum(pubkey []byte) ([]byte, error) {
   hash, err := blake2b.New(5, nil)
   var checksum []byte
   if err != nil {
      return nil, fmt.Errorf("NanoAddressChecksum: %w", err)
   }
   hash.Write(pubkey)
   for _, b := range hash.Sum(nil) {
      checksum = append([]byte{b}, checksum...)
   }
   return checksum, nil
}

// NanoED25519PublicKey uses the ED25519 curve to derive public key. Nano uses
// blake2b instead of sha.
func NanoED25519PublicKey(privateKey []byte) ([]byte, error){
   h := blake2b.Sum512(privateKey)
   s, err := edwards25519.NewScalar().SetBytesWithClamping(h[:BYTES_IN_KEY])
   if (err != nil) {
      return nil, fmt.Errorf("NanoED25519PublicKey: %w", err)
   }
   A := (&edwards25519.Point{}).ScalarBaseMult(s)

   return A.Bytes(), nil
}

func ReturnActiveSeed() *Key {
   return &activeSeed
}

func WalletVerbose(setting bool) {
   verbose = setting
}

// ReinitSeed returns the given Key to all its default values
func ReinitSeed(activeSeed *Key) {
   activeSeed.initialized = false
   activeSeed.keyType = 0
   activeSeed.seed = make([]byte, 0)
   activeSeed.index = 0
   activeSeed.mnemonic = ""
   activeSeed.privateKey = make([]byte, 0)
   activeSeed.publicKey = make([]byte, 0)
   activeSeed.nanoAddress = ""
}

