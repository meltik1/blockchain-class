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
	privateKey, err := crypto.LoadECDSA("zblock/accounts/cesar.ecdsa")
	if err != nil {
		return fmt.Errorf("error while loading private key %w", err)

	}

	simpleTrx := Tx{
		FromID: "Molly",
		ToID:   "Aaron",
		Value:  "123",
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

	pub, err := crypto.SigToPub(hashData, signature)
	if err != nil {
		return err
	}

	fmt.Println(crypto.PubkeyToAddress(*pub).String())

	// simple trx 2

	simpleTrx2 := Tx{
		FromID: "Molly2",
		ToID:   "Aaron",
		Value:  "123",
	}

	data2, err := json.Marshal(simpleTrx2)
	if err != nil {
		return fmt.Errorf("error while loading private key %w", err)
	}

	hashData2 := crypto.Keccak256(data2)
	signature2, err := crypto.Sign(hashData2, privateKey)
	if err != nil {
		return fmt.Errorf("error while signing: %w", err)

	}

	fmt.Println(hexutil.Encode(signature2))

	// detecting fraud
	pub2, err := crypto.SigToPub(hashData2, signature)
	if err != nil {
		return err
	}

	fmt.Println(crypto.PubkeyToAddress(*pub2).String())
	return nil
}
