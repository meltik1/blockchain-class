// Package v1 contains the full set of handler functions and routes
// supported by the v1 web api.
package v1

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/ardanlabs/blockchain/app/services/node/handlers/v1/private"
	"github.com/ardanlabs/blockchain/app/services/node/handlers/v1/public"
	"github.com/ardanlabs/blockchain/foundation/blockchain/state"
	"github.com/ardanlabs/blockchain/foundation/web"
)

const version = "v1"

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log   *zap.SugaredLogger
	State *state.State
}

// PublicRoutes binds all the version 1 public routes.
func PublicRoutes(app *web.App, cfg Config) {
	pbl := public.Handlers{
		Log:   cfg.Log,
		State: cfg.State,
	}

	app.Handle(http.MethodGet, version, "/sample", pbl.Sample)
	app.Handle(http.MethodGet, version, "/genesis/list", pbl.Genesis)
	app.Handle(http.MethodGet, version, "/accounts/list", pbl.GetAccounts)
	app.Handle(http.MethodGet, version, "/accounts/list/:account", pbl.GetAccounts)
	app.Handle(http.MethodGet, version, "/tx/uncommitted/list", pbl.MemPool)
	app.Handle(http.MethodGet, version, "/tx/uncommitted/list/:account", pbl.MemPool)
	app.Handle(http.MethodPost, version, "/tx/submit", pbl.SubmitWalletTransaction)
	app.Handle(http.MethodPost, version, "/tx/cancel", pbl.Cancel)
}

// PrivateRoutes binds all the version 1 private routes.
func PrivateRoutes(app *web.App, cfg Config) {
	prv := private.Handlers{
		Log: cfg.Log,
	}

	app.Handle(http.MethodGet, version, "/node/sample", prv.Sample)
}
