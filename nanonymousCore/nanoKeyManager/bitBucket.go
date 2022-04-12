package nanoKeyManager

import (
   "fmt"
   "errors"
   "math"
)

type bitBucket struct {
   bits []byte
   currentBit int
   maxBits int
   slurpLocation int
}

// setBitSquirtPosition sets the internal variables of the bitBucket so that the
// next time squirtBits is used it will start at the given index.
func (storedData *bitBucket) setBitSquirtPosition(position int) {
   storedData.currentBit = position;
}

// squirtBits returns the next numBits bits in the bucket. If there were less
// than numBits left, it will return the remaining bits and how many there were.
func (storedData *bitBucket) squirtBits(numBits int) (int64, int) {
   var bits int64
   var tmp byte

   for i := 0; i < numBits; i++ {

      // If there are no bits left, return early
      if (storedData.currentBit >= storedData.maxBits) {
         return bits, i
      }

      index := storedData.currentBit / 8

      // The very last byte must be treated differently because it might not be
      // completly full.
      var mask byte
      if (index == (len(storedData.bits) - 1)) {
         // This case only matters if we have a number of bits not divisible by 8.
         mask  = byte(math.Pow(float64(2), float64(((storedData.maxBits - 1) % 8 + 1) - (storedData.currentBit % 8 + 1))))
      } else {
         mask  = byte(math.Pow(float64(2), float64(8 - (storedData.currentBit % 8 + 1))))
      }
      tmp = storedData.bits[index] & mask

      bits = bits << 1
      if (tmp >= 1) {
         bits |= 1
      }

      storedData.currentBit++

   }

   return bits, numBits
}

// slurpBits takes an arbitrary number of bits and stores them into its internal
// storage.
func (storedData *bitBucket) slurpBits(bitSoup int64, numBits int) error{
   var tmp int64

   for i := 0; i < numBits; i++ {
      index := storedData.slurpLocation / 8
      mask  := (int64)(1 << (numBits - 1 - i))
      tmp = bitSoup & mask

      // Add additional storage space if we need
      if (len(storedData.bits) <= index) {
         storedData.bits = append(storedData.bits, 0)
      }

      // Can only get here if slurpLocation has been messed with outside this
      // function.
      if (len(storedData.bits) <= index) {
         err := errors.New(fmt.Sprint("index ", index, " out of bounds for bitBucket"))
         return fmt.Errorf("slurpBits: %w", err)
      }

      storedData.bits[index] = storedData.bits[index] << 1
      if (tmp >= 1) {
         storedData.bits[index] |= 1
      }

      storedData.slurpLocation++
   }

   storedData.maxBits += numBits

   return nil
}
