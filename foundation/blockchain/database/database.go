package database

import (
	"sync"

	"emperror.dev/errors"

	"github.com/ardanlabs/blockchain/foundation/blockchain/genesis"
)

type Datatabase struct {
	mx       sync.RWMutex
	genesis  genesis.Genesis
	accounts map[AccountID]Account
}

func NewDatatabase(mx sync.RWMutex, genesis genesis.Genesis, accounts map[AccountID]int64) (*Datatabase, error) {
	db := Datatabase{mx: mx,
		genesis:  genesis,
		accounts: make(map[AccountID]Account),
	}

	for accountIString, balances := range genesis.Balances {
		accountId, err := ToAccountID(accountIString)
		if err != nil {
			return nil, errors.Wrap(err, "Error while converting accountID")
		}

		db.accounts[accountId] = newAccount(accountId, balances)
	}

	return &db, nil
}

func (db *Datatabase) Remove(id AccountID) {
	db.mx.Lock()
	defer db.mx.Unlock()

	delete(db.accounts, id)

}

func (db *Datatabase) Query(id AccountID) (Account, error) {
	db.mx.RLock()
	defer db.mx.RUnlock()

	account, exists := db.accounts[id]
	if !exists {
		return Account{}, errors.New("Account not found")
	}

	return account, nil

}
