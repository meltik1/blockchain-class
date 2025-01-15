package state

import (
	"sync"

	"emperror.dev/errors"

	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
	"github.com/ardanlabs/blockchain/foundation/blockchain/genesis"
	"github.com/ardanlabs/blockchain/foundation/blockchain/mempool"
)

// Config represents the configuration required to start
// the blockchain node.

type Config struct {
	BeneficiaryID   database.AccountID // Аккаунт, который получает вохзнограждеение за майнинг или ГАЗ
	Genesis         genesis.Genesis
	EvHandler       EventHandler
	MemPoolStrategy string
}

// Worker interface represents the behavior required to be implemented by any
// package providing support for mining, peer updates, and transaction sharing.
type Worker interface {
	Shutdown()
	Sync()
	SignalStartMining()
	SignalCancelMining()
	SignalShareTx(blockTx database.BlockTx)
}

// EventHandler defines a function that is called when events
// occur in the processing of persisting blocks
// Мы ее будем юзать для логирования
type EventHandler func(v string, args ...any)

// State manages the blockchain database.
type State struct {
	Mu sync.RWMutex

	BeneficiaryID database.AccountID
	EvHandler     EventHandler

	Genesis genesis.Genesis
	Db      *database.Database
	memPool *mempool.MemPool

	Worker Worker
}

func NewState(cfg Config) (*State, error) {
	// Игнорируем панику
	ev := func(v string, args ...any) {
		if cfg.EvHandler != nil {
			cfg.EvHandler(v, args...)
		}

	}

	db, err := database.NewDatabase(
		cfg.Genesis,
		ev,
	)
	if err != nil {
		return nil, errors.Wrap(err, "Error while creating new database")
	}

	pool, err := mempool.NewWithStrategy(cfg.MemPoolStrategy)
	if err != nil {
		return nil, errors.Wrap(err, "Error while creating new mempool")
	}

	return &State{
		BeneficiaryID: cfg.BeneficiaryID,
		EvHandler:     ev,
		Genesis:       cfg.Genesis,
		Db:            db,
		memPool:       pool,
	}, nil
}

func (s *State) Shutdown() error {
	s.EvHandler("Shutting down the state")

	return nil
}

func (s *State) GetGenesis() genesis.Genesis {
	return s.Genesis
}

// MempoolLength returns the current length of the mempool.
func (s *State) MempoolLength() int64 {
	return s.memPool.Count()
}

// Mempool returns a copy of the mempool.
func (s *State) Mempool() []database.BlockTx {
	return s.memPool.PickBest()
}

// Accepting transaction
func (s *State) SubmitTx(tx database.SignedTx) error {
	// CORE NOTE: It's up to the wallet to make sure the account has a proper
	// balance and this transaction has a proper nonce. Fees will be taken if
	// this transaction is mined into a block it doesn't have enough money to
	// pay or the nonce isn't the next expected nonce for the account.

	// Check the signed transaction has a proper signature, the from matches the
	// signature, and the from and to fields are properly formatted.
	if err := tx.IsValid(); err != nil {
		return errors.Wrap(err, "Invalid transaction")
	}

	const oneUnitOfGas = 1
	blockTx := database.NewBlockTx(tx, s.Genesis.GasPrice, oneUnitOfGas)
	if err := s.memPool.Upsert(blockTx); err != nil {
		return err
	}

	if s.MempoolLength() >= int64(s.Genesis.TransPerBlock) {
		s.Worker.SignalStartMining()
	}

	return nil
}

func (s *State) GetStateRoot() string {
	return s.Db.GetStateRoot()
}

func (s *State) GetLastBlock() database.Block {
	return database.Block{}
}
