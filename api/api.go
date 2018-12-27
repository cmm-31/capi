package api

import (
	"net/http"
	"log"
	"strings"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"encoding/json"
	"strconv"
	"os"
	"github.com/btcsuite/btcd/rpcclient"
	"time"
	"io/ioutil"
	"html/template"
	"github.com/boltdb/bolt"
	"fmt"
	"capi/db"
	"encoding/binary"
	"bytes"
)


type Config struct {
	Coin				string	`json:"Coin"`
	Ticker				string	`json:"Ticker"`
	Daemon				string	`json:"Daemon"`
	RPCUser				string	`json:"RPCUser"`
	RPCPassword	 		string	`json:"RPCPassword"`
	HTTPPostMode		bool
	DisableTLS			bool
	EnableCoinCodexAPI 	bool	`json:"EnableCoinCodexAPI"`
	CapiPort			string	`json:"capi_port"`
}

// Block Structs

// Block struct
type Block struct {
	Hash              string   `json:"hash"`
	Confirmations     int64    `json:"confirmations"`
	Size              int32    `json:"size"`
	StrippedSize  	  int32	   `json:"strippedSize"`
	Weight      	  int32	   `json:"weight"`
	Height            int64    `json:"height"`
	Version           int32	   `json:"version"`
	VersionHex  	  string   `json:"versionHex"`
	MerkleRoot        string   `json:"merkleRoot"`
	BlockTransactions []string `json:"tx"`
	Time              int64    `json:"time"`
	Nonce             uint32   `json:"nonce"`
	Bits              string   `json:"bits"`
	Difficulty        float64  `json:"difficulty"`
	PreviousHash      string   `json:"previousBlockHash"`
	NextHash          string   `json:"nextBlockHash"`

}

type TxRaw struct {
	Hex           string `json:"hex"`
	Txid          string `json:"txid"`
	Hash          string `json:"hash,omitempty"`
	Size          int32  `json:"size,omitempty"`
	Vsize         int32  `json:"vsize,omitempty"`
	Version       int32  `json:"version"`
	LockTime      uint32 `json:"locktime"`
	Vin           []Vin  `json:"vin"`
	Vout          []Vout `json:"vout"`
	BlockHash     string `json:"blockhash,omitempty"`
	Confirmations uint64 `json:"confirmations,omitempty"`
	Time          int64  `json:"time,omitempty"`
	Blocktime     int64  `json:"blocktime,omitempty"`
}

type Vin struct {
	Coinbase  string     `json:"coinbase"`
	Txid      string     `json:"txid"`
	Vout      uint32     `json:"vout"`
	ScriptSig *ScriptSig `json:"scriptSig"`
	Sequence  uint32     `json:"sequence"`
	Witness   []string   `json:"txinwitness"`
}

type Vout struct {
	Value        float64            `json:"value"`
	N            uint32             `json:"n"`
	ScriptPubKey ScriptPubKeyResult `json:"scriptPubKey"`
}

type ScriptPubKeyResult struct {
	Asm       string   `json:"asm"`
	Hex       string   `json:"hex,omitempty"`
	ReqSigs   int32    `json:"reqSigs,omitempty"`
	Type      string   `json:"type"`
	Addresses []string `json:"addresses,omitempty"`
}

type ScriptSig struct {
	Asm string `json:"asm"`
	Hex string `json:"hex"`
}



var Blocks []Block



/// Function to Load Config file from disk as type Config struct

func LoadConfig(file string) (Config) {
	// get the local config from disk
	//filename is the path to the json config file

	var config Config
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		log.Fatal("ERROR: Could not find config file \n GoLang Error:  " , err)
	}

	decoder := json.NewDecoder(configFile)
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatal("ERROR: Could not decode json config  \n GoLang Error:  " , err)
	}
	return config
}

// config file from disk using loadConfig function
var configFile  = LoadConfig("./config/config.json")

// coin client using coinClientConfig
var coinClient, _ = rpcclient.New(coinClientConfig, nil)

// coin client config for coinClient, loads values from configFile
var coinClientConfig = &rpcclient.ConnConfig {
	Host: 			configFile.Daemon,
	User: 			configFile.RPCUser,
	Pass: 			configFile.RPCPassword,
	HTTPPostMode:	configFile.HTTPPostMode,
	DisableTLS:		configFile.DisableTLS,
}


//GetBlockObject
func GetBlock(w http.ResponseWriter, r *http.Request)  {


	urlBlock := r.URL.Path
	if len(urlBlock) > 60{

		urlBlock = strings.TrimPrefix(urlBlock, "/block/")

		log.Println("Block Hash", urlBlock)

		hash, err := chainhash.NewHashFromStr(urlBlock)
		if err != nil {
			log.Print("Error with hash")
		}

		log.Println("Trying to get block hash: ", urlBlock)
		block, err := coinClient.GetBlockVerbose(hash)
		if err != nil {
			log.Print("Error with hash requested: ", urlBlock)
			http.Error(w, "ERROR: invalid block hash requested \n"+
				"Please use a block hash, eg: 4b6c3362e2f2a6b6317c85ecaa0f5415167e2bb333d2bf3d3699d73df613b91f", 500)
			return
		}


		jsonBlock, err := json.Marshal(&block)
		data := json.RawMessage(jsonBlock)
		json.NewEncoder(w).Encode(data)


	} else {
		urlBlock = strings.TrimPrefix(urlBlock, "/block/")
		log.Println("Parsed Block Object from the URL", urlBlock)

		blockHeight, err := strconv.ParseInt(urlBlock, 10, 64)
		if err != nil {
			log.Println("ERROR: invalid block height specified" + " -- Go Error:" ,err)
			http.Error(w, "ERROR: invalid block height specified \n"+"Please chose a number like '0' for the genesis block or '444' for block 444", 404)
			return
		}

		log.Println("Block converted to int64", blockHeight)
		blockHash, err := coinClient.GetBlockHash(blockHeight)
		if err != nil {
			log.Println(err)
			http.Error(w, "ERROR Getting Block Hash from Height: \n"+err.Error(), 500)

		}

		block, err := coinClient.GetBlockVerbose(blockHash)
		if err != nil {
			log.Println(err)
			http.Error(w, "ERROR Getting Block from Block Hash:  "+err.Error(), 500)
		}


		jsonBlock, err := json.Marshal(&block)
		data := json.RawMessage(jsonBlock)
		json.NewEncoder(w).Encode(data)
	}
}



//GetTX
func GetTX(w http.ResponseWriter, r *http.Request) {

	request := r.URL.Path
	request = strings.TrimPrefix(request, "/tx/")

	log.Println("Parsed txid from request ", request)

	requestHash, err := chainhash.NewHashFromStr(request)
	if err != nil {
		log.Println("ERROR:", err)
		return

	}

	txhash, err := coinClient.GetRawTransactionVerbose(requestHash)
	if err != nil {
		log.Println("ERROR:", err)
		http.Error(w, "ERROR: invalid transaction id specified \n"+err.Error(), 404)
		return

	}

	json.NewEncoder(w).Encode(txhash)

}



//GetBlockchainInfo
func GetBlockchainInfo(w http.ResponseWriter, r *http.Request) {

	getblockchaininfo, err := coinClient.GetBlockChainInfo()
	if err != nil {
		log.Println("ERROR: ", err)
		http.Error(w, "ERROR: \n"+ err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(getblockchaininfo)

}



/// CoinCodex.com API for prices

type coincodexapi struct {
	Symbol 					string	`json:"symbol"`
	CoinName 				string  `json:"coin_name"`
	LastPrice				float64	`json:"last_price_usd"`
	Price_TodayOpenUSD		float64	`json:"today_open"`
	Price_HighUSD			float64	`json:"price_high_24_usd"`
	Price_LowUSD			float64	`json:"price_low_24_usd"`
	Volume24USD				float64	`json:"volume_24_usd"`
	DataProvider			string	`json:"data_provider"`
}

func GetCoinCodexData(w http.ResponseWriter, r *http.Request) {
	if configFile.EnableCoinCodexAPI == false {
		return
	} else {

		url := "https://coincodex.com/api/coincodex/get_coin/"+configFile.Ticker

		client := http.Client{
			Timeout:time.Second * 5, // 5 second timeout
		}

		request, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			log.Println("ERROR: ", err)
		}
		request.Header.Set("User-Agent", "capi v0.1")

		response, getError := client.Do(request)
		if getError != nil {
			log.Println("ERROR: ", getError)

		}
		body, readError := ioutil.ReadAll(response.Body)
		if readError != nil {
			log.Println("ERROR:", readError)
		}

		jsonData := coincodexapi{
			DataProvider:"CoinCodex.com",
		}
		jsonError := json.Unmarshal(body, &jsonData)
		if jsonError != nil {
			log.Println("ERROR: ", jsonError)
		}

		json.NewEncoder(w).Encode(jsonData)
	}
}


func IndexRoute (w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.ParseFiles("templates/index.tmpl")
		if err != nil {
			log.Println("ERROR: Parsing template file index.tmpl", err)
		}


	url := "https://coincodex.com/api/coincodex/get_coin/"+configFile.Ticker

	client := http.Client{
		Timeout:time.Second * 5, // 5 second timeout
	}

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Println("ERROR: ", err)
	}
	request.Header.Set("User-Agent", "capi v0.1")

	response, getError := client.Do(request)
	if getError != nil {
		log.Println("ERROR: ", getError)

	}
	body, readError := ioutil.ReadAll(response.Body)
	if readError != nil {
		log.Println("ERROR:", readError)
	}

	jsonData := coincodexapi{
		DataProvider:"CoinCodex.com",
	}
	jsonError := json.Unmarshal(body, &jsonData)
	if jsonError != nil {
		log.Println("ERROR: ", jsonError)
	}

	tmpl.Execute(w, jsonData)
}


func GetAllBlocks(w http.ResponseWriter, r *http.Request) {

		db := database.OpenRead()
		defer db.Close()

		blocks, err := GetAll(db)
		if err != nil {
			log.Println("Failed to list blocks from DB", err)
		}
		json.NewEncoder(w).Encode(blocks)

}





/// Block Ranger stuff


var blockBucket = []byte("block")

func AddBlock(db *bolt.DB, height string, block string) error {
	return db.Update(func(tx *bolt.Tx) error {
		bk, err := tx.CreateBucketIfNotExists(blockBucket)
		if err != nil {
			return fmt.Errorf("Failed to create bucket: %v", err)
		}

		if err := bk.Put([]byte(height), []byte(block)); err != nil {
			return fmt.Errorf("Failed to insert '%s': %v", block, err)
		}
		return nil
	})
}


var ErrNoBucket = fmt.Errorf("failed to find bucket")

func GetAll(db *bolt.DB) ([]Block, error) {
	var blocks []Block
	err := db.View(func(tx *bolt.Tx) error {
		bk := tx.Bucket([]byte(blockBucket))
		if bk == nil {
			fmt.Print(ErrNoBucket, "failed to get 'Blocks' bucket")
		}

		c := bk.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			buf := bytes.NewBuffer(k)
			Blockint64, err := binary.ReadVarint(buf)
			if err != nil{
				log.Println(err)
			}
			blocks = append(blocks, Block{
				Hash:              string(v),
				Confirmations:     0,
				Size:              0,
				StrippedSize:      0,
				Weight:            0,
				Height:            Blockint64,
				Version:           0,
				VersionHex:        "",
				MerkleRoot:        "",
				BlockTransactions: nil,
				Time:              0,
				Nonce:             0,
				Bits:              "",
				Difficulty:        0,
				PreviousHash:      "",
				NextHash:          "",
			})
			//log.Println(string(k),string(v))

		}
		return nil

	})
	return blocks, err
}


func GoBlockRanger() {

	// Get the current block count.
	blockCount, err := coinClient.GetBlockCount()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Block count: %d", blockCount)

	startIndex := int64(0)
	endIndex := int64(blockCount)

	if (blockCount > 500) {
		endIndex = int64(499);
	}
	BlockRanger(coinClient, startIndex, endIndex, blockCount);

	log.Println("All Done! Block Hight is ", blockCount)
}


func BlockRanger (client *rpcclient.Client, startIndex int64, endIndex int64, blockCount int64) {

	//For each item (block) in the array index, print block details
	log.Println("starting block ranger at", strconv.FormatInt(startIndex,10) )

	logIndex := startIndex
	blockArray := make([]Block, 500)
	blockArrayIndex := 0
	for ; startIndex <= endIndex; startIndex++{
		blockHash, err := client.GetBlockHash(startIndex)
		if err != nil {
			log.Println("Error getting block hash from height ", err)
		}

		block, err := client.GetBlockVerbose(blockHash)
		if err != nil {
			log.Println("Error getting block hash ", err)
		}

		blockArray[blockArrayIndex] = Block{
			Hash:              block.Hash,
			Confirmations:     block.Confirmations,
			Size:              block.Size,
			StrippedSize:      block.StrippedSize,
			Weight:            block.Weight,
			Height:            block.Height,
			Version:           block.Version,
			VersionHex:        block.VersionHex,
			MerkleRoot:        block.MerkleRoot,
			BlockTransactions: block.Tx,
			Time:              block.Time,
			Nonce:             block.Nonce,
			Bits:              block.Bits,
			Difficulty:        block.Difficulty,
			PreviousHash:      block.PreviousHash,
			NextHash:          block.NextHash,
		}
		blockArrayIndex = blockArrayIndex +1

	}
	for block, BlockArray := range blockArray {
		key := strconv.FormatInt(BlockArray.Height, 10)
		value := BlockArray.Hash
		db := database.OpenWrite()
		defer db.Close()
		if err := AddBlock(db, key, value); err != nil {
			log.Println("Failed to insert new item: %v\n", err, block)
			os.Exit(1)
		}
		log.Println("A new item has been inserted   HEIGHT: ", key,"HASH:", value)
		db.Close()
	}
	log.Println("Batch completed from:", logIndex, " to ", endIndex)
	log.Println("Sending to DB")

	if (endIndex < blockCount) {
		newEndIndex := int64(0)
		if ((endIndex + 500) > blockCount) {
			newEndIndex = blockCount
		} else {
			newEndIndex = endIndex + 500
		}
		BlockRanger(client, endIndex + 1, newEndIndex, blockCount)
	}

}





