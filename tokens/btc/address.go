package btc

import (
	"fmt"
	"math/big"

	"github.com/anyswap/CrossChain-Bridge/common"
	"github.com/btcsuite/btcutil"
)

// DecodeAddress decode address
func (b *Bridge) DecodeAddress(addr string) (address btcutil.Address, err error) {
	chainConfig := b.GetChainParams()
	address, err = btcutil.DecodeAddress(addr, chainConfig)
	if err != nil {
		return
	}
	if !address.IsForNet(chainConfig) {
		err = fmt.Errorf("invalid address for net")
		return
	}
	return
}

// NewAddressPubKeyHash encap
func (b *Bridge) NewAddressPubKeyHash(pkData []byte) (*btcutil.AddressPubKeyHash, error) {
	return btcutil.NewAddressPubKeyHash(btcutil.Hash160(pkData), b.GetChainParams())
}

// NewAddressScriptHash encap
func (b *Bridge) NewAddressScriptHash(redeemScript []byte) (*btcutil.AddressScriptHash, error) {
	return btcutil.NewAddressScriptHash(redeemScript, b.GetChainParams())
}

// IsValidAddress check address
func (b *Bridge) IsValidAddress(addr string) bool {
	_, err := b.DecodeAddress(addr)
	return err == nil
}

// IsP2pkhAddress check p2pkh addrss
func (b *Bridge) IsP2pkhAddress(addr string) bool {
	address, err := b.DecodeAddress(addr)
	if err != nil {
		return false
	}
	_, ok := address.(*btcutil.AddressPubKeyHash)
	return ok
}

// IsP2shAddress check p2sh addrss
func (b *Bridge) IsP2shAddress(addr string) bool {
	address, err := b.DecodeAddress(addr)
	if err != nil {
		return false
	}
	_, ok := address.(*btcutil.AddressScriptHash)
	return ok
}

// DecodeWIF decode wif
func DecodeWIF(wif string) (*btcutil.WIF, error) {
	return btcutil.DecodeWIF(wif)
}

// GetBip32InputCode get bip32 input code
func (b *Bridge) GetBip32InputCode(addr string) (string, error) {
	address, err := b.DecodeAddress(addr)
	if err != nil {
		return "", err
	}
	bsData := btcutil.Hash160([]byte(address.EncodeAddress()))
	index := new(big.Int).SetBytes(bsData)
	index.Add(index, common.BigPow(2, 31))
	return fmt.Sprintf("m/%s", index.String()), nil
}

// PublicKeyToAddress public key to address
func (b *Bridge) PublicKeyToAddress(hexPubkey string) (string, error) {
	cpkData, err := b.GetCompressedPublicKey(hexPubkey, false)
	if err != nil {
		return "", err
	}
	address, err := b.NewAddressPubKeyHash(cpkData)
	if err != nil {
		return "", err
	}
	return address.EncodeAddress(), nil
}
