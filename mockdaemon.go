package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// MockDaemon mocks a balance for hardware testing
type MockDaemon struct {
	sync.Mutex
	balance Balance

	// true for up, false for down
	direction bool
}

// NewMockDaemon returns a new http handler
func NewMockDaemon() *MockDaemon {
	d := MockDaemon{direction: true}
	go d.loop()
	return &d
}

func (d *MockDaemon) loop() {
	for {
		d.Lock()
		if d.direction {
			// up
			d.balance.Balance += 500
		} else {
			// down
			d.balance.Balance -= 500
		}

		// hardware scale is 0..10000 - and it should be able to handle
		// offscale values without exploding
		if d.balance.Balance >= 13500 || d.balance.Balance <= -2000 {
			d.direction = !d.direction
		}

		d.Unlock()
		time.Sleep(time.Second * 1)

	}
}

func (d *MockDaemon) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	d.Lock()
	defer d.Unlock()

	response, err := json.Marshal(d.balance)
	if err != nil {
		log.Printf("unable to marshal balance: %s", err)
		http.Error(w, fmt.Sprintf("unable to marshal balance: %s", err.Error()), http.StatusInternalServerError)
	}

	w.Write(response)
}
