package filecoin

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/anyswap/CrossChain-Bridge/log"
	filApi "github.com/filecoin-project/lotus/api"

	"github.com/anyswap/CrossChain-Bridge/common"
	"github.com/anyswap/CrossChain-Bridge/params"
	"github.com/anyswap/CrossChain-Bridge/tokens"
)

var (
	b                   *Bridge
	configFile          string
	chainCfg            *tokens.ChainConfig
	gatewayCfg          *tokens.GatewayConfig
	defServerConfigFile = "filecoinTestConfig.toml"
)

func init() {
	flag.StringVar(&configFile, "configFile", "./filecoinTestConfig.toml", "")
}

func InitTestBridge() {
	flag.Parse()
	if b == nil {
		scfg := LoadConfig(configFile, false)
		chainCfg = scfg.SrcChain
		gatewayCfg = scfg.SrcGateway
		b = NewCrossChainBridge(true)
		b.SetChainAndGateway(chainCfg, gatewayCfg)
	}
}

// TestGetLatestBlockNumberOf
func TestGetLatestBlockNumberOf(t *testing.T) {
	t.Logf("\n\nTestGetLatestBlockNumberOf\n\n")

	InitTestBridge()

	h, err := b.GetLatestBlockNumberOf(gatewayCfg.AuthAPIs[0].Address)
	if err != nil {
		t.Fatalf("TestGetLatestBlockNumberOf fail: %v\n", err)
	}
	fmt.Printf("Latest height: %v\n", h)
}

// TestGetTransactionByHash
func TestGetTransactionByHash(t *testing.T) {
	t.Logf("\n\nTestGetTransactionByHash\n\n")

	InitTestBridge()

	msg, err := b.GetTransactionByHash("bafy2bzaceacfbbpeb35zhrlxfjae7zdzmqf3rag62roxmgjyogsbnhxtcgxd2")
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

	msgLookup, err := b.GetTransactionReceipt("bafy2bzaceacfbbpeb35zhrlxfjae7zdzmqf3rag62roxmgjyogsbnhxtcgxd2")
	if err != nil {
		t.Fatalf("TestGetTransactionReceipt fail: %v\n", err)
	}
	fmt.Printf("\n\nMsgLookup: %+v\n\n", msgLookup)
}

// TestGetAddressNonce
func TestGetAddressNonce(t *testing.T) {
	t.Logf("\n\nTestGetTransactionReceipt\n\n")

	InitTestBridge()

	nonce, err := b.getAddressNonce("f16vqklv5ijzcq4r7cvwesldr3bfdyl6yf4enxnsy")
	if err != nil {
		t.Fatalf("TestGetAddressNonce fail: %v\n", err)
	}
	fmt.Printf("\n\nnonce: %+v\n\n", nonce)
}

// TestGetBalance
func TestGetBalance(t *testing.T) {
	t.Logf("\n\nTestGetBalance\n\n")

	InitTestBridge()

	balance, err := b.GetBalance("f16vqklv5ijzcq4r7cvwesldr3bfdyl6yf4enxnsy")
	if err != nil {
		t.Fatalf("TestGetBalance fail: %v\n", err)
	}
	fmt.Printf("\n\nbalance: %+v\n\n", balance)
}

// TestGetTipsetByNumber
func TestGetTipsetByNumber(t *testing.T) {
	t.Logf("\n\nTestGetTipsetByNumber\n\n")

	InitTestBridge()

	tipset, err := b.GetTipsetByNumber(219767)
	if err != nil {
		t.Fatalf("TestGetTipsetByNumber fail: %v\n", err)
	}
	fmt.Printf("\n\ntipset block number: %v\n\n", len(tipset.Blocks())) // 7
}

// TestGetBlockMessages
func TestGetBlockMessages(t *testing.T) {
	t.Logf("\n\nTestGetBlockMessages\n\n")

	InitTestBridge()

	msgids := b.GetBlockMessages("bafy2bzacedbkvdtzeerkf7kzdsmlp6subtlkip67zt7nnqqwbyf4mc6dgiwi2")
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
		err := b.StartListenMpool(ctx, func(mup filApi.MpoolUpdate) {
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
