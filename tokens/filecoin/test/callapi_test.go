package test

import (
	"fmt"
	"testing"

	"github.com/anyswap/CrossChain-Bridge/tokens"
	"github.com/anyswap/CrossChain-Bridge/tokens/filecoin"
)

var (
	confirmations uint64 = 2
	initialHeight uint64 = 219000
)

func TestGetTransactionByHash(t *testing.T) {
	t.Logf("\n\nTestGetTransactionByHash\n\n")
	chainCfg := &tokens.ChainConfig{
		BlockChain:    "Filecoin",
		NetID:         "0", // mainnet
		Confirmations: &confirmations,
		InitialHeight: &initialHeight,
		EnableScan:    true,
	}
	gatewayCfg := &tokens.GatewayConfig{
		APIAddress: []string{"wss://filecoin.infura.io"},
		AuthAPIs: []tokens.AuthAPI{
			tokens.AuthAPI{
				AuthType:  "Basic",
				AuthToken: "MWp1REFvZ0VJM1E0b0pRNnI4dnhBUlNuZm5MOmIyNzI5OGNkNmIzZDFiMzJlYWQ0MDQ0N2FhNDI3YTg0",
				Address:   "wss://filecoin.infura.io",
			},
		},
	}
	b := filecoin.NewCrossChainBridge(true)
	b.SetChainAndGateway(chainCfg, gatewayCfg)
	msg, err := b.GetTransactionByHash("bafy2bzaceacfbbpeb35zhrlxfjae7zdzmqf3rag62roxmgjyogsbnhxtcgxd2")
	if err != nil {
		t.Fatalf("GetTransactionByHash fail: %v\n", err)
	}
	msgJSON, _ := msg.MarshalJSON()
	fmt.Printf("message: %+v\n", string(msgJSON))
}
