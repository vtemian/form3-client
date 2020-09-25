package pkg

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"

	"github.com/vtemian/form3/pkg/api"
)

type Client interface {
	Fetch(context.Context, api.Object) error
	List(context.Context, api.Object) error
	Create(context.Context, api.Object) error
	Delete(context.Context, api.Object) error
}

type Form3Client struct {
	BaseURL string
	Version string
}

type HttpClient struct {
	RetryCount int
}

// TODO: implement retry/backoff
// TODO: implement checks for valid object types

// TODO: handle all errors from upstream
var RespErrors = map[int]string{
	http.StatusBadRequest:          "invalid request: %s",
	http.StatusUnauthorized:        "not authorized: %s",
	http.StatusNotFound:            "not found: %s",
	http.StatusInternalServerError: "server error %s",
	http.StatusBadGateway:          "bad gateway %s",
	http.StatusGatewayTimeout:      "gateway timeout %s",
}

var InvalidObjectTypeErr = errors.New("invalid object type")

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

func (c *Form3Client) Fetch(ctx context.Context, obj api.Object) error {
	if obj.GetID() == "" {
		return fmt.Errorf(MissingOrInvalidArgumentFmt, "uuid")
	}

	endpoint, err := api.Schema.GetEndpointForObj(obj)
	if err != nil {
		return err
	}

	if strings.Contains(endpoint, "%s") {
		endpoint = fmt.Sprintf(endpoint, obj.GetID())
	}

	resp, err := http.Get(fmt.Sprintf("%s/%s", c.baseURL(), endpoint))
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if !c.isOK(resp) {
		return c.err(resp)
	}

	dataObj := api.WrapObject(obj)

	// TODO: extract client logic in a separate pkg
	parseErr := json.NewDecoder(resp.Body).Decode(&dataObj)

	return parseErr
}

func (c *Form3Client) List(ctx context.Context, obj api.Object) error {
	// TODO: implement pagination

	v, err := api.EnforcePtr(obj)
	if err != nil {
		return err
	}

	items := v.FieldByName("Items")
	if !items.IsValid() {
		return InvalidObjectTypeErr
	}

	endpoint, err := api.Schema.GetEndpointForObj(obj)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/%s", c.baseURL(), endpoint)
	resp, err := http.Get(url)
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
			Type: items.Type(),
			Tag:  `json:"data"`,
		},
	})
	objList := reflect.New(objListType).Elem()

	result := objList.Addr().Interface()

	// TODO: extract client logic in a separate pkg
	bodyBytes, err := ioutil.ReadAll(resp.Body)
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

	items.Set(store)

	return nil
}

func (c *Form3Client) Create(ctx context.Context, obj api.Object) error {
	// TODO: implement a .validate() function

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
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonObj))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
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

	// TODO: extract client logic in a separate pkg
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

	endpoint, err := api.Schema.GetEndpointForObj(obj)
	if err != nil {
		return err
	}

	if strings.Contains(endpoint, "%s") {
		endpoint = fmt.Sprintf(endpoint, obj.GetID())
	}

	client := &http.Client{}

	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/%s?version=%d", c.baseURL(), endpoint, obj.GetVersion()), nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
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
