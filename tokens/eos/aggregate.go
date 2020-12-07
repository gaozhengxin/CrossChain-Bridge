package eos

import (
	"github.com/anyswap/CrossChain-Bridge/tokens"
)

// StartAggregateJob is not implemented
func (b *Bridge) StartAggregateJob() {
	return
}

// VerifyAggregateMsgHash is not implemented
func VerifyAggregateMsgHash(msgHash []string, args *tokens.BuildTxArgs) error {
	return nil
}
