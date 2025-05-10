package main

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"time"
)

type Book struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	Author        string `json:"author"`
	PublishedDate string `json:"published_date"`
	ISBN          string `json:"isbn"`
}

type BookCheckout struct {
	BookID       string `json:"book_id"`
	User         string `json:"user"`
	CheckoutDate string `json:"checkout_date"`
	IsGenesis    bool   `json:"is_genesis"`
}

type Block struct {
	Position  int
	Data      BookCheckout
	TimeStamp string
	Hash      string
	PrevHash  string
}

// generate new hash for the new block
func (b *Block) generateHash() {
	bytes, err := json.Marshal(b.Data)
	if err != nil {
		log.Printf("could generate hash: %v", err)
		return
	}
	data := string(b.Position) + b.TimeStamp + string(bytes) + b.PrevHash
	hash := sha256.New()
	hash.Write([]byte(data))
	b.Hash = hex.EncodeToString(hash.Sum(nil))
}

func (b *Block) validateHash(hash string) bool {
	b.generateHash()
	if b.Hash != hash {
		return false
	}
	return true
}

type Blockchain struct {
	blocks []*Block
}

// CreateBlock function - creates a new block
func CreateBlock(prevBlock *Block, checkoutItem BookCheckout) *Block {
	block := &Block{}
	block.Position = prevBlock.Position + 1
	block.PrevHash = prevBlock.Hash
	block.TimeStamp = time.Now().String()
	block.generateHash()
	return block
}

// AddBlock method function - add new block with the previous blocks
func (bc *Blockchain) AddBlock(data BookCheckout) {
	// add a block to a blockchain
	prevBlock := bc.blocks[len(bc.blocks)-1]
	block := CreateBlock(prevBlock, data)
	if validBlock(block, prevBlock) {
		bc.blocks = append(bc.blocks, block)
	}
}

var blockchain *Blockchain

// main Function
func main() {
	fmt.Println("Hello Blockchain😍")

	blockchain = newBlockchain()

	router := mux.NewRouter()
	router.HandleFunc("/", getBlockchain).Methods("GET")
	router.HandleFunc("/", writeBlock).Methods("POST")
	router.HandleFunc("/new", newBook).Methods("POST")

	go func() {
		for _, block := range blockchain.blocks {
			fmt.Printf("Prev Hash: %x", block.Hash)
			bytes, _ := json.MarshalIndent(block.Data, "", " ")
			fmt.Printf("Data:%v\n", string(bytes))
			fmt.Printf("Hash:%v\n", block.Hash)
			fmt.Println()
		}
	}()

	log.Println("Listening on port 3000")

	log.Fatal(http.ListenAndServe(":3000", router))
}

// newBlockchain function - to create a new blockchain
func newBlockchain() *Blockchain {
	return &Blockchain{[]*Block{
		GenesisBlock(),
	}}
}

// GenesisBlock function - to generate the IsGenesis
func GenesisBlock() *Block {
	return CreateBlock(&Block{}, BookCheckout{IsGenesis: true})
}

// validBlock function
func validBlock(block, prevBlock *Block) bool {
	if prevBlock.Hash != block.PrevHash {
		return false
	}
	if !block.validateHash(block.Hash) {
		return false
	}
	if prevBlock.Position+1 != block.Position {
		return false
	}
	return true
}

// newBook function - create a new book
func newBook(res http.ResponseWriter, req *http.Request) {
	var book Book

	// unable to decode the post request
	if err := json.NewDecoder(req.Body).Decode(&book); err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		log.Printf("could not create the book: %v", err)
		res.Write([]byte("could not create the book"))
		return
	}

	h := md5.New()
	_, err := io.WriteString(h, book.ISBN+book.PublishedDate)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		log.Printf("unable to write string: %v", err)
		res.Write([]byte("unable to write string"))
		return
	}
	book.ID = fmt.Sprintf("%x", h.Sum(nil))

	data, err := json.MarshalIndent(book, "", " ")
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		log.Printf("unable to marshall payload: %v", err)
		res.Write([]byte("could not save book data"))
		return
	}
	res.WriteHeader(http.StatusOK)
	res.Write(data)
}

// writeBlock function
func writeBlock(res http.ResponseWriter, req *http.Request) {
	var checkoutItem BookCheckout

	if err := json.NewDecoder(req.Body).Decode(&checkoutItem); err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		log.Printf("could not write the block: %v", err)
		res.Write([]byte("could not write the block"))
		return
	}
	blockchain.AddBlock(checkoutItem)

	resp, err := json.MarshalIndent(checkoutItem, "", " ")
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		log.Printf("could not marshal payload: %v", err)
		res.Write([]byte("could not write block"))
		return
	}
	res.WriteHeader(http.StatusOK)
	res.Write(resp)
}

// getBlockchain function - get all the blockchains
func getBlockchain(res http.ResponseWriter, req *http.Request) {
	jbytes, err := json.MarshalIndent(blockchain.blocks, "", " ")
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(res).Encode(err)
		return
	}
	io.WriteString(res, string(jbytes))
}
