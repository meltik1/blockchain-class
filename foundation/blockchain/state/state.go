package state

import (
	"sync"

	"emperror.dev/errors"

	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
	"github.com/ardanlabs/blockchain/foundation/blockchain/genesis"
)

// Config represents the configuration required to start
// the blockchain node.

type Config struct {
	BeneficiaryID database.AccountID // Аккаунт, который получает вохзнограждеение за майнинг или ГАЗ
	Genesis       genesis.Genesis
	EvHandler     EventHandler
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

	return &State{
		BeneficiaryID: cfg.BeneficiaryID,
		EvHandler:     ev,
		Genesis:       cfg.Genesis,
		Db:            db,
	}, nil
}

func (s *State) Shutdown() error {
	s.EvHandler("Shutting down the state")

	return nil
}
