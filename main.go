// Copyright (c) 2017 Philipp Weber
// Use of this source code is governed by the MIT license
// which can be found in the repositorys LICENSE file.

package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	// token is used to authenticate at SPDyn API
	token string
	// hostname contains host which will be updated
	hostname string
)

// transport includes multiple timeouts for reliable http clients.
var transport = &http.Transport{
	Dial: (&net.Dialer{
		Timeout: 5 * time.Second,
	}).Dial,
	TLSHandshakeTimeout: 5 * time.Second,
}

// client contains default client for http requests with timeouts.
var client = &http.Client{
	Timeout:   time.Second * 10,
	Transport: transport,
}

// defined errors for logging
var (
	errNoEnvFlags         = errors.New("env SPDYNTOKEN and SPDYNHOSTNAME not set, command line flags for token and hostname not specified")
	errNoIPFound          = errors.New("no public IP found")
	errSPDynAbuse         = errors.New("host is locked because of too many failed updates")
	errSPDynBadauth       = errors.New("given username or password / token is not accepted")
	errSPDynNotYours      = errors.New("given host could not be managed by your account")
	errSPDynNotFQDN       = errors.New("given host is not a FQDN")
	errSPDynNumHost       = errors.New("tried to update more then 20 hosts in one query")
	errSPDynNoChg         = errors.New("the IP has not changed since last update")
	errSPDynNoHost        = errors.New("given host does not exist or was deleted")
	errSPDynFatal         = errors.New("given host was manually deactivated")
	errSPDynUnknown       = errors.New("received unknown response code from SPDyn")
	errSPDynEmptyResponse = errors.New("received an empty response from SPDyn")
)

// IPs contains slice of detected ips
type IPs struct {
	IPInfo []struct {
		IP     string `json:"ip"`
		Source string `json:"source"`
	} `json:"ipinfo"`
}

func getPublicIP() (string, error) {
	// http://wiki.securepoint.de/index.php/SPDyn/meineIP
	var ips IPs
	res, err := client.Get("http://checkip.spdyn.de/json")
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(body, &ips)
	if err != nil {
		return "", err
	}

	// search for REMOTE_ADDR
	for _, ip := range ips.IPInfo {
		if ip.Source == "REMOTE_ADDR" {
			return ip.IP, nil
		}
	}

	return "", errNoIPFound
}

func updateIP(ip string) error {
	// http://wiki.securepoint.de/index.php/SPDyn/Hostverwenden
	req, err := http.NewRequest("GET", "https://update.spdyn.de/nic/update", nil)
	if err != nil {
		return err
	}
	q := req.URL.Query()
	q.Add("hostname", hostname)
	q.Add("myip", ip)
	q.Add("user", hostname)
	q.Add("pass", token)
	req.URL.RawQuery = q.Encode()
	res, err := client.Get(req.URL.String())
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return fmt.Errorf("%v: Server status %s", errSPDynEmptyResponse, res.Status)
	}
	resCode := strings.Fields(string(body))[0]

	// http://wiki.securepoint.de/index.php/SPDyn/Rückgabecodes
	switch resCode {
	case "abuse":
		return errSPDynAbuse
	case "badauth":
		return errSPDynBadauth
	case "good":
		return nil
	case "!yours":
		return errSPDynNotYours
	case "notfqdn":
		return errSPDynNotFQDN
	case "numhost":
		return errSPDynNumHost
	case "nochg":
		return errSPDynNoChg
	case "nohost":
		return errSPDynNoHost
	case "fatal":
		return errSPDynFatal
	default:
		return fmt.Errorf("%v: %s", errSPDynUnknown, resCode)
	}
}

func init() {
	// read in flags
	var flagToken, flagHostname string
	flag.StringVar(&flagToken, "token", "", "SPDyn API token")
	flag.StringVar(&flagHostname, "host", "", "hostname to update")
	flag.Parse()

	// read in env variables
	envToken := os.Getenv("SPDYNTOKEN")
	envHostname := os.Getenv("SPDYNHOSTNAME")

	if envToken != "" && envHostname != "" {
		token = envToken
		hostname = envHostname
	} else if flagToken != "" && flagHostname != "" {
		token = flagToken
		hostname = flagHostname
	} else {
		log.Fatal(errNoEnvFlags)
	}
}

func main() {
	myIP, err := getPublicIP()
	if err != nil {
		log.Fatal(err)
	}

	err = updateIP(myIP)
	if err != nil {
		if err.Error() == errSPDynNoChg.Error() {
			log.Printf("%v: %s - %s\n", err, hostname, myIP)
			os.Exit(0)
		}

		log.Fatal(err)
	}

	log.Printf("IP successfully updated: %s - %s\n", hostname, myIP)
}
