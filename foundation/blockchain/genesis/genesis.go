package genesis

import (
	"encoding/json"
	"os"
	"time"
)

type Genesis struct {
	Date          time.Time        `json:"date"`
	ChainID       uint16           `json:"chain_id"`
	TransPerBlock uint16           `json:"trans_per_block"`
	Difficulty    uint16           `json:"difficulty"`
	MiningReward  int64            `json:"mining_reward"`
	GasPrice      int64            `json:"gas_price"`
	Balances      map[string]int64 `json:"balances"`
}

func Load() (Genesis, error) {
	path := "zblock/genesis.json"
	open, err := os.ReadFile(path)
	if err != nil {
		return Genesis{}, err
	}

	var g Genesis
	err = json.Unmarshal(open, &g)
	if err != nil {
		return Genesis{}, err
	}

	return g, nil
}
