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

This simple implementation is replaced by the *Proof of Work* algorithm

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

## Consensus Algorithm

Also known as "Proof Of" algorithms

### Proof of Work

We want to secure our blockchain by making the server do computational work to add the next block

Called *mining* - running the Proof Of Work algorithm, 
powering the system, adding the next block and making the blockchain more secure,
and getting rewarded for their work

But once a miner signs a block, they need to show proof that they performed this work

*Work must be hard to do, but proving it should be relatively easy*

#### Requirements

- `The first few bytes must contain 0s`

For Bitcoin original proof of work specifications (HashCash), 
difficulty was **20** consecutive bits of the hash as zeroes

- `The difficulty changes over time`

Difficulty goes up, meaning more 0s in front to be valid

To account for increasing amount of miners, and the increasing
computational power of computers doing the mining. 

Want the relative time to compute a block stay the same over time,
as well as the creation rate of blocks stay the same

```
Difficulty = 12

func NewProof(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-Difficulty))

	pow := &ProofOfWork{b, target}

	return pow
}
```

raising the difficulty scales quickly

- Difficulty 12 takes ~241 ms to run 4 sha256 hashes
- Difficulty 15 takes ~1.25 seconds
- Difficulty 18 takes ~12 seconds
- Difficulty 21 takes ~1 minute 24 seconds

### Proof of Stake

## Persistence (Database)

Original specifications for bitcoin didn't call for a specific database

-Bitcoin and other cryptocurrencies use LevelDB

    -LevelDB is a low level key-value store database
    
- Bitcoin core specification: Two main groups of data
    
    - Blocks object
        
        - Stored with Metadata which describes all blocks on the chain
        
    - Chain State object
        
        - Stores the state of a chain and all current unspent transaction outputs, with some metadata
            
    - Bitcoin specifications has each block be it's own separate file on the DB
    
        - For Performance: With each on it's own file, don't have to open up multiple
        blocks just to read one

Using BadgerDB for tutorial blockchain

### BadgerDB

Native Golang, Key-Value storage database based off of LevelDB

Only accepts slices of bytes - `[]byte`

## Transactions

Inputs and Outputs

Because blockchains are public, there are no sensitive data in the block's data

All of the sensitive data are done through inputs and outputs

In Bitcoin, PubKey is derived from scripting language ("script")

### First transaction - Genesis (Coinbase)

The first transaction is the creation of the genesis block - 
known as the *Coinbase* transaction

- only one input and one output

- has a reward attached to it - given to the miner that has processed it

## Wallets

A wallet is made up of 2 keys: Private and Public Keys

- The *Private* key is the identifier for each of the accounts in the blockchain

    - This means that each private key must be globally unique
    
    - Using **Elliptical Curve Digital Signing Algorithm** (ecdsa)
    
- The Public key is given to other users, and derives the address used to send and receive data in the chain

```
                            [Private Key] 
                                 |
                                 V
                               [ecdsa] 
                                |
                                V
                            [public key]
                                   |
                                   V 
                             [sha 256]    
                                    |
                                    V 
                                [ripemd160]
                                    |
                                    V      
                              [public key hash]
                              /       \
                             /         \ 
                      [sha 256]         |
                          |             |
                          V             |
                      [sha 256]         |
                          |             |
                          V             |
                  [1st 4 bytes]         |
                          |             |
                          V             |
                      [Checksum]        |       [version]
                               \        |       /
                                \       |      /
                                 \      |     /
                                  \     |    /
                                   [ Base 58 ]
                                        |
                                        V
                                    [address]
```

*Base 58* was invented with bitcoin, derivative of base 64 algorithm

- Characters missing from Base 58:  `0 O l I + /`

    - they are easily confused with one another, 
    helps with human readability when one user gives their key to someone else

## Merkel Tree

A Merkel Tree is another optimization method much like the UTXO Set

Because Bitcoin is decentralized, every node must have its own independent and 
self sufficient node which will store a copy of the blockchain

- Now a node of Bitcoin is about 200 GB

There needs to be a way to verify a block transaction without downloading the block
itself.

Merkel Trees obtain transaction hashes, saved in a block header, 
considered by the proof of work system

```
                                        Merkle Root
                                     /               \
                                    /                 \
                Sha256 Branch A + B                      Sha256 Branch C + D
                /           \                               /               \
               /             \                             /                 \
    Branch A: sha256 tx1    Branch B: sha256 tx2    Branch C: sha256 tx3      Branch D: sha256 tx4
            |                       |                        |                        |
            |                       |       Merkle Tree      |                        |
    --------------------------------------------------------------------------------------------------
            |                       |Serialized Transactions |                        |
            |                       |                        |                        |
        Transaction One        Transaction Two       Transaction Three          Transaction Four                   
```

At each level, the parent of the two leaves below it is the combined hash of its children
- Branch A is a sha256 hash of the transaction
    - The lowest parts of the tree essentially have its children leaves set to nil
- Branch A + B is the sha256 hash of the two hashes of Branch A and Branch B
- The Merkle Root is the hash of Branch A + B and Branch C + D

One rule of a Merkel Tree is that the leaves of the tree must be even

- If there was no Transaction 4, then Branch D is just another hash of Transaction Three
- Branch C + D is basically the Hash of Transaction 3 + Hash of Transaction 3

The benefit of a Merkle Tree is that you can look at the Merkle Root to see if a block is inside of the tree