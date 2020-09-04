package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/lng50k/booster-backend/config"
	"github.com/lng50k/booster-backend/db"
	"github.com/lng50k/booster-backend/server"
)

func main() {
	environment := flag.String("e", "development", "")
	flag.Usage = func() {
		fmt.Println("Usage: server -e {mode}")
		os.Exit(1)
	}
	flag.Parse()
	config.Init(*environment)
	db.Init()
	server.Init()
}
