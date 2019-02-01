package main

import (
	"encoding/json"
	"time"

	"github.com/spf13/viper"
)

func init() {
	// lets a default monthy allowance of 10k
	viper.SetDefault("monthallowance", 10000)
}

// Balance is the json we will be sending to the gauge
type Balance struct {
	Balance float64
}

// MarshalJSON will add a allowance field to the payload
// this is a calculated field
func (b Balance) MarshalJSON() ([]byte, error) {
	return json.Marshal(
		struct {
			Balance   float64 `json:"balance"`
			Allowance float64 `json:"allowance"`
		}{
			Balance:   b.Balance,
			Allowance: b.Allowance(),
		},
	)
}

// Allowance calculates how much money you are allowed to use
// today by forecasting how much money is available each day in the month
func (b *Balance) Allowance() float64 {
	// how much are we allowed to spend pr day
	dayAllowance := viper.GetInt("monthallowance") / daysInCurrentMonth()
	allowedUseToday := time.Now().Day() * dayAllowance

	return b.Balance - (float64(allowedUseToday)-viper.GetFloat64("monthallowance"))*-1
}

func daysInCurrentMonth() int {
	now := time.Now()
	return time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
