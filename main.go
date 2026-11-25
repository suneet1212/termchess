package main

import (
	"flag"
	"fmt"
	"termchess/src"

	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("Starting Termchess")
	godotenv.Load()

	isServerPtr := flag.Bool("server", false, "is server true")
	flag.Parse()

	if *isServerPtr {
		src.StartServer()
	} else {
		src.StartClient()
	}
}
