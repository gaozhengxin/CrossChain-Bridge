package filecoin

import (
	"fmt"
	"math/big"
	"testing"

	filTypes "github.com/filecoin-project/lotus/chain/types"

	"github.com/anyswap/CrossChain-Bridge/tokens"
	"github.com/anyswap/CrossChain-Bridge/tools/crypto"
)

var (
	myaddr    = "f16vqklv5ijzcq4r7cvwesldr3bfdyl6yf4enxnsy"
	myprivhex = "4a45ece58f230a25a777cb52cf1726f7955a9400135a33aa50a1d804c0917498"
	//toaddr    = "f3saxx3wzfaedmijfrj7z2ihxni7i3nuztthbsqry5dpvdd2ogxtyay3y35podeadtcjdqzpv4lsbilknivbza"
	toaddr = "f1mptlz5tip3mt4qaeq76y5vfar5q7rxdvtlvnhoy"
)

// TestBuildAndSendTx
func TestBuildAndSendTx(t *testing.T) {
	t.Logf("\n\nTestSignTx\n\n")

	InitTestBridge()

	rawTx := testBuildTx(t)
	sm := testSignTx(rawTx, t)
	testSendTx(sm, t)
}

// testBuildSendTx
func testBuildTx(t *testing.T) interface{} {
	t.Logf("\n\ntestBuildTx\n\n")

	//value := big.NewInt(1234)
	value := big.NewInt(1000000000000000)

	args := &tokens.BuildTxArgs{
		SwapInfo: tokens.SwapInfo{
			PairID:   "filecoin",
			SwapType: tokens.SwapoutType,
			Bind:     toaddr,
		},
		From:  myaddr,
		Value: value,
		Extra: &tokens.AllExtras{
			FilExtra: &tokens.FilExtraArgs{},
		},
	}

	msg, err := tb.BuildRawTransaction(args)
	if err != nil {
		t.Fatalf("testBuildTx fail: %v\n", err)
	}

	fmt.Printf("\n\nRaw tx: %+v\n\n", msg)
	return msg
}

// testSignTx
func testSignTx(rawTx interface{}, t *testing.T) (sm interface{}) {
	if rawTx == nil {
		return
	}

	t.Logf("\n\ntestSignTx\n\n")

	priv, _ := crypto.HexToECDSA(myprivhex)

	sm, txhash, err := tb.SignTransactionWithPrivateKey(rawTx, priv)
	if err != nil {
		t.Fatalf("\n\ntestSignTx fail: %v\n\n", err)
	}

	smJSON, _ := sm.(*filTypes.SignedMessage).MarshalJSON()
	fmt.Printf("\n\nSigned msg: %+v\ntxhash: %v\n\n", string(smJSON), txhash)

	return sm
}

// testSendTx
func testSendTx(sm interface{}, t *testing.T) {
	t.Logf("\n\ntestSendTx\n\n")

	txhash, err := tb.SendTransaction(sm)

	if err != nil {
		t.Fatalf("\n\ntestSendTx fail: %v\n\n", err)
	}

	fmt.Printf("\n\nTx cid: %v\n\n", txhash)
}
