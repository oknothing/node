package main

import (
	"net"
	"os"
	"sync"
	"time"
)

const (
	KeysFilename = "keys.txt"
)

const (
	DEBUG LogLevel = iota
	INFO
	WARNING
	ERROR
	CRITICAL
)

var GlobalPeers map[NodeID]*Peer = make(map[NodeID]*Peer)

var (
	logLevelNames = []string{
		"DEBUG",
		"INFO",
		"WARNING",
		"ERROR",
		"CRITICAL",
	}
	logFile     *os.File
	logFileName = "app.log"
	logLevel    LogLevel
)

var MyNodeID NodeID
var MyPublicKey PublicKey
var port int

type LogLevel int

type NodeID string
type Hash [32]byte
type Nonce [32]byte
type PublicKey [32]byte
type PrivateKey [64]byte
type Signature [64]byte
type SuperBlock [100]Block

type MessageType int

const (
	MessageTypeBlock MessageType = iota
	MessageTypeTransaction
	MessageTypeDiscoverPeersRequest
	MessageTypeDiscoverPeersResponse
	MessageTypeHelloRequest
	MessageTypeHelloResponse
)

type Message struct {
	Type        MessageType
	Block       *Block
	Transaction *Transaction
	Request     *DiscoverPeersRequest
	Response    *DiscoverPeersResponse
	HelloReq    *HelloRequest
	HelloRes    *HelloResponse
}
type HelloRequest struct {
	NodeID    NodeID
	PublicKey PublicKey
}

type HelloResponse struct {
	NodeID    NodeID
	PublicKey PublicKey
}
type Peer struct {
	NodeID    NodeID    // This is a globally unique peer identifier.
	Address   net.IP    // This is the IP address for this peer.
	Port      uint16    // This is a port number for this peer.
	PublicKey PublicKey // This is the public key associated with this peer.
	Conn      net.Conn  // This is the TCP connection associated with this peer.
}

type PeerList struct {
	Peers map[NodeID]*Peer // This is a map representing our peer list.
}

type PeerManager struct {
	Peers  map[NodeID]*Peer // These are the managed active peers.
	Mutex  *sync.Mutex      // This is a mutex to ensure consistency when managing our peer list.
	MyNode *Peer            // This is our own peer information.
}

type Block struct {
	Height       uint64        // This is this block's height.
	Nonce        Nonce         // This is the nonce for this block.
	BlockHash    Hash          // This is the hash of this block.
	ParentHash   Hash          // This is the hash of the previous block.
	Version      uint64        // This is the current block template version.
	Timestamp    time.Time     // This is the timestamp of when this block was added to the chain.
	Issuer       PublicKey     // This is who minted this block.
	Signature    Signature     // This is the signature from the issuer of this block's contents.
	Transactions []Transaction // These are the transactions in this block.
}

type Transaction struct {
	Nonce     Nonce     // This is the nonce for this transaction.
	TxHash    Hash      // This is the hash of this tx.
	Sender    PublicKey // This is who sent this tx.
	Recipient PublicKey // This is the tx recipient.
	Amount    uint64    // This is the number of units being sent.
	TxFee     uint64    // This is the number of units for fee.
	Timestamp time.Time // This is the time for when this transaction was first seen by the Issuer.
	Signature Signature // This is the transaction signature from this Sender.
}

type Chain struct {
	FeeBasis            uint64        // This is the minimum fee amount.
	BlockInterval       time.Time     // This is the minimum amount of time between blocks. Blocks may not be produced in less than this amount of time.
	SuperBlockSize      uint16        // A single issuer consolidates their blocks into a compound block called a 'SuperBlock' consisting of this many normal blocks.
	BlockHistory        []Block       // This is a list of blocks on the chain.
	PendingTransactions []Transaction // This is a list of the transactions pending inclusion into a block.
}

type KeyPair struct {
	PrivateKey PrivateKey //  This is the private signing key.
	PublicKey  PublicKey  // This is the public key derived from the private key.
}

type DiscoverPeersRequest struct {
	KnownPeers []NodeID
}

type DiscoverPeersResponse struct {
	Peers []NodeID
}
