package main

import (
	"fmt"
	"net"
	"os"
)

func main() {

	// Check that we have keys, or make them
	myKeys := initKeypair()

	// Create a new blockchain
	chain := initChain()

	// Validate the blockchain
	err := chain.Validate()
	Log(DEBUG, "validating blockchain...")
	if err != nil {
		Log(CRITICAL, err.Error())
	}

	// Start the peer-to-peer network
	peerManager := StartPeerNetwork(myKeys)

	// Discover new peers
	peerManager.DiscoverPeers()

	// Generate some demo transactions
	err = generateDemoTXData(myKeys, chain)
	if err != nil {
		Log(ERROR, err.Error())
	}

	// Broadcast the first demo transaction to all peers
	if len(chain.PendingTransactions) > 0 {
		demoTx := &chain.PendingTransactions[0]
		for _, peer := range peerManager.Peers {
			if err := peer.SendTransaction(demoTx); err != nil {
				Log(ERROR, fmt.Sprintf("Failed to send transaction to peer %s", peer.NodeID))
			}
		}
	}

	Log(DEBUG, "blockchain loaded and validated")

	// Start the listener in a new goroutine
	go func() {
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			Log(ERROR, fmt.Sprintf("Failed to listen on port %d: %v", port, err))
			os.Exit(1)
		}
		defer listener.Close()

		for {
			conn, err := listener.Accept()
			if err != nil {
				Log(ERROR, fmt.Sprintf("Failed to accept connection: %v", err))
				continue
			}

			// Handle each connection in a new goroutine
			go handleConnection(conn)
		}
	}()

	select {}
}
