package nanoKeyManager

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
   _"embed"
)

const NANO_ADDRESS_ENCODING = "13456789abcdefghijkmnopqrstuwxyz"
const BYTES_IN_KEY = 32
const NUMBER_OF_MNEMONICS = 2048
const MNEMNOIC_WORDS = 24
const BITS_IN_ONE_WORD = 11

type Key struct {
   Initialized bool
   KeyType     int    // 0 - Full seed; 1 - private key; 2 - public key
   Seed        []byte
   Index       int
   Mnemonic    string
   PrivateKey  []byte
   PublicKey   []byte
   NanoAddress string
}

var ChecksumMismatch = errors.New("checksum mismatch")

var activeSeed Key
var verbose bool

//go:embed bip39-English.txt
var bipWordFile string

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

      WalletVerbose(true)
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
         if (activeSeed.Initialized) {
            if (activeSeed.KeyType < 2) {
               activeSeed.Index++
               err := SeedToKeys(&activeSeed)
               if (err != nil) {
                  fmt.Println(err.Error())
                  break
               }

               fmt.Print("Index ", activeSeed.Index, ":\n")
               fmt.Print("Private key:  0x", strings.ToUpper(hex.EncodeToString(activeSeed.PrivateKey[:])), "\n")
               fmt.Print("Public  key:  0x", strings.ToUpper(hex.EncodeToString(activeSeed.PublicKey[:])), "\n")
               fmt.Print("Nano Address: ", activeSeed.NanoAddress, "\n")
            } else {
               fmt.Println("ERROR: Key doesn't support this operatoin!")
            }
         } else {
            fmt.Println("ERROR: No active seed!")
         }
      case "4":
         if (activeSeed.Initialized) {
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
      WalletVerbose(false)
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

   if (newKey.Initialized) {
      err = errors.New("Key already initialized. Delete key first!")
      return fmt.Errorf("GenerateSeed: %w", err)
   } else {
      // Make sure we start with a clean slate
      ReinitSeed(newKey)
   }

   for i := 0; i < len(seed); i++ {
      seed[i] = entropy[i]
   }
   newKey.Seed = append(newKey.Seed, seed...)

   newKey.Mnemonic, err = SeedToMnemonic(seed)
   if (err != nil) {
      return fmt.Errorf("GenerateSeed: %w", err)
   }

   saveBehaviour := verbose
   verbose = false
   err = SeedToKeys(newKey);
   verbose = saveBehaviour

   if (err != nil) {
      return fmt.Errorf("GenerateSeed: %w", err)
   }

   if (verbose) {
      fmt.Println("mnemonic is:\"", newKey.Mnemonic, "\"")
      fmt.Print("Seed: 0x", strings.ToUpper(hex.EncodeToString(newKey.Seed)), "\n")
      fmt.Print("Index 0\n")
      fmt.Print("Private key:  0x", strings.ToUpper(hex.EncodeToString(newKey.PrivateKey[:])), "\n")
      fmt.Print("Public  key:  0x", strings.ToUpper(hex.EncodeToString(newKey.PublicKey[:])), "\n")
      fmt.Print("Nano Address: ", newKey.NanoAddress, "\n")
   }

   newKey.Initialized = true
   return nil
}

// SeedToMnemonic takes a nano seed and returns the corresponding BIP39
// compliant mnemonic.
func SeedToMnemonic(seed []byte) (string, error) {
   var wordlist [NUMBER_OF_MNEMONICS]string
   var mnemonic = make([]string, 0)
   var err error

   // Read wordlist from file
   for i, word := range strings.Split(bipWordFile, "\n") {
      // Splitting on newline gives us an extra blank entry at the end.
      if (word == "") {
         continue
      }
      if (i >= NUMBER_OF_MNEMONICS) {
         return "", fmt.Errorf("SeedToMnemonic: index out of bounds: %d", i)
      }
      wordlist[i] = strings.Trim(word, "\r\n")
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
   var err error

   if (newKey.Initialized) {
      err = errors.New("cannot overwrite active seed; delete current active seed first")
      return fmt.Errorf("GenerateSeedFromMnemonic: %w", err);
   }

   if (len(strings.Split(mnemonic, " ")) != MNEMNOIC_WORDS) {
      err = errors.New(fmt.Sprint("Invalid Mnemonic,", MNEMNOIC_WORDS, " entries required!"))
      return fmt.Errorf("GenerateSeedFromMnemonic: %w", err);
   }

   mnemonic = strings.Trim(mnemonic, "\n\r")
   mnemonicArray := strings.Split(mnemonic, " ")
   var wordlist = make(map[string]int)

   if (bipWordFile == "") {
      return fmt.Errorf("GenerateSeedFromMnemonic: file not found");
   }

   // Read wordlist from file
   for i, word := range strings.Split(bipWordFile, "\n") {
      // Splitting on newline gives us an extra blank entry at the end.
      if (word == "") {
         continue
      }
      wordlist[strings.Trim(word, "\r\n")] = i
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
   newKey.Seed = seed
   newKey.Mnemonic = mnemonic
   err = SeedToKeys(newKey)
   if (err != nil) {
      return fmt.Errorf("GenerateSeedFromMnemonic: %w", err)
   }

   if (verbose) {
      fmt.Println("mnemonic is:\"", newKey.Mnemonic, "\"")
      fmt.Print("Seed: 0x", strings.ToUpper(hex.EncodeToString(newKey.Seed)), "\n")
      fmt.Print("Index 0\n")
      fmt.Print("Private key:  0x", strings.ToUpper(hex.EncodeToString(newKey.PrivateKey[:])), "\n")
      fmt.Print("Public  key:  0x", strings.ToUpper(hex.EncodeToString(newKey.PublicKey[:])), "\n")
      fmt.Print("Nano Address: ", newKey.NanoAddress, "\n")
   }

   newKey.Initialized = true
   return nil
}

// SeedToKeys takes a prepopulated seed from a key struct and gernerates the
// corresponding private key, public key, and public nano address.
func SeedToKeys(seed *Key) error {

   var index = seed.Index

   // Temporary spot to store the seed + index
   var seedIndex []byte

   // Append 32 bit integer form of index to the seed
   seedIndex = append(seed.Seed, (byte)((index & 0xFF000000) >> 24))
   seedIndex = append(seedIndex, (byte)((index & 0x00FF0000) >> 16))
   seedIndex = append(seedIndex, (byte)((index & 0x0000FF00) >> 8))
   seedIndex = append(seedIndex, (byte)(index & 0x000000FF))

   // blake2b hash the seed + index
   var address = blake2b.Sum256(seedIndex)
   seed.PrivateKey = make([]byte, len(address))
   copy(seed.PrivateKey, address[:])

   var err error
   seed.PublicKey, err = NanoED25519PublicKey(seed.PrivateKey)
   if (err != nil) {
      return fmt.Errorf("SeedToKeys: %w", err)
   }

   // Make a copy that we can manuipulate
   var pubCopy = make([]byte, len(seed.PublicKey))
   copy(pubCopy, seed.PublicKey)

   checksum, err := NanoAddressChecksum(pubCopy)
   pubCopy = append([]byte{0, 0, 0}, pubCopy...)
   b32 := base32.NewEncoding(NANO_ADDRESS_ENCODING)

   if (err != nil) {
      return fmt.Errorf("SeedToKeys: %w", err)
   }

   seed.NanoAddress = "nano_" + b32.EncodeToString(pubCopy)[4:] + b32.EncodeToString(checksum)

   seed.Initialized = true

   if (verbose) {
      fmt.Println("mnemonic is:\"", seed.Mnemonic, "\"")
      fmt.Print("Seed: 0x", strings.ToUpper(hex.EncodeToString(seed.Seed)), "\n")
      fmt.Print("Index: ", seed.Index, "\n")
      fmt.Print("Private key:  0x", strings.ToUpper(hex.EncodeToString(seed.PrivateKey[:])), "\n")
      fmt.Print("Public  key:  0x", strings.ToUpper(hex.EncodeToString(seed.PublicKey[:])), "\n")
      fmt.Print("Nano Address: ", seed.NanoAddress, "\n")
   }

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

func WalletVerbose(setting bool) {
   verbose = setting
}

// ReinitSeed returns the given Key to all its default values
func ReinitSeed(activeSeed *Key) {
   activeSeed.Initialized = false
   activeSeed.KeyType = 0
   activeSeed.Seed = make([]byte, 0)
   activeSeed.Index = 0
   activeSeed.Mnemonic = ""
   activeSeed.PrivateKey = make([]byte, 0)
   activeSeed.PublicKey = make([]byte, 0)
   activeSeed.NanoAddress = ""
}
