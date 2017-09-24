package main

import (
	"bitbucket.org/pdedkov/rachek"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	var (
		config = flag.String("config", fmt.Sprintf("%s%s.rachek.toml", os.Getenv("HOME"), string(os.PathSeparator)), "config path")
		port   = flag.String("port", "8080", "Server http port")
	)
	flag.Parse()

	d, err := rachek.NewDaemon(*config)
	if err != nil {
		panic(err)
	}

	log.Fatal(http.ListenAndServe(":"+*port, d))
}
