package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"sort"
	"time"

	"github.com/anyswap/CrossChain-Bridge/cmd/utils"
	"github.com/anyswap/CrossChain-Bridge/log"
	"github.com/anyswap/CrossChain-Bridge/params"
	"github.com/anyswap/CrossChain-Bridge/tokens"
	"github.com/anyswap/CrossChain-Bridge/worker"
	"github.com/urfave/cli/v2"
)

var (
	clientIdentifier = "swaporacle"
	// Git SHA1 commit hash of the release (set via linker flags)
	gitCommit = ""
	// The app that holds all commands and flags.
	app = utils.NewApp(clientIdentifier, gitCommit, "the swaporacle command line interface")
)

func initApp() {
	// Initialize the CLI app and start action
	app.Action = swaporacle
	app.HideVersion = true // we have a command to print the version
	app.Copyright = "Copyright 2017-2020 The CrossChain-Bridge Authors"
	app.Commands = []*cli.Command{
		utils.LicenseCommand,
		utils.VersionCommand,
	}
	app.Flags = []cli.Flag{
		utils.ConfigFileFlag,
		utils.TokenPairsDirFlag,
		utils.LogFileFlag,
		utils.LogRotationFlag,
		utils.LogMaxAgeFlag,
		utils.VerbosityFlag,
		utils.JSONFormatFlag,
		utils.ColorFormatFlag,
	}
	sort.Sort(cli.CommandsByName(app.Commands))
}

func main() {
	initApp()
	if err := app.Run(os.Args); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func swaporacle(ctx *cli.Context) error {
	utils.SetLogger(ctx)
	if ctx.NArg() > 0 {
		return fmt.Errorf("invalid command: %q", ctx.Args().Get(0))
	}
	exitCh := make(chan struct{})
	configFile := utils.GetConfigFilePath(ctx)
	params.LoadConfig(configFile, false)

	tokens.SetTokenPairsDir(utils.GetTokenPairsDir(ctx))

	worker.StartWork(false)

	// cpuprofile
	f, err := os.Create("./cpuprofile")
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	if err := pprof.StartCPUProfile(f); err != nil { //监控cpu
		log.Fatal("could not start CPU profile: ", err)
	}
	defer pprof.StopCPUProfile()

	// memprofile
	go func() {
		for {
			f, err := os.Create("./memprofile")
			if err != nil {
				log.Fatal("could not create memory profile: ", err)
			}
			runtime.GC()                                      // GC，获取最新的数据信息
			if err := pprof.WriteHeapProfile(f); err != nil { // 写入内存信息
				log.Fatal("could not write memory profile: ", err)
			}
			f.Close()
			time.Sleep(10 * time.Second)
		}
	}()

	<-exitCh
	return nil
}
