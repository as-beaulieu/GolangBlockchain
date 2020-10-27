package blockchain

import (
	"encoding/hex"
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

func (chain *BlockChain) AddBlock(transactions []*Transaction) {
	var lastHash []byte

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

		newBlock := CreateBlock(transactions, lastHash)

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

func (chain *BlockChain) FindUnspentTransactions(address string) []Transaction {
	var unspentTxs []Transaction

	spentTXOs := make(map[string][]int)

	iterator := chain.Iterator()

	for {
		block := iterator.Next()

		for _, transaction := range block.Transactions {
			txID := hex.EncodeToString(transaction.ID)

		Outputs: //Lable for this for loop so we can break this labeled loop and not the others
			for outIdx, out := range transaction.Outputs {
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				if out.CanBeUnlocked(address) {
					unspentTxs = append(unspentTxs, *transaction)
				}
			}
			if transaction.IsCoinbase() == false {
				for _, in := range transaction.Inputs {
					if in.CanUnlock(address) {
						inTxID := hex.EncodeToString(in.ID)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Out)
					}
				}
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}
	return unspentTxs
}

func (chain *BlockChain) FindUnspentTransactionOutputs(address string) []TxOutput {
	var UTXOs []TxOutput
	unspentTransactions := chain.FindUnspentTransactions(address)

	for _, tx := range unspentTransactions {
		for _, out := range tx.Outputs {
			if out.CanBeUnlocked(address) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

//FindSpendableOutputs will enable normal transactions that are not coinbase transactions
func (chain *BlockChain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOuts := make(map[string][]int)
	unspentTxs := chain.FindUnspentTransactions(address)
	accumulated := 0

Work:
	for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Outputs {
			if out.CanBeUnlocked(address) && accumulated < amount {
				accumulated += out.Value
				unspentOuts[txID] = append(unspentOuts[txID], outIdx)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	return accumulated, unspentOuts
}
