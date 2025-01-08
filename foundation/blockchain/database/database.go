package database

import (
	"sync"

	"emperror.dev/errors"

	"github.com/ardanlabs/blockchain/foundation/blockchain/genesis"
)

type Database struct {
	mx        sync.RWMutex
	genesis   genesis.Genesis
	accounts  map[AccountID]Account
	evHandler func(v string, args ...interface{})
}

func NewDatabase(genesis genesis.Genesis, ev func(v string, args ...interface{})) (*Database, error) {
	db := Database{
		evHandler: ev,
		genesis:   genesis,
		accounts:  make(map[AccountID]Account),
	}

	for accountIString, balances := range genesis.Balances {
		accountId, err := ToAccountID(accountIString)
		if err != nil {
			return nil, errors.Wrap(err, "Error while converting accountID")
		}

		db.accounts[accountId] = newAccount(accountId, balances)
		ev("Account : %s, Balance : %d", accountIString, balances)
	}

	return &db, nil
}

func (db *Database) Remove(id AccountID) {
	db.mx.Lock()
	defer db.mx.Unlock()

	delete(db.accounts, id)

}

func (db *Database) Query(id AccountID) (Account, error) {
	db.mx.RLock()
	defer db.mx.RUnlock()

	account, exists := db.accounts[id]
	if !exists {
		return Account{}, errors.New("Account not found")
	}

	return account, nil

}
