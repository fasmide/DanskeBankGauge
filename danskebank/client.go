package danskebank

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"text/template"
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

// JavascriptEvaluator should read javascript and return results
type JavascriptEvaluator func(io.Reader) ([]byte, error)

// NewRequest initializes requests with required headers
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

	// the endpoint does not seem to care about these
	// r.Header["x-app-version"] = []string{"MobileBank android DK 1201367"}
	// r.Header["referer"] = []string{"MobileBanking3 DK"}
	// r.Header["x-app-culture"] = []string{"da-DK"}
	// r.Header.Set("User-Agent", "okhttp/3.11.0")

	return r, nil
}

// SignerURL is used to fetch "javascript sealer" - some obfuscated javascript
// providing a performLogonServiceCode_v2 which takes social security number and
// a service code and provides a LogonPackage which must be posted to the `LogonURL`
const SignerURL = "https://apiebank.danskebank.com/ebanking/ext/Functions?stage=LogonStep1&secsystem=SC&brand=DB&channel=MOB"

// LogonURL is used to post the result of the above sealer
const LogonURL = "https://apiebank.danskebank.com/ebanking/ext/logon"

// Logon creates a new session with the mobile api
// 	This is quite a long process, involving:
//	* Fetching some javascript from their api (they seem to call it Javascript Sealer)
// 	* Evaluating this javascript
// 	* Using a method from the evaluated javascript to create a LogonPackage
//	* Posting the logonpackage to their api, to receive an auth token
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

	// save temp auth token, it must be used on the next request
	tempAuthToken := resp.Header.Get("Persistent-Auth")
	if tempAuthToken == "" {
		return fmt.Errorf("when fetching javascript sealer, there was no auth token")
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

	// parse the received logon package
	logonPackage := LogonPackage{}
	err = json.Unmarshal(result, &logonPackage)
	if err != nil {
		return fmt.Errorf("unable to parse json from sealer: %s", err)
	}

	// try to figure out if the package is valid
	if !logonPackage.Valid() {
		return fmt.Errorf("LogonPackage from sealer was not valid: %s", result)
	}

	payload, err := json.Marshal(logonPackage)
	if err != nil {
		return fmt.Errorf("unable to marshal logonpackage: %s", err)
	}

	// so far so good, we should now take the above logon package and post it to LogonURL
	req, err = c.NewRequest(http.MethodPost, LogonURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("unable create post logonpackage request: %s", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header["authorization"] = []string{tempAuthToken}

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to post logon package: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("unexpected statuscode when posting logonpackage (%d): %s", resp.StatusCode, body)
	}

	// at this point we should have a working auth token
	c.auth = resp.Header.Get("Persistent-Auth")

	if c.auth == "" {
		return fmt.Errorf("we failed somewhere - you will never see this error :)")
	}
	return nil
}

// AccountListURL is the url to POST when asking for accounts
const AccountListURL = "https://apiebank.danskebank.com/ebanking/ext/e4/account/list"

// AccountListBody is posted to the list URL - not sure if explicitly required
const AccountListBody = "{\n    \"languageCode\": \"DA\"\n}"

// AccountList lists all accounts
func (c *Client) AccountList() ([]Account, error) {

	if c.auth == "" {
		return nil, fmt.Errorf("This client has not been logged on yet")
	}

	req, err := c.NewRequest(http.MethodPost, AccountListURL, bytes.NewBuffer([]byte(AccountListBody)))
	if err != nil {
		return nil, fmt.Errorf("unable to create account list request: %s", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to post request to accountlist endpoint: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected statuscode from accountlist endpoint (%d): %s", resp.StatusCode, body)
	}

	// i have not seen this change between requests but one could imagine its supposed to - to fight replay attacks
	if c.auth != resp.Header.Get("Persistent-Auth") {
		c.auth = resp.Header.Get("Persistent-Auth")
	}

	accountResponse := AccountListResponse{}
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&accountResponse)
	if err != nil {
		return nil, fmt.Errorf("unable to parse json from account list: %s", err)
	}

	return accountResponse.Accounts, nil
}

// LogoffURL is the url to POST when signing off
const LogoffURL = "https://apiebank.danskebank.com/ebanking/ext/logoff"

// Logoff closes a session
func (c *Client) Logoff() error {
	req, err := c.NewRequest(http.MethodPost, LogoffURL, bytes.NewBuffer([]byte("{}")))
	if err != nil {
		return fmt.Errorf("unable to create logoff request: %s", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable send logoff request: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		return fmt.Errorf("unexpected statuscode when logging off (%d): %s", resp.StatusCode, body)
	}

	c.auth = ""
	return nil
}
