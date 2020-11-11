package filecoin

import (
	"encoding/json"

	"github.com/BurntSushi/toml"

	"github.com/anyswap/CrossChain-Bridge/common"
	"github.com/anyswap/CrossChain-Bridge/log"
	"github.com/anyswap/CrossChain-Bridge/params"
	"github.com/anyswap/CrossChain-Bridge/tokens"
)

var (
	tb              *Bridge
	configFile      = "filecoinTestConfig.toml"
	chainCfg        *tokens.ChainConfig
	gatewayCfg      *tokens.GatewayConfig
	tokenConfigFile = "tokenConfigs"
)

func InitTestBridge() {
	if tb == nil {
		scfg := LoadConfig(configFile, false)
		chainCfg = scfg.SrcChain
		gatewayCfg = scfg.SrcGateway
		tb = NewCrossChainBridge(true)
		tb.SetChainAndGateway(chainCfg, gatewayCfg)
		LoadTokenConfig(tokenConfigFile)
	}
}

// LoadConfig load config
func LoadConfig(configFile string, isServer bool) *params.ServerConfig {
	log.Println("Config file is", configFile)
	if !common.FileExist(configFile) {
		log.Fatalf("LoadConfig error: config file %v not exist", configFile)
	}
	config := &params.ServerConfig{}
	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		log.Fatalf("LoadConfig error (toml DecodeFile): %v", err)
	}

	var bs []byte
	if log.JSONFormat {
		bs, _ = json.Marshal(config)
	} else {
		bs, _ = json.MarshalIndent(config, "", "  ")
	}
	params.SetConfig(config)
	log.Println("LoadConfig finished.", string(bs))
	return config
}

// LoadTokenConfig
func LoadTokenConfig(tokenConfigFile string) {
	log.Println("Token config file is", configFile)
	if !common.FileExist(tokenConfigFile) {
		log.Fatalf("LoadTokenConfig error: config file %v not exist", tokenConfigFile)
	}
	cfg, err := tokens.LoadTokenPairsConfigInDir(tokenConfigFile, false)
	if err != nil {
		log.Fatalf("LoadTokenConfig: %v", err)
	}
	tokens.SetTokenPairsConfig(cfg, false)
}
