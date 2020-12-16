package main

import (
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"math/big"

	"github.com/anyswap/CrossChain-Bridge/tokens"
	"github.com/anyswap/CrossChain-Bridge/tokens/block"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcwallet/wallet/txauthor"
)

var (
	b                                                  *block.Bridge
	fromAddress, wif, toAddress, changeAddress, amount string
	dryrun                                             bool
)

func init() {
	flag.StringVar(&fromAddress, "fromAddress", "", "from address")
	flag.StringVar(&wif, "wif", "", "private key in blocknet wif format")
	flag.StringVar(&toAddress, "toAddress", "", "to address")
	flag.StringVar(&amount, "amount", "", "send amount in wei")
	flag.BoolVar(&dryrun, "dryrun", true, "dry run")
	initBridge()
}

func initBridge() {
	b = block.NewCrossChainBridge(true)

	b.ChainConfig = &tokens.ChainConfig{
		BlockChain: "Block",
		NetID:      "mainnet",
	}

	b.GatewayConfig = &tokens.GatewayConfig{
		APIAddress: []string{"5.189.139.168:51515"},
		Extras: &tokens.GatewayExtras{
			BlockExtra: &tokens.BlockExtraArgs{
				CoreAPIs: []tokens.BlocknetCoreAPIArgs{
					{
						APIAddress:  "5.189.139.168:51515",
						RPCUser:     "xxmm",
						RPCPassword: "123456",
						DisableTLS:  true,
					},
				},
				UTXOAPIAddresses: []string{"https://plugin-dev.core.cloudchainsinc.com"},
			},
		},
	}
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	flag.Parse()

	changeAddress = fromAddress

	amt, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		log.Fatal(fmt.Errorf("invalid amount"))
	}

	pkwif, err := btcutil.DecodeWIF(wif)
	checkError(err)
	privkey := pkwif.PrivKey.ToECDSA()

	// build tx
	utxos, err := b.FindUtxos(fromAddress)
	checkError(err)

	txOuts, err := b.GetTxOutputs(toAddress, amt, "")
	checkError(err)

	inputSource := func(target btcAmountType) (total btcAmountType, inputs []*wireTxInType, inputValues []btcAmountType, scripts [][]byte, err error) {
		return b.GetUtxosFromElectUtxos(target, fromAddress, utxos)
	}

	changeSource := func() ([]byte, error) {
		return b.GetPayToAddrScript(changeAddress)
	}

	relayFeePerKb := 10000

	authoredTx, err = b.NewUnsignedTransaction(txOuts, btcAmountType(relayFeePerKb), inputSource, changeSource, true)
	checkError(err)

	// signTx
	signedTx, _, err := b.SignTransactionWithPrivateKey(authoredTx, privkey)
	checkError(err)

	tx := signedTx.(*txauthor.AuthoredTx).Tx

	buf := bytes.NewBuffer(make([]byte, 0, tx.SerializeSize()))
	err = tx.Serialize(buf)
	checkError(err)
	txHex := hex.EncodeToString(buf.Bytes())
	if dryrun {
		log.Printf("Bridge create tx: %v\n", txHex)
		log.Printf("Tx hash: %v\n", tx.TxHash())
		log.Println("dry run does not push tx to network")
		return
	}
	txHash, err := b.PostTransaction(txHex)
	checkError(err)
	log.Println("Bridge send tx, hash: %v", txHash)
	log.Println("Done")
}

type btcAmountType = btcutil.Amount
type wireTxInType = wire.TxIn

var authoredTx *txauthor.AuthoredTx
