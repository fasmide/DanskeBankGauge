package main

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/fasmide/ProjectDanskeBankGauge/nodejs"
)

func main() {
	n, err := nodejs.NewNodejs()
	if err != nil {
		panic(err)
	}

	n.Write([]byte("console.log('her er javascript', true, 3245); process.exit(0);"))
	n.Close()

	result, _ := ioutil.ReadAll(n)
	log.Printf("output: %s", result)

	result, _ = ioutil.ReadAll(n.Stderr)

	log.Printf("err: %s", result)

	http.ListenAndServe("localhost:1234", nil)

}
