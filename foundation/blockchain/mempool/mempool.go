package mempool

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
	"github.com/ardanlabs/blockchain/foundation/blockchain/mempool/selector"
)

type MemPool struct {
	mw       sync.RWMutex
	pool     map[string]database.BlockTx
	selectFn selector.Func
}

func NewMemPool() (*MemPool, error) {
	return NewStrategy(selector.StrategyTip)
}

func NewStrategy(strategy string) (*MemPool, error) {
	selectFn, err := selector.Retrieve(strategy)
	if err != nil {
		return nil, err
	}

	mp := MemPool{
		pool:     make(map[string]database.BlockTx),
		selectFn: selectFn,
	}

	return &mp, nil
}

func (mp *MemPool) Count() int64 {
	mp.mw.RLock()
	defer mp.mw.RUnlock()

	return int64(len(mp.pool))
}

func (mp *MemPool) Remove(tx database.BlockTx) {
	mp.mw.Lock()
	defer mp.mw.Unlock()

	key := mapKey(tx)
	delete(mp.pool, key)
}

func (mp *MemPool) Truncate() {
	mp.mw.Lock()
	defer mp.mw.Unlock()

	mp.pool = make(map[string]database.BlockTx)
}

func (mp *MemPool) Upsert(tx database.BlockTx) error {
	mp.mw.Lock()
	defer mp.mw.Unlock()

	// CORE NOTE: Different blockchains have different algorithms to limit the
	// size of the mempool. Some limit based on the amount of memory being
	// consumed and some may limit based on the number of transaction. If a limit
	// is met, then either the transaction that has the least return on investment
	// or the oldest will be dropped from the pool to make room for new the transaction.

	key := mapKey(tx)

	// Ethereum requires a 10% bump in the tip to replace an existing
	// transaction in the mempool and so do we. We want to limit users
	// from this sort of behavior.
	if etx, exists := mp.pool[key]; exists {
		if tx.Tip < uint64(math.Round(float64(etx.Tip)*1.10)) {
			return errors.New("replacing a transaction requires a 10% bump in the tip")
		}
	}

	mp.pool[key] = tx

	return nil
}

// PickBest uses the configured sort strategy to return a set of transactions.
// If 0 is passed, all transactions in the mempool will be returned.
func (mp *MemPool) PickBest(howMany ...uint16) []database.BlockTx {
	number := 0
	if len(howMany) > 0 {
		number = int(howMany[0])
	}

	// CORE NOTE: Most blockchains do set a max block size limit and this size
	// will determined which transactions are selected. When picking the best
	// transactions for the next block, the Ardan blockchain is currently not
	// focused on block size but a max number of transactions.
	//
	// When the selection algorithm does need to consider sizing, picking the
	// right transactions that maximize profit gets really hard. On top of this,
	// today a miner gets a mining reward for each mined block. In the future
	// this could go away leaving just fees for the transactions that are
	// selected as the only form of revenue. This will change how transactions
	// need to be selected.

	// Copy all the transactions for each account into separate slices.
	m := make(map[database.AccountID][]database.BlockTx)
	mp.mw.RLock()
	{
		if number == 0 {
			number = len(mp.pool)
		}

		for key, tx := range mp.pool {
			account := accountFromMapKey(key)
			m[account] = append(m[account], tx)
		}
	}
	mp.mw.RUnlock()

	return mp.selectFn(m, number)

}

func mapKey(tx database.BlockTx) string {
	return fmt.Sprintf("%s:%d", tx.FromID, tx.Nonce)
}

// accountFromMapKey extracts the account information from the mapkey.
func accountFromMapKey(key string) database.AccountID {
	return database.AccountID(strings.Split(key, ":")[0])
}
