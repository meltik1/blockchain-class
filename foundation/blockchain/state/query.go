package state

import (
	"emperror.dev/errors"

	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
)

func (s *State) Query(account database.AccountID) (database.Account, error) {
	query, err := s.Db.Query(account)
	if err != nil {
		return database.Account{}, errors.Wrap(err, "Error while querying account")
	}

	return query, nil
}

func (s *State) Accounts() map[database.AccountID]database.Account {
	all := s.Db.All()
	accounts := make(map[database.AccountID]database.Account, len(all))
	for _, account := range all {
		accounts[account.AccountID] = account
	}

	return accounts
}
