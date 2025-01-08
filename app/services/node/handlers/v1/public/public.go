// Package public maintains the group of handlers for public access.
package public

import (
	"context"
	"net/http"

	"go.uber.org/zap"

	"github.com/ardanlabs/blockchain/foundation/blockchain/state"
	"github.com/ardanlabs/blockchain/foundation/web"
)

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
