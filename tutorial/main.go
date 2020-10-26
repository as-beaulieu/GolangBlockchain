package main

import (
	"GolangBlockchain/tutorial/blockchain"
	"fmt"
	"strconv"
	"time"
)

func main() {
	startTime := time.Now()

	Run()

	elapsed := time.Since(startTime)
	fmt.Printf("Finished! Application took %s\n", elapsed)
}

func Run() {
	chain := blockchain.InitializeBlockChain()

	chain.AddBlock("First block after genesis")
	chain.AddBlock("Second block after genesis")
	chain.AddBlock("Third block after genesis")

	for i, block := range chain.Blocks {
		fmt.Println("Block #", i)
		fmt.Printf("previous hash: %x\n", block.PrevHash)
		fmt.Printf("data in block: %s\n", block.Data)
		fmt.Printf("hash: %x\n", block.Hash)

		pow := blockchain.NewProof(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()
	}
}
