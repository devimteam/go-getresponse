package getresponse

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"reflect"

	"github.com/healthimation/go-client/client"
)

func testClient(handler http.HandlerFunc, timeout time.Duration) (Client, *httptest.Server) {
	ts := httptest.NewServer(handler)
	finder := func(serviceName string, useTLS bool) (url.URL, error) {
		ret, err := url.Parse(ts.URL)
		if err != nil || ret == nil {
			return url.URL{}, err
		}
		return *ret, err
	}
	c := &getResponseClient{
		c:      client.NewBaseClient(finder, "gr", true, timeout, nil),
		apiKey: "",
	}
	return c, ts
}

func makeInt32Ptr(v int32) *int32 {
	return &v
}
func makeStringPtr(v string) *string {
	return &v
}

func TestUnit_CreateContact(t *testing.T) {

	type testcase struct {
		name            string
		handler         http.HandlerFunc
		timeout         time.Duration
		ctx             context.Context
		email           string
		contactName     *string
		dayOfCycle      *int32
		campaignID      string
		customFields    []CustomField
		ipAddress       *string
		expectedErrCode *string
	}

	testcases := []testcase{
		testcase{
			name:            "base path",
			handler:         http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
			timeout:         5 * time.Second,
			ctx:             context.Background(),
			email:           "foo@bar.baz",
			contactName:     makeStringPtr("foobar"),
			dayOfCycle:      makeInt32Ptr(5),
			customFields:    []CustomField{CustomField{CustomFieldID: "some_key", Value: []string{"some_value"}}},
			ipAddress:       makeStringPtr("127.0.0.1"),
			expectedErrCode: nil,
		},
		testcase{
			name: "unmarshal error",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, `{"not json"`)
			}),
			timeout:         5 * time.Second,
			ctx:             context.Background(),
			email:           "foo@bar.baz",
			contactName:     makeStringPtr("foobar"),
			dayOfCycle:      makeInt32Ptr(5),
			customFields:    []CustomField{CustomField{CustomFieldID: "some_key", Value: []string{"some_value"}}},
			ipAddress:       makeStringPtr("127.0.0.1"),
			expectedErrCode: makeStringPtr("ERROR_DECODING_ERROR"),
		},
		testcase{
			name: "error response",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusConflict)
				fmt.Fprint(w, `{"code":1008}`)
			}),
			timeout:         5 * time.Second,
			ctx:             context.Background(),
			email:           "foo@bar.baz",
			contactName:     makeStringPtr("foobar"),
			dayOfCycle:      makeInt32Ptr(5),
			customFields:    []CustomField{CustomField{CustomFieldID: "some_key", Value: []string{"some_value"}}},
			ipAddress:       makeStringPtr("127.0.0.1"),
			expectedErrCode: makeStringPtr("1008"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			c, ts := testClient(tc.handler, tc.timeout)
			defer ts.Close()
			err := c.CreateContact(tc.ctx, tc.email, tc.contactName, tc.dayOfCycle, tc.campaignID, tc.customFields, tc.ipAddress)
			if tc.expectedErrCode != nil || err != nil {
				if tc.expectedErrCode == nil {
					t.Fatalf("Unexpected error occurred (%#v)", err)
				}
				if err == nil {
					t.Fatalf("Expected error did not occur")
				}
			}
		})
	}
}

func TestUnit_GetContacts(t *testing.T) {

	type testcase struct {
		name             string
		handler          http.HandlerFunc
		timeout          time.Duration
		ctx              context.Context
		queryHash        map[string]string
		fields           []string
		sortHash         map[string]string
		page             int32
		perPage          int32
		additionalFlags  *string
		expectedErrCode  *string
		expectedResponse []Contact
	}

	testcases := []testcase{
		testcase{
			name: "base path",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `[{"name": "foobar", "email": "foo@bar.baz"}]`)
			}),
			timeout:          5 * time.Second,
			ctx:              context.Background(),
			queryHash:        map[string]string{"campaignId": "123"},
			fields:           []string{"name", "email"},
			sortHash:         map[string]string{"name": "asc"},
			page:             1,
			perPage:          10,
			additionalFlags:  nil,
			expectedErrCode:  nil,
			expectedResponse: []Contact{Contact{Email: makeStringPtr("foo@bar.baz"), Name: makeStringPtr("foobar")}},
		},
		testcase{
			name: "unmarshal error",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, `{"not json"`)
			}),
			timeout:         5 * time.Second,
			ctx:             context.Background(),
			expectedErrCode: makeStringPtr("ERROR_DECODING_ERROR"),
		},
		testcase{
			name: "error response",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusConflict)
				fmt.Fprint(w, `{"code":1008}`)
			}),
			timeout:         5 * time.Second,
			ctx:             context.Background(),
			expectedErrCode: makeStringPtr("1008"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			c, ts := testClient(tc.handler, tc.timeout)
			defer ts.Close()
			ret, err := c.GetContacts(tc.ctx, tc.queryHash, tc.fields, tc.sortHash, tc.page, tc.perPage, tc.additionalFlags)
			if err == nil && tc.expectedErrCode == nil {
				if !reflect.DeepEqual(tc.expectedResponse, ret) {
					t.Fatalf("Actual response (%#v) did not match expected (%#v)", ret, tc.expectedResponse)
				}
			} else {
				if tc.expectedErrCode == nil {
					t.Fatalf("Unexpected error occurred (%#v)", err)
				}
				if err == nil {
					t.Fatalf("Expected error did not occur")
				}
			}
		})
	}
}

func TestUnit_GetContact(t *testing.T) {

	type testcase struct {
		name             string
		handler          http.HandlerFunc
		timeout          time.Duration
		ctx              context.Context
		id               string
		fields           []string
		expectedErrCode  *string
		expectedResponse Contact
	}

	testcases := []testcase{
		testcase{
			name: "base path",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{"name": "foobar", "email": "foo@bar.baz"}`)
			}),
			timeout:          5 * time.Second,
			ctx:              context.Background(),
			id:               "foo",
			fields:           []string{"name", "email"},
			expectedErrCode:  nil,
			expectedResponse: Contact{Email: makeStringPtr("foo@bar.baz"), Name: makeStringPtr("foobar")},
		},
		testcase{
			name: "unmarshal error",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, `{"not json"`)
			}),
			timeout:         5 * time.Second,
			ctx:             context.Background(),
			expectedErrCode: makeStringPtr("ERROR_DECODING_ERROR"),
		},
		testcase{
			name: "error response",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusConflict)
				fmt.Fprint(w, `{"code":1008}`)
			}),
			timeout:         5 * time.Second,
			ctx:             context.Background(),
			expectedErrCode: makeStringPtr("1008"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			c, ts := testClient(tc.handler, tc.timeout)
			defer ts.Close()
			ret, err := c.GetContact(tc.ctx, tc.id, tc.fields)
			if err == nil && tc.expectedErrCode == nil {
				if !reflect.DeepEqual(tc.expectedResponse, ret) {
					t.Fatalf("Actual response (%#v) did not match expected (%#v)", ret, tc.expectedResponse)
				}
			} else {
				if tc.expectedErrCode == nil {
					t.Fatalf("Unexpected error occurred (%#v)", err)
				}
				if err == nil {
					t.Fatalf("Expected error did not occur")
				}
			}
		})
	}
}

func TestUnit_UpdateContact(t *testing.T) {

	type testcase struct {
		name             string
		handler          http.HandlerFunc
		timeout          time.Duration
		ctx              context.Context
		id               string
		newData          Contact
		expectedErrCode  *string
		expectedResponse Contact
	}

	testcases := []testcase{
		testcase{
			name: "base path",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{"name": "foobar", "email": "foo@bar.baz"}`)
			}),
			timeout:          5 * time.Second,
			ctx:              context.Background(),
			id:               "foo",
			newData:          Contact{Name: makeStringPtr("foobar")},
			expectedErrCode:  nil,
			expectedResponse: Contact{Email: makeStringPtr("foo@bar.baz"), Name: makeStringPtr("foobar")},
		},
		testcase{
			name: "unmarshal error",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, `{"not json"`)
			}),
			timeout:         5 * time.Second,
			ctx:             context.Background(),
			expectedErrCode: makeStringPtr("ERROR_DECODING_ERROR"),
		},
		testcase{
			name: "error response",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusConflict)
				fmt.Fprint(w, `{"code":1008}`)
			}),
			timeout:         5 * time.Second,
			ctx:             context.Background(),
			expectedErrCode: makeStringPtr("1008"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			c, ts := testClient(tc.handler, tc.timeout)
			defer ts.Close()
			ret, err := c.UpdateContact(tc.ctx, tc.id, tc.newData)
			if err == nil && tc.expectedErrCode == nil {
				if !reflect.DeepEqual(tc.expectedResponse, ret) {
					t.Fatalf("Actual response (%#v) did not match expected (%#v)", ret, tc.expectedResponse)
				}
			} else {
				if tc.expectedErrCode == nil {
					t.Fatalf("Unexpected error occurred (%#v)", err)
				}
				if err == nil {
					t.Fatalf("Expected error did not occur")
				}
			}
		})
	}
}

func TestUnit_UpdateContactCustomFields(t *testing.T) {

	type testcase struct {
		name             string
		handler          http.HandlerFunc
		timeout          time.Duration
		ctx              context.Context
		id               string
		customFields     []CustomField
		expectedErrCode  *string
		expectedResponse Contact
	}

	testcases := []testcase{
		testcase{
			name: "base path",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{"name": "foobar", "email": "foo@bar.baz"}`)
			}),
			timeout:          5 * time.Second,
			ctx:              context.Background(),
			id:               "foo",
			customFields:     []CustomField{CustomField{CustomFieldID: "some_key", Value: []string{"some_value"}}},
			expectedErrCode:  nil,
			expectedResponse: Contact{Email: makeStringPtr("foo@bar.baz"), Name: makeStringPtr("foobar")},
		},
		testcase{
			name: "unmarshal error",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, `{"not json"`)
			}),
			timeout:         5 * time.Second,
			ctx:             context.Background(),
			expectedErrCode: makeStringPtr("ERROR_DECODING_ERROR"),
		},
		testcase{
			name: "error response",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusConflict)
				fmt.Fprint(w, `{"code":1008}`)
			}),
			timeout:         5 * time.Second,
			ctx:             context.Background(),
			expectedErrCode: makeStringPtr("1008"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			c, ts := testClient(tc.handler, tc.timeout)
			defer ts.Close()
			ret, err := c.UpdateContactCustomFields(tc.ctx, tc.id, tc.customFields)
			if err == nil && tc.expectedErrCode == nil {
				if !reflect.DeepEqual(tc.expectedResponse, ret) {
					t.Fatalf("Actual response (%#v) did not match expected (%#v)", ret, tc.expectedResponse)
				}
			} else {
				if tc.expectedErrCode == nil {
					t.Fatalf("Unexpected error occurred (%#v)", err)
				}
				if err == nil {
					t.Fatalf("Expected error did not occur")
				}
			}
		})
	}
}

func TestUnit_DeleteContact(t *testing.T) {

	type testcase struct {
		name            string
		handler         http.HandlerFunc
		timeout         time.Duration
		ctx             context.Context
		id              string
		messageID       string
		ipAddress       string
		expectedErrCode *string
	}

	testcases := []testcase{
		testcase{
			name:            "base path",
			handler:         http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
			timeout:         5 * time.Second,
			ctx:             context.Background(),
			id:              "123",
			messageID:       "hello world",
			ipAddress:       "127.0.0.1",
			expectedErrCode: nil,
		},
		testcase{
			name: "unmarshal error",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, `{"not json"`)
			}),
			timeout:         5 * time.Second,
			ctx:             context.Background(),
			expectedErrCode: makeStringPtr("ERROR_DECODING_ERROR"),
		},
		testcase{
			name: "error response",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusConflict)
				fmt.Fprint(w, `{"code":1008}`)
			}),
			timeout:         5 * time.Second,
			ctx:             context.Background(),
			expectedErrCode: makeStringPtr("1008"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			c, ts := testClient(tc.handler, tc.timeout)
			defer ts.Close()
			err := c.DeleteContact(tc.ctx, tc.id, tc.messageID, tc.ipAddress)
			if tc.expectedErrCode != nil || err != nil {
				if tc.expectedErrCode == nil {
					t.Fatalf("Unexpected error occurred (%#v)", err)
				}
				if err == nil {
					t.Fatalf("Expected error did not occur")
				}
			}
		})
	}
}
