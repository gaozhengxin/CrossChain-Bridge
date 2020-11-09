package test

import (
	"encoding/json"
	"flag"
	"fmt"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/anyswap/CrossChain-Bridge/log"

	"github.com/anyswap/CrossChain-Bridge/common"
	"github.com/anyswap/CrossChain-Bridge/params"
	"github.com/anyswap/CrossChain-Bridge/tokens"
	"github.com/anyswap/CrossChain-Bridge/tokens/filecoin"
)

var (
	configFile          string
	chainCfg            *tokens.ChainConfig
	gatewayCfg          *tokens.GatewayConfig
	defServerConfigFile = "filecoinTestConfig.toml"
)

func init() {
	flag.StringVar(&configFile, "configFile", "./filecoinTestConfig.toml", "")
}

func LoadTestConfig() {
	flag.Parse()
	scfg := LoadConfig(configFile, false)
	chainCfg = scfg.SrcChain
	gatewayCfg = scfg.SrcGateway
}

func TestGetTransactionByHash(t *testing.T) {
	t.Logf("\n\nTestGetTransactionByHash\n\n")

	LoadTestConfig()

	b := filecoin.NewCrossChainBridge(true)

	b.SetChainAndGateway(chainCfg, gatewayCfg)

	msg, err := b.GetTransactionByHash("bafy2bzaceacfbbpeb35zhrlxfjae7zdzmqf3rag62roxmgjyogsbnhxtcgxd2")
	if err != nil {
		t.Fatalf("GetTransactionByHash fail: %v\n", err)
	}
	msgJSON, _ := msg.MarshalJSON()
	fmt.Printf("message: %+v\n", string(msgJSON))
}

// LoadConfig load config
func LoadConfig(configFile string, isServer bool) *params.ServerConfig {
	if configFile == "" {
		// find config file in the execute directory (default).
		dir, err := common.ExecuteDir()
		if err != nil {
			log.Fatalf("LoadConfig error (get ExecuteDir): %v", err)
		}
		configFile = common.AbsolutePath(dir, defServerConfigFile)
	}
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
	log.Println("LoadConfig finished.", string(bs))
	return config
}
