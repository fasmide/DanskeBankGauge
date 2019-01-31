package main

import (
	"log"
	"net/http"

	"github.com/spf13/viper"
)

func init() {
	viper.SetDefault("listen", "localhost:1234")
	viper.SetDefault("mock", false)
}

func main() {

	viper.AutomaticEnv()

	// use mock daemon or the real deal
	if viper.GetBool("mock") {
		http.Handle("/balance", NewMockDaemon())
	} else {
		http.Handle("/balance", NewDaemon())
	}

	log.Printf("Listening on %s", viper.GetString("listen"))
	http.ListenAndServe(viper.GetString("listen"), nil)

}
