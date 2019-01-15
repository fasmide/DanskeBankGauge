package danskebank

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
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
type JavascriptEvaluator func(io.Reader) ([]byte, error)

// SignerURL is used to fetch "javascript sealer" - some obfuscated javascript
// providing a performLogonServiceCode_v2 which takes social security number and
// a service code and provides a LogonPackage which must be posted to the `LogonURL`
// const SignerURL = "https://apiebank.danskebank.com/ebanking/ext/Functions?stage=LogonStep1&secsystem=SC&brand=DB&channel=MOB"
const SignerURL = "http://localhost/signer.js"

// LogonURL is used to post the result of the above sealer
//const LogonURL = "https://apiebank.danskebank.com/ebanking/ext/logon"
const LogonURL = "http://localhost/logon"

// Logon creates a new session with the mobile api
func (c *Client) Logon(cpr, sc string) error {
	if c.auth != "" {
		return fmt.Errorf("this client is already logged on")
	}

	req, err := c.NewRequest(http.MethodGet, SignerURL, nil)
	if err != nil {
		return fmt.Errorf("unable to create request: %s", err)
	}

	// add content-type to default headers
	req.Header["content-type"] = []string{"text/plain"}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to fetch javascript sealer: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code (%d) when fetching sealer: %s", resp.StatusCode, body)
	}

	code, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to read sealer: %s", err)
	}

	// we now have downloaded the javascript sealer, we need to
	// render and evaluate JSTemplate inside the JavascriptEvaluator
	t := template.Must(template.New("jstemplate").Parse(JSTemplate))
	buffer := &bytes.Buffer{}

	t.Execute(buffer, struct {
		Signer, SSN, SC string
	}{
		Signer: string(code),
		SSN:    cpr,
		SC:     sc,
	})

	result, err := c.Evaluator(buffer)
	if err != nil {
		return fmt.Errorf("could not compute logon package: %s", err)
	}
	log.Printf("logonPackage : %s", result)
	return nil
}

// NewRequest initializes a request with required headers
func (c *Client) NewRequest(method string, url string, body io.Reader) (*http.Request, error) {
	r, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	// if we have an auth token, add header
	if c.auth != "" {
		r.Header["authorization"] = []string{c.auth}
	}
	// Not using the Set here to preserve header case
	r.Header["x-ibm-client-id"] = []string{c.IbmID}
	r.Header["x-ibm-client-secret"] = []string{c.IbmSecret}
	r.Header["x-app-version"] = []string{"MobileBank android DK 1201367"}
	r.Header["referer"] = []string{"MobileBanking3 DK"}
	r.Header["x-app-culture"] = []string{"da-DK"}

	r.Header.Set("User-Agent", "okhttp/3.11.0")
	return r, nil
}
