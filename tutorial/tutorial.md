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

