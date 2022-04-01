package main

import (
   "crypto/rand"
   "crypto/sha256"
   "math"
)

//TODO look into import crypto/rand for entropy generation
// Generate array of 11 bit ints to be used for bip39 word generation
// TODO Don't generate checksum in entropy
func GetBipEntropy() []byte {

   // Seeds are 32 bytes long
   var entropy = make([]byte, 32)
   rand.Read(entropy)

   hash := sha256.New()
   hash.Write(entropy)
   checksum := hash.Sum(nil)

   entropy = append(entropy, checksum[0])

   structContainer := newBitSquirt(entropy)
   currentEntropy = *structContainer

   return entropy
}

type bitSquirt struct {
   bits []byte
   currentBit int
   maxBits int
}

var currentEntropy bitSquirt

func newBitSquirt(newBits []byte) *bitSquirt {
   newBS := bitSquirt{newBits, 0, 0}
   newBS.maxBits = len(newBS.bits) * 8
   return &newBS
}

// TODO possible memory leak??
//func resetBitSquirt() {
   //currentBit = 0;
//}

func squirtBits(numBits int) int {
   var bits int
   var tmp byte

   for i := 0; i < numBits; i++ {
      index := currentEntropy.currentBit / 8
      mask  := byte(math.Pow(float64(2), float64(8 - (currentEntropy.currentBit % 8 + 1))))
      tmp = currentEntropy.bits[index] & mask

      bits = bits << 1
      if (tmp >= 1){
         bits |= 1
      }

      currentEntropy.currentBit++

      if (currentEntropy.currentBit > currentEntropy.maxBits) {
         currentEntropy.currentBit = 0
         break
      }
   }

   return bits
}
