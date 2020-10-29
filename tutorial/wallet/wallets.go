package wallet

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"io/ioutil"
	"log"
)

const walletFile = "./tmp/wallets.data"

type Wallets struct {
	Wallets map[string]*Wallet
}

func (ws *Wallets) SaveFile() {
	var content bytes.Buffer
	gob.Register(elliptic.P256())

	encoder := gob.NewEncoder(&content)
	if err := encoder.Encode(ws); err != nil {
		log.Panic(err)
	}

	if err := ioutil.WriteFile(walletFile, content.Bytes(), 0644); err != nil {
		log.Panic(err)
	}
}
