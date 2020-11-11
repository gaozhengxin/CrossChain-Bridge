package filecoin

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	filApi "github.com/filecoin-project/lotus/api"
)

// TestGetLatestBlockNumberOf
func TestGetLatestBlockNumberOf(t *testing.T) {
	t.Logf("\n\nTestGetLatestBlockNumberOf\n\n")

	InitTestBridge()

	h, err := tb.GetLatestBlockNumberOf(gatewayCfg.AuthAPIs[0].Address)
	if err != nil {
		t.Fatalf("TestGetLatestBlockNumberOf fail: %v\n", err)
	}
	fmt.Printf("Latest height: %v\n", h)
}

// TestGetTransactionByHash
func TestGetTransactionByHash(t *testing.T) {
	t.Logf("\n\nTestGetTransactionByHash\n\n")

	InitTestBridge()

	msg, err := tb.GetTransactionByHash("bafy2bzaceacfbbpeb35zhrlxfjae7zdzmqf3rag62roxmgjyogsbnhxtcgxd2")
	if err != nil {
		t.Fatalf("TestGetTransactionByHash fail: %v\n", err)
	}
	msgJSON, _ := msg.MarshalJSON()
	fmt.Printf("\n\nmessage: %+v\n\n", string(msgJSON))
}

// TODO
// infura: websocket connection closed
// TestGetTransactionReceipt
func TestGetTransactionReceipt(t *testing.T) {
	t.Logf("\n\nTestGetTransactionReceipt\n\n")

	InitTestBridge()

	msgLookup, err := tb.GetTransactionReceipt("bafy2bzaceacfbbpeb35zhrlxfjae7zdzmqf3rag62roxmgjyogsbnhxtcgxd2")
	if err != nil {
		t.Fatalf("TestGetTransactionReceipt fail: %v\n", err)
	}
	fmt.Printf("\n\nMsgLookup: %+v\n\n", msgLookup)
}

// TestGetAddressNonce
func TestGetAddressNonce(t *testing.T) {
	t.Logf("\n\nTestGetTransactionReceipt\n\n")

	InitTestBridge()

	nonce, err := tb.getAddressNonce("f16vqklv5ijzcq4r7cvwesldr3bfdyl6yf4enxnsy")
	if err != nil {
		t.Fatalf("TestGetAddressNonce fail: %v\n", err)
	}
	fmt.Printf("\n\nnonce: %+v\n\n", nonce)
}

// TestGetBalance
func TestGetBalance(t *testing.T) {
	t.Logf("\n\nTestGetBalance\n\n")

	InitTestBridge()

	balance, err := tb.GetBalance("f16vqklv5ijzcq4r7cvwesldr3bfdyl6yf4enxnsy")
	if err != nil {
		t.Fatalf("TestGetBalance fail: %v\n", err)
	}
	fmt.Printf("\n\nbalance: %+v\n\n", balance)
}

// TestGetTipsetByNumber
func TestGetTipsetByNumber(t *testing.T) {
	t.Logf("\n\nTestGetTipsetByNumber\n\n")

	InitTestBridge()

	tipset, err := tb.GetTipsetByNumber(219767)
	if err != nil {
		t.Fatalf("TestGetTipsetByNumber fail: %v\n", err)
	}
	fmt.Printf("\n\ntipset block number: %v\n\n", len(tipset.Blocks())) // 7
}

// TestGetBlockMessages
func TestGetBlockMessages(t *testing.T) {
	t.Logf("\n\nTestGetBlockMessages\n\n")

	InitTestBridge()

	msgids := tb.GetBlockMessages("bafy2bzacedbkvdtzeerkf7kzdsmlp6subtlkip67zt7nnqqwbyf4mc6dgiwi2")
	fmt.Printf("\n\nmsgids: %+v\n\n", len(msgids)) // 230
}

// TestStartListenMpool
func TestStartListenMpool(t *testing.T) {
	t.Logf("\n\nTestStartListenMpool\n\n")

	InitTestBridge()

	ctx := context.Background()

	ok := make(chan bool, 1)
	cnt := 0

	go func() {
		err := tb.StartListenMpool(ctx, func(mup filApi.MpoolUpdate) {
			if cnt > 3 {
				ok <- true
				return
			}
			mupJSON, _ := json.Marshal(mup)
			fmt.Printf("mpool update: %+v\n", string(mupJSON))
			cnt++
		})
		if err != nil {
			t.Errorf("TestStartListenMpool error: %v\n", err)
		}
	}()

	timer := time.NewTimer(time.Second * 60)

	select {
	case <-ok:
		ctx.Done()
		time.Sleep(time.Millisecond * 100)
		return
	case <-timer.C:
		fmt.Println("time out")
		ctx.Done()
		time.Sleep(time.Millisecond * 100)
		return
	}
}
