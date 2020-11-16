package filecoin

import (
	"fmt"
	"math/big"

	"github.com/anyswap/CrossChain-Bridge/common"
	"github.com/anyswap/CrossChain-Bridge/tokens"

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

// GetBip32InputCode get bip32 input code
func (b *Bridge) GetBip32InputCode(addr string) (string, error) {
	if !tokens.DstBridge.IsValidAddress(addr) {
		return "", fmt.Errorf("invalid address")
	}
	address := common.HexToAddress(addr)
	index := new(big.Int).SetBytes(address.Bytes())
	index.Add(index, common.BigPow(2, 31))
	return fmt.Sprintf("m/%s", index.String()), nil
}

// PublicKeyToAddress public key to address
func (b *Bridge) PublicKeyToAddress(hexPubkey string) (string, error) {
	pkData := common.FromHex(hexPubkey)
	if len(pkData) != 65 {
		return "", fmt.Errorf("wrong length of public key")
	}
	if pkData[0] != 4 {
		return "", fmt.Errorf("wrong public key, shoule be uncompressed")
	}
	address, err := filAddress.NewSecp256k1Address(pkData)
	if err != nil {
		return "", err
	}
	return address.String(), nil
}
