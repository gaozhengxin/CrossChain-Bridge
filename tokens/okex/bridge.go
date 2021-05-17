package okex

import (
	"github.com/anyswap/CrossChain-Bridge/tokens/eth"
)

// Bridge etc bridge inherit from eth bridge
type Bridge struct {
	*eth.Bridge
}

// NewCrossChainBridge new etc bridge
func NewCrossChainBridge(isSrc bool) *Bridge {
	return &Bridge{Bridge: eth.NewCrossChainBridge(isSrc)}
}