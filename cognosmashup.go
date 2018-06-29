// Package cognosmashup implements methods for retrieving data from IBM Cognos Mashup Service.
package cognosmashup

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"

	"golang.org/x/net/publicsuffix"
)

// CognosSession ...
type CognosSession struct {
	DispatcherURL          string
	Namespace              string
	Username               string
	Password               string
	CredentialTemplatePath string
	jar                    *cookiejar.Jar
}

// Credentials ...
type Credentials struct {
	Credentialelements []CredentialElement `xml:"credentialElements"`
}

// CredentialElement ...
type CredentialElement struct {
	ActualValue string `xml:"value>actualValue"`
	Name        string `xml:"name"`
	Label       string `xml:"label"`
}

// Report ...
type Report struct {
	DataSet DataSet
}

// DataSet ...
type DataSet struct {
	DataTable []DataTable `json:"dataTable"`
}

// DataTable ...
type DataTable struct {
	ID  string        `json:"id"`
	Row []interface{} `json:"row"`
}

// Logon ...
func (cs *CognosSession) Logon() bool {
	xmlCredentials, err := cs.parseCredentialFile()

	xmlCredentials = url.QueryEscape(xmlCredentials)

	reqStr := cs.DispatcherURL + "/rds/auth/logon?xmlData=" + xmlCredentials

	cs.jar, err = cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})

	client := &http.Client{Jar: cs.jar}
	req, err := http.NewRequest("GET", reqStr, nil)
	resp, err := client.Do(req)

	var retVal bool

	if err == nil && resp.StatusCode == 200 {
		retVal = true
	} else {
		retVal = false
	}

	return retVal
}

// GetReportDataByID ...
func (cs *CognosSession) GetReportDataByID(reportID string, dataSetID int) (bool, *DataTable) {
	reqStr := cs.DispatcherURL + "/rds/reportData/report/" + reportID + "?fmt=DataSetJSON"

	client := &http.Client{Jar: cs.jar}
	req, err := http.NewRequest("GET", reqStr, nil)
	resp, err := client.Do(req)

	var retVal bool

	report := Report{}

	if err == nil && resp.StatusCode == 200 {
		retVal = true

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)

		err = json.Unmarshal(body, &report)

		if err != nil {
			retVal = false
		}
	} else {
		retVal = false
	}

	if err != nil {
		retVal = false
	}

	return retVal, &report.DataSet.DataTable[dataSetID-1]
}

// Logoff ...
func (cs *CognosSession) Logoff() bool {
	reqStr := cs.DispatcherURL + "/rds/auth/logoff"

	client := &http.Client{Jar: cs.jar}
	req, err := http.NewRequest("GET", reqStr, nil)
	resp, err := client.Do(req)

	if err != nil {
		return false
	}

	switch resp.StatusCode {
	case 200:
		return true
	default:
		return false
	}
}

// parseCredentialFile ...
func (cs *CognosSession) parseCredentialFile() (string, error) {
	xmlFile, err := os.Open(cs.CredentialTemplatePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return "Error opening file:", err
	}
	defer xmlFile.Close()

	b, _ := ioutil.ReadAll(xmlFile)

	var c Credentials
	xml.Unmarshal(b, &c)

	for i := range c.Credentialelements {
		switch c.Credentialelements[i].Name {
		case "CAMNamespace":
			c.Credentialelements[i].ActualValue = cs.Namespace
		case "CAMUsername":
			c.Credentialelements[i].ActualValue = cs.Username
		case "CAMPassword":
			c.Credentialelements[i].ActualValue = cs.Password
		}
	}

	x, _ := xml.Marshal(c)

	return string(x), nil
}
