package state

import (
	"emperror.dev/errors"

	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
)

func (s *State) ValidateBlock(block *database.Block) error {
	s.Mu.RLock()
	defer s.Mu.RUnlock()

	lastBlock := s.GetLastBlock()

	if block.Header.Number != lastBlock.Header.Number+1 {
		return errors.New("Invalid index")
	}

	if block.Header.PrevBlockHash != lastBlock.Hash() {
		return errors.New("Invalid previous hash")
	}

	if block.MerkleTree.Verify() != nil && block.MerkleTree.RootHex() != block.Header.TransRoot {
		return errors.New("Transaction hashes and TransRoot Does not match")
	}

	if block.Header.Difficulty < lastBlock.Header.Difficulty {
		return errors.New("Difficulty level can't be lower, that in previous block")
	}

	if !database.IsHashSolved(block.Header.Difficulty, block.Hash()) {
		return errors.New("Hash is not solved")
	}

	if block.Header.TimeStamp < lastBlock.Header.TimeStamp {
		return errors.New("Wrong time")
	}

	return nil
}

func (s *State) UpdateBlock(block *database.Block) error {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	if err := s.ValidateBlock(block); err != nil {
		return errors.Wrap(err, "Error while validating block")
	}

	for _, tx := range block.MerkleTree.Values() {
		err := s.Db.ApplyTransaction(tx, block.Header.BeneficiaryID)
		if err != nil {
			return errors.Wrap(err, "Error while applying transaction")
		}
		s.memPool.Remove(tx)
	}

	s.Db.ApplyMiningReward(block.Header.BeneficiaryID)
	return nil
}
