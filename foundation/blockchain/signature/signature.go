package signature

import (
	"crypto/sha256"
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

const ArdanID = 27

// ZeroHash represents a hash code of zeros.
const ZeroHash string = "0x0000000000000000000000000000000000000000000000000000000000000000"

type Signature []byte

func (s Signature) ToVrs() (*big.Int, *big.Int, *big.Int) {
	R := new(big.Int).SetBytes(s[0:32])
	S := new(big.Int).SetBytes(s[32:64])
	V := new(big.Int).SetBytes([]byte{s[64] + ArdanID})

	return R, S, V
}

func FromVRSToSignature(v, r, s *big.Int) Signature {
	signatureBytes := make([]byte, 64)

	signatureBytes = append(r.Bytes(), s.Bytes()...)
	signatureBytes = append(signatureBytes, byte(v.Int64()-ArdanID))

	return signatureBytes
}

// SignatureString returns the signature as a string.
func SignatureString(v, r, s *big.Int) string {
	return hexutil.Encode(ToSignatureBytesWithArdanID(v, r, s))
}

// ToSignatureBytes converts the r, s, v values into a slice of bytes
// with the removal of the ardanID.
func ToSignatureBytes(v, r, s *big.Int) []byte {
	sig := make([]byte, crypto.SignatureLength)

	rBytes := make([]byte, 32)
	r.FillBytes(rBytes)
	copy(sig, rBytes)

	sBytes := make([]byte, 32)
	s.FillBytes(sBytes)
	copy(sig[32:], sBytes)

	sig[64] = byte(v.Uint64() - ArdanID)

	return sig
}

// ToSignatureBytesWithArdanID converts the r, s, v values into a slice of bytes
// keeping the Ardan id.
func ToSignatureBytesWithArdanID(v, r, s *big.Int) []byte {
	sig := ToSignatureBytes(v, r, s)
	sig[64] = byte(v.Uint64())

	return sig
}

// Hash returns a unique string for the value.
func Hash(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return ZeroHash
	}

	hash := sha256.Sum256(data)
	return hexutil.Encode(hash[:])
}

func ValidateSignatureValues(v, r, s *big.Int) bool {
	if !(v.Int64() == ArdanID || v.Int64() == ArdanID+1) {
		return false
	}

	if !crypto.ValidateSignatureValues(byte(v.Uint64()-ArdanID), r, s, false) {
		return false
	}

	return true
}
