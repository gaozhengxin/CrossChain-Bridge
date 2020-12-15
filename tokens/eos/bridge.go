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

	mainnet     = "mainnet" //"aca376f206b8fc25a6ed44dbdc66547c36c6c33e3a119ffbeaef943642f0e906"
	jungletest  = "jungle"  //"2a02a0053e5a8cf73a56ba0fda11e4d92e0238a4a2aa74fccf46d5a910746840"
	jungle2test = "jungle2" //"e70aaab8997e1dfce58fbfac80cbbb8fecec7b99cf982a9444273cbc64c41473"
	kylintest   = "kylin"   //"5fff1dae8dc8e2fc4d5b23b2c7665c97f9e9d8edf2b6485a86ba311c25639191"
	netCustom   = "custom"
)

// ChainID eos chain id
var ChainID = "aca376f206b8fc25a6ed44dbdc66547c36c6c33e3a119ffbeaef943642f0e906"

// Bridge eth bridge
type Bridge struct {
	*tokens.CrossChainBridgeBase
}

// BridgeInstance is eos bridge instance
var BridgeInstance *Bridge

// NewCrossChainBridge new bridge
func NewCrossChainBridge(isSrc bool) *Bridge {
	BridgeInstance = &Bridge{
		CrossChainBridgeBase: tokens.NewCrossChainBridgeBase(isSrc),
	}
	return BridgeInstance
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
	case mainnet, jungletest, jungle2test, kylintest:
	case netCustom:
	default:
		log.Fatalf("unsupported eos network: %v", b.ChainConfig.NetID)
	}

	for {
		// call NetworkID instead of ChainID as ChainID may return 0x0 wrongly
		chainID, err := b.GetChainID()
		if err == nil {
			ChainID = chainID
			break
		}
		log.Errorf("can not get gateway chainID. %v", err)
		log.Println("retry query gateway", b.GatewayConfig.APIAddress)
		time.Sleep(3 * time.Second)
	}

	panicMismatchChainID := func() {
		log.Fatalf("gateway chainID %v is not %v", ChainID, b.ChainConfig.NetID)
	}

	switch networkID {
	case mainnet:
		if ChainID != "aca376f206b8fc25a6ed44dbdc66547c36c6c33e3a119ffbeaef943642f0e906" {
			panicMismatchChainID()
		}
		log.Info("VerifyChainID succeed", "chain", "mainnet", "networkID", networkID)
	case jungletest:
		if ChainID != "2a02a0053e5a8cf73a56ba0fda11e4d92e0238a4a2aa74fccf46d5a910746840" {
			panicMismatchChainID()
		}
		log.Info("VerifyChainID succeed", "chain", "jungle test", "networkID", networkID)
	case jungle2test:
		if ChainID != "e70aaab8997e1dfce58fbfac80cbbb8fecec7b99cf982a9444273cbc64c41473" {
			panicMismatchChainID()
		}
		log.Info("VerifyChainID succeed", "chain", "jungle2 test", "networkID", networkID)
	case kylintest:
		if ChainID != "5fff1dae8dc8e2fc4d5b23b2c7665c97f9e9d8edf2b6485a86ba311c25639191" {
			panicMismatchChainID()
		}
		log.Info("VerifyChainID succeed", "chain", "kylin", "networkID", networkID)
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

// GetChainID get eos chain id
func (b *Bridge) GetChainID() (string, error) {
	cli := b.GetClient()
	resp, err := cli.GetInfo(context.Background())
	if err != nil {
		return "", err
	}
	return resp.ChainID.String(), nil
}

// GetIrreversible get eos last irriversible block number
func (b *Bridge) GetIrreversible() (uint64, error) {
	cli := b.GetClient()
	resp, err := cli.GetInfo(context.Background())
	if err != nil {
		return 0, err
	}
	return uint64(resp.LastIrreversibleBlockNum), nil
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
