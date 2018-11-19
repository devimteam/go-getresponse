package getresponse

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/healthimation/go-client/client"
)

//Error codes
const (
	XAuthTokenHeader = "X-Auth-Token"
	XDomainHeader    = "X-Domain"

	ErrorAPI = "ERROR_API"

	// described @ https://apidocs.getresponse.com/v3/errors
	ErrorInternalError          = 1
	ErrorValidationError        = 1000
	errorelatedResourceNotFound = 1001
	ErrorForbidden              = 1002
	ErrorInvalidParameterFormat = 1003
	ErrorInvalidHash            = 1004
	ErrorMissingParameter       = 1005
	ErrorInvalidParameterType   = 1006
	ErrorInvalidParameterLength = 1007
	erroresourceAlreadyExists   = 1008
	erroresourceInUse           = 1009
	ErrorExternalError          = 1010
	ErrorMessageAlreadySending  = 1011
	ErrorMessageParsing         = 1012
	erroresourceNotFound        = 1013
	ErrorAuthenticationFailure  = 1014
	errorequestQuotaReached     = 1015
	ErrorTemporarilyBlocked     = 1016
	ErrorPermanentlyBlocked     = 1017
	ErrorIPBlocked              = 1018
	ErrorInvalidRequestHeaders  = 1021
)

var (
	ErrCouldNotUnmarshal = errors.New("could not unmarshal")
)

// Client can make requests to the GR api
type Client interface {
	// CreateContact - https://apidocs.getresponse.com/v3/resources/contacts#contacts.create
	CreateContact(ctx context.Context, email string, name *string, dayOfCycle *int32, campaignID string, customFields []CustomField, ipAddress *string) error

	// GetContacts - https://apidocs.getresponse.com/v3/resources/contacts#contacts.get.all
	GetContacts(ctx context.Context, queryHash map[string]string, fields []string, sortHash map[string]string, page int32, perPage int32, additionalFlags *string) ([]Contact, error)

	// Get Contact - https://apidocs.getresponse.com/v3/resources/contacts#contacts.get
	GetContact(ctx context.Context, ID string, fields []string) (Contact, error)

	// UpdateContact - https://apidocs.getresponse.com/v3/resources/contacts#contacts.update
	UpdateContact(ctx context.Context, ID string, newData Contact) (Contact, error)

	// UpdateContactCustomFields - https://apidocs.getresponse.com/v3/resources/contacts#contacts.upsert.custom-fields
	UpdateContactCustomFields(ctx context.Context, ID string, customFields []CustomField) (Contact, error)

	// DeleteContact - https://apidocs.getresponse.com/v3/resources/contacts#contacts.delete
	DeleteContact(ctx context.Context, ID string, messageID string, ipAddress string) error
}

type getResponseClient struct {
	c      client.BaseClient
	apiKey string
	domain string
	url    string
}

func (g *getResponseClient) buildDefaultHeaders() http.Header {
	h := http.Header{}
	h.Set(XAuthTokenHeader, fmt.Sprintf("api-key %s", g.apiKey))
	h.Set("Content-type", "application/json")
	if g.domain != "" {
		h.Set(XDomainHeader, g.domain)
	}

	return h
}

// NewClient returns a new pushy client
func NewClient(apiKey string, apiUrl string, timeout time.Duration) Client {
	return &getResponseClient{
		c:      client.NewBaseClient(buildSvcFinder(apiUrl), "getresponse", true, timeout, nil),
		apiKey: apiKey,
	}
}

func (g *getResponseClient) CreateContact(ctx context.Context, email string, name *string, dayOfCycle *int32, campaignID string, customFields []CustomField, ipAddress *string) error {
	slug := "/v3/contacts"

	bodyObj := createContactRequest{
		Email:             email,
		Name:              name,
		DayOfCycle:        dayOfCycle,
		Campaign:          Campaign{CampaignID: campaignID},
		CustomFieldValues: customFields,
		IPAddress:         ipAddress,
	}

	body, err := client.ObjectToJSONReader(bodyObj)
	if err != nil {
		return err
	}

	status, ret, err := g.c.MakeRequest(ctx, http.MethodPost, slug, nil, g.buildDefaultHeaders(), body)
	if err != nil {
		return err
	}

	if status < 200 || status >= 400 {
		//parse error
		return g.parseError(ret)
	}

	return nil
}

func (g *getResponseClient) GetContacts(ctx context.Context, queryHash map[string]string, fields []string, sortHash map[string]string, page int32, perPage int32, additionalFlags *string) ([]Contact, error) {
	slug := "/v3/contacts"

	query := url.Values{}
	for k, v := range queryHash {
		query.Set(fmt.Sprintf("query[%s]", k), v)
	}

	for k, v := range sortHash {
		query.Set(fmt.Sprintf("sort[%s]", k), v)
	}

	if len(fields) > 0 {
		query.Set("fields", strings.Join(fields, ","))
	}

	query.Set("page", strconv.Itoa(int(page)))
	query.Set("perPage", strconv.Itoa(int(perPage)))

	if additionalFlags != nil {
		query.Set("additionalFlags", *additionalFlags)
	}

	result := make([]Contact, 0)
	status, ret, err := g.c.MakeRequest(ctx, http.MethodGet, slug, query, g.buildDefaultHeaders(), nil)
	if err != nil {
		return result, err
	}

	if status < 200 || status >= 400 {
		//parse error
		return result, g.parseError(ret)
	}

	jErr := json.Unmarshal(ret, &result)
	if jErr != nil {
		return result, ErrCouldNotUnmarshal
	}

	return result, nil
}

func (g *getResponseClient) GetContact(ctx context.Context, ID string, fields []string) (Contact, error) {
	slug := fmt.Sprintf("/v3/contacts/%s", ID)

	query := url.Values{}
	if len(fields) > 0 {
		query.Set("fields", strings.Join(fields, ","))
	}

	result := Contact{}
	status, ret, err := g.c.MakeRequest(ctx, http.MethodGet, slug, query, g.buildDefaultHeaders(), nil)
	if err != nil {
		return result, err
	}

	if status < 200 || status >= 400 {
		//parse error
		return result, g.parseError(ret)
	}

	jErr := json.Unmarshal(ret, &result)
	if jErr != nil {
		return result, ErrCouldNotUnmarshal
	}

	return result, nil
}

func (g *getResponseClient) UpdateContact(ctx context.Context, ID string, newData Contact) (Contact, error) {
	result := Contact{}
	slug := fmt.Sprintf("/v3/contacts/%s", ID)

	body, err := client.ObjectToJSONReader(newData)
	if err != nil {
		return result, err
	}

	status, ret, err := g.c.MakeRequest(ctx, http.MethodPost, slug, nil, g.buildDefaultHeaders(), body)
	if err != nil {
		return result, err
	}

	if status < 200 || status >= 400 {
		//parse error
		return result, g.parseError(ret)
	}

	jErr := json.Unmarshal(ret, &result)
	if jErr != nil {
		return result, ErrCouldNotUnmarshal
	}

	return result, nil
}

func (g *getResponseClient) UpdateContactCustomFields(ctx context.Context, ID string, customFields []CustomField) (Contact, error) {
	result := Contact{}
	slug := fmt.Sprintf("/v3/contacts/%s/custom-fields", ID)

	bodyObj := updateCustomFieldRequest{customFields}
	body, err := client.ObjectToJSONReader(bodyObj)
	if err != nil {
		return result, err
	}

	status, ret, err := g.c.MakeRequest(ctx, http.MethodPost, slug, nil, g.buildDefaultHeaders(), body)
	if err != nil {
		return result, err
	}

	if status < 200 || status >= 400 {
		//parse error
		return result, g.parseError(ret)
	}

	jErr := json.Unmarshal(ret, &result)
	if jErr != nil {
		return result, ErrCouldNotUnmarshal
	}

	return result, nil
}

func (g *getResponseClient) DeleteContact(ctx context.Context, ID string, messageID string, ipAddress string) error {
	slug := fmt.Sprintf("/v3/contacts/%s", ID)

	query := url.Values{}
	query.Set("messageId", messageID)
	query.Set("ipAddress", ipAddress)

	status, ret, err := g.c.MakeRequest(ctx, http.MethodDelete, slug, query, g.buildDefaultHeaders(), nil)
	if err != nil {
		return err
	}

	if status < 200 || status >= 400 {
		//parse error
		return g.parseError(ret)
	}

	return nil
}

// Todo refact
func (g *getResponseClient) parseError(resp []byte) error {
	errRet := ErrorResponse{}
	err := json.Unmarshal(resp, &errRet)
	if err != nil {
		return errors.New("could not unmarshal error response")
	}
	return errors.New(errRet.Message)
}

func buildSvcFinder(link string) func(string, bool) (url.URL, error) {
	return func(serviceName string, tls bool) (url.URL, error) {
		ret, err := url.Parse(link)
		if err != nil || ret == nil {
			return url.URL{}, err
		}
		return *ret, err
	}
}
