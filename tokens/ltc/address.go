package ltc

import (
	"fmt"
	"math/big"

	"github.com/anyswap/CrossChain-Bridge/common"
	"github.com/ltcsuite/ltcutil"
)

// DecodeAddress decode address
func (b *Bridge) DecodeAddress(addr string) (address ltcutil.Address, err error) {
	chainConfig := b.GetChainParams()
	address, err = ltcutil.DecodeAddress(addr, chainConfig)
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
func (b *Bridge) NewAddressPubKeyHash(pkData []byte) (*ltcutil.AddressPubKeyHash, error) {
	return ltcutil.NewAddressPubKeyHash(ltcutil.Hash160(pkData), b.GetChainParams())
}

// NewAddressScriptHash encap
func (b *Bridge) NewAddressScriptHash(redeemScript []byte) (*ltcutil.AddressScriptHash, error) {
	return ltcutil.NewAddressScriptHash(redeemScript, b.GetChainParams())
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
	_, ok := address.(*ltcutil.AddressPubKeyHash)
	return ok
}

// IsP2shAddress check p2sh addrss
func (b *Bridge) IsP2shAddress(addr string) bool {
	address, err := b.DecodeAddress(addr)
	if err != nil {
		return false
	}
	_, ok := address.(*ltcutil.AddressScriptHash)
	return ok
}

// DecodeWIF decode wif
func DecodeWIF(wif string) (*ltcutil.WIF, error) {
	return ltcutil.DecodeWIF(wif)
}

// GetBip32InputCode get bip32 input code
func (b *Bridge) GetBip32InputCode(addr string) (string, error) {
	address, err := b.DecodeAddress(addr)
	if err != nil {
		return "", err
	}
	bsData := ltcutil.Hash160([]byte(address.EncodeAddress()))
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
