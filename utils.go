package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"time"
)

func Log(level LogLevel, message string) {
	if level >= logLevel {
		log.Println(logLevelNames[level]+":", message)
		if level >= ERROR {
			fmt.Fprintln(os.Stderr, logLevelNames[level]+":", message)
		} else {
			fmt.Println(logLevelNames[level]+":", message)
		}
	}
}

// GenerateKeyPair creates a new ed25519 public-private key pair.
// The public key is derived from the private key.
func GenerateKeyPair() (*KeyPair, error) {
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	// Copying the generated keys into our key types
	var publicKey PublicKey
	copy(publicKey[:], pubKey)

	var privateKey [64]byte
	copy(privateKey[:], privKey)

	return &KeyPair{PrivateKey: privateKey, PublicKey: publicKey}, nil
}

func GenerateDemoTransactions(sender, recipient KeyPair, count int) ([]Transaction, error) {
	var transactions []Transaction

	for i := 0; i < count; i++ {
		// Create a new transaction
		tx := Transaction{
			Sender:    sender.PublicKey,
			Recipient: recipient.PublicKey,
			Amount:    uint64(i+1) * 100, // just an example amount
			TxFee:     uint64(i+1) * 10,  // just an example transaction fee
			Timestamp: time.Now(),
		}

		// Sign the transaction
		err := tx.Sign(sender.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to sign transaction: %w", err)
		}

		transactions = append(transactions, tx)
	}

	return transactions, nil
}

func ToPublicKey(edPubKey ed25519.PublicKey) (PublicKey, error) {
	var pubKey PublicKey
	if len(edPubKey) != len(pubKey) {
		return pubKey, fmt.Errorf("invalid public key length")
	}
	copy(pubKey[:], edPubKey)
	return pubKey, nil
}
