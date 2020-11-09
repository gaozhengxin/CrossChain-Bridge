package filecoin

import (
	filAddress "github.com/filecoin-project/go-address"
)

// IsValidAddress check address
func (b *Bridge) IsValidAddress(address string) bool {
	addr, err := filAddress.NewFromString(address)
	if err != nil {
		return false
	}
	var ok bool
	switch addr.Protocol() {
	case filAddress.ID:
		ok = true
	case filAddress.SECP256K1:
		ok = true
	case filAddress.Actor:
		ok = false
	case filAddress.BLS:
		ok = false
	case filAddress.Unknown:
		ok = false
	default:
		ok = false
	}
	return ok
}
