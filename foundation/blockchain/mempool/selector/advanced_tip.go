package selector

import (
	"sort"

	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
)

// advancedTipSelect returns transactions with the best tip while respecting the nonce
// for each account/transaction. This strategy takes into account high-value transactions
// that happens to be stuck on a low-nonce transaction with a low tip price.
var advancedTipSelect = func(m map[database.AccountID][]database.BlockTx, howMany int) []database.BlockTx {
	final := []database.BlockTx{}

	// Sort the transactions per account by nonce.
	for key := range m {
		if len(m[key]) > 1 {
			sort.Sort(byNonce(m[key]))
		}
	}

	at := newAdvancedTips(m, howMany)
	for accountID, transactionNumber := range at.findBest() {
		for i := 0; i < transactionNumber; i++ {
			final = append(final, m[accountID][i])
		}
	}

	return final
}

// =============================================================================

type advancedTips struct {
	howMany                     int
	bestTip                     uint64
	accountToTransactionsNumber map[database.AccountID]int
	accountCummilativeTips      map[database.AccountID][]uint64
	accountsList                []database.AccountID
}

func newAdvancedTips(accountsToTransactionsMap map[database.AccountID][]database.BlockTx, howMany int) *advancedTips {
	accountCummilativeTips := map[database.AccountID][]uint64{}
	accounts := []database.AccountID{}

	for accountID := range accountsToTransactionsMap {
		accountCummilativeTips[accountID] = []uint64{0}
		accounts = append(accounts, accountID)
	}

	for accountID, transactions := range accountsToTransactionsMap {
		for cummilativeTipIndex, tx := range transactions {
			if cummilativeTipIndex > howMany {
				break
			}
			accountCummilativeTips[accountID] = append(accountCummilativeTips[accountID], tx.Tip+accountCummilativeTips[accountID][cummilativeTipIndex])
		}
	}

	return &advancedTips{
		howMany:                howMany,
		accountCummilativeTips: accountCummilativeTips,
		accountsList:           accounts,
	}
}

func (advancedTip *advancedTips) findBest() map[database.AccountID]int {
	advancedTip.findBestTransactions(0, 0, advancedTip.howMany, advancedTip.accountToTransactionsNumber, 0)
	return advancedTip.accountToTransactionsNumber
}

func (advancedTip *advancedTips) findBestTransactions(accountIndex, pos int, transactionsInBlock int, accountTransactions map[database.AccountID]int, prevTip uint64) {
	if prevTip > advancedTip.bestTip {
		advancedTip.bestTip = prevTip
		advancedTip.accountToTransactionsNumber = accountTransactions
	}

	if accountIndex >= len(advancedTip.accountsList) {
		return
	}
	accountID := advancedTip.accountsList[accountIndex]

	for tipIndex, cummilativeTip := range advancedTip.accountCummilativeTips[accountID] {
		if transactionsInBlock-tipIndex < 0 {
			break
		}

		newTipIndex := copyMap(accountTransactions)
		newTipIndex[accountID] = tipIndex
		advancedTip.findBestTransactions(accountIndex+1, tipIndex, transactionsInBlock-tipIndex, newTipIndex, prevTip+cummilativeTip)
	}
}

// =============================================================================

func copyMap(m map[database.AccountID]int) map[database.AccountID]int {
	newCurrPos := map[database.AccountID]int{}
	for from, pos := range m {
		newCurrPos[from] = pos
	}

	return newCurrPos
}
