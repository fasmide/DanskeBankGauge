package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/carlescere/scheduler"
	"github.com/fasmide/DanskeBankGauge/danskebank"
	"github.com/fasmide/DanskeBankGauge/nodejs"
	"github.com/spf13/viper"
)

// Daemon caches balance and initiates communication to the bank when needed
type Daemon struct {
	sync.Mutex
	balance *Balance
}

// NewDaemon returns a initialized daemon
func NewDaemon() *Daemon {
	d := Daemon{}

	// these are the 3 times a day we would like updated data
	scheduler.Every().Day().At("07:00").Run(d.Clear)
	scheduler.Every().Day().At("13:00").Run(d.Clear)
	scheduler.Every().Day().At("19:00").Run(d.Clear)

	// check required vars is available
	if !viper.IsSet("ssn") ||
		!viper.IsSet("sc") ||
		!viper.IsSet("ibmid") ||
		!viper.IsSet("ibmsecret") ||
		!viper.IsSet("accountno") {
		panic("required vars not set, need ssn, sc, ibmid, ibmsecret and accountno")
	}

	return &d
}

// Clear clears the balance
func (d *Daemon) Clear() {
	d.Lock()
	d.balance = nil
	d.Unlock()
	log.Printf("Cleared balance cache")
}

func (d *Daemon) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	d.Lock()
	defer d.Unlock()

	// if ?reload is set to something - reload balance
	q := r.URL.Query()
	if q.Get("reload") != "" {
		log.Printf("Reload set, refreshing balance")
		err := d.fetch()
		if err != nil {
			err = fmt.Errorf("could not fetch balance: %s", err)
			log.Print(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// if the balance is nil - reload balance
	if d.balance == nil {
		log.Printf("balance timed out, refreshing balance")
		err := d.fetch()
		if err != nil {
			err = fmt.Errorf("could not fetch balance: %s", err)
			log.Print(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	payload, err := json.Marshal(d.balance)
	if err != nil {
		http.Error(w, fmt.Sprintf("unable to marshal json: %s", err), http.StatusInternalServerError)
	}

	headers := w.Header()
	headers["Content-Type"] = []string{"application/json"}

	_, err = w.Write(payload)
	if err != nil {
		log.Printf("could not write response to client: %s", err)
	}
}

func (d *Daemon) fetch() error {
	db := danskebank.Client{
		IbmID:     viper.GetString("ibmid"),
		IbmSecret: viper.GetString("ibmsecret"),
		Evaluator: nodejs.Eval,
	}

	err := db.Logon(viper.GetString("ssn"), viper.GetString("sc"))
	if err != nil {
		return fmt.Errorf("unable to logon: %s", err)
	}

	accounts, err := db.AccountList()
	if err != nil {
		return fmt.Errorf("unable to receive accounts: %s", err)

	}
	log.Printf("we have accounts: %+v", accounts)

	err = db.Logoff()
	if err != nil {
		// we are not returning an error here - just logging we failed to logoff
		log.Printf("unable to logoff: %s", err)
	}

	// populate d.balance
	var found bool
	for _, account := range accounts {
		if account.AccountNoExt == viper.GetString("accountno") {
			// we found the account we where looking for
			d.balance = &Balance{Balance: account.Balance}
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("unable to find account %s", viper.GetString("accountno"))
	}

	return nil

}
