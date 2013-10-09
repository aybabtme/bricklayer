package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/aybabtme/bricklayer/util"
	"github.com/aybabtme/color/brush"
	"github.com/aybabtme/dskvs"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"strconv"
)

const (
	defaultPort = 3000
	partsPath   = "parts"
	indexPath   = "index"
	partsIndex  = "parts"
)

var (
	port   *int
	doSeed *bool

	sout *log.Logger
	serr *log.Logger

	dbPath = "db"
	db     *dskvs.Store
)

func init() {
	port = flag.Int("port", defaultPort, "port on which to listen for request")
	doSeed = flag.Bool("seedDB", false, "seed the DB with data from the iGem server")
	flag.Parse()
}

func main() {

	sout = log.New(os.Stdout, "["+brush.Green("OK ").String()+"]\t", log.LstdFlags)
	serr = log.New(os.Stderr, "["+brush.Red("ERR").String()+"]\t", log.LstdFlags)

	sout.Printf("opening DB at path '%s'", dbPath)
	store, err := dskvs.Open(dbPath)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	db = store

	if *doSeed {
		sout.Printf("seeding DB with iGem API")
		if err := seedDB(db); err != nil {
			panic(err)
		}
	}

	sout.Printf("registering HTTP endpoints")

	router := mux.NewRouter()
	router.HandleFunc(fmt.Sprintf("/api/%s/", partsPath), AllPartsHandler)
	router.HandleFunc(fmt.Sprintf("/api/%s/{name}", partsPath), PartsHandler)
	http.Handle("/", router)

	listenAddr := fmt.Sprintf(":%d", *port)
	sout.Printf("listening on %s", brush.Blue(listenAddr))
	if err := http.ListenAndServe(listenAddr, router); err != nil {
		panic(err)
	}
}

func seedDB(db *dskvs.Store) error {
	allBricks, err := util.DownloadAllBiobricks()
	if err != nil {
		return fmt.Errorf("failed to download parts in seed of DB, %v", err)
	}

	allParts, err := db.GetAll(partsPath)
	if err != nil {
		return fmt.Errorf("could not verify if DB is empty, %v", err)
	}

	if len(allParts) != 0 {
		err = db.DeleteAll(partsPath)
		if err != nil {
			return fmt.Errorf("could not delete all parts before seeding DB, %v", err)
		}
	}

	// Optimization for index query
	allBioPartName := make([]string, len(allBricks))
	for i, biobrick := range allBricks {
		brickData, err := json.Marshal(&biobrick)
		if err != nil {
			return fmt.Errorf("could not get JSON from biobrick '%s', %v", biobrick.PartName, err)
		}

		err = db.Put(fmt.Sprintf("%s/%s", partsPath, biobrick.PartName), brickData)
		if err != nil {
			return fmt.Errorf("could not persist biobrick '%s' to DB, %v", biobrick.PartName, err)
		}
		allBioPartName[i] = biobrick.PartName
	}

	// Save the extracted index
	indexData, err := json.Marshal(allBioPartName)
	if err != nil {
		return fmt.Errorf("could not serialize all part names, %v", err)
	}
	return db.Put(fmt.Sprintf("%s/%s", indexPath, partsIndex), indexData)
}

func getPort() int {

	if port != nil {
		return *port
	}

	envPortStr := os.Getenv("BRICKLAYER_PORT")
	if envPortStr == "" {
		envPort, err := strconv.Atoi(envPortStr)
		if err == nil {
			return envPort
		}
		log.Printf("error parsing BRICKLAYER_PORT variable, %v, falling back to default %d", err, defaultPort)
	}

	return defaultPort
}
