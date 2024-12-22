package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

type Tx struct {
	FromID string `json:"from_id"`
	ToID   string `json:"to_id"`
	Value  string `json:"value"`
}

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {

	simpleTrx := Tx{
		FromID: "Molly",
		ToID:   "Aaron",
		Value:  "123",
	}

	privateKey, err := crypto.LoadECDSA("zblock/accounts/cesar.ecdsa")
	if err != nil {
		return fmt.Errorf("error while loading private key %w", err)

	}

	data, err := json.Marshal(simpleTrx)
	if err != nil {
		return fmt.Errorf("error while loading private key %w", err)
	}

	hashData := crypto.Keccak256(data)

	signature, err := crypto.Sign(hashData, privateKey)
	if err != nil {
		return fmt.Errorf("error while signing: %w", err)

	}

	fmt.Println(hexutil.Encode(signature))
	return nil
}
