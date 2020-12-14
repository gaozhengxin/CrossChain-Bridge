package eos

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"

	"github.com/eoscanada/eos-go/btcsuite/btcd/btcec"
	"github.com/eoscanada/eos-go/btcsuite/btcutil"
	"github.com/eoscanada/eos-go/ecc"
)

// HexToPubKey converts pubkey hex to eos pubkey
func HexToPubKey(pubKeyHex string) (ecc.PublicKey, error) {
	fmt.Printf("hex is %v\nlen(hex) is %v\n\n", pubKeyHex, len(pubKeyHex))
	if pubKeyHex[:2] == "0x" || pubKeyHex[:2] == "0X" {
		pubKeyHex = pubKeyHex[2:]
	}
	if len(pubKeyHex) == 130 {
		uBytes, err := hex.DecodeString(pubKeyHex)
		if err != nil {
			return ecc.PublicKey{}, err
		}
		pubkey, err := btcec.ParsePubKey(uBytes, btcec.S256())
		if err != nil {
			return ecc.PublicKey{}, err
		}
		pubkeyBytes := pubkey.SerializeCompressed()
		pubkeyBytes = append([]byte{0}, pubkeyBytes...) // byte{0} 表示 curve K1, byte{1} 表示 curve R1
		return ecc.NewPublicKeyFromData(pubkeyBytes)
	}

	if len(pubKeyHex) == 66 {
		pubkeyBytes, _ := hex.DecodeString(pubKeyHex)
		pubkeyBytes = append([]byte{0}, pubkeyBytes...)
		return ecc.NewPublicKeyFromData(pubkeyBytes)
	}

	return ecc.PublicKey{}, fmt.Errorf("unexpected public key length  %v", len(pubKeyHex))
}

// WifToECDSA converts eos wif to ecdsa private key
func WifToECDSA(wif string) (*ecdsa.PrivateKey, error) {
	w, err := btcutil.DecodeWIF(wif)
	if err != nil {
		return nil, err
	}
	if w != nil && w.PrivKey != nil {
		privateKey := w.PrivKey.ToECDSA()
		return privateKey, nil
	}
	return nil, fmt.Errorf("Decode wif unknown error")
}
