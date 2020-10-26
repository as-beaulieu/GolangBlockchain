package blockchain

import (
	"fmt"
	"github.com/dgraph-io/badger"
	"log"
)

const (
	dbPath = "./tmp/blocks"
)

type BlockChain struct {
	LastHash []byte
	Database *badger.DB
}

type BlockChainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

func InitializeBlockChain() *BlockChain {
	var lastHash []byte

	opts := badger.DefaultOptions(dbPath)

	db, err := badger.Open(opts)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(transaction *badger.Txn) error {
		if _, err := transaction.Get([]byte("lh")); err == badger.ErrKeyNotFound {
			fmt.Println("No existing blockchain found. Creating Genesis block")
			genesis := Genesis()
			fmt.Println("Genesis Proved")
			if err = transaction.Set(genesis.Hash, genesis.Serialize()); err != nil {
				log.Panic(err)
			} //Hash is the key, and serialize the whole block

			if err = transaction.Set([]byte("lh"), genesis.Hash); err != nil {
				log.Panic(err)
			}

			lastHash = genesis.Hash

			return err
		} else { //If we already have a database, and already has a blockchain inside
			fmt.Println("database found, getting lastHash block")
			item, err := transaction.Get([]byte("lh"))
			if err != nil {
				log.Panic(err)
			}
			err = item.Value(func(val []byte) error {
				lastHash = append([]byte{}, val...)
				return nil
			})
			return err
		}
	})

	if err != nil {
		log.Panic(err)
	}

	blockchain := BlockChain{lastHash, db} //new blockchain in memory
	return &blockchain
}

func (chain *BlockChain) AddBlock(data string) {
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

		newBlock := CreateBlock(data, lastHash)

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
