package eos

import (
	"context"
	"regexp"
)

var (
	// EOSNameReg is regular expression for EOS account name
	EOSNameReg = "^([1-5a-z]{12})$"
	// EOSUserKeyReg is regular expression for EOS user key
	EOSUserKeyReg = "^(d[1-5a-z]{32,33})$"
)

// IsValidAddress check address
// Check EOS user name is valid and already exists on EOS
func (b *Bridge) IsValidAddress(address string) bool {
	if !b.IsValidName(address) {
		return false
	}

	account, err := GetClient().GetAccount(context.Background(), address)
	if account != nil && err == nil && string(account.AccountName) == address {
		return true
	}
	return false
}

// IsValidName check if EOS account name is valid
// Only to check address is a well-formed EOS username
// NOT ensuring address is registered on EOS
func (b *Bridge) IsValidName(name string) bool {
	match, _ := regexp.MatchString(EOSNameReg, name)
	return match
}

// IsValidUserKey check if EOS user key is valid
func (b *Bridge) IsValidUserKey(userkey string) bool {
	match, _ := regexp.MatchString(EOSUserKeyReg, userkey)
	return match
}

// GetBip32InputCode get bip32 input code
func (b *Bridge) GetBip32InputCode(addr string) (string, error) {
	return "", nil
}

// PublicKeyToAddress public key to address
func (b *Bridge) PublicKeyToAddress(hexPubkey string) (string, error) {
	return "", nil
}
