// (C) Copyright 2021 Hewlett Packard Enterprise Development LP

//go:generate go run github.com/golang/mock/mockgen -source ./client.go -package client -destination ./client_mock.go

package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"time"
)

var (
	jsonCheck = regexp.MustCompile("(?i:[application|text]/json)")
	xmlCheck  = regexp.MustCompile("(?i:[application|text]/xml)")
)

type APIClientHandler interface {
	ChangeBasePath(path string)
	prepareRequest(
		ctx context.Context,
		path string, method string,
		postBody interface{},
		headerParams map[string]string,
		queryParams url.Values,
		formParams url.Values,
		fileName string,
		fileBytes []byte) (localVarRequest *http.Request, err error)
	decode(v interface{}, b []byte, contentType string) (err error)
	callAPI(request *http.Request) (*http.Response, error)
}

// APIClient manages communication with the GreenLake Private Cloud VMaaS CMP API API v1.0.0
// In most cases there should be only one, shared, APIClient.
type APIClient struct {
	cfg    *Configuration
	common service // Reuse a single struct instead of allocating one for each service on the heap.

	// API Services

	/*CloudsAPI *CloudsAPIService

	GroupsAPI *GroupsAPIService

	InstancesAPI *InstancesAPIService

	KeysCertsAPI *KeysCertsAPIService

	LibraryAPI *LibraryAPIService

	NetworksAPI *NetworksAPIService

	PlansAPI *PlansAPIService

	PoliciesAPI *PoliciesAPIService

	RolesAPI *RolesAPIService*/
}

type service struct {
	client *APIClient
}

// NewAPIClient creates a new API Client. Requires a userAgent string describing your application.
// optionally a custom http.Client to allow for advanced features such as caching.
func NewAPIClient(cfg *Configuration, secure bool) *APIClient {
	if !secure {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = http.DefaultClient
	}

	c := &APIClient{}
	c.cfg = cfg
	c.common.client = c

	// API Services
	/*c.CloudsAPI = (*CloudsAPIService)(&c.common)
	c.GroupsAPI = (*GroupsAPIService)(&c.common)
	c.InstancesAPI = (*InstancesAPIService)(&c.common)
	c.KeysCertsAPI = (*KeysCertsAPIService)(&c.common)
	c.LibraryAPI = (*LibraryAPIService)(&c.common)
	c.NetworksAPI = (*NetworksAPIService)(&c.common)
	c.PlansAPI = (*PlansAPIService)(&c.common)
	c.PoliciesAPI = (*PoliciesAPIService)(&c.common)
	c.RolesAPI = (*RolesAPIService)(&c.common)
	*/

	return c
}

// selectHeaderContentType select a content type from the available list.
func selectHeaderContentType(contentTypes []string) string {
	if len(contentTypes) == 0 {
		return ""
	}
	if contains(contentTypes, "application/json") {
		return "application/json"
	}

	return contentTypes[0] // use the first content type specified in 'consumes'
}

// selectHeaderAccept join all accept types and return
func selectHeaderAccept(accepts []string) string {
	if len(accepts) == 0 {
		return ""
	}

	if contains(accepts, "application/json") {
		return "application/json"
	}

	return strings.Join(accepts, ",")
}

// contains is a case insenstive match, finding needle in a haystack
func contains(haystack []string, needle string) bool {
	for _, a := range haystack {
		if strings.EqualFold(a, needle) {
			return true
		}
	}

	return false
}

// parameterToString convert interface{} parameters to string, using a delimiter if format is provided.
func parameterToString(obj interface{}, collectionFormat string) string {
	var delimiter string

	switch collectionFormat {
	case "pipes":
		delimiter = "|"
	case "ssv":
		delimiter = " "
	case "tsv":
		delimiter = "\t"
	case "csv":
		delimiter = ","
	}

	if reflect.TypeOf(obj).Kind() == reflect.Slice {
		return strings.Trim(strings.ReplaceAll(fmt.Sprint(obj), " ", delimiter), "[]")
	}

	return fmt.Sprintf("%v", obj)
}

// callAPI do the request.
func (c *APIClient) callAPI(request *http.Request) (*http.Response, error) {
	return c.cfg.HTTPClient.Do(request)
}

// Change base path to allow switching to mocks
func (c *APIClient) ChangeBasePath(path string) {
	c.cfg.BasePath = path
}

// prepareRequest build the request
//nolint
func (c *APIClient) prepareRequest(
	ctx context.Context,
	path string, method string,
	postBody interface{},
	headerParams map[string]string,
	queryParams url.Values,
	formParams url.Values,
	fileName string,
	fileBytes []byte) (localVarRequest *http.Request, err error) {
	var body *bytes.Buffer
	// Detect postBody type and post.
	if postBody != nil {
		contentType := headerParams["Content-Type"]
		if contentType == "" {
			contentType = detectContentType(postBody)
			headerParams["Content-Type"] = contentType
		}
		body, err = setBody(postBody, contentType)
		if err != nil {
			return nil, err
		}
	}
	// add form parameters and file if available.
	if strings.HasPrefix(headerParams["Content-Type"], "multipart/form-data") && len(formParams) > 0 ||
		(len(fileBytes) > 0 && fileName != "") {
		if body != nil {
			return nil, errors.New("cannot specify postBody and multipart form at the same time")
		}
		body = &bytes.Buffer{}
		w := multipart.NewWriter(body)
		for k, v := range formParams {
			for _, iv := range v {
				if strings.HasPrefix(k, "@") { // file
					err = addFile(w, k[1:], iv)
					if err != nil {
						return nil, err
					}
				} else { // form value
					err := w.WriteField(k, iv)
					if err != nil {
						continue
					}
				}
			}
		}
		if len(fileBytes) > 0 && fileName != "" {
			w.Boundary()
			// _, fileNm := filepath.Split(fileName)
			part, err := w.CreateFormFile("file", filepath.Base(fileName))
			if err != nil {
				return nil, err
			}
			_, err = part.Write(fileBytes)
			if err != nil {
				return nil, err
			}
			// Set the Boundary in the Content-Type
			headerParams["Content-Type"] = w.FormDataContentType()
		}
		// Set Content-Length
		headerParams["Content-Length"] = fmt.Sprintf("%d", body.Len())
		w.Close()
	}

	if strings.HasPrefix(headerParams["Content-Type"], "application/x-www-form-urlencoded") &&
		len(formParams) > 0 {
		if body != nil {
			return nil, errors.New("cannot specify postBody and x-www-form-urlencoded form at the same time")
		}
		body = &bytes.Buffer{}
		body.WriteString(formParams.Encode())
		// Set Content-Length
		headerParams["Content-Length"] = fmt.Sprintf("%d", body.Len())
	}

	// Setup path and query parameters
	url, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	// Adding Query Param
	query := url.Query()
	for k, v := range queryParams {
		for _, iv := range v {
			query.Add(k, iv)
		}
	}

	// Encode the parameters.
	url.RawQuery = query.Encode()

	// Generate a new request
	if body != nil {
		localVarRequest, err = http.NewRequest(method, url.String(), body)
	} else {
		localVarRequest, err = http.NewRequest(method, url.String(), nil)
	}
	if err != nil {
		return nil, err
	}

	// add header parameters, if any
	if len(headerParams) > 0 {
		headers := http.Header{}
		for h, v := range headerParams {
			headers.Set(h, v)
		}
		localVarRequest.Header = headers
	}

	// Add the user agent to the request.
	localVarRequest.Header.Add("User-Agent", c.cfg.UserAgent)

	// Add the Authentication Token to header.
	if ctx != nil {
		// add context to the request
		localVarRequest = localVarRequest.WithContext(ctx)
		// Basic HTTP Authentication
		if auth, ok := ctx.Value(ContextBasicAuth).(BasicAuth); ok {
			localVarRequest.SetBasicAuth(auth.UserName, auth.Password)
		}
		// AccessToken Authentication
		if auth, ok := ctx.Value(ContextAccessToken).(string); ok {
			localVarRequest.Header.Set("Authorization", "Bearer "+auth)
		}
	}

	for header, value := range c.cfg.DefaultHeader {
		if value != "" && value != " " {
			localVarRequest.Header.Add(header, value)
		}
	}

	return localVarRequest, nil
}

func (c *APIClient) decode(v interface{}, b []byte, contentType string) (err error) {
	if strings.Contains(contentType, "application/xml") {
		if err = xml.Unmarshal(b, v); err != nil {
			return err
		}

		return nil
	} else if strings.Contains(contentType, "application/json") {
		if err = json.Unmarshal(b, v); err != nil {
			return err
		}

		return nil
	}

	return errors.New("undefined response type")
}

// Add a file to the multipart request
func addFile(w *multipart.Writer, fieldName, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	part, err := w.CreateFormFile(fieldName, filepath.Base(path))
	if err != nil {
		return err
	}
	_, err = io.Copy(part, file)

	return err
}

// Prevent trying to import "fmt"
func reportError(format string, a ...interface{}) error {
	return fmt.Errorf(format, a...)
}

// Set request body from an interface{}
func setBody(body interface{}, contentType string) (bodyBuf *bytes.Buffer, err error) {
	if bodyBuf == nil {
		bodyBuf = &bytes.Buffer{}
	}

	switch v := body.(type) {
	case io.Reader:
		_, err = bodyBuf.ReadFrom(v)
	case []byte:
		_, err = bodyBuf.Write(v)
	case string:
		_, err = bodyBuf.WriteString(v)
	case *string:
		_, err = bodyBuf.WriteString(*v)
	default:
		if jsonCheck.MatchString(contentType) {
			err = json.NewEncoder(bodyBuf).Encode(body)
		} else if xmlCheck.MatchString(contentType) {
			err = xml.NewEncoder(bodyBuf).Encode(body)
		}
	}

	if err != nil {
		return nil, err
	}

	if bodyBuf.Len() == 0 {
		err = fmt.Errorf("invalid body type %s", contentType)

		return nil, err
	}

	return bodyBuf, nil
}

// detectContentType method is used to figure out `Request.Body` content type for request header
func detectContentType(body interface{}) string {
	contentType := "text/plain; charset=utf-8"
	kind := reflect.TypeOf(body).Kind()

	switch kind {
	case reflect.Struct, reflect.Map, reflect.Ptr:
		contentType = "application/json; charset=utf-8"
	case reflect.String:
		contentType = "text/plain; charset=utf-8"
	default:
		if b, ok := body.([]byte); ok {
			contentType = http.DetectContentType(b)
		} else if kind == reflect.Slice {
			contentType = "application/json; charset=utf-8"
		}
	}

	return contentType
}

// Ripped from https://github.com/gregjones/httpcache/blob/master/httpcache.go
type cacheControl map[string]string

func parseCacheControl(headers http.Header) cacheControl {
	cc := cacheControl{}
	ccHeader := headers.Get("Cache-Control")
	for _, part := range strings.Split(ccHeader, ",") {
		part = strings.Trim(part, " ")
		if part == "" {
			continue
		}
		if strings.ContainsRune(part, '=') {
			keyval := strings.Split(part, "=")
			cc[strings.Trim(keyval[0], " ")] = strings.Trim(keyval[1], ",")
		} else {
			cc[part] = ""
		}
	}

	return cc
}

// CacheExpires helper function to determine remaining time before repeating a request.
func CacheExpires(r *http.Response) time.Time {
	// Figure out when the cache expires.
	var expires time.Time
	now, err := time.Parse(time.RFC1123, r.Header.Get("date"))
	if err != nil {
		return time.Now()
	}
	respCacheControl := parseCacheControl(r.Header)

	// if maxAge, ok := respCacheControl["max-age"]; ok {
	// 	lifetime, err := time.ParseDuration(maxAge + "s")
	// 	if err != nil {
	// 		expires = now
	// 	} else {
	// 		expires = now.Add(lifetime)
	// 	}
	// } else {
	// 	expiresHeader := r.Header.Get("Expires")
	// 	if expiresHeader != "" {
	// 		expires, err = time.Parse(time.RFC1123, expiresHeader)
	// 		if err != nil {
	// 			expires = now
	// 		}
	// 	}
	// }
	if maxAge, ok := respCacheControl["max-age"]; ok {
		if lifetime, err := time.ParseDuration(maxAge + "s"); err != nil {
			expires = now
		} else {
			expires = now.Add(lifetime)
		}
	} else if expiresHeader := r.Header.Get("Expires"); expiresHeader != "" {
		expires, err = time.Parse(time.RFC1123, expiresHeader)
		if err != nil {
			expires = now
		}
	}

	return expires
}

// GenericSwaggerError Provides access to the body, error and model on returned errors.
type GenericSwaggerError struct {
	body  []byte
	error string
	model interface{}
}

// Error returns non-empty string if there was an error.
func (e GenericSwaggerError) Error() string {
	return e.error
}

// Body returns the raw bytes of the response
func (e GenericSwaggerError) Body() []byte {
	return e.body
}

// Model returns the unpacked model of the error
func (e GenericSwaggerError) Model() interface{} {
	return e.model
}
