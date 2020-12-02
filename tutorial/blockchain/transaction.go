package blockchain

import (
	"GolangBlockchain/tutorial/wallet"
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strings"
)

const (
	ErrorPreviousTransactionNotExist = "ERROR: Previous transaction is not correct"
)

type Transaction struct {
	ID      []byte
	Inputs  []TxInput
	Outputs []TxOutput
}

func (tx Transaction) Serialize() []byte {
	var encoded bytes.Buffer
	enc := gob.NewEncoder(&encoded)
	if err := enc.Encode(tx); err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}

func (tx *Transaction) Hash() []byte {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.Serialize())

	return hash[:]
}

func CoinbaseTx(to, data string) *Transaction {
	if data == "" {
		randomData := make([]byte, 24)
		_, err := rand.Read(randomData)
		if err != nil {
			log.Panic(err)
		}
		data = fmt.Sprintf("%x", randomData)
	}

	txin := TxInput{[]byte{}, -1, nil, []byte(data)}
	txout := NewTxOutput(20, to) //The reward for mining the coinbase

	tx := Transaction{nil, []TxInput{txin}, []TxOutput{*txout}}
	tx.Hash()

	return &tx
}

func (tx *Transaction) IsCoinbase() bool {
	singleInput := len(tx.Inputs) == 1
	noInputID := len(tx.Inputs[0].ID) == 0
	inputOutIsNegative := tx.Inputs[0].Out == -1
	return singleInput && noInputID && inputOutIsNegative
}

func (tx *Transaction) Sign(privateKey ecdsa.PrivateKey, previousTXs map[string]Transaction) {
	if tx.IsCoinbase() {
		return
	}

	for _, in := range tx.Inputs {
		if previousTXs[hex.EncodeToString(in.ID)].ID == nil {
			log.Panic(ErrorPreviousTransactionNotExist)
		}
	}

	txCopy := tx.TrimmedCopy()

	for inId, in := range txCopy.Inputs {
		previousTransaction := previousTXs[hex.EncodeToString(in.ID)]
		txCopy.Inputs[inId].Signature = nil
		txCopy.Inputs[inId].PubKey = previousTransaction.Outputs[in.Out].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Inputs[inId].PubKey = nil

		r, s, err := ecdsa.Sign(rand.Reader, &privateKey, txCopy.ID)
		if err != nil {
			log.Panic(err)
		}
		signature := append(r.Bytes(), s.Bytes()...)

		tx.Inputs[inId].Signature = signature
	}
}

func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	for _, in := range tx.Inputs {
		inputs = append(inputs, TxInput{in.ID, in.Out, nil, nil})
	}

	for _, out := range tx.Outputs {
		outputs = append(outputs, TxOutput{out.Value, out.PubKeyHash})
	}

	return Transaction{tx.ID, inputs, outputs}
}

func (tx *Transaction) Verify(previousTXs map[string]Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	for _, in := range tx.Inputs {
		if previousTXs[hex.EncodeToString(in.ID)].ID == nil {
			log.Panic(ErrorPreviousTransactionNotExist)
		}
	}

	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	for inId, in := range tx.Inputs {
		previousTransaction := previousTXs[hex.EncodeToString(in.ID)]
		txCopy.Inputs[inId].Signature = nil
		txCopy.Inputs[inId].PubKey = previousTransaction.Outputs[in.Out].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Inputs[inId].PubKey = nil

		//Unpack all the data
		r := big.Int{}
		s := big.Int{}
		signatureLength := len(in.Signature)
		r.SetBytes(in.Signature[:(signatureLength / 2)])
		s.SetBytes(in.Signature[(signatureLength / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLength := len(in.PubKey)
		x.SetBytes(in.PubKey[:(keyLength / 2)])
		y.SetBytes(in.PubKey[(keyLength / 2):])

		rawPubKey := ecdsa.PublicKey{curve, &x, &y}
		if ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s) == false {
			return false
		}
	}

	return true
}

func (tx Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.ID))
	for i, input := range tx.Inputs {
		lines = append(lines, fmt.Sprintf("	Input %d:", i))
		lines = append(lines, fmt.Sprintf("		TXID:	%x", input.ID))
		lines = append(lines, fmt.Sprintf("		Out:	%d", input.Out))
		lines = append(lines, fmt.Sprintf("		Signature:	%x", input.Signature))
		lines = append(lines, fmt.Sprintf("		PubKey:		%x", input.PubKey))
	}

	for i, output := range tx.Outputs {
		lines = append(lines, fmt.Sprintf("	Output %d", i))
		lines = append(lines, fmt.Sprintf("		Value: %d", output.Value))
		lines = append(lines, fmt.Sprintf("		Script: %x", output.PubKeyHash))
	}

	return strings.Join(lines, "\n")
}

func NewTransaction(from, to string, amount int, UTXO *UTXOSet) *Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	wallets, err := wallet.CreateWallets()
	if err != nil {
		log.Panic(err)
	}
	w := wallets.GetWallet(from)
	pubKeyHash := wallet.PublicKeyHash(w.PublicKey)

	accumulator, validOutputs := UTXO.FindSpendableOutputs(pubKeyHash, amount)

	if accumulator < amount {
		log.Panic("Error: not enough funds")
	}

	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			log.Panic(err)
		}

		for _, out := range outs {
			input := TxInput{txID, out, nil, w.PublicKey}
			inputs = append(inputs, input)
		}
	}

	outputs = append(outputs, *NewTxOutput(amount, to))
	if accumulator > amount {
		outputs = append(outputs, *NewTxOutput(accumulator-amount, from))
	}

	tx := Transaction{nil, inputs, outputs}
	tx.ID = tx.Hash()
	UTXO.BlockChain.SignTransaction(&tx, w.PrivateKey)

	return &tx
}
