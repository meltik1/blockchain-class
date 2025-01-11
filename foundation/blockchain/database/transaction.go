package database

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"emperror.dev/errors"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ardanlabs/blockchain/foundation/blockchain/signature"
)

type Tx struct {
	FromID  AccountID `json:"from_id"`
	ToID    AccountID `json:"to_id"`
	Value   uint64    `json:"value"`
	Tip     uint64    `json:"tip"`
	ChainId uint16    `json:"chain_id"`
	Data    []byte    `json:"data"`
	Nonce   uint64    `json:"nonce"` // Сколько транзакций уже соверщил отправитель
}

func NewTx(fromID AccountID, toID AccountID, value uint64, tip uint64, chainId uint16, data []byte, nonce uint64) (Tx, error) {
	if !fromID.IsValid() {
		return Tx{}, errors.New("Invalid fromID account")
	}

	if !toID.IsValid() {
		return Tx{}, errors.New("Invalid toID account")
	}

	return Tx{FromID: fromID,
		ToID:    toID,
		Value:   value,
		Tip:     tip,
		ChainId: chainId,
		Data:    data,
		Nonce:   nonce}, nil
}

func stamp(tx Tx) ([]byte, error) {
	marshal, err := json.Marshal(tx)
	if err != nil {
		return nil, errors.Wrap(err, "Error while marshalling TX")
	}

	salt := fmt.Sprintf("Ardan blockchain salt %d", len(marshal))
	saltInBytes, err := json.Marshal(salt)
	if err != nil {
		return nil, errors.Wrap(err, "error while marsahalling salt")
	}

	txHashWithSalt := crypto.Keccak256(marshal, saltInBytes)

	return txHashWithSalt, nil
}

func (tx Tx) Sign(key *ecdsa.PrivateKey) (SignedTx, error) {
	txHashWithSalt, err := stamp(tx)
	if err != nil {
		return SignedTx{}, errors.Wrap(err, "stamp:")
	}

	sign, err := crypto.Sign(txHashWithSalt, key)
	if err != nil {
		return SignedTx{}, errors.Wrap(err, "error while Siging Tx")
	}

	signed := signature.Signature(sign)
	v, r, s := signed.ToVrs()

	signedTx := SignedTx{
		tx,
		v,
		r,
		s,
	}

	return signedTx, nil
}

type SignedTx struct {
	Tx
	V *big.Int
	R *big.Int
	S *big.Int
}

func (tx SignedTx) IsValid() error {
	if !tx.FromID.IsValid() {
		return errors.New("Invalid fromID account")
	}

	if !tx.ToID.IsValid() {
		return errors.New("Invalid toID account")
	}

	if tx.Value == 0 {
		return errors.New("Value must be greater than 0")
	}

	if tx.FromID == tx.ToID {
		return errors.New("FromID and ToID must be different")
	}

	if !signature.ValidateSignatureValues(tx.V, tx.R, tx.S) {
		return errors.New("Invalid signature values")
	}

	address, err := tx.fromSignToAddress()

	if !(address == tx.FromID) || err != nil {
		return errors.Wrap(err, "Invalid signature")
	}

	return nil
}

func (tx SignedTx) fromSignToAddress() (AccountID, error) {
	signat := signature.ToSignatureBytes(tx.V, tx.R, tx.S)

	hashedMassage, err := stamp(tx.Tx)
	if err != nil {
		return "", errors.Wrap(err, "error while hashing message")
	}

	pub, err := crypto.SigToPub(hashedMassage, signat)
	if err != nil {
		return "", errors.Wrap(err, "error while converting sign to public key")
	}

	return AccountID(crypto.PubkeyToAddress(*pub).String()), nil
}

// SignatureString returns the signature as a string.
func (tx SignedTx) SignatureString() string {
	return signature.SignatureString(tx.V, tx.R, tx.S)
}

// String implements the Stringer interface for logging.
func (tx SignedTx) String() string {
	return fmt.Sprintf("%s:%d", tx.FromID, tx.Nonce)
}

// BlockTx represents the transaction as it's recorded inside a block. This
// includes a timestamp and gas fees.
type BlockTx struct {
	SignedTx
	TimeStamp uint64 `json:"timestamp"` // Ethereum: The time the transaction was received.
	GasPrice  uint64 `json:"gas_price"` // Ethereum: The price of one unit of gas to be paid for fees.
	GasUnits  uint64 `json:"gas_units"` // Ethereum: The number of units of gas used for this transaction.
}

// NewBlockTx creates a new BlockTx value.
func NewBlockTx(tx SignedTx, gasPrice, gasUnits uint64) BlockTx {
	return BlockTx{
		SignedTx:  tx,
		TimeStamp: uint64(time.Now().UTC().UnixMilli()),
		GasPrice:  gasPrice,
		GasUnits:  gasUnits,
	}
}

// Hash implements the merkle Hashable interface for providing a hash
// of a block transaction.
func (tx BlockTx) Hash() ([]byte, error) {
	str := signature.Hash(tx)

	// Need to remove the 0x prefix from the hash.
	return hex.DecodeString(str[2:])
}

// Equals implements the merkle Hashable interface for providing an equality
// check between two block transactions. If the nonce and signatures are the
// same, the two blocks are the same.
func (tx BlockTx) Equals(otherTx BlockTx) bool {
	txSig := signature.ToSignatureBytes(tx.V, tx.R, tx.S)
	otherTxSig := signature.ToSignatureBytes(otherTx.V, otherTx.R, otherTx.S)

	return tx.Nonce == otherTx.Nonce && bytes.Equal(txSig, otherTxSig)
}
