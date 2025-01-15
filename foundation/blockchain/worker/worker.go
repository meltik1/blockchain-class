package worker

import (
	"context"
	"sync"

	"emperror.dev/errors"

	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
	"github.com/ardanlabs/blockchain/foundation/blockchain/state"
)

type Worker struct {
	shutDown     chan struct{}
	startMining  chan bool
	cancelMining chan bool

	s  *state.State
	ev state.EventHandler
}

func newWorker(s *state.State, handler state.EventHandler) *Worker {
	return &Worker{
		startMining:  make(chan bool, 1),
		cancelMining: make(chan bool, 1),
		s:            s,
		ev:           handler,
	}
}

func Init(s *state.State, ev state.EventHandler) {
	worker := newWorker(s, ev)
	s.Worker = worker

	hasStarted := make(chan bool)
	go func() {
		hasStarted <- true
		worker.Run()
	}()

	<-hasStarted
	worker.ev("Worker started")
	return

}

func (w *Worker) Run() {
	for {
		select {
		case <-w.startMining:
			w.mine()
		case <-w.shutDown:
			w.ev("Worker: Shutdown requested")
			return
		}
	}
}

func (w *Worker) mine() {
	if w.s.MempoolLength() == 0 {
		w.ev("No transactions in mempool")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	args := database.POWArgs{
		BeneficiaryID: w.s.BeneficiaryID,
		Difficulty:    w.s.GetGenesis().Difficulty,
		MiningReward:  uint64(w.s.GetGenesis().MiningReward),
		PrevBlock:     w.s.GetLastBlock(),
		StateRoot:     w.s.GetStateRoot(),
		Trans:         w.s.Mempool(),
		EvHandler:     w.ev,
	}

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		block, err := database.POW(ctx, args)
		if err != nil {
			w.ev(errors.Wrap(err, "Error while mining").Error())
		}

		w.ev("!!!! We ve mined block: %s !!!", block.Hash())
	}()

	go func() {
		defer func() {
			wg.Done()
			cancel()
		}()
		select {
		case <-w.cancelMining:
			w.ev("worker: runMiningOperation: MINING: CANCEL: requested")
		case <-ctx.Done():
		}
	}()

	wg.Wait()
}

func (w *Worker) Shutdown() {
	w.shutDown <- struct{}{}
}

func (w *Worker) Sync() {
	//TODO implement me
	panic("implement me")
}

func (w *Worker) SignalStartMining() {
	select {
	case w.startMining <- true:
		w.ev("Start mining signal sent")
	default:
		w.ev("Start mining signal already sent")

	}
	return
}

func (w *Worker) SignalCancelMining() {
	w.cancelMining <- true
}

func (w *Worker) SignalShareTx(blockTx database.BlockTx) {
	//TODO implement me
	panic("implement me")
}
