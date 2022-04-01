package main

import (
   "fmt"
   "bufio"
   "os"
   "strings"
   "golang.org/x/crypto/pbkdf2"
   "golang.org/x/crypto/blake2b"
   "crypto/sha512"
   "encoding/hex"
   "filippo.io/edwards25519"
   "encoding/base32"
   "errors"
)

// ED25519
const PublicKeyCompressedLength = 33
const NANO_ADDRESS_ENCODING = "13456789abcdefghijkmnopqrstuwxyz"

func main () {

   var usr string

   menu:
   for {
      fmt.Print("1. Generate Seed\n",
                "2. Input Mnomonic\n",
                "3. Get next address\n")
      fmt.Scan(&usr)

      switch (usr) {
      case "1":
         GenerateSeed()
      case "2":
         inputReader := bufio.NewReader(os.Stdin)
         fmt.Print("Mnemonic: ")
         input1, _ := inputReader.ReadString('\n')
         input1 += input1
         input, _ := inputReader.ReadString('\n')
         GenerateSeedFromMnemonic(input)
      case "3":
         fmt.Println("3")
      default:
         fmt.Println("uh?")
         break menu
      }
   }

   fmt.Println("We out!")
}


// Get entropy and generate a new mnemonic/seed pair
func GenerateSeed() []byte {
   var entropy = GetBipEntropy()
   var seed = make([]byte, 32)

   for i := 0; i < len(seed); i++ {
      seed[i] = entropy[i]
   }

   var mnemonic = SeedToMnemonic(entropy)

   var index = 41
   var address, addressPublic, nanoAddress = SeedToPublicAddress(seed, index);

   fmt.Println("mnemonic is:\"", mnemonic, "\"")
   fmt.Print("Seed: 0x", strings.ToUpper(hex.EncodeToString(seed)), "\n")
   fmt.Print("Index ", index, ":\n")
   fmt.Print("Private key:  0x", strings.ToUpper(address), "\n")
   fmt.Print("Public  key:  0x", strings.ToUpper(addressPublic), "\n")
   fmt.Print("Nano Address: ", nanoAddress, "\n")

   return seed
}

func SeedToMnemonic(seed []byte) string {
   var file, err = os.Open("bip39-English.txt")
   var wordlist [2048]string
   var mnemonic = make([]string, 0)
   //var bigInt int64

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

   // get the mnemoic
   for i := 0; i < 24; i++ {
      index := squirtBits(11)
      mnemonic = append(mnemonic, wordlist[index])
   }

   return strings.Join(mnemonic, " ")
}

// Convert an exisiting BIP39 mnemonic to a seed
func GenerateSeedFromMnemonic(mnemonic string) []byte {
   var seed []byte

   if (len(strings.Split(mnemonic, " ")) != 24){
      fmt.Println("Invalid Mnemonic, 24 entries required!")
      return nil
   } else {
      seed = pbkdf2.Key([]byte(mnemonic), []byte("mnemonic"+"p"), 2048, 32, sha512.New)

      fmt.Print("Seed is: 0x", strings.ToUpper(hex.EncodeToString(seed)), "\n")

      return seed
   }
}

func SeedToPublicAddress(seed []byte, index int) (string, string, string) {

   // Append 32 bit integer form of index to the seed
   seed = append(seed, (byte)((index & 0xFF000000) >> 24))
   seed = append(seed, (byte)((index & 0x00FF0000) >> 16))
   seed = append(seed, (byte)((index & 0x0000FF00) >> 8))
   seed = append(seed, (byte)(index & 0x000000FF))

   // blake2b hash the seed + index
   var address = blake2b.Sum256(seed)

   h := blake2b.Sum512(address[:])
   s, _ := edwards25519.NewScalar().SetBytesWithClamping(h[:32])
   A := (&edwards25519.Point{}).ScalarBaseMult(s)
   // TODO Error check _

   publicKey := A.Bytes()
   var pubCopy = make([]byte, len(publicKey))
   _ = copy(pubCopy, publicKey)

   checksum := checksum(pubCopy)
   pubCopy = append([]byte{0, 0, 0}, pubCopy...)
   b32 := base32.NewEncoding(NANO_ADDRESS_ENCODING)

   nanoAddress := "nano_" + b32.EncodeToString(pubCopy)[4:] + b32.EncodeToString(checksum)

   return hex.EncodeToString(address[:]), hex.EncodeToString(publicKey[:]), nanoAddress
}

func checksum(pubkey []byte) (checksum []byte) {
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

   checksum := checksum(pubKey)

   fmt.Println("checksum:", b32.EncodeToString(checksum))
   fmt.Println("us      :", address[52:])
   if (b32.EncodeToString(checksum) != address[52:]) {
      err = errors.New("checksum mismatch")
   }

   return
}
