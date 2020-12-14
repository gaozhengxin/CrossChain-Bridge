package eos

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"strconv"
	"testing"

	"github.com/anyswap/CrossChain-Bridge/tokens"
	eosgo "github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/btcsuite/btcd/btcec"
	"github.com/eoscanada/eos-go/btcsuite/btcutil"
	"github.com/eoscanada/eos-go/ecc"
	"github.com/eoscanada/eos-go/token"
)

var b *Bridge

func initKylin() {
	b = NewCrossChainBridge(true)
	chaincfg := &tokens.ChainConfig{
		BlockChain:    "EOS",
		NetID:         "kylin",
		Confirmations: new(uint64),
		InitialHeight: new(uint64),
		EnableScan:    true,
	}
	*chaincfg.Confirmations = uint64(2)
	*chaincfg.InitialHeight = uint64(139667837)
	gatewaycfg := &tokens.GatewayConfig{
		APIAddress: []string{"https://api.kylin.alohaeos.com:443", "https://api-kylin.eoslaomao.com:443"},
	}
	b.SetChainAndGateway(chaincfg, gatewaycfg)
}

func initMainnet() {
	b = NewCrossChainBridge(true)
	chaincfg := &tokens.ChainConfig{
		BlockChain:    "EOS",
		NetID:         "mainnet",
		Confirmations: new(uint64),
		InitialHeight: new(uint64),
		EnableScan:    true,
	}
	*chaincfg.Confirmations = uint64(2)
	*chaincfg.InitialHeight = uint64(152750000)
	gatewaycfg := &tokens.GatewayConfig{
		APIAddress: []string{"https://openapi.eos.ren:443", "https://api.eoslaomao.com:443", "https://api.eossweden.se:443", "https://eos.greymass.com:443", "https://nodes.get-scatter.com:443"},
	}
	b.SetChainAndGateway(chaincfg, gatewaycfg)
}

func checkError(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func TestConvertKey(t *testing.T) {
	t.Logf("TestConvertKey")

	t.Logf("\n==== 1. Pubkey hex to Eos pubkey ====\n")
	pubkeyHex := "0489ff22501cfeea513cbbb6456af044b8def7417c27d9b2ac83afde660425a8634f05adb77528eea6df15eb26be151045fed984097e8ae16f89ed81a4018dd96a"
	pubkey, err := HexToPubKey(pubkeyHex)
	checkError(t, err)
	t.Logf("Eos pubkey: %v\n", pubkey) // EOS7bkj2uy4YJbZmh2PqGJsu6B9GrpxifYgRhw8U7av6ymNDM9YLx

	t.Logf("\n==== 2. Privkey hex to Eos pubkey ====\n")
	privkeyHex := "02d1532a7637ed632fc299a2699d16589850ee92cf5ecd945ed7691641e79ef2"
	pkData, err := hex.DecodeString(privkeyHex)
	checkError(t, err)

	privkey, _ := btcec.PrivKeyFromBytes(btcec.S256(), pkData)

	wif, err := btcutil.NewWIF(privkey, 0x80, false)
	checkError(t, err)
	t.Logf("Eos wif: %v\n", wif) // 5HqXZA63VbUfxr7raBsDf6ZTJCdbdnyCF92vmm6fXtgNUFgnrKX

	t.Logf("\n==== 3. Read from wif ====\n")
	eosprivkey, err := ecc.NewPrivateKey(wif.String())
	checkError(t, err)
	t.Logf("Eos private: %v\n", eosprivkey) // 5HqXZA63VbUfxr7raBsDf6ZTJCdbdnyCF92vmm6fXtgNUFgnrKX

	t.Logf("\n==== 4. Sign ====\n")
	hash, _ := hex.DecodeString("8999ad62a7e1f159f060fbcaf8e462054effe1246596c1e6d6b6ca0f24a2b5d4")
	sig, err := eosprivkey.Sign(hash)
	checkError(t, err)
	t.Logf("signature: %v\n", sig) // SIG_K1_K5DqZMn4ndnBtVsygbdYepMosZCwPGY43ZRjoJV8Fs3GV7bUByhUnnN3QKUCmkXSaSQfrM8Qc3wapnsrvnbr9drSv5GFnj
}

func TestGetInfo(t *testing.T) {
	initKylin()
	t.Logf("TestGetBasic\n")
	chainid, err := b.GetChainID()
	checkError(t, err)
	t.Logf("chain id: %v\n", chainid)
	irr, err := b.GetIrreversible()
	checkError(t, err)
	t.Logf("last irreversible block number: %v\n", irr)
	blk, err := b.GetLatestBlockNumber()
	checkError(t, err)
	t.Logf("last block number: %v\n", blk)
	blk, err = b.GetLatestBlockNumberOf("https://api-kylin.eoslaomao.com:443")
	checkError(t, err)
	t.Logf("last block number of eoslaomao: %v\n", blk)
}

func TestGetBalance(t *testing.T) {
	initKylin()
	t.Logf("TestGetBalance\n")
	acctname := "gzx123454321"
	bal, err := b.GetBalance(acctname)
	checkError(t, err)
	t.Logf("account %v has balance %v\n", acctname, bal)
}

func TestGetTransaction(t *testing.T) {
	initMainnet()
	t.Logf("TestGetTransaction (mainnet)\n")
	txid := "2801580d37d0ed93411532ca0b43d1623d906f39d056335acdaaf3c83f57774a"
	tx, err := b.GetTransaction(txid)
	checkError(t, err)
	t.Logf("GetTransaction result: %+v\n", tx)
	txstatus := b.GetTransactionStatus(txid)
	t.Logf("GetTransactionStatus result: %+v\n", txstatus)
}

var dryrun = flag.Bool("dryrun", true, "dry run")

func TestBuildTransaction(t *testing.T) {
	flag.Parse()
	fmt.Printf("dry run: %v\n", *dryrun)

	initKylin()

	cli := b.GetClient()

	// build swapout transaction
	// From gzx123454321
	// To 222222222ppp
	// Value 0.0001 EOS

	from := "gzx123454321"
	to := "222222222ppp"
	s := strconv.FormatFloat(float64(1)/10000, 'f', 4, 64) + " EOS"
	quantity, _ := eosgo.NewAsset(s)

	transfer := &eosgo.Action{
		Account: eosgo.AN("eosio.token"),
		Name:    eosgo.ActN("transfer"),
		Authorization: []eosgo.PermissionLevel{
			{
				Actor:      eosgo.AccountName(from),
				Permission: eosgo.PN("active"),
			},
		},
		ActionData: eosgo.NewActionData(token.Transfer{
			From:     eosgo.AccountName(from),
			To:       eosgo.AccountName(to),
			Quantity: quantity,
			Memo:     "Test transfer",
		}),
	}

	err := cli.FillFromChain(context.Background(), opts)
	checkError(t, err)

	var actions []*eosgo.Action
	actions = append(actions, transfer)

	rawTx := eosgo.NewTransaction(actions, opts)

	t.Logf("Build raw tx result: %+v\n", rawTx)

	// active key
	priv, err := WifToECDSA("5JqB...")
	checkError(t, err)

	stx, txhash, err := b.SignTransactionWithPrivateKey(rawTx, priv)
	checkError(t, err)

	t.Logf("Sign tx result: %+v\n", stx)
	t.Logf("Tx hash: %+v\n", txhash)

	if *dryrun == false {
		senttxhash, err := b.SendTransaction(stx)
		checkError(t, err)
		t.Logf("Tx sent: %v\n", senttxhash)
	} else {
		t.Logf("dryrun does not push tx to network")
	}
}
