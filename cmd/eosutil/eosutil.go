package main

import (
	"fmt"
	"log"

	"github.com/anyswap/CrossChain-Bridge/tokens/eos"
)

func main() {
	pubkeyhex := "04c030e0d3e40fa06d94dae0122f0d472caaf0e27536286dfe4803d51d2db6d78e802a86fe77d319138fdbe3bda226782722b136b29c0c5ee80223dc143de1fa1b"
	pubkey, err := eos.HexToPubKey(pubkeyhex)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(pubkey)
}
