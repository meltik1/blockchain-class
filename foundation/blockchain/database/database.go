package database

import (
	"sync"

	"emperror.dev/errors"

	"github.com/ardanlabs/blockchain/foundation/blockchain/genesis"
	"github.com/ardanlabs/blockchain/foundation/blockchain/signature"
)

type Database struct {
	mx        sync.RWMutex
	genesis   genesis.Genesis
	accounts  map[AccountID]Account
	evHandler func(v string, args ...interface{})
	st        Storage
}

type Storage interface {
	Save(Block) error
	Delete(blockNumber uint64) error
	Find(blockNumber uint64) (Block, error)
	List() ([]Block, error)
}

var NotFound = errors.New("Account not found")

func NewDatabase(genesis genesis.Genesis, st Storage, ev func(v string, args ...interface{})) (*Database, error) {
	db := Database{
		evHandler: ev,
		genesis:   genesis,
		accounts:  make(map[AccountID]Account),
		st:        st,
	}

	for accountIString, balances := range genesis.Balances {
		accountId, err := ToAccountID(accountIString)
		if err != nil {
			return nil, errors.Wrap(err, "Error while converting accountID")
		}

		db.accounts[accountId] = newAccount(accountId, balances)
		ev("Account : %s, Balance : %d", accountIString, balances)
	}

	list, err := st.List()
	if err != nil {
		return nil, errors.Wrap(err, "Error while listing blocks")
	}

	for _, block := range list {
		for _, tx := range block.MerkleTree.Values() {
			if err := db.ApplyTransaction(tx, block.Header.BeneficiaryID); err != nil {
				return nil, errors.Wrap(err, "Error while applying transaction")
			}
		}

	}

	return &db, nil
}

func (db *Database) Remove(id AccountID) {
	db.mx.Lock()
	defer db.mx.Unlock()

	delete(db.accounts, id)
}

func (db *Database) All() []Account {
	var accounts = make([]Account, 0, len(db.accounts))

	db.mx.RLock()
	defer db.mx.RUnlock()

	for _, account := range db.accounts {
		accounts = append(accounts, account)
	}

	return accounts
}

func (db *Database) Query(id AccountID) (Account, error) {
	db.mx.RLock()
	defer db.mx.RUnlock()

	account, exists := db.accounts[id]
	if !exists {
		return Account{}, NotFound
	}

	return account, nil

}

func (db *Database) GetStateRoot() string {
	return signature.Hash(db.All())
}

func (db *Database) Save(block Block) error {
	return db.st.Save(block)
}

func (db *Database) ApplyTransaction(tx BlockTx, beneficiaryID AccountID) error {
	from := db.accounts[tx.FromID]
	to := db.accounts[tx.ToID]
	beneficiary := db.accounts[beneficiaryID]

	if uint64(from.Balance) < tx.Value+tx.Tip+tx.GasUnits*tx.GasPrice {
		from.Balance -= int64(tx.GasUnits * tx.GasPrice)

		return errors.New("Not enough balance. However we've charged extra money for gas.")
	}

	from.Balance -= int64(tx.Value + tx.Tip + tx.GasUnits*tx.GasPrice)
	to.Balance += int64(tx.Value)

	beneficiary.Balance += int64(tx.Tip)
	beneficiary.Balance += int64(tx.GasUnits * tx.GasPrice)

	db.accounts[tx.FromID] = from
	db.accounts[tx.ToID] = to
	db.accounts[beneficiaryID] = beneficiary

	return nil
}

func (db *Database) ApplyMiningReward(beneficiaryID AccountID) {
	beneficiary := db.accounts[beneficiaryID]
	beneficiary.Balance += db.genesis.MiningReward
}
