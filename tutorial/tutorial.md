# Part 1

Basic blockchain concepts and setup

# Part 2

Separated Blockchain logic into separate module

Proof of Work

# Part 3

Adding CLI

`go run main.go print`

`go run main.go add -block "block data"`

Serializing/Deserializing data into []byte

Persistent blockchain with BadgerDB

# Part 4

Basic Transactions

New CLI commands

`go run main.go createblockchain -address "john"`

`go run main.go printchain`

`go run main.go getbalance -address "john"`

`go run main.go send -from "john" -to "fred" -amount 50`

`go run main.go getbalance -address "fred"`

# Part 5

Adding wallet module

Refactoring CLI logic to be separate from main

Setting persistence for wallets - BadgerDB exclusively for blocks

`go run main.go createwallet`

`go run main.go listaddresses`

# Part 6

## Signing

Signing is to protect the transactions from modification
while waiting for a user to collect their unspent transaction and 
that a user can't deny that they had sent coins to another user

Genesis Block does not have signing

User A essentially unlocks the genesis block to collect the coinbase

When User A sends X coins to User B, a signature is created in the input
of the next block, with two outputs

-   User B gets an output in this block for the 10 coins being sent to them

-   As if User A paid with a hundred dollar bill, the second output is the 
90 coins in change going back to User A

-   The input of this first block tells the coinbase output from the genesis block
that it's okay to unlock and become spendable

Now when User B sends 5 tokens to User C, 
the output from Block 1 that held User B's 10 coins, and Block 2 is created
with that input

- Block 2 will have 2 outputs

    - 5 coins to User C
    
    - 5 coins back to User B

## Verification

```
Address: {address hash}
FullHash: {address hash}
[Version] 00
[Pub Key Hash] {hash}
[Checksum] 2bc6bc767
```

Take the address, and convert back into the full hash

- By passing it through the Base58 Decoder

    - Remove the version portion (1st 2 characters of the hash)
    
    - Remove the Pub Key Hash
    
    - Take the CheckSum out
    
    - Attach a new constant, and then the checksum, and compare again



## Locking of Transactions

## Integrating Wallets into our Blockchain

## CLI integration of new features

`go run main.go createwallet` 

- 2x to create 2 wallets to transfer between

```
New address is: 1DDUHF6ZhFCFH8V6e8wjWd7mtAXZVncKDc

New address is: 18cpTLjwkBMjVmatvTTuGUMfJFTNHRX4FS
```

`go run main.go createblockchain -address 1DDUHF6ZhFCFH8V6e8wjWd7mtAXZVncKDc`

```
0008f0a15a32e91bfa0fe6acdb5e97867e102b7e1a3f9fd1dfcc5fe660b8a3ea
Genesis Created
```

`go run main.go printchain`

```
previous hash: 
data in block: 0008f0a15a32e91bfa0fe6acdb5e97867e102b7e1a3f9fd1dfcc5fe660b8a3ea
PoW: true
--- Transaction 0ed53eb4acf99d7aa8a720d132604e55bea2bc259fbf472eed97a1b50be37aad:
        Input 0:
                TXID:   
                Out:    -1
                Signature:      
                PubKey:         4669727374205472616e73616374696f6e2066726f6d2047656e65736973
        Output 0
                Value: 100
                Script: 85fd45efa0580a2de939a4db6e8f16d86cdc2de5

```

`go run main.go send -to 18cpTLjwkBMjVmatvTTuGUMfJFTNHRX4FS -from 1DDUHF6ZhFCFH8V6e8wjWd7mtAXZVncKDc -amount 30`

```
000af036fa243d01612470b4a907c7391a6116c8057859c741eacfa196a76ab9
Successful send
```

`go run main.go printchain`

```
previous hash: 0008f0a15a32e91bfa0fe6acdb5e97867e102b7e1a3f9fd1dfcc5fe660b8a3ea
data in block: 000af036fa243d01612470b4a907c7391a6116c8057859c741eacfa196a76ab9
PoW: true
--- Transaction 10ba4a07140959539ac6f71faa34d46ed3e15732ad45b6ca0bb65ef001b7443b:
        Input 0:
                TXID:   0ed53eb4acf99d7aa8a720d132604e55bea2bc259fbf472eed97a1b50be37aad
                Out:    0
                Signature:      1e2d0e4a92f653e903c38153db8a3124eabde380a1f59c611260a4528d9ce4c99d9aaa72cccbd0483211ccd5d178cc1c137acd9fb5bc0adbce6cc876bc461c1b
                PubKey:         d2a17ba6cb2e5d144fe409775e069f2edd2ceacef67ab5253be33cb310403cf473e10b9617002acab56d21fdf701d8a9432b80a572b61e861240b8c9fc12dfda
        Output 0
                Value: 30
                Script: 538f0d4796823aa4f2ee244b4f3fc26f55b46cbb
        Output 1
                Value: 70
                Script: 85fd45efa0580a2de939a4db6e8f16d86cdc2de5

previous hash: 
data in block: 0008f0a15a32e91bfa0fe6acdb5e97867e102b7e1a3f9fd1dfcc5fe660b8a3ea
PoW: true
--- Transaction 0ed53eb4acf99d7aa8a720d132604e55bea2bc259fbf472eed97a1b50be37aad:
        Input 0:
                TXID:   
                Out:    -1
                Signature:      
                PubKey:         4669727374205472616e73616374696f6e2066726f6d2047656e65736973
        Output 0
                Value: 100
                Script: 85fd45efa0580a2de939a4db6e8f16d86cdc2de5
```

`go run main.go getbalance -address 18cpTLjwkBMjVmatvTTuGUMfJFTNHRX4FS`

```
Balance of 18cpTLjwkBMjVmatvTTuGUMfJFTNHRX4FS: 30
```

`go run main.go getbalance -address 1DDUHF6ZhFCFH8V6e8wjWd7mtAXZVncKDc`

```
Balance of 1DDUHF6ZhFCFH8V6e8wjWd7mtAXZVncKDc: 70
```

# Part 7

Adding Unspent Transaction Outputs Set

- Searching through all transactions to find one is inefficient

- Bitcoin's blocks takes up 200 GB

- Solution is to index the unspent transaction outputs, and search them for
specific things

    - Unspent Outputs are important, because they can tell how much a user has,
    and how much coin they can actually move around
    
    - Utilizing the same database for the Blockchain
    
        - Creating another layer in it for the Unspent Transactions
        
        - BadgerDB does not have tables. Only way to separate data is by **Prefixes**
        
`go run main.go createwallet`

```
New address is: 186FcUiLto18VrSDjm388M2vG22cxdF7Gq
```

`go run main.go createwallet`

```
New address is: 1M7cJeHzD9w1i4mLYLr82rCxDrNeqYGT6k
```

`go run main.go createblockchain -address 186FcUiLto18VrSDjm388M2vG22cxdF7Gq`

```
000bb2892af7423316a1c270364f84a098d9e4f4659669ed6796e8de77e3f02d
Genesis Created
```

`go run main.go getbalance -address 186FcUiLto18VrSDjm388M2vG22cxdF7Gq`

```
Balance of 186FcUiLto18VrSDjm388M2vG22cxdF7Gq: 100
```

# After Tutorial Refactor

Refactor the Network Module

Cleanup on Blockchain package

Switch from Badger to LevelDB

Split into Microservices

Docker and Kubernetes