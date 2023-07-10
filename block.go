package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
)

func (b *Block) Validate() error {
	// Compute the hash of the block
	blockHash := b.Hash()

	// Verify the signature of the block
	if !ed25519.Verify(ed25519.PublicKey(b.Issuer[:]), blockHash[:], b.Signature[:]) {
		return errors.New("block signature is invalid")
	}

	// Additional block validation checks
	if b.Height < 1 {
		return errors.New("block height must be greater than zero")
	}

	if len(b.Transactions) == 0 {
		return errors.New("block must have at least one transaction")
	}

	// Check that the transactions in the block are valid
	for _, tx := range b.Transactions {
		err := tx.Validate()
		if err != nil {
			return fmt.Errorf("invalid transaction in block: %w", err)
		}
	}

	return nil
}

func (b *Block) Hash() Hash {
	h := sha256.New()

	// Add block fields to hash
	binary.Write(h, binary.LittleEndian, b.Height)
	h.Write(b.Nonce[:])
	h.Write(b.ParentHash[:])
	binary.Write(h, binary.LittleEndian, b.Version)
	binary.Write(h, binary.LittleEndian, b.Timestamp.Unix())
	h.Write(b.Issuer[:])

	// Add each transaction hash to block hash
	for _, tx := range b.Transactions {
		txHash := tx.Hash()
		h.Write(txHash[:])
	}

	return Hash(sha256.Sum256(h.Sum(nil)))
}

func (b *Block) Sign(key PrivateKey) error {
	// Generate a new nonce for this block
	_, err := rand.Read(b.Nonce[:])
	if err != nil {
		return err
	}

	// Compute the hash of the block
	blockHash := b.Hash()

	// Generate the signature
	signature := ed25519.Sign(ed25519.PrivateKey(key[:]), blockHash[:])

	// Check that the signature is correct length
	if len(signature) != len(b.Signature) {
		return errors.New("signature generation failed")
	}

	// Copy the signature into the block
	copy(b.Signature[:], signature)

	return nil
}
