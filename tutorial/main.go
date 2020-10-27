package main

import (
	"GolangBlockchain/tutorial/blockchain"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"time"
)

type CommandLine struct{}

func (cli *CommandLine) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("print :: prints the blocks in the blockchain")
	fmt.Println("getbalance -address ADDRESS :: get the balance for the address")
	fmt.Println("createblockchain -address ADDRESS :: creates a blockchain for the address")
	fmt.Println("send -from FROM -to TO -amount AMOUNT :: send amount from address to another address")
}

func (cli *CommandLine) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		runtime.Goexit() //exits the application by shutting down the goroutines
	}
}

func (cli *CommandLine) printChain() {
	chain := blockchain.ContinueBlockChain("")
	defer chain.Database.Close()
	iterator := chain.Iterator()

	for {
		block := iterator.Next()

		fmt.Printf("previous hash: %x\n", block.PrevHash)
		fmt.Printf("data in block: %x\n", block.Hash)
		pow := blockchain.NewProof(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(block.PrevHash) == 0 {
			break
		}
	}
}

func (cli *CommandLine) createBlockChain(address string) {
	chain := blockchain.InitializeBlockChain(address)
	chain.Database.Close()
	fmt.Println("finished")
}

func (cli *CommandLine) getBalance(address string) {
	chain := blockchain.ContinueBlockChain(address)
	defer chain.Database.Close()

	balance := 0
	UTXOs := chain.FindUnspentTransactionOutputs(address)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of %s: %d\n", address, balance)
}

func (cli *CommandLine) run() {
	cli.validateArgs()

	addBlockCmd := flag.NewFlagSet("add", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("print", flag.ExitOnError)
	addBlockData := addBlockCmd.String("block", "", "Block Data")

	switch os.Args[1] {
	case "add":
		if err := addBlockCmd.Parse(os.Args[2:]); err != nil {
			log.Panic(err)
		}
	case "print":
		if err := printChainCmd.Parse(os.Args[2:]); err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		runtime.Goexit()
	}

	if addBlockCmd.Parsed() {
		if *addBlockData == "" {
			addBlockCmd.Usage()
			runtime.Goexit()
		}
		cli.addBlock(*addBlockData)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}
}

func main() {
	startTime := time.Now()

	Run()

	elapsed := time.Since(startTime)
	fmt.Printf("Finished! Application took %s\n", elapsed)
}

func Run() {
	defer os.Exit(0)
	chain := blockchain.InitializeBlockChain()
	defer chain.Database.Close()

	cli := CommandLine{chain}
	cli.run()
}
