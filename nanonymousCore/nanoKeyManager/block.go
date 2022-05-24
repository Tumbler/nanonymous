package nanoKeyManager

import (
   "fmt"
   "golang.org/x/crypto/blake2b"

   // Local packages
   nt "nanoTypes"

   // Third party packages
   "github.com/hectorchu/gonano/wallet/ed25519"

)

// Block corresponds to the JSON representation of a block.
type Block struct {
   Type           string     `json:"type"`
   Account        string     `json:"account"`
   Previous       nt.BlockHash  `json:"previous"`
   Representative string     `json:"representative"`
   Balance        *nt.Raw       `json:"balance"`
   Link           nt.BlockHash  `json:"link"`
   LinkAsAccount  string     `json:"link_as_account"`
   Signature      nt.HexData    `json:"signature"`
   Work           nt.HexData    `json:"work"`
   SubType        string
   Seed           Key
}

// Hash calculates the block hash.
func (b *Block) Hash() (hash nt.BlockHash, err error) {
   h, err := blake2b.New256(nil)
   if err != nil {
      return
   }
   h.Write(make([]byte, 31))
   h.Write([]byte{6})
   pubkey, err := AddressToPubKey(b.Account)
   if err != nil {
      return
   }
   h.Write(pubkey)
   h.Write(b.Previous)
   pubkey, err = AddressToPubKey(b.Representative)
   if err != nil {
      return
   }
   h.Write(pubkey)
   h.Write(b.Balance.FillBytes(make([]byte, 16)))
   h.Write(b.Link)
   return h.Sum(nil), nil
}


func (b *Block) Sign() ([]byte, error) {

   if (b.Seed.KeyType > 1 || !b.Seed.Initialized) {
      return nil, fmt.Errorf("Sign: no private key in key struct")
   }

   hash , err := b.Hash()
   if (err != nil) {
      return nil, fmt.Errorf("Sign: %w", err)
   }

   keyPair := append(b.Seed.PrivateKey, b.Seed.PublicKey...)
   return ed25519.Sign(keyPair, hash), nil
}

