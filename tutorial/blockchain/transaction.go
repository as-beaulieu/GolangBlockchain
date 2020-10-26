package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
)

type Transaction struct {
	ID      []byte
	Inputs  []TxInput
	Outputs []TxOutput
}

type TxOutput struct {
	Value  int
	PubKey string
}

type TxInput struct {
	ID  []byte
	Out int
	Sig string
}

func CoinbaseTx(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Coins to %s", to)
	}

	txin := TxInput{[]byte{}, -1, data}
	txout := TxOutput{100, to} //The reward for mining the coinbase

	return &Transaction{nil, []TxInput{txin}, []TxOutput{txout}}
}

func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte

	encode := gob.NewEncoder(&encoded)
	if err := encode.Encode(tx); err != nil {
		log.Panic(err)
	}

	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

func (tx *Transaction) IsCoinbase() bool {
	singleInput := len(tx.Inputs) == 1
	noInputID := len(tx.Inputs[0].ID) == 0
	inputOutIsNegative := tx.Inputs[0].Out == -1
	return singleInput && noInputID && inputOutIsNegative
}

func (in *TxInput) CanUnlock(data string) bool {
	return in.Sig == data
}

func (out *TxOutput) CanBeUnlocked(data string) bool {
	return out.PubKey == data
}
