// Package public maintains the group of handlers for public access.
package public

import (
	"context"
	"fmt"
	"net/http"

	"emperror.dev/errors"
	"go.uber.org/zap"

	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
	"github.com/ardanlabs/blockchain/foundation/blockchain/state"
	"github.com/ardanlabs/blockchain/foundation/web"
)

const FIXED_GAS_PRICE = 1
const FIXED_GAS_AMOUNT = 1

// Handlers manages the set of bar ledger endpoints.
type Handlers struct {
	Log   *zap.SugaredLogger
	State *state.State
}

// Sample just provides a starting point for the class.
func (h Handlers) Sample(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	resp := struct {
		Status string
	}{
		Status: "OK",
	}

	return web.Respond(ctx, w, resp, http.StatusOK)
}

func (h Handlers) Genesis(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	genesis := h.State.GetGenesis()

	return web.Respond(ctx, w, genesis, http.StatusOK)
}

func (h Handlers) GetAccounts(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	accountStr := web.Param(r, "account")

	var accounts map[database.AccountID]database.Account
	switch accountStr {
	case "":
		accounts = h.State.Accounts()

	default:
		accountID, err := database.ToAccountID(accountStr)
		if err != nil {
			return err
		}
		account, err := h.State.Query(accountID)
		if err != nil {
			if errors.Is(err, database.NotFound) {
				return web.Respond(ctx, w, nil, http.StatusNotFound)
			}
			return err
		}
		accounts = map[database.AccountID]database.Account{accountID: account}
	}

	resp := make([]accountDTO, 0, len(accounts))
	for account, info := range accounts {
		act := accountDTO{
			Account: account,
			Balance: info.Balance,
			Nonce:   info.Nonce,
		}
		resp = append(resp, act)
	}

	return web.Respond(ctx, w, resp, http.StatusOK)
}

func (h Handlers) MemPool(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	mempool := h.State.Mempool()

	var resultTx = make([]txDTO, 0, len(mempool))

	for _, tx := range mempool {
		resultTx = append(resultTx, txDTO{
			FromAccount: tx.FromID,
			To:          tx.ToID,
			Value:       tx.Value,
			Nonce:       tx.Nonce,
			ChainID:     tx.ChainId,
			Tip:         tx.Tip,
			GasPrice:    tx.GasPrice,
			GasUnits:    tx.GasUnits,
			Data:        tx.Data,
			TimeStamp:   tx.TimeStamp,
			Sig:         tx.SignatureString(),
		})
	}

	return web.Respond(ctx, w, mempool, http.StatusOK)
}

func (h Handlers) SubmitWalletTransaction(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	// Decode the JSON in the post call into a Signed transaction.
	var signedTx database.SignedTx
	if err := web.Decode(r, &signedTx); err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}

	err := h.State.SubmitTx(signedTx)
	if err != nil {
		h.Log.Error(errors.Wrap(err, "h.State.SubmitTx"))
		return web.Respond(ctx, w, nil, http.StatusInternalServerError)
	}

	resp := struct {
		Status string `json:"status"`
	}{
		Status: "transactions added to mempool",
	}

	return web.Respond(ctx, w, resp, http.StatusOK)
}
