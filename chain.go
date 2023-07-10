package main

import (
	"crypto/ed25519"
	"encoding/binary"
	"errors"
	"sort"
	"time"
)

func (c *Chain) AddBlock(block Block) error {
	err := block.Validate()
	if err != nil {
		return err
	}
	c.BlockHistory = append(c.BlockHistory, block)
	return nil
}

func (c *Chain) Validate() error {
	for _, block := range c.BlockHistory {
		err := block.Validate()
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Chain) AddTransaction(tx Transaction) error {
	err := tx.Validate()
	if err != nil {
		return err
	}
	c.PendingTransactions = append(c.PendingTransactions, tx)
	return nil
}

func (c *Chain) MineBlock(miner PrivateKey) error {
	// Check for transactions to mine
	if len(c.PendingTransactions) == 0 {
		return errors.New("no transactions to mine")
	}

	// Max block size 1MB
	const MaxBlockSize = 1000000

	// Sort the transactions by fees in descending order
	sort.Slice(c.PendingTransactions, func(i, j int) bool {
		return c.PendingTransactions[i].TxFee > c.PendingTransactions[j].TxFee
	})

	// Select transactions for the new block
	var blockTransactions []Transaction
	var totalBlockSize int

	for _, tx := range c.PendingTransactions {
		txSize := binary.Size(tx)
		if totalBlockSize+txSize > MaxBlockSize {
			break
		}

		blockTransactions = append(blockTransactions, tx)
		totalBlockSize += txSize
	}

	// Derive public key from miner private key
	edPublicKey := ed25519.PublicKey(miner[32:])
	publicKey, err := ToPublicKey(edPublicKey)
	if err != nil {
		return err
	}

	// Create a new block
	block := Block{
		Height:       uint64(len(c.BlockHistory) + 1),
		Timestamp:    time.Now(),
		Issuer:       publicKey,
		Transactions: blockTransactions,
	}

	// Sign the block
	err = block.Sign(miner)
	if err != nil {
		return err
	}

	// Validate the block
	err = block.Validate()
	if err != nil {
		return err
	}

	// Add the block to the chain
	c.BlockHistory = append(c.BlockHistory, block)

	// Remove mined transactions from the pool
	c.PendingTransactions = c.PendingTransactions[len(blockTransactions):]

	return nil
}
