package main

import (
	"encoding/json"
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

func handleConnection(conn net.Conn) {
	decoder := json.NewDecoder(conn)
	for {
		var message Message
		err := decoder.Decode(&message)
		if err != nil {
			Log(ERROR, fmt.Sprintf("Failed to decode message from peer: %v", err))
			break
		}

		switch message.Type {
		case MessageTypeHelloRequest:
			// If we received a HelloRequest, verify the peer's public key (add this functionality)
			// For this example, we're assuming all HelloRequest messages have valid keys and NodeID

			// Create a new peer and add it to the GlobalPeers map
			newPeer := &Peer{
				NodeID:    message.HelloReq.NodeID,
				PublicKey: message.HelloReq.PublicKey,
				Conn:      conn,
			}
			GlobalPeers[newPeer.NodeID] = newPeer

			// Generate a HelloResponse and send it back
			response := &HelloResponse{
				NodeID:    MyNodeID,
				PublicKey: MyPublicKey,
			}
			respMessage := &Message{
				Type:     MessageTypeHelloResponse,
				HelloRes: response,
			}
			encoder := json.NewEncoder(conn)
			err = encoder.Encode(respMessage)
			if err != nil {
				Log(ERROR, fmt.Sprintf("Failed to send HelloResponse to peer: %v", err))
			}
		default:
			// If we received a different message type, log a message and do nothing
			Log(INFO, fmt.Sprintf("Received unexpected message type: %v", message.Type))
		}
	}

	conn.Close()
}
