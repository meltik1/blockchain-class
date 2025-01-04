package blockchain

import (
	"errors"
)

type AccountID string

func ToAccountID(hex string) (AccountID, error) {
	acc := AccountID(hex)
	if !acc.IsValid() {
		return "", errors.New("Wrong format of AccountID")
	}

	return acc, nil
}

func (a AccountID) IsValid() bool {
	const addressLength = 20

	if has0xPrefix(a) {
		a = a[2:]
	}

	return len(a) == 2*addressLength && isHex(a)
}

// =============================================================================

// has0xPrefix validates the account starts with a 0x.
func has0xPrefix(a AccountID) bool {
	return len(a) >= 2 && a[0] == '0' && (a[1] == 'x' || a[1] == 'X')
}

// isHex validates whether each byte is valid hexadecimal string.
func isHex(a AccountID) bool {
	if len(a)%2 != 0 {
		return false
	}

	for _, c := range []byte(a) {
		if !isHexCharacter(c) {
			return false
		}
	}

	return true
}

// isHexCharacter returns bool of c being a valid hexadecimal.
func isHexCharacter(c byte) bool {
	return ('0' <= c && c <= '9') || ('a' <= c && c <= 'f') || ('A' <= c && c <= 'F')
}
