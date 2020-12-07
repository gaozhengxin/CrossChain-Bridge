package eos

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/anyswap/CrossChain-Bridge/log"
	"github.com/anyswap/CrossChain-Bridge/tokens"
	eosgo "github.com/eoscanada/eos-go"
)

const (
	// PairID is EOS pair id
	PairID = "eos"
	// ChainID is EOS chain id
	ChainID = "aca376f206b8fc25a6ed44dbdc66547c36c6c33e3a119ffbeaef943642f0e906"
)

// Bridge eth bridge
type Bridge struct {
	*tokens.CrossChainBridgeBase
}

// NewCrossChainBridge new bridge
func NewCrossChainBridge(isSrc bool) *Bridge {
	return &Bridge{
		CrossChainBridgeBase: tokens.NewCrossChainBridgeBase(isSrc),
	}
}

// SetChainAndGateway set chain and gateway config
func (b *Bridge) SetChainAndGateway(chainCfg *tokens.ChainConfig, gatewayCfg *tokens.GatewayConfig) {
	b.CrossChainBridgeBase.SetChainAndGateway(chainCfg, gatewayCfg)
	b.VerifyChainID()
	b.Init()
}

// Init init after verify
func (b *Bridge) Init() {
	b.VerifyChainID()
	b.InitLatestBlockNumber()
}

// VerifyChainID verify chain id
func (b *Bridge) VerifyChainID() {
	networkID := strings.ToLower(b.ChainConfig.NetID)

	switch networkID {
	case ChainID:
		log.Info("VerifyChainID succeed", "networkID", networkID)
	default:
		log.Fatalf("unsupported eos network %v", networkID)
	}
}

// VerifyTokenConfig verify token config
func (b *Bridge) VerifyTokenConfig(tokenCfg *tokens.TokenConfig) (err error) {
	if !b.IsValidAddress(tokenCfg.DcrmAddress) {
		return fmt.Errorf("invalid dcrm address: %v", tokenCfg.DcrmAddress)
	}
	if b.IsSrc && !b.IsValidAddress(tokenCfg.DepositAddress) {
		return fmt.Errorf("invalid deposit address: %v", tokenCfg.DepositAddress)
	}
	if *tokenCfg.Decimals != 18 {
		return fmt.Errorf("invalid decimals for EOS: want 18 but have %v", *tokenCfg.Decimals)
	}
	return nil
}

// InitLatestBlockNumber init latest block number
func (b *Bridge) InitLatestBlockNumber() {
	var (
		latest uint64
		err    error
	)

	for {
		latest, err = b.GetLatestBlockNumber()
		if err == nil {
			tokens.SetLatestBlockHeight(latest, b.IsSrc)
			log.Info("get latst block number succeed.", "number", latest, "BlockChain", b.ChainConfig.BlockChain, "NetID", b.ChainConfig.NetID)
			break
		}
		log.Error("get latst block number failed.", "BlockChain", b.ChainConfig.BlockChain, "NetID", b.ChainConfig.NetID, "err", err)
		log.Println("retry query gateway", b.GatewayConfig.APIAddress)
		time.Sleep(3 * time.Second)
	}
}

// GetLatestBlockNumber get latest block number
func (b *Bridge) GetLatestBlockNumber() (uint64, error) {
	cli := b.GetClient()
	resp, err := cli.GetInfo(context.Background())
	if err != nil {
		return 0, err
	}
	return uint64(resp.HeadBlockNum), nil
}

// GetLatestBlockNumberOf get latest block number of apiAddress
func (b *Bridge) GetLatestBlockNumberOf(apiAddress string) (uint64, error) {
	cli := NewClient(apiAddress)
	resp, err := cli.GetInfo(context.Background())
	if err != nil {
		return 0, err
	}
	return uint64(resp.HeadBlockNum), nil
}

// GetBalance returns EOS balance
func (b *Bridge) GetBalance(accountAddress string) (*big.Int, error) {
	cli := b.GetClient()
	asset, err := cli.GetCurrencyBalance(context.Background(), eosgo.AccountName(accountAddress), "EOS", eosgo.AccountName("eosio.token"))
	if err != nil {
		return big.NewInt(0), err
	}
	if len(asset) < 1 {
		return big.NewInt(0), fmt.Errorf("EOS balance not found")
	}
	balance := big.NewInt(int64(asset[0].Amount))
	return balance, nil
}

// GetTokenBalance not supported
func (b *Bridge) GetTokenBalance(tokenType, tokenAddress, accountAddress string) (*big.Int, error) {
	return nil, fmt.Errorf("Tokens not supported on EOS bridge")
}

// GetTokenSupply not supported
func (b *Bridge) GetTokenSupply(tokenType, tokenAddress string) (*big.Int, error) {
	return nil, fmt.Errorf("Tokens not supported on EOS bridge")
}

// VerifyAggregateMsgHash not supported
func (b *Bridge) VerifyAggregateMsgHash(msgHash []string, args *tokens.BuildTxArgs) error {
	return fmt.Errorf("EOS bridge does not support aggregation")
}
