package database

import (
	"sync"

	"github.com/ardanlabs/blockchain/foundation/blockchain"
	"github.com/ardanlabs/blockchain/foundation/blockchain/genesis"
)

type Datatabase struct {
	mx       sync.RWMutex
	genesis  genesis.Genesis
	accounts map[blockchain.AccountID]int64
}
