package danskebank

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

// Client represents a session with the mobile API
type Client struct {
	auth string

	// Evaluator is needed to evaluate javascript when logging in
	Evaluator JavascriptEvaluator

	// IbmID sets the x-ibm-client-id header
	IbmID string
	// IbmSecret sets the x-ibm-client-secret header
	IbmSecret string
}

// JavascriptEvaluator should accept javascript written to it and
// allow reading of result
type JavascriptEvaluator func([]byte) ([]byte, error)

// SignerURL is used to fetch "javascript sealer" - some obfuscated javascript
// providing a performLogonServiceCode_v2 which takes social security number and
// a service code and provides a LogonPackage which must be posted to the `LogonURL`
const SignerURL = "https://apiebank.danskebank.com/ebanking/ext/Functions?stage=LogonStep1&secsystem=SC&brand=DB&channel=MOB"

// LogonURL is used to post the result of the above sealer
const LogonURL = "https://apiebank.danskebank.com/ebanking/ext/logon"

// Logon creates a new session with the mobile api
func (c *Client) Logon(cpr, sc string) error {
	if c.auth != "" {
		return fmt.Errorf("this client is already logged on")
	}

	req, err := http.NewRequest(http.MethodGet, "http://localhost", nil)
	if err != nil {
		return fmt.Errorf("unable to create request: %s", err)
	}

	c.setHeaders(req)
	// add content-type to default headers
	req.Header["content-type"] = []string{"text/plain"}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to fetch javascript sealer: %s", err)
	}
	defer resp.Body.Close()

	code, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to read sealer: %s", err)
	}

	return nil
}

func (c *Client) setHeaders(r *http.Request) {

	// Not using the Set here to preserve header case
	r.Header["x-ibm-client-id"] = []string{c.IbmID}
	r.Header["x-ibm-client-secret"] = []string{c.IbmSecret}
	r.Header["x-app-version"] = []string{"MobileBank android DK 1201367"}
	r.Header["referer"] = []string{"MobileBanking3 DK"}
	r.Header["x-app-culture"] = []string{"da-DK"}

	r.Header.Set("User-Agent", "okhttp/3.11.0")

}
