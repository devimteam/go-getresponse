package getresponse

import (
	"bytes"
	"context"
	"github.com/healthimation/go-glitch/glitch"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"encoding/json"

	"time"
)

// Error codes
const (
	ErrorCantFind          = "CANT_FIND_SERVICE"
	ErrorRequestCreation   = "CANT_CREATE_REQUEST"
	ErrorRequestError      = "ERROR_MAKING_REQUEST"
	ErrorDecodingError     = "ERROR_DECODING_ERROR"
	ErrorDecodingResponse  = "ERROR_DECODING_RESPONSE"
	ErrorMarshallingObject = "ERROR_MARSHALLING_OBJECT"
)

// ServiceFinder can find a service's base URL
type ServiceFinder func(serviceName string, useTLS bool) (url.URL, error)

// BaseClient can do requests
type BaseClient interface {
	// MakeRequest does the request and returns the status, body, and any error
	// This should be used only if the api doesn't return glitch.DataErrors
	MakeRequest(ctx context.Context, method string, slug string, query url.Values, headers http.Header, body io.Reader) (int, []byte, glitch.DataError)
}

type beforeFunc func(ctx context.Context, r *http.Request) context.Context
type afterFunc func(ctx context.Context, r *http.Request, resp *http.Response) context.Context

type client struct {
	finder      ServiceFinder
	useTLS      bool
	serviceName string
	client      *http.Client
	beforeFunc  beforeFunc
	afterFunc   afterFunc
}

// NewBaseClient creates a new BaseClient
func NewBaseClient(finder ServiceFinder, serviceName string, useTLS bool, timeout time.Duration, rt http.RoundTripper, beforeFunc beforeFunc, afterFunc afterFunc) BaseClient {

	if rt == nil {
		rt = http.DefaultTransport
	}
	c := &http.Client{
		Timeout:   timeout,
		Transport: rt,
	}

	return &client{finder: finder, serviceName: serviceName, useTLS: useTLS, client: c, beforeFunc: beforeFunc, afterFunc: afterFunc}
}

func (c *client) MakeRequest(ctx context.Context, method string, slug string, query url.Values, headers http.Header, body io.Reader) (int, []byte, glitch.DataError) {
	u, err := c.finder(c.serviceName, c.useTLS)
	if err != nil {
		return 0, nil, glitch.NewDataError(err, ErrorCantFind, "Error finding service")
	}
	u.Path = slug
	u.RawQuery = query.Encode()

	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return 0, nil, glitch.NewDataError(err, ErrorRequestCreation, "Error creating request object")
	}

	req.Header = headers

	if ctx != nil {
		req = req.WithContext(ctx)
	}

	if c.beforeFunc != nil {
		ctx = c.beforeFunc(ctx, req)
	}

	resp, err := c.client.Do(req)

	if c.afterFunc != nil {
		ctx = c.afterFunc(ctx, req, resp)
	}

	if err != nil {
		return 0, nil, glitch.NewDataError(err, ErrorRequestError, "Could not make the request")
	}
	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()

	ret, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, glitch.NewDataError(err, ErrorDecodingResponse, "Could not read response body")
	}

	return resp.StatusCode, ret, nil
}

// ObjectToJSONReader will v to a io.Reader of the JSON representation of v
func ObjectToJSONReader(v interface{}) (io.Reader, glitch.DataError) {
	if by, ok := v.([]byte); ok {
		return bytes.NewBuffer(by), nil
	}
	by, err := json.Marshal(v)
	if err != nil {
		return nil, glitch.NewDataError(err, ErrorMarshallingObject, "Error marshalling object to json")
	}
	return bytes.NewBuffer(by), nil
}
