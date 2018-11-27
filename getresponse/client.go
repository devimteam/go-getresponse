package getresponse

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Error codes
const (
	XAuthTokenHeader = "X-Auth-Token"
	XDomainHeader    = "X-Domain"

	// described @ https://apidocs.getresponse.com/v3/errors
	ErrInternalError           = 1
	ErrValidationError         = 1000
	ErrRelatedResourceNotFound = 1001
	ErrForbidden               = 1002
	ErrInvalidParameterFormat  = 1003
	ErrInvalidHash             = 1004
	ErrMissingParameter        = 1005
	ErrInvalidParameterType    = 1006
	ErrInvalidParameterLength  = 1007
	ErrResourceAlreadyExists   = 1008
	ErrResourceInUse           = 1009
	ErrExternalError           = 1010
	ErrMessageAlreadySending   = 1011
	ErrMessageParsing          = 1012
	ErrResourceNotFound        = 1013
	ErrAuthenticationFailure   = 1014
	ErrequestQuotaReached      = 1015
	ErrTemporarilyBlocked      = 1016
	ErrPermanentlyBlocked      = 1017
	ErrIPBlocked               = 1018
	ErrInvalidRequestHeaders   = 1021
	ErrRequestForbidden        = 1023
)

var (
	ErrCouldNotUnmarshal = errors.New("could not unmarshal")
)

// Client can make requests to the GR api
type Client interface {
	// CreateContact - https://apidocs.getresponse.com/v3/resources/contacts#contacts.create
	CreateContact(ctx context.Context, request *CreateContactRequest) error

	// GetContacts - https://apidocs.getresponse.com/v3/resources/contacts#contacts.get.all
	GetContacts(ctx context.Context, request *GetContactsRequest) (*GetContactsResponse, error)

	// Get Contact - https://apidocs.getresponse.com/v3/resources/contacts#contacts.get
	GetContact(ctx context.Context, request *GetContactRequest) (*GetContactResponse, error)

	// UpdateContact - https://apidocs.getresponse.com/v3/resources/contacts#contacts.update
	UpdateContact(ctx context.Context, request *UpdateContactRequest) (*UpdateContactResponse, error)

	// UpdateContactCustomFields - https://apidocs.getresponse.com/v3/resources/contacts#contacts.upsert.custom-fields
	UpdateContactCustomFields(ctx context.Context, request *UpdateContactCustomFieldsRequest) (*UpdateContactCustomFieldsResponse, error)

	// DeleteContact - https://apidocs.getresponse.com/v3/resources/contacts#contacts.delete
	DeleteContact(ctx context.Context, request *DeleteContactRequest) error
}

type getResponseClient struct {
	c      *http.Client
	apiKey string
	domain string
	apiUrl string
}

// NewClient returns a new pushy client
func NewClient(apiUrl, apiKey, domain string, client *http.Client) Client {
	if client == nil {
		client = http.DefaultClient
	}

	return &getResponseClient{
		c:      client,
		apiKey: apiKey,
		apiUrl: apiUrl,
		domain: domain,
	}
}

func (g *getResponseClient) CreateContact(ctx context.Context, request *CreateContactRequest) error {
	body, err := json.Marshal(request)
	if err != nil {
		return err
	}

	status, ret, err := g.roundTrip(ctx, http.MethodPost, "/v3/contacts", nil, body)
	return g.checkGetResponseError(status, ret, err)
}

func (g *getResponseClient) GetContacts(ctx context.Context, req *GetContactsRequest) (*GetContactsResponse, error) {
	query := url.Values{}
	for k, v := range req.QueryHash {
		query.Set(fmt.Sprintf("query[%s]", k), v)
	}

	for k, v := range req.SortHash {
		query.Set(fmt.Sprintf("sort[%s]", k), v)
	}

	if len(req.Fields) > 0 {
		query.Set("fields", strings.Join(req.Fields, ","))
	}

	query.Set("page", strconv.Itoa(int(req.Page)))
	query.Set("perPage", strconv.Itoa(int(req.PerPage)))

	if req.AdditionalFlags != nil {
		query.Set("additionalFlags", *req.AdditionalFlags)
	}

	status, ret, err := g.roundTrip(ctx, http.MethodGet, "/v3/contacts", query, nil)
	err = g.checkGetResponseError(status, ret, err)
	if err != nil {
		return nil, err
	}

	res := &GetContactsResponse{}
	jErr := json.Unmarshal(ret, &res.Contacts)
	if jErr != nil {
		return nil, ErrCouldNotUnmarshal
	}

	return res, nil
}

func (g *getResponseClient) GetContact(ctx context.Context, request *GetContactRequest) (*GetContactResponse, error) {
	query := url.Values{}
	if len(request.Fields) > 0 {
		query.Set("fields", strings.Join(request.Fields, ","))
	}

	status, ret, err := g.roundTrip(ctx, http.MethodGet, fmt.Sprintf("/v3/contacts/%s", request.ID), query, nil)
	err = g.checkGetResponseError(status, ret, err)
	if err != nil {
		return nil, err
	}

	c := Contact{}
	jErr := json.Unmarshal(ret, &c)
	if jErr != nil {
		return nil, ErrCouldNotUnmarshal
	}

	return &GetContactResponse{
		Contact: c,
	}, nil
}

func (g *getResponseClient) UpdateContact(ctx context.Context, req *UpdateContactRequest) (*UpdateContactResponse, error) {
	body, err := json.Marshal(req.NewData)
	if err != nil {
		return nil, err
	}

	status, ret, err := g.roundTrip(ctx, http.MethodPost, fmt.Sprintf("/v3/contacts/%s", req.ID), nil, body)
	err = g.checkGetResponseError(status, ret, err)
	if err != nil {
		return nil, err
	}

	result := &UpdateContactResponse{}
	jErr := json.Unmarshal(ret, &result.Contact)
	if jErr != nil {
		return result, ErrCouldNotUnmarshal
	}

	return result, nil
}

func (g *getResponseClient) UpdateContactCustomFields(ctx context.Context, request *UpdateContactCustomFieldsRequest) (*UpdateContactCustomFieldsResponse, error) {

	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	status, ret, err := g.roundTrip(ctx, http.MethodPost, fmt.Sprintf("/v3/contacts/%s/custom-fields", request.ID), nil, body)
	err = g.checkGetResponseError(status, ret, err)
	if err != nil {
		return nil, err
	}

	result := &UpdateContactCustomFieldsResponse{}
	jErr := json.Unmarshal(ret, &result.Contact)
	if jErr != nil {
		return nil, ErrCouldNotUnmarshal
	}

	return result, nil
}

func (g *getResponseClient) DeleteContact(ctx context.Context, request *DeleteContactRequest) error {
	query := url.Values{}
	query.Set("messageId", request.MessageID)
	query.Set("ipAddress", request.IpAddress)

	status, ret, err := g.roundTrip(ctx, http.MethodDelete, fmt.Sprintf("/v3/contacts/%s", request.ID), query, nil)
	return g.checkGetResponseError(status, ret, err)
}

func (g *getResponseClient) checkGetResponseError(status int, ret []byte, err error) error {
	if err != nil {
		return err
	}

	if status >= 200 && status < 400 {
		return nil
	}

	grErr := &GetResponseError{}
	jsonErr := json.Unmarshal(ret, grErr)
	if jsonErr != nil {
		return jsonErr
	}

	return grErr
}

func (g *getResponseClient) roundTrip(ctx context.Context, method string, path string, query url.Values, body []byte) (int, []byte, error) {
	u, err := url.Parse(g.apiUrl + path)
	if err != nil {
		return 0, nil, err
	}
	u.RawQuery = query.Encode()

	req, err := http.NewRequest(method, u.String(), bytes.NewBuffer(body))
	if err != nil {
		return 0, nil, err
	}

	h := http.Header{}
	h.Set(XAuthTokenHeader, fmt.Sprintf("api-key %s", g.apiKey))
	h.Set("Content-type", "application/json")
	if g.domain != "" {
		h.Set(XDomainHeader, g.domain)
	}

	if ctx != nil {
		req = req.WithContext(ctx)
	}

	resp, err := g.c.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	ret, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}

	return resp.StatusCode, ret, nil
}
