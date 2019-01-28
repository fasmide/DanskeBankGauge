package main

import (
	"log"
	"net/http"

	"github.com/spf13/viper"
)

func init() {
	viper.SetDefault("listen", "localhost:1234")
}

func main() {

	viper.AutomaticEnv()

	// lets just go with the default servemux today
	http.Handle("/balance", NewDaemon())

	log.Printf("Listening on %s", viper.GetString("listen"))
	http.ListenAndServe(viper.GetString("listen"), nil)

}
