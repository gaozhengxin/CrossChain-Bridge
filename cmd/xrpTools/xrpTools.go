package main

import (
	"fmt"
	"log"
	"os"

	"github.com/anyswap/CrossChain-Bridge/cmd/utils"
	"github.com/anyswap/CrossChain-Bridge/tokens"
	"github.com/anyswap/CrossChain-Bridge/tokens/eth"
	"github.com/anyswap/CrossChain-Bridge/tokens/xrp"
	"github.com/rubblelabs/ripple/websockets"
	"github.com/urfave/cli/v2"
)

var (
	clientIdentifier = "xrptools"
	gitCommit        = ""
	app              = utils.NewApp(clientIdentifier, gitCommit, "the xrptools command line interface")

	seed       string
	keyseq     uint
	to         string
	memo       string
	amount     string
	apiAddress string
	net        string
	b          *xrp.Bridge
	startScan  uint64

	seedFlag = &cli.StringFlag{
		Name:  "seed",
		Usage: "private key seed",
	}
	keyseqFlag = &cli.UintFlag{
		Name:  "keyseq",
		Usage: "private key sequence",
		Value: 0,
	}
	toFlag = &cli.StringFlag{
		Name:  "to",
		Usage: "send xrp to",
	}
	amountFlag = &cli.StringFlag{
		Name:  "amount",
		Usage: "send xrp amount (in drop)",
	}
	memoFlag = &cli.StringFlag{
		Name:  "memo",
		Usage: "swapin bind address",
	}
	netFlag = &cli.StringFlag{
		Name:  "net",
		Usage: "submit on network",
		Value: "testnet",
	}
	apiAddressFlag = &cli.StringFlag{
		Name:  "remote",
		Usage: "ripple api provider",
	}
	startScanFlag = &cli.Uint64Flag{
		Name:  "startscan",
		Usage: "start scan",
		Value: uint64(13794220),
	}

	sendXRPCommand = &cli.Command{
		Action: sendXrpAction,
		Name:   "sendxrp",
		Usage:  "sendxrp",
		Flags: []cli.Flag{
			seedFlag,
			keyseqFlag,
			toFlag,
			amountFlag,
			memoFlag,
			netFlag,
			apiAddressFlag,
		},
	}
	scanCommand = &cli.Command{
		Action: scanTxAction,
		Name:   "scan",
		Usage:  "scan ripple ledgers and txs",
		Flags: []cli.Flag{
			netFlag,
			apiAddressFlag,
			startScanFlag,
		},
	}
)

func main() {
	initApp()
	if err := app.Run(os.Args); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func initApp() {
	app.Action = xrptools
	app.Commands = []*cli.Command{
		sendXRPCommand,
		scanCommand,
	}
	app.Flags = []cli.Flag{
		//utils.VerbosityFlag,
		//utils.JSONFormatFlag,
		//utils.ColorFormatFlag,
	}
}

func initBridge() {
	tokens.DstBridge = eth.NewCrossChainBridge(false)
	b = xrp.NewCrossChainBridge(true)

	b.Remotes[apiAddress] = &websockets.Remote{}
}

func xrptools(ctx *cli.Context) error {
	utils.SetLogger(ctx)
	if ctx.NArg() > 0 {
		return fmt.Errorf("invalid command: %q", ctx.Args().Get(0))
	}

	_ = cli.ShowAppHelp(ctx)
	fmt.Println()
	log.Fatalf("please specify a sub command to run")
	return nil
}
