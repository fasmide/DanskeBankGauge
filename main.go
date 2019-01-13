package main

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/fasmide/DanskeBankGauge/nodejs"
	"github.com/spf13/viper"
)

func init() {
	viper.SetDefault("listen", "localhost:1234")
}

func main() {

	viper.AutomaticEnv()

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

	// lets just go with the default servemux today
	log.Printf("Listening on %s", viper.GetString("listen"))
	http.ListenAndServe(viper.GetString("listen"), nil)

}
