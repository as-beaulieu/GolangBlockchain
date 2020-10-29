package main

import (
	"GolangBlockchain/tutorial/wallet"
	"fmt"
	"time"
)

func main() {
	startTime := time.Now()

	Run()

	elapsed := time.Since(startTime)
	fmt.Printf("Finished! Application took %s\n", elapsed)
}

func Run() {
	//defer os.Exit(0)
	//command := cli.CommandLine{}
	//command.Run()

	w := wallet.MakeWallet()
	w.Address()
}
