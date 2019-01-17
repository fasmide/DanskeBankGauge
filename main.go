package main

import (
	"fmt"
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

	err := db.Logon(viper.GetString("ssn"), viper.GetString("sc"))
	if err != nil {
		panic(fmt.Sprintf("unable to logon: %s", err))
	}

	accounts, err := db.AccountList()
	if err != nil {
		panic(fmt.Sprintf("unable to receive accounts: %s", err))

	}
	log.Printf("we have accounts: %+v", accounts)

	err = db.Logoff()
	if err != nil {
		log.Printf("unable to logoff: %s", err)
	}
	log.Printf("session closed")
	// lets just go with the default servemux today
	log.Printf("Listening on %s", viper.GetString("listen"))
	http.ListenAndServe(viper.GetString("listen"), nil)

}
