package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/anyswap/CrossChain-Bridge/cmd/utils"
	"github.com/anyswap/CrossChain-Bridge/log"
	"github.com/anyswap/CrossChain-Bridge/rpc/client"
	tronclient "github.com/fbsobreira/gotron-sdk/pkg/client"
	"github.com/anyswap/CrossChain-Bridge/tokens"
	"github.com/anyswap/CrossChain-Bridge/tokens/eth"
	"github.com/anyswap/CrossChain-Bridge/tokens/tron"
	"github.com/anyswap/CrossChain-Bridge/tokens/tools"
	"github.com/fsn-dev/fsn-go-sdk/efsn/common"
	"github.com/urfave/cli/v2"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/api"
	tronaddress "github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/golang/protobuf/ptypes"
)

var (
	scanTronCommand = &cli.Command{
		Action:    scanTron,
		Name:      "scantron",
		Usage:     "scan swap on tron",
		ArgsUsage: " ",
		Description: `
scan swap on tron
`,
		Flags: []cli.Flag{
			utils.GatewayFlag,
			utils.SwapServerFlag,
			utils.SwapTypeFlag,
			utils.DepositAddressSliceFlag,
			utils.TokenAddressSliceFlag,
			utils.PairIDSliceFlag,
			utils.StartHeightFlag,
			utils.EndHeightFlag,
			utils.StableHeightFlag,
			utils.JobsFlag,
			isSwapoutType2Flag,
			scanReceiptFlag,
			isProxyFlag,
		},
	}
)

type tronSwapScanner struct {
	gateway          string
	swapServer       string
	swapType         string
	depositAddresses []string
	tokenAddresses   []string
	pairIDs          []string
	startHeight      uint64
	endHeight        uint64
	stableHeight     uint64
	jobCount         uint64
	isSwapoutType2   bool
	scanReceipt      bool
	isProxy          bool

	client *tronclient.GrpcClient
	ctx    context.Context

	rpcInterval   time.Duration
	rpcRetryCount int

	isSwapin bool
}

func scanTron(ctx *cli.Context) error {
	utils.SetLogger(ctx)
	scanner := &tronSwapScanner{
		ctx:           context.Background(),
		rpcInterval:   3 * time.Second,
		rpcRetryCount: 3,
	}
	scanner.gateway = ctx.String(utils.GatewayFlag.Name)
	scanner.swapServer = ctx.String(utils.SwapServerFlag.Name)
	scanner.swapType = ctx.String(utils.SwapTypeFlag.Name)
	scanner.depositAddresses = ctx.StringSlice(utils.DepositAddressSliceFlag.Name)
	scanner.tokenAddresses = ctx.StringSlice(utils.TokenAddressSliceFlag.Name)
	scanner.pairIDs = ctx.StringSlice(utils.PairIDSliceFlag.Name)
	scanner.startHeight = ctx.Uint64(utils.StartHeightFlag.Name)
	scanner.endHeight = ctx.Uint64(utils.EndHeightFlag.Name)
	scanner.stableHeight = ctx.Uint64(utils.StableHeightFlag.Name)
	scanner.jobCount = ctx.Uint64(utils.JobsFlag.Name)
	scanner.isSwapoutType2 = ctx.Bool(isSwapoutType2Flag.Name)
	scanner.scanReceipt = ctx.Bool(scanReceiptFlag.Name)
	scanner.isProxy = ctx.Bool(isProxyFlag.Name)

	switch strings.ToLower(scanner.swapType) {
	case "swapin":
		scanner.isSwapin = true
	case "swapout":
		scanner.isSwapin = false
	default:
		log.Fatalf("unknown swap type: '%v'", scanner.swapType)
	}

	log.Info("get argument success",
		"gateway", scanner.gateway,
		"swapServer", scanner.swapServer,
		"swapType", scanner.swapType,
		"depositAddress", scanner.depositAddresses,
		"tokenAddress", scanner.tokenAddresses,
		"pairID", scanner.pairIDs,
		"scanReceipt", scanner.scanReceipt,
		"isProxy", scanner.isProxy,
		"start", scanner.startHeight,
		"end", scanner.endHeight,
		"stable", scanner.stableHeight,
		"jobs", scanner.jobCount,
	)

	scanner.verifyOptions()
	scanner.init()
	scanner.run()
	return nil
}

func (scanner *tronSwapScanner) verifyOptions() {
	if scanner.isSwapin && len(scanner.depositAddresses) != len(scanner.pairIDs) {
		log.Fatalf("count of depositAddresses and pairIDs mismatch")
	}
	if len(scanner.tokenAddresses) != len(scanner.pairIDs) {
		log.Fatalf("count of tokenAddresses and pairIDs mismatch")
	}
	if !scanner.isSwapin && len(scanner.tokenAddresses) == 0 {
		log.Fatal("must sepcify token address for swapout scan")
	}
	for i, pairID := range scanner.pairIDs {
		if pairID == "" {
			log.Fatal("must specify pairid")
		}
		if scanner.isSwapin && !common.IsHexAddress(scanner.depositAddresses[i]) {
			log.Fatalf("invalid deposit address '%v'", scanner.depositAddresses[i])
		}
		if scanner.tokenAddresses[i] != "" && !common.IsHexAddress(scanner.tokenAddresses[i]) {
			log.Fatalf("invalid token address '%v'", scanner.tokenAddresses[i])
		}
		switch strings.ToLower(pairID) {
		case "btc", "ltc":
			scanner.isSwapoutType2 = true
		}
	}
	if scanner.gateway == "" {
		log.Fatal("must specify gateway address")
	}
	if scanner.swapServer == "" {
		log.Fatal("must specify swap server address")
	}
	scanner.verifyJobsOption()
}

func (scanner *tronSwapScanner) verifyJobsOption() {
	if scanner.endHeight != 0 && scanner.startHeight >= scanner.endHeight {
		log.Fatalf("wrong scan range [%v, %v)", scanner.startHeight, scanner.endHeight)
	}
	if scanner.jobCount == 0 {
		log.Fatal("zero jobs specified")
	}
}

func (scanner *tronSwapScanner) init() {
	scanner.client = tronclient.NewGrpcClientWithTimeout(scanner.gateway, scanner.rpcInterval)
	if scanner.client == nil {
		log.Fatal("ethclient.Dail failed", "gateway", scanner.gateway)
	}

	tron.InitExtCodePartsWithFlag(scanner.isSwapoutType2)
	logSwapoutTopic = tron.ExtCodeParts["LogSwapoutTopic"]

	/*
	for _, tokenAddr := range scanner.tokenAddresses {
		if scanner.isSwapin && tokenAddr == "" {
			continue
		}
		var code []byte
		code, err = ethcli.CodeAt(scanner.ctx, common.HexToAddress(tokenAddr), nil)
		if err != nil {
			log.Fatalf("get contract code of '%v' failed, %v", tokenAddr, err)
		}
		if len(code) == 0 {
			log.Fatalf("'%v' is not contract address", tokenAddr)
		}
		if scanner.isSwapin {
			err = eth.VerifyErc20ContractCode(code)
		} else {
			err = eth.VerifySwapContractCode(code)
		}
		if err != nil {
			if scanner.isProxy {
				log.Warn("verify contract code failed. please ensure it's proxy contract", "contract", tokenAddr, "err", err)
			} else {
				log.Fatalf("wrong contract address '%v', %v", tokenAddr, err)
			}
		}
	}
	*/
}

func (scanner *tronSwapScanner) run() {
	start := scanner.startHeight
	wend := scanner.endHeight
	if wend == 0 {
		wend = scanner.loopGetLatestBlockNumber()
	}
	if start == 0 {
		start = wend
	}

	scanner.doScanRangeJob(start, wend)

	if scanner.endHeight == 0 {
		scanner.scanLoop(wend)
	}
}

// nolint:dupl // in diff sub command
func (scanner *tronSwapScanner) doScanRangeJob(start, end uint64) {
	if start >= end {
		return
	}
	jobs := scanner.jobCount
	count := end - start
	step := count / jobs
	if step == 0 {
		jobs = 1
		step = count
	}
	wg := new(sync.WaitGroup)
	for i := uint64(0); i < jobs; i++ {
		from := start + i*step
		to := start + (i+1)*step
		if i+1 == jobs {
			to = end
		}
		wg.Add(1)
		go scanner.scanRange(i+1, from, to, wg)
	}
	if scanner.endHeight != 0 {
		wg.Wait()
	}
}

func (scanner *tronSwapScanner) scanRange(job, from, to uint64, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Info(fmt.Sprintf("[%v] start scan range", job), "from", from, "to", to)

	for {
		res, err := scanner.client.GetBlockByLimitNext(int64(from), int64(to))
		if err != nil {
			log.Warn("Get block by limit failed", "err", err)
			continue
		}
		for _, block := range  res.Block {
			for _,  tx := range block.Transactions {
				scanner.scanTransaction(tx)
			}
		}
		break
	}

	log.Info(fmt.Sprintf("[%v] scan range finish", job), "from", from, "to", to)
}

func (scanner *tronSwapScanner) scanLoop(from uint64) {
	stable := scanner.stableHeight
	log.Info("start scan loop", "from", from, "stable", stable)
	for {
		latest := scanner.loopGetLatestBlockNumber()
		res, err := scanner.client.GetBlockByLimitNext(int64(from), int64(latest))
		if err != nil {
			log.Warn("Get block by limit failed", "err", err)
			continue
		}
		for _, block := range  res.Block {
			for _,  tx := range block.Transactions {
				scanner.scanTransaction(tx)
			}
		}
		if from+stable < latest {
			from = latest - stable
		}
		time.Sleep(5 * time.Second)
	}
}

func (scanner *tronSwapScanner) loopGetLatestBlockNumber() uint64 {
	for {
		res, err := scanner.client.GetNowBlock()
		if err == nil {
			height := uint64(res.BlockHeader.RawData.Number)
			log.Info("get latest block number success", "height", height)
			return height
		}
		log.Warn("get latest block number failed", "err", err)
		time.Sleep(scanner.rpcInterval)
	}
}

func (scanner *tronSwapScanner) scanTransaction(txext *api.TransactionExtention) {
	tx := txext.Transaction
	ret := tx.GetRet()
	if len(ret) != 1 || ret[0].GetRet() != core.Transaction_Result_SUCESS {
		return
	}
	var err error
	for i, pairID := range scanner.pairIDs {
		tokenAddress := scanner.tokenAddresses[i]
		if scanner.isSwapin {
			depositAddress := scanner.depositAddresses[i]
			if tokenAddress != "" {
				err = scanner.verifyTrc20SwapinTx(tx, tokenAddress, depositAddress)
			} else {
				err = scanner.verifySwapinTx(tx, depositAddress)
			}
		} else {
			err = scanner.verifySwapoutTx(tx, tokenAddress)
		}
		if !tokens.ShouldRegisterSwapForError(err) {
			continue
		}
		scanner.postSwap(fmt.Sprintf("%x", txext.GetTxid()), pairID)
		break
	}
}

func (scanner *tronSwapScanner) postSwap(txid, pairID string) {
	var subject, rpcMethod string
	if scanner.isSwapin {
		subject = "post swapin register"
		rpcMethod = "swap.Swapin"
	} else {
		subject = "post swapout register"
		rpcMethod = "swap.Swapout"
	}
	log.Info(subject, "txid", txid, "pairID", pairID)

	var result interface{}
	args := map[string]interface{}{
		"txid":   txid,
		"pairid": pairID,
	}
	for i := 0; i < scanner.rpcRetryCount; i++ {
		err := client.RPCPost(&result, scanner.swapServer, rpcMethod, args)
		if tokens.ShouldRegisterSwapForError(err) {
			break
		}
		if tools.IsSwapAlreadyExistRegisterError(err) {
			break
		}
		log.Warn(subject+" failed", "txid", txid, "pairID", pairID, "err", err)
	}
}

func (scanner *tronSwapScanner) verifyTrc20SwapinTx(tx *core.Transaction, tokenAddress, depositAddress string) error {
	if len(tx.RawData.Contract) != 1 {
		return fmt.Errorf("Tron transaction contract number is not 1")
	}

	contract := tx.RawData.Contract[0]
	if contract.Type != core.Transaction_Contract_TriggerSmartContract {
		return fmt.Errorf("Not a trigger smart contract contract")
	}

	var c core.TriggerSmartContract
	err := ptypes.UnmarshalAny(contract.GetParameter(), &c)
	if err != nil {
		return fmt.Errorf("Decode trigger smart contract contract error: %v", err)
	}

	contractAddress := fmt.Sprintf("%v", tronaddress.Address(c.ContractAddress))

	if !strings.EqualFold(contractAddress, tokenAddress) {
		return tokens.ErrTxWithWrongContract
	}

	inputData := c.Data

	_, _, value, err := eth.ParseErc20SwapinTxInput(&inputData, depositAddress)
	if err != nil {
		return err
	}

	if value.Sign() <= 0 {
		return tokens.ErrTxWithWrongValue
	}

	return nil
}

func (scanner *tronSwapScanner) verifySwapinTx(tx *core.Transaction, depositAddress string) error {
	if len(tx.RawData.Contract) != 1 {
		return fmt.Errorf("Tron transaction contract number is not 1")
	}
	contract := tx.RawData.Contract[0]
	if contract.Type != core.Transaction_Contract_TransferContract {
		return fmt.Errorf("Not a TRX transfer contract")
	}
	var c core.TransferContract
	err := ptypes.UnmarshalAny(contract.GetParameter(), &c)
	if err != nil {
		return fmt.Errorf("Decode transfer error: %v", err)
	}
	toAddress := fmt.Sprintf("%v", tronaddress.Address(c.ToAddress))
	if strings.EqualFold(toAddress, depositAddress) == false {
		return tokens.ErrTxWithWrongReceiver
	}

	return nil
}

func (scanner *tronSwapScanner) verifySwapoutTx(tx *core.Transaction, tokenAddress string) (err error) {
	return fmt.Errorf("verify swapout not implemented on tron")
}
