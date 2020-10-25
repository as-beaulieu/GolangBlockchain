# Blockchain basics

## Blockchain and its blocks

```
type BlockChain struct {
	blocks []*Block
}
```

A *Blockchain* is a series of blocks of data connected by the one before it

```
type Block struct {
	Hash     []byte // Contains a hash of all information in the block
	Data     []byte // The actual public data
	PrevHash []byte // The hash of the block before it, setting up the block chain
}
```

Other features that go into a block is a timestamp, blockheight, and others

## Deriving hash

A hash is derived by joining all of the contents of the block, and then running a 
hashing algorithm on it. Here, we're running a SHA256 to create the hash from the block contents

```
func (b *Block) DeriveHash() {
	info := bytes.Join([][]byte{b.Data, b.PrevHash}, []byte{})
	hash := sha256.Sum256(info)
	b.Hash = hash[:]
}
```

**Note:** SHA256 is a rather simple way to derive the hash instead of the real way, which is perfect for
this demonstration

## Adding blocks

Adding blocks work by creating a new block, 
and referencing the hash of the block before it in the blockchain

```
func CreateBlock(data string, prevHash []byte) *Block {
	block := &Block{[]byte{}, []byte(data), prevHash}
	block.DeriveHash()
	return block
}

func (chain *BlockChain) AddBlock(data string) {
	prevBlock := chain.blocks[len(chain.blocks)-1]
	new := CreateBlock(data, prevBlock.Hash)
	chain.blocks = append(chain.blocks, new)
}
```

There is a problem with this simple implementation. 
Adding a block requires a block to exist before it.
What about initializing a new blockchain with a first block?

## Genesis Block

```
func Genesis() *Block {
	return CreateBlock("Genesis", []byte{})
}

func InitializeBlockChain() *BlockChain {
	return &BlockChain{[]*Block{Genesis()}}
}
```

## Comparing Blocks

With a real blockchain, you have multiple copies of the blockchain *ledger*.
To check for corruption, you must compare the hashes and seeing how they've changed

```
Block # 0
previous hash: 
data in block: Genesis
hash: 81ddc8d248b2dccdd3fdd5e84f0cad62b08f2d10b57f9a831c13451e5c5c80a5
Block # 1
previous hash: 81ddc8d248b2dccdd3fdd5e84f0cad62b08f2d10b57f9a831c13451e5c5c80a5
data in block: First block after genesis
hash: cb62069ecc6cfce5add8040d0b2a2da7a622382112b4fc9588788fe80e3d2bbe
Block # 2
previous hash: cb62069ecc6cfce5add8040d0b2a2da7a622382112b4fc9588788fe80e3d2bbe
data in block: Second block after genesis
hash: fc1494229b5818d6d62520e39dc8bb168ed4195348d6c4615942f9996f9a72a8
Block # 3
previous hash: fc1494229b5818d6d62520e39dc8bb168ed4195348d6c4615942f9996f9a72a8
data in block: Third block after genesis
hash: ec19d50693a7c6ace4b8a31eabd49d1636d7b7dfe4376bfc194adfdf1408e8b4
```

# Features

Wallets

Merkel Tree

Consensus Algorithm

