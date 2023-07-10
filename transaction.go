package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
)

func (t *Transaction) Validate() error {
	// Compute the hash of the transaction
	txHash := t.Hash()

	// Verify the signature of the transaction
	if !ed25519.Verify(ed25519.PublicKey(t.Sender[:]), txHash[:], t.Signature[:]) {
		return errors.New("transaction signature is invalid")
	}

	// Additional transaction validation checks
	if t.Amount <= 0 {
		return errors.New("transaction amount must be greater than zero")
	}

	if t.Sender == t.Recipient {
		return errors.New("sender and recipient cannot be the same")
	}

	return nil
}

func (t *Transaction) Hash() Hash {
	h := sha256.New()

	// Add transaction fields to hash
	h.Write(t.Nonce[:])
	h.Write(t.Sender[:])
	h.Write(t.Recipient[:])
	binary.Write(h, binary.LittleEndian, t.Amount)
	binary.Write(h, binary.LittleEndian, t.TxFee)
	binary.Write(h, binary.LittleEndian, t.Timestamp.Unix())

	return Hash(sha256.Sum256(h.Sum(nil)))
}

func (t *Transaction) Sign(key PrivateKey) error {
	// Generate a new nonce for this transaction
	_, err := rand.Read(t.Nonce[:])
	if err != nil {
		return err
	}

	// Compute the hash of the transaction
	txHash := t.Hash()

	// Generate the signature
	signature := ed25519.Sign(ed25519.PrivateKey(key[:]), txHash[:])

	// Check that the signature is correct length
	if len(signature) != len(t.Signature) {
		return errors.New("signature generation failed")
	}

	// Copy the signature into the transaction
	copy(t.Signature[:], signature)

	return nil
}
