package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/dgraph-io/badger"
	"log"
	"os"
	"runtime"
)

const (
	dbPath      = "./tmp/blocks"
	dbFile      = "./tmp/blocks/MANIFEST"
	genesisData = "First Transaction from Genesis"
)

type BlockChain struct {
	LastHash []byte
	Database *badger.DB
}

type BlockChainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

func InitializeBlockChain(address string) *BlockChain {
	var lastHash []byte

	if DBexists() {
		fmt.Println("Blockchain already exists")
		runtime.Goexit()
	}

	opts := badger.DefaultOptions(dbPath)

	db, err := badger.Open(opts)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(transaction *badger.Txn) error {
		coinbaseTransaction := CoinbaseTx(address, genesisData) //The address that is passed will be rewarded for mining the block
		genesis := Genesis(coinbaseTransaction)
		fmt.Println("Genesis Created")
		if err = transaction.Set(genesis.Hash, genesis.Serialize()); err != nil {
			return err
		}

		if err = transaction.Set([]byte("lh"), genesis.Hash); err != nil {
			return err
		}

		lastHash = genesis.Hash

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	blockchain := BlockChain{lastHash, db} //new blockchain in memory
	return &blockchain
}

func ContinueBlockChain(address string) *BlockChain {
	if DBexists() == false {
		fmt.Println("No existing blockchain found. Need to create one")
		runtime.Goexit()
	}

	var lastHash []byte

	opts := badger.DefaultOptions(dbPath)

	db, err := badger.Open(opts)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			lastHash = append([]byte{}, val...)
			return nil
		})

		return err
	})

	return &BlockChain{lastHash, db}
}

func (chain *BlockChain) AddBlock(transactions []*Transaction) *Block {
	var lastHash []byte
	var newBlock *Block

	for _, tx := range transactions {
		if chain.VerifyTransaction(tx) != true {
			log.Panic("Invalid Transaction")
		}
	}

	err := chain.Database.View(func(transaction *badger.Txn) error {
		item, err := transaction.Get([]byte("lh"))
		if err != nil {
			log.Panic(err)
		}
		err = item.Value(func(val []byte) error {
			lastHash = append([]byte{}, val...)
			return nil
		})
		if err != nil {
			log.Panic(err)
		}

		newBlock = CreateBlock(transactions, lastHash)

		err = chain.Database.Update(func(transaction *badger.Txn) error {
			if err := transaction.Set(newBlock.Hash, newBlock.Serialize()); err != nil {
				log.Panic(err)
			}

			if err = transaction.Set([]byte("lh"), newBlock.Hash); err != nil { //Set the new blocks hash as our latest lastHash
				log.Panic(err)
			}

			chain.LastHash = newBlock.Hash

			return err
		})
		if err != nil {
			log.Panic(err)
		}

		return err
	})

	if err != nil {
		log.Panic(err)
	}

	return newBlock
}

func (chain *BlockChain) Iterator() *BlockChainIterator {
	return &BlockChainIterator{chain.LastHash, chain.Database}
}

//Because we start with the BlockChain's LastHash, we're iterating backwards through the blocks (Newest -> Genesis)

func (iterator *BlockChainIterator) Next() *Block {
	var block *Block

	err := iterator.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(iterator.CurrentHash)
		if err != nil {
			log.Panic(err)
		}
		var encodedBlock []byte
		err = item.Value(func(val []byte) error {
			encodedBlock = append([]byte{}, val...)
			return nil
		})
		if err != nil {
			log.Panic(err)
		}

		block = Deserialize(encodedBlock)

		return err
	})
	if err != nil {
		log.Panic(err)
	}

	iterator.CurrentHash = block.PrevHash //because each block points to its previous block, this sets the next step in the iterator

	return block
}

func DBexists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}

func (chain *BlockChain) FindUnspentTransactions() map[string]TxOutputs {
	UTXO := make(map[string]TxOutputs)
	spentTXOs := make(map[string][]int)

	iterator := chain.Iterator()

	for {
		block := iterator.Next()

		for _, transaction := range block.Transactions {
			txID := hex.EncodeToString(transaction.ID)

		Outputs: //Label for this for loop so we can break this labeled loop and not the others
			for outIdx, out := range transaction.Outputs {
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}
			if transaction.IsCoinbase() == false {
				for _, in := range transaction.Inputs {
					inTxID := hex.EncodeToString(in.ID)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Out)
				}
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}
	return UTXO
}

func (chain *BlockChain) FindTransaction(ID []byte) (Transaction, error) {
	iterator := chain.Iterator()

	for {
		block := iterator.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}
	return Transaction{}, errors.New("transaction does not exist")
}

func (chain *BlockChain) SignTransaction(tx *Transaction, privateKey ecdsa.PrivateKey) {
	previousTXs := make(map[string]Transaction)

	for _, in := range tx.Inputs {
		previousTX, err := chain.FindTransaction(in.ID)
		if err != nil {
			log.Panic(err)
		}
		previousTXs[hex.EncodeToString(previousTX.ID)] = previousTX
	}

	tx.Sign(privateKey, previousTXs)
}

func (chain *BlockChain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	previousTXs := make(map[string]Transaction)

	for _, in := range tx.Inputs {
		previousTX, err := chain.FindTransaction(in.ID)
		if err != nil {
			log.Panic(err)
		}
		previousTXs[hex.EncodeToString(previousTX.ID)] = previousTX
	}

	return tx.Verify(previousTXs)
}
