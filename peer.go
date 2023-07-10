package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"sync"
)

func (pl *PeerList) InitializePeers(peerParams []struct {
	NodeID    NodeID
	Address   net.IP
	Port      uint16
	PublicKey PublicKey
}) {
	for _, param := range peerParams {
		pl.Peers[param.NodeID] = &Peer{
			NodeID:    param.NodeID,
			Address:   param.Address,
			Port:      param.Port,
			PublicKey: param.PublicKey,
		}
	}
}

// Generate a random NodeID
func GenerateNodeID() NodeID {
	b := make([]byte, 16)
	rand.Read(b)
	Log(DEBUG, "nodeID: "+hex.EncodeToString(b))
	return NodeID(hex.EncodeToString(b))
}

func NewPeerList() *PeerList {
	Log(DEBUG, "creating new peer list..")
	return &PeerList{
		Peers: make(map[NodeID]*Peer),
	}
}

func NewPeerManager(myNode *Peer) *PeerManager {
	Log(DEBUG, "starting peer manager..")
	return &PeerManager{
		Peers:  make(map[NodeID]*Peer),
		Mutex:  &sync.Mutex{},
		MyNode: myNode,
	}
}

func StartPeerNetwork(myKeys KeyPair) *PeerManager {
	// Instantiate our PeerManager and our own Peer
	Log(DEBUG, "starting peer networking..")
	myNode := &Peer{
		NodeID:    GenerateNodeID(),
		PublicKey: myKeys.PublicKey,
		// fill in Address and Port fields
	}

	MyNodeID = myNode.NodeID
	MyPublicKey = myNode.PublicKey
	// Initialize the PeerManager with our node
	peerManager := &PeerManager{
		Peers:  make(map[NodeID]*Peer),
		Mutex:  new(sync.Mutex),
		MyNode: myNode,
	}

	// Hardcoded bootstrap peers
	bootstrapPeers := []struct {
		ip   string
		port uint16
	}{
		{"170.64.168.154", 19876},
		{"159.65.11.179", 19876},
		{"165.22.9.57", 19876},
	}

	for _, s := range bootstrapPeers {
		Log(DEBUG, fmt.Sprintf("bootstrap peer: %s:%d", s.ip, s.port))
	}

	// Connect to each bootstrap peer
	for _, info := range bootstrapPeers {
		peer := &Peer{
			Address: net.ParseIP(info.ip),
			Port:    info.port,
		}
		if err := peer.Connect(); err != nil {
			Log(ERROR, fmt.Sprintf("failed to connect to peer %s:%d", info.ip, info.port))
			continue
		}

		// Exchange HelloRequest and HelloResponse to get NodeID
		helloRequest := &HelloRequest{
			NodeID:    myNode.NodeID,
			PublicKey: myNode.PublicKey,
		}

		helloResponse, err := peer.SendHelloRequest(helloRequest)
		if err != nil {
			Log(ERROR, fmt.Sprintf("failed to send HelloRequest to peer %s:%d", info.ip, info.port))
			continue
		}

		// After successfully getting HelloResponse, set NodeID and PublicKey
		peer.NodeID = helloResponse.NodeID
		peer.PublicKey = helloResponse.PublicKey

		// Add to PeerManager's peers
		peerManager.Peers[peer.NodeID] = peer
	}

	// Discover new peers
	peerManager.DiscoverPeers()

	return peerManager
}

// Establishes a connection to the peer.
func (p *Peer) Connect() error {
	address := net.JoinHostPort(p.Address.String(), strconv.Itoa(int(p.Port)))
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}

	p.Conn = conn
	return nil
}

// Sends a block to the peer.
func (p *Peer) SendBlock(block *Block) error {
	message := Message{
		Type:  MessageTypeBlock,
		Block: block,
	}
	return p.sendMessage(&message)
}

// Sends a transaction to the peer.
func (p *Peer) SendTransaction(tx *Transaction) error {
	message := Message{
		Type:        MessageTypeTransaction,
		Transaction: tx,
	}
	return p.sendMessage(&message)
}

// SendDiscoverPeersRequest Sends a DiscoverPeersRequest to the peer.
func (p *Peer) SendDiscoverPeersRequest(request *DiscoverPeersRequest) (*DiscoverPeersResponse, error) {
	message := Message{
		Type:    MessageTypeDiscoverPeersRequest,
		Request: request,
	}
	err := p.sendMessage(&message)
	if err != nil {
		return nil, err
	}

	// Assuming we get a response immediately after a request
	responseMessage, err := p.receiveMessage()
	if err != nil {
		return nil, err
	}

	// Check if we received the correct message type
	if responseMessage.Type != MessageTypeDiscoverPeersResponse {
		return nil, fmt.Errorf("unexpected message type received: %v", responseMessage.Type)
	}

	return responseMessage.Response, nil
}

func (p *Peer) sendMessage(message *Message) error {
	encoder := json.NewEncoder(p.Conn)
	return encoder.Encode(message)
}

func (p *Peer) receiveMessage() (*Message, error) {
	decoder := json.NewDecoder(p.Conn)
	var message Message
	err := decoder.Decode(&message)
	if err != nil {
		return nil, err
	}
	return &message, nil
}

// AddPeer Adds a new peer to the peer list.
func (pm *PeerManager) AddPeer(peer *Peer) {
	pm.Mutex.Lock()
	defer pm.Mutex.Unlock()

	if _, exists := pm.Peers[peer.NodeID]; !exists {
		pm.Peers[peer.NodeID] = peer
	}
}

// RemovePeer Removes a peer from the peer list.
func (pm *PeerManager) RemovePeer(nodeID NodeID) {
	pm.Mutex.Lock()
	defer pm.Mutex.Unlock()

	delete(pm.Peers, nodeID)
}

func NodeIDToPeer(nodeID NodeID) (*Peer, error) {
	peer, exists := GlobalPeers[nodeID]
	if !exists {
		return nil, fmt.Errorf("peer with NodeID %s not found", nodeID)
	}

	return peer, nil
}

// Discovers new peers by sending a DiscoverPeersRequest to all known peers
func (pm *PeerManager) DiscoverPeers() {
	Log(DEBUG, "discovering new peers..")

	pm.Mutex.Lock()
	defer pm.Mutex.Unlock()

	knownPeers := make([]NodeID, 0, len(pm.Peers))
	for nodeID := range pm.Peers {
		Log(DEBUG, "adding peer: "+string(nodeID))
		knownPeers = append(knownPeers, nodeID)
	}

	request := &DiscoverPeersRequest{KnownPeers: knownPeers}

	for _, peer := range pm.Peers {
		response, err := peer.SendDiscoverPeersRequest(request)
		if err != nil {
			continue
		}

		for _, nodeID := range response.Peers {
			if _, exists := pm.Peers[nodeID]; !exists {
				newPeer, err := NodeIDToPeer(nodeID)
				if err == nil {
					pm.Peers[nodeID] = newPeer
				}
			}
		}
	}
}

// SendHelloRequest sends a HelloRequest to the peer.
func (p *Peer) SendHelloRequest(request *HelloRequest) (*HelloResponse, error) {
	message := Message{
		Type:     MessageTypeHelloRequest,
		HelloReq: request,
	}
	err := p.sendMessage(&message)
	if err != nil {
		return nil, err
	}

	// Assuming we get a response immediately after a request
	responseMessage, err := p.receiveMessage()
	if err != nil {
		return nil, err
	}

	// Check if we received the correct message type
	if responseMessage.Type != MessageTypeHelloResponse {
		return nil, fmt.Errorf("unexpected message type received: %v", responseMessage.Type)
	}

	return responseMessage.HelloRes, nil
}

// SendHelloResponse sends a HelloResponse to the peer.
func (p *Peer) SendHelloResponse(response *HelloResponse) error {
	message := Message{
		Type:     MessageTypeHelloResponse,
		HelloRes: response,
	}
	return p.sendMessage(&message)
}
