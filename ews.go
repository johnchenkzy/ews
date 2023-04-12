package ews

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strconv"

	httpntlm "github.com/vadimi/go-http-ntlm/v2"
)

const (
	soapStart = `<?xml version="1.0" encoding="utf-8" ?>
<soap:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" 
		xmlns:m="http://schemas.microsoft.com/exchange/services/2006/messages" 
		xmlns:t="http://schemas.microsoft.com/exchange/services/2006/types" 
		xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  		<soap:Header>
    		<t:RequestServerVersion Version="Exchange2010_SP2" />
  		</soap:Header>
  		<soap:Body>
`
	soapEnd = `
</soap:Body></soap:Envelope>`
)

type Config struct {
	Dump    bool
	NTLM    bool
	SkipTLS bool
	Http2   bool
}

// DefaultConfig sets the default configuration
// use http1.1 instead of http2
func GetDefaultConfig() Config {
	c := Config{}
	c.Dump = true
	c.NTLM = true
	c.SkipTLS = true
	c.Http2 = false
	return c
}

type Client interface {
	SendAndReceive(body []byte) ([]byte, error)
	GetEWSAddr() string
	GetUsername() string
}

type client struct {
	EWSAddr  string
	Domain   string
	Username string
	Password string
	config   *Config
}

func (c *client) GetEWSAddr() string {
	return c.EWSAddr
}

func (c *client) GetUsername() string {
	return c.Username
}

func NewClient(ewsAddr, username, password string, config *Config) Client {
	return &client{
		EWSAddr:  ewsAddr,
		Username: username,
		Password: password,
		config:   config,
	}
}

func (c *client) SendAndReceive(body []byte) ([]byte, error) {

	bb := []byte(soapStart)
	bb = append(bb, body...)
	bb = append(bb, soapEnd...)

	req, err := http.NewRequest("POST", c.EWSAddr, bytes.NewReader(bb))
	if err != nil {
		return nil, err
	}
	defer req.Body.Close()
	req.Header.Set("Content-Type", "text/xml; charset=UTF-8")
	req.Header.Set("User-Agent", "ExchangeServicesClient/0.0.0.0")
	req.Header.Set("Accept", "text/xml")
	req.Header.Set("Keep-Alive", "300")
	req.Header.Set("Connection", "Keep-Alive")
	req.Header.Set("Content-Length", strconv.Itoa(len(bb)))
	logRequest(c, req)

	req.SetBasicAuth(c.Username, c.Password)

	var client *http.Client

	if c.config.Http2 {
		client = &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	} else {
		client = &http.Client{
			Transport: &http.Transport{
				TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
			},
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	}

	applyConfig(client, c)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	logResponse(c, resp)

	if resp.StatusCode != http.StatusOK {
		return nil, NewError(resp)
	}

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return respBytes, err
}

func applyConfig(client *http.Client, c *client) {
	config := c.config
	if config.NTLM {
		client.Transport = &httpntlm.NtlmTransport{
			Domain:   c.Domain,
			User:     c.Username,
			Password: c.Password,
			// Configure RoundTripper if necessary, otherwise DefaultTransport is used
			RoundTripper: &http.Transport{
				// provide tls config
				TLSClientConfig: &tls.Config{},
				// other properties RoundTripper, see http.DefaultTransport
			},
		}
	}
	// if config.Https && config.NTLM {
	// 	xprt := client.Transport
	// 	xprt.TLSNextProto = make(map[string]func(authority string, c *tls.Conn) http.RoundTripper)
	// 	client.Transport = &xprt
	// }
	if config.SkipTLS {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
}

func logRequest(c *client, req *http.Request) {
	if c.config != nil && c.config.Dump {
		dump, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("Request:\n%v\n----\n", string(dump))
	}
}

func logResponse(c *client, resp *http.Response) {
	if c.config != nil && c.config.Dump {
		dump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("Response:\n%v\n----\n", string(dump))
	}
}
