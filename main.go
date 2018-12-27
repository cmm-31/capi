package main

import (

	"log"
	"github.com/gorilla/mux"
	"net/http"
	"capi/api"
)


var configFile  = api.LoadConfig("./config/config.json")



func main() {

	log.Println("Running Block Ranger")
	//api.GoBlockRanger()

	router := mux.NewRouter()
	router.HandleFunc("/test", api.GetBlockBytes).Methods("GET")
	router.HandleFunc("/block/{HeightOrHash}", api.GetBlock).Methods("GET")
	router.HandleFunc("/tx/{tx}", api.GetTX).Methods("GET")
	router.HandleFunc("/market", api.GetCoinCodexData).Methods("GET")
	router.HandleFunc("/blockchaininfo", api.GetBlockchainInfo).Methods("GET")
	router.HandleFunc("/getallblocks", api.GetAllBlocks).Methods("GET")
	router.HandleFunc("/", api.IndexRoute).Methods("GET")
	log.Println("capi v0.1 is running!")
	log.Fatal(http.ListenAndServe(configFile.CapiPort, router))


}
