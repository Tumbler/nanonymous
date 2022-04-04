package main

import (
   "math"
)

type bitSquirt struct {
   bits []byte
   currentBit int
   maxBits int
}

// TODO possible memory leak??
func (storedData *bitSquirt) resetBitSquirt() {
   storedData.bits = make([]byte, 0)
   storedData.currentBit = 0;
   storedData.maxBits = 0;
}

func (storedData *bitSquirt) restartBitSquirt() {
   storedData.currentBit = 0;
}

func (storedData *bitSquirt) squirtBits(numBits int) int {
   var bits int
   var tmp byte

   for i := 0; i < numBits; i++ {
      index := storedData.currentBit / 8
      mask  := byte(math.Pow(float64(2), float64(8 - (storedData.currentBit % 8 + 1))))
      tmp = storedData.bits[index] & mask

      bits = bits << 1
      if (tmp >= 1) {
         bits |= 1
      }

      storedData.currentBit++

      if (storedData.currentBit > storedData.maxBits) {
         storedData.currentBit = 0
         break
      }
   }
   // TODO what if maxbits isn't divisible by 8?

   return bits
}

func (storedData *bitSquirt) slurpBits(bitSoup int64, numBits int) {
   var tmp int64
   for i := 0; i < numBits; i++ {
      index := storedData.currentBit / 8
      mask  := (int64)(1 << (numBits - 1 - i))
      tmp = bitSoup & mask

      // Add additional storage space if we need
      if (len(storedData.bits) <= index) {
         storedData.bits = append(storedData.bits, 0)
      }

      storedData.bits[index] = storedData.bits[index] << 1
      if (tmp >= 1) {
         storedData.bits[index] |= 1
      }

      storedData.currentBit++
   }

   storedData.maxBits += numBits
}

func (storedData *bitSquirt) getBitSquirtLength() int {
   if (storedData.maxBits > 0) {
      return len(storedData.bits)
   } else {
      return 0
   }
}

// TODO make squirtbits return value so can evaulate if done with bit
