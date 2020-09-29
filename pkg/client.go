package pkg

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"

	"github.com/vtemian/form3/pkg/api"
)

type Client interface {
	Fetch(context.Context, api.Object) error
	List(context.Context, api.Object, *ListOptions) error
	Create(context.Context, api.Object) error
	Delete(context.Context, api.Object) error
}

type Form3Client struct {
	BaseURL string
	Version string
}

type ListFilter struct {
	BankIDCode    string
	BankID        string
	AccountNumber string
	IBAN          string
	CustomerID    string
	Country       string
}

func (l *ListFilter) Build() string {
	query := ""

	if l.BankIDCode != "" {
		query = fmt.Sprintf("%s&filter[bank_id_code]=%s", query, l.BankIDCode)
	}

	if l.BankID != "" {
		query = fmt.Sprintf("%s&filter[bank_id]=%s", query, l.BankID)
	}

	if l.AccountNumber != "" {
		query = fmt.Sprintf("%s&filter[account_number]=%s", query, l.AccountNumber)
	}

	if l.IBAN != "" {
		query = fmt.Sprintf("%s&filter[iban]=%s", query, l.IBAN)
	}

	if l.CustomerID != "" {
		query = fmt.Sprintf("%s&filter[customer_id]=%s", query, l.IBAN)
	}

	if l.Country != "" {
		query = fmt.Sprintf("%s&filter[country]=%s", query, l.IBAN)
	}

	return query
}

type ListOptions struct {
	PageNumber int
	PageSize   int
	Filter     *ListFilter
}

func (l *ListOptions) Build() string {
	query := "?"

	if l.PageNumber != 0 {
		query = fmt.Sprintf("%s&page[number]=%d", query, l.PageNumber)
	}

	if l.PageSize != 0 {
		query = fmt.Sprintf("%s&page[size]=%d", query, l.PageSize)
	}

	if l.Filter != nil {
		query = fmt.Sprintf("%s%s", query, l.Filter.Build())
	}

	return query
}

// TODO: implement retry/backoff

// TODO: handle all errors from upstream

var RespErrors = map[int]string{
	http.StatusBadRequest:          "invalid request: %s",
	http.StatusUnauthorized:        "not authorized: %s",
	http.StatusNotFound:            "not found: %s",
	http.StatusInternalServerError: "server error %s",
	http.StatusBadGateway:          "bad gateway %s",
	http.StatusGatewayTimeout:      "gateway timeout %s",
}

var ErrInvalidObjectType = errors.New("invalid object type")

const MissingOrInvalidArgumentFmt = "missing or invalid argument: %s"
const DefaultResponseErrorFmt = "error: %s"

type errorResponse struct {
	ErrorMessage string `json:"error_message"`
}

func (c *Form3Client) baseURL() string {
	return fmt.Sprintf("%s/%s", c.BaseURL, c.Version)
}

func (c *Form3Client) isOK(resp *http.Response) bool {
	return http.StatusOK <= resp.StatusCode && resp.StatusCode <= http.StatusIMUsed
}

func (c *Form3Client) err(resp *http.Response) error {
	if c.isOK(resp) {
		return nil
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf(DefaultResponseErrorFmt, "couldn't read response from server")
	}

	errMsg := ""
	var errResponse = &errorResponse{}

	err = json.Unmarshal(body, errResponse)
	if err == nil {
		errMsg = errResponse.ErrorMessage
	}

	respError, exists := RespErrors[resp.StatusCode]
	if !exists {
		return fmt.Errorf(DefaultResponseErrorFmt, body)
	}

	return fmt.Errorf(respError, errMsg)
}

func (c *Form3Client) execute(ctx context.Context, method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Form3Client) url(obj api.Object) (string, error) {
	endpoint, err := api.Schema.GetEndpointForObj(obj)
	if err != nil {
		return "", err
	}

	if strings.Contains(endpoint, "%s") {
		endpoint = fmt.Sprintf(endpoint, obj.GetID())
	}

	url := fmt.Sprintf("%s/%s", c.baseURL(), endpoint)

	return url, nil
}

func (c *Form3Client) Fetch(ctx context.Context, obj api.Object) error {
	if obj.GetID() == "" {
		return fmt.Errorf(MissingOrInvalidArgumentFmt, "uuid")
	}

	url, err := c.url(obj)
	if err != nil {
		return err
	}

	resp, err := c.execute(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if !c.isOK(resp) {
		return c.err(resp)
	}

	dataObj := api.WrapObject(obj)

	parseErr := json.NewDecoder(resp.Body).Decode(&dataObj)

	return parseErr
}

func (c *Form3Client) List(ctx context.Context, obj api.Object, listOptions *ListOptions) error {
	v, err := api.EnforcePtr(obj)
	if err != nil {
		return err
	}

	items := v.FieldByName("Items")
	if !items.IsValid() {
		return ErrInvalidObjectType
	}

	url, err := c.url(obj)
	if err != nil {
		return err
	}

	if listOptions != nil {
		url = fmt.Sprintf("%s%s", url, listOptions.Build())
	}

	objListType := reflect.StructOf([]reflect.StructField{
		{
			Name: "Data",
			Type: items.Type(),
			Tag:  `json:"data"`,
		},
		{
			Name: "Links",
			Type: reflect.TypeOf(api.Links{}),
			Tag:  `json:"links"`,
		},
	})
	objList := reflect.New(objListType).Elem()

	results := reflect.MakeSlice(items.Type(), 0, 1)

	for {
		resp, err := c.execute(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}

		defer resp.Body.Close()

		if !c.isOK(resp) {
			return c.err(resp)
		}

		result := objList.Addr().Interface()

		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		if err := json.Unmarshal(bodyBytes, result); err != nil {
			return err
		}

		data := objList.FieldByName("Data")
		store := reflect.MakeSlice(items.Type(), data.Len(), data.Len()+1)

		for i := 0; i < data.Len(); i++ {
			dest := store.Index(i)
			item := data.Index(i).Interface().(api.Object)
			dest.Set(reflect.ValueOf(item))
		}

		results = reflect.AppendSlice(results, store)

		links := objList.FieldByName("Links").Interface().(api.Links)
		if links.Next == "" || links.Next == links.Self {
			break
		}

		url = fmt.Sprintf("%s/%s", c.BaseURL, links.Next)
	}

	items.Set(results)

	return nil
}

func (c *Form3Client) Create(ctx context.Context, obj api.Object) error {
	dataObj := api.WrapObject(obj)

	jsonObj, err := json.Marshal(dataObj)
	if err != nil {
		return err
	}

	endpoint, err := api.Schema.GetEndpointForObj(obj)
	if err != nil {
		return err
	}

	if strings.HasSuffix(endpoint, "%s") {
		endpoint = endpoint[:len(endpoint)-2]
	}

	url := fmt.Sprintf("%s/%s", c.baseURL(), endpoint)
	resp, err := c.execute(ctx, http.MethodPost, url, bytes.NewBuffer(jsonObj))
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if !c.isOK(resp) {
		return c.err(resp)
	}

	objListType := reflect.StructOf([]reflect.StructField{
		{
			Name: "Data",
			Type: reflect.TypeOf(obj),
			Tag:  `json:"data"`,
		},
	})
	objList := reflect.New(objListType).Elem()
	result := objList.Addr().Interface()

	parsedErr := json.NewDecoder(resp.Body).Decode(result)
	if parsedErr != nil {
		return parsedErr
	}

	reflect.ValueOf(&obj).Elem().Set(objList.FieldByName("Data"))

	return nil
}

func (c *Form3Client) Delete(ctx context.Context, obj api.Object) error {
	if obj.GetID() == "" {
		return fmt.Errorf(MissingOrInvalidArgumentFmt, "ID")
	}

	if obj.GetVersion() < 0 {
		return fmt.Errorf(MissingOrInvalidArgumentFmt, "Version")
	}

	url, err := c.url(obj)
	if err != nil {
		return err
	}

	resp, err := c.execute(ctx, http.MethodDelete,
		fmt.Sprintf("%s?version=%d", url, obj.GetVersion()), nil)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if !c.isOK(resp) {
		return c.err(resp)
	}

	return nil
}

func NewClient(baseURL string) (Client, error) {
	return &Form3Client{BaseURL: baseURL, Version: "v1"}, nil
}
