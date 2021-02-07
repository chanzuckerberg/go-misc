package databricks

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const (
	APIVersion = "2.0" //APIVersion is the version of the RESTful API of DataBricks
)

// DBClientOption is used to configure the DataBricks Client
type DBClientOption struct {
	User           string
	Password       string
	Host           string
	Token          string
	DefaultHeaders map[string]string
	TimeoutSeconds int
	client         http.Client
}

// Init initializes the client
func (o *DBClientOption) Init() {
	if o.TimeoutSeconds == 0 {
		o.TimeoutSeconds = 10
	}

	o.client = http.Client{
		Timeout: time.Duration(o.TimeoutSeconds) * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
		},
	}
}

func (o *DBClientOption) getHTTPClient() http.Client {
	return o.client
}

func (o *DBClientOption) getAuthHeader() map[string]string {
	auth := make(map[string]string)

	if o.User != "" && o.Password != "" {
		encodedAuth := []byte(o.User + ":" + o.Password)
		userHeaderData := "Basic " + base64.StdEncoding.EncodeToString(encodedAuth)
		auth["Authorization"] = userHeaderData
		auth["Content-Type"] = "application/json"
	} else if o.Token != "" {
		auth["Authorization"] = "Bearer " + o.Token
		auth["Content-Type"] = "application/json"
	}

	return auth
}

func (o *DBClientOption) getUserAgentHeader() map[string]string {
	return map[string]string{
		"User-Agent": "go-misc:databricks",
	}
}

func (o *DBClientOption) getDefaultHeaders() map[string]string {
	auth := o.getAuthHeader()
	userAgent := o.getUserAgentHeader()

	defaultHeaders := make(map[string]string)
	for k, v := range auth {
		defaultHeaders[k] = v
	}

	for k, v := range o.DefaultHeaders {
		defaultHeaders[k] = v
	}

	for k, v := range userAgent {
		defaultHeaders[k] = v
	}

	return defaultHeaders
}

func (o *DBClientOption) getRequestURI(path string) (string, error) {
	parsedURI, err := url.Parse(o.Host)
	if err != nil {
		return "", err
	}

	requestURI := fmt.Sprintf("%s://%s/api/%s%s", parsedURI.Scheme, parsedURI.Host, APIVersion, path)

	return requestURI, nil
}
