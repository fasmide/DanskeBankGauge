package danskebank

import "testing"

func TestLogon(t *testing.T) {
	db := Client{IbmID: "ibmid", IbmSecret: "ibmsecret"}
	err := db.Logon("2020202121", "1234")
	if err != nil {
		t.Fatalf("unable to logon: %s", err)
	}
}
