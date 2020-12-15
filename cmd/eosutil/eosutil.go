package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/anyswap/CrossChain-Bridge/tokens/eos"
)

var pubkeyhex string

func init() {
	flag.StringVar(&pubkeyhex, "pubkey", "", "pubkey hex")
}

func main() {
	flag.Parse()
	pubkey, err := eos.HexToPubKey(pubkeyhex)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(pubkey)
}
