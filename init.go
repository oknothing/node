package main

import (
	"bufio"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

func init() {

	initFlags()

	// Open the log file
	var err error
	logFile, err = os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}

	// Set the log output to the file
	log.SetOutput(logFile)
	Log(DEBUG, "================== init ==================")
}

func initFlags() {
	// Define the log level flag
	flag.Func("loglevel", "Set the log level (DEBUG, INFO, WARNING, ERROR, CRITICAL)", func(s string) error {
		for i, name := range logLevelNames {
			if s == name {
				logLevel = LogLevel(i)
				return nil
			}
		}
		return fmt.Errorf("invalid log level %q", s)
	})
	flag.IntVar(&port, "port", 19876, "Port number to listen on")
	// Parse the flags
	flag.Parse()
}

func initChain() *Chain {
	chain := &Chain{
		FeeBasis:       10,
		SuperBlockSize: 100,
	}
	Log(DEBUG, "creating new blockchain")
	Log(DEBUG, "FeeBasis "+strconv.Itoa(int(chain.FeeBasis)))
	Log(DEBUG, "SuperBlockSize "+strconv.Itoa(int(chain.SuperBlockSize)))
	return chain
}

func generateDemoTXData(myKeys KeyPair, chain *Chain) error {
	recipient, err := GenerateKeyPair()
	if err != nil {
		panic(err)
	}

	Log(DEBUG, "generating demo transactions..")
	transactions, err := GenerateDemoTransactions(myKeys, *recipient, 500)
	if err != nil {
		panic(err)
	}

	for _, tx := range transactions {
		err := chain.AddTransaction(tx)
		if err != nil {
			panic(err)
		}

	}

	Log(DEBUG, "processing transactions..")
	err = chain.MineBlock(myKeys.PrivateKey)
	if err != nil {
		panic(err)
	}
	return err
}

func initKeypair() KeyPair {
	var myKeys KeyPair

	if _, err := os.Stat("keys.txt"); os.IsNotExist(err) {

		publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			panic(err)
		}

		Log(DEBUG, "generating new keys, no keys.txt found")

		copy(myKeys.PublicKey[:], publicKey)
		copy(myKeys.PrivateKey[:], privateKey)

		file, err := os.Create("keys.txt")
		if err != nil {
			panic(err)
		}
		defer file.Close()

		_, err = file.WriteString("PublicKey: " + hex.EncodeToString(myKeys.PublicKey[:]) + "\n")
		if err != nil {
			panic(err)
		}
		_, err = file.WriteString("PrivateKey: " + hex.EncodeToString(myKeys.PrivateKey[:]) + "\n")
		if err != nil {
			panic(err)
		}
		Log(DEBUG, "PublicKey: "+hex.EncodeToString(myKeys.PublicKey[:]))
	} else {

		file, err := os.Open("keys.txt")
		Log(DEBUG, "loading keys.txt")
		if err != nil {
			Log(DEBUG, "loading keys.txt failed")
			Log(ERROR, err.Error())
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "PublicKey: ") {
				pubKey, err := hex.DecodeString(line[len("PublicKey: "):])
				if err != nil {
					panic(err)
				}
				copy(myKeys.PublicKey[:], pubKey)
			} else if strings.HasPrefix(line, "PrivateKey: ") {
				privKey, err := hex.DecodeString(line[len("PrivateKey: "):])
				if err != nil {
					panic(err)
				}
				copy(myKeys.PrivateKey[:], privKey)
			}
		}
		if err := scanner.Err(); err != nil {
			panic(err)
		}
		Log(DEBUG, "keys loaded successfully")
	}
	return myKeys
}
