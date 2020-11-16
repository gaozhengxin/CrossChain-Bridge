package eos

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/anyswap/CrossChain-Bridge/log"
	"github.com/anyswap/CrossChain-Bridge/tokens"
)

const (
	PairID = "eos"
)

// Bridge eth bridge
type Bridge struct {
	*tokens.CrossChainBridgeBase
	*NonceSetterBase
}

// NewCrossChainBridge new bridge
func NewCrossChainBridge(isSrc bool) *Bridge {
	return &Bridge{
		CrossChainBridgeBase: tokens.NewCrossChainBridgeBase(isSrc),
		NonceSetterBase:      NewNonceSetterBase(),
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
	b.SetChainID()
	b.InitLatestBlockNumber()
}

// VerifyChainID verify chain id
func (b *Bridge) VerifyChainID() {
	networkID := strings.ToLower(b.ChainConfig.NetID)
	id, err := strconv.Atoi(networkID)
	if err != nil {
		log.Fatalf("unsupported eos network %v", networkID)
	}

	/*switch byte(id) {
	case filAddress.Mainnet: // byte(0)
	case filAddress.Testnet: // byte(1)
	default:
		log.Fatalf("unsupported eos network %v", networkID)
	}*/

	log.Info("VerifyChainID succeed", "networkID", networkID)
}

func (b *Bridge) SetChainID() {
	/*networkID := strings.ToLower(b.ChainConfig.NetID)
	id, err := strconv.Atoi(networkID)
	if err != nil {
		log.Fatalf("unsupported filecoin network %v", networkID)
	}
	filAddress.CurrentNetwork = byte(id)*/
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
