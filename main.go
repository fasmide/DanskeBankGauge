package main

import (
	"log"
	"net/http"

	"github.com/fasmide/DanskeBankGauge/danskebank"
	"github.com/fasmide/DanskeBankGauge/nodejs"
	"github.com/spf13/viper"
)

func init() {
	viper.SetDefault("listen", "localhost:1234")
}

func main() {

	viper.AutomaticEnv()

	db := danskebank.Client{
		IbmID:     viper.GetString("ibmid"),
		IbmSecret: viper.GetString("ibmsecret"),
		Evaluator: nodejs.Eval,
	}

	err := db.Logon("2020202121", "1234")
	if err != nil {
		log.Printf("unable to logon: %s", err)
	}
	// lets just go with the default servemux today
	log.Printf("Listening on %s", viper.GetString("listen"))
	http.ListenAndServe(viper.GetString("listen"), nil)

}
