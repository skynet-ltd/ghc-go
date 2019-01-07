package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"sync"

	"github.com/skynet-ltd/ghc-go/response"

	"github.com/skynet-ltd/ghc-go/request"
)

var httpClient *http.Client

// Options ...
type Options struct {
	HTTPOpts interface{}
	WsOpts   interface{}
}

// Client ....
type Client struct {
	u  *url.URL
	mu sync.Mutex
}

// New ...
func New(apiURL string, opts *Options) (*Client, error) {
	var err error
	var url *url.URL

	if opts != nil {
		if err = initHTTPClient(opts); err != nil {
			return nil, errors.New("can't configure client:" + err.Error())
		}
	}

	u, err := url.Parse(apiURL)
	if err != nil {
		return nil, err
	}
	return &Client{u: u}, nil
}

// Execute ...
func (c *Client) Execute(req *request.Request) (*response.Response, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := json.Marshal(req)
	if err != nil {
		return nil, errors.New("execute: marshal error: " + err.Error())
	}

	switch req.Type() {
	case reflect.TypeOf(request.Query("")), reflect.TypeOf(request.Mutation("")):
		return httpExecute(c.u, data)
	case reflect.TypeOf(request.Subscription("")):
		return wsExecute(c.u, data)
	default:
		return nil, errors.New("execute: unsupported request type: " + req.Type().String())
	}
}

func httpExecute(u *url.URL, data []byte) (*response.Response, error) {
	req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewReader(data))
	if err != nil {
		return nil, errors.New("execute: new request: " + err.Error())
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, errors.New("execute: request: " + err.Error())
	}

	data, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.New("execute: read response: " + err.Error())
	}

	defer res.Body.Close()

	resp := &response.Response{}

	decoder := json.NewDecoder(bytes.NewReader(data))

	if err = decoder.Decode(resp); err != nil {
		return nil, errors.New("execute: decode response: " + err.Error())
	}

	if resp.Errors != nil {
		return resp, errors.New("execute: server returned error: see response.Errors for details")
	}

	return resp, nil
}

func wsExecute(u *url.URL, data []byte) (*response.Response, error) {
	panic("isn't supported yet")
}

func initHTTPClient(opts *Options) error {
	if opts.HTTPOpts != nil {
		return nil
	}
	return nil
}

func init() {
	httpClient = http.DefaultClient
}