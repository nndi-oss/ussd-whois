package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"bitbucket.org/nndi/phada"
)

var (
	sessionStore           = make(map[string]*phada.UssdRequestSession, 0)
	ErrFailedToCheckDomain = errors.New("Failed to check if domain is available")
	ApiKey                 = "__YOUR_API_KEY__" // changed during+ build with build tags -ldApiKey=""
	bindAddress            string
)

const (
	JSON_WHOIS_API = "https://api.jsonwhois.io/whois/domain"

	STATE_ENTRY        = 1
	STATE_PROMPT       = 2
	STATE_PROMPT_INPUT = 3
	STATE_ABOUT        = 4
	ENTRY_MENU         = `Welcome to USSD WhoIs

1. Check for domain
2. About
`
	CHECK_FOR_DOMAIN_PROMPT = `Check for domain
	
Enter the name of a domain. e.g. mywebsite.com
`
	DOMAIN_ALREADY_REGISTERED_TPL = `%s is already registered.
	
Owner: %s
Updated: %s
Name Servers: %s
`
	DOMAIN_AVAILABLE = `The domain %s is available.
Purchase now before someone else gets their hands on it.
`
	ABOUT_SERVICE = `This service is brought to you by Payrope and NNDI, who specialize in developing SMS, USSD and Web applications.
	
Find us online: https://payrope.com  https://nndi-tech.com
`
)

type JsonWhoIsResult struct {
	Data struct {
		Name        string   `json:"name"`
		Created     string   `json:"created"`
		Changed     string   `json:"changed"`
		Expires     string   `json:"expires"`
		Dnssec      string   `json:"dnssec"`
		Registered  bool     `json:"registered"`
		NameServers []string `json:"nameservers"`
	} `json:"result"`
}

func (j *JsonWhoIsResult) NameServersList() string {
	return strings.Join(j.Data.NameServers, ",")
}

func ussdContinue(text string) string {
	return fmt.Sprintf("CON %s", text)
}

func ussdEnd(text string) string {
	return fmt.Sprintf("END %s", text)
}

func checkDomainAvailability(domainName string) (*JsonWhoIsResult, error) {
	client := &http.Client{}
	whoIsResult := &JsonWhoIsResult{}
	requestUrl, _ := url.Parse(JSON_WHOIS_API)
	q := requestUrl.Query()
	q.Add("key", ApiKey)
	q.Add("domain", domainName)
	requestUrl.RawQuery = q.Encode()
	log.Print("Sending domain request check for domain: " + domainName)
	res, err := client.Get(requestUrl.String())
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(res.Body)
	log.Print(fmt.Sprintf(
		"Got Status: %s for domain request check for: %s: body=%s",
		res.StatusCode,
		domainName,
		data,
	))

	if err != nil {
		return nil, err
	}
	json.Unmarshal(data, whoIsResult)
	if whoIsResult == nil {
		return nil, ErrFailedToCheckDomain
	}
	return whoIsResult, nil
}

func handlerFunc(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	session, err := phada.ParseUssdRequest(req)
	if err != nil {
		log.Print("Failed to parse request to UssdRequestSession")
		fmt.Fprintf(w, ussdEnd("Failed to process request"))
		return
	}

	existingSession, ok := sessionStore[session.SessionID]

	if ok {
		existingSession.RecordHop(session.Text)
		session = existingSession
	} else {
		sessionStore[session.SessionID] = session // store/persist the session
	}

	prevState := session.State
	session.SetState(STATE_ENTRY)

	if prevState == STATE_PROMPT {
		session.SetState(STATE_PROMPT_INPUT)
	}

	if session.ReadIn() == "1" {
		session.SetState(STATE_PROMPT)
	}
	if session.ReadIn() == "2" {
		session.SetState(STATE_ABOUT)
	}

	// fmt.Println("Text", session.ReadIn(), "Current state", session.State, "Previous state", prevState)

	switch session.State {
	case STATE_ENTRY:
		fmt.Fprintf(w, ussdContinue(ENTRY_MENU))
		break
	case STATE_PROMPT:
		fmt.Fprintf(w, ussdContinue(CHECK_FOR_DOMAIN_PROMPT))
		break
	case STATE_PROMPT_INPUT:
		domainName := session.ReadIn()
		domainWhoIs, err := checkDomainAvailability(domainName)
		if err != nil {
			log.Println("Error while processing request", err)
			fmt.Fprintf(w, ussdEnd("Failed to process request"))
			break
		}
		if !domainWhoIs.Data.Registered {
			fmt.Fprintf(w, ussdEnd(fmt.Sprintf(DOMAIN_AVAILABLE, domainName)))
			break
		}
		fmt.Fprintf(w, ussdEnd(fmt.Sprintf(
			DOMAIN_ALREADY_REGISTERED_TPL,
			domainName,
			domainWhoIs.Data.Name,
			domainWhoIs.Data.Changed,
			domainWhoIs.NameServersList(),
		)))
		break
	case STATE_ABOUT:
		fmt.Fprintf(w, ussdEnd(ABOUT_SERVICE))
		break
	default:
		fmt.Fprintf(w, ussdEnd("Failed to process"+session.ReadIn()))
		break
	}

	sessionStore[session.SessionID] = session
}

func init() {
	flag.StringVar(&bindAddress, "b", "localhost:8773", "interface and port to bind server to, e.g. localhost:8080")
}

func main() {
	flag.Parse()
	if bindAddress == "" {
		bindAddress = "localhost:8773"
	}
	http.HandleFunc("/", handlerFunc)
	log.Fatalf("Failed to start server. Error %s", http.ListenAndServe(bindAddress, nil))
}
