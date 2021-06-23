package pokt

import (
	"fmt"
	"time"

	"github.com/anyswap/CrossChain-Bridge/log"
	"github.com/anyswap/CrossChain-Bridge/tokens"
)

var (
	ErrPOKTTxType = fmt.Errorf("Pokt tx type error")
)

type Bridge struct {
	/*
		CrossChainBridgeBase implements following functions
		SetChainAndGateway(*ChainConfig, *GatewayConfig)
		GetChainConfig() *ChainConfig
		GetGatewayConfig() *GatewayConfig
		GetTokenConfig(pairID string) *TokenConfig
	*/
	*tokens.CrossChainBridgeBase
}

// VerifyChainConfig verify chain config
func (b *Bridge) VerifyChainConfig() {
	// Panic if anything incorrect in config
}

// VerifyTokenConfig verify token config
func (b *Bridge) VerifyTokenConfig(tokenCfg *tokens.TokenConfig) error {
	return nil
}

// InitLatestBlockNumber init latest block number
func (b *Bridge) InitLatestBlockNumber() {
	chainCfg := b.ChainConfig
	gatewayCfg := b.GatewayConfig
	var latest uint64
	var err error
	for {
		latest, err = b.GetLatestBlockNumber()
		if err == nil {
			tokens.SetLatestBlockHeight(latest, b.IsSrc)
			log.Info("get latst block number succeed.", "number", latest, "BlockChain", chainCfg.BlockChain, "NetID", chainCfg.NetID)
			break
		}
		log.Error("get latst block number failed.", "BlockChain", chainCfg.BlockChain, "NetID", chainCfg.NetID, "err", err)
		log.Println("retry query gateway", gatewayCfg.APIAddress)
		time.Sleep(3 * time.Second)
	}
}
