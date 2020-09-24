package pkg

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/vtemian/form3/pkg/api"
)

type Client interface {
	Fetch(context.Context, string, api.Object) error
	List(context.Context, api.Object) error
	Create(ctx context.Context)
	Delete(ctx context.Context)
}

type Form3Client struct {
	BaseURL string
	Version string
}

type HttpClient struct {
	RetryCount int
}

var RespErrors = map[int]string{
	http.StatusBadRequest:          "invalid request: %s",
	http.StatusUnauthorized:        "not authorized: %s",
	http.StatusNotFound:            "not found: %s",
	http.StatusInternalServerError: "server error %s",
	http.StatusBadGateway:          "bad gateway %s",
	http.StatusGatewayTimeout:      "gateway timeout %s",
}

var MissingOrInvalidArgumentFmt = "missing or invalid argument: %s"
var defaultResponseErrorFmt = "error: %s"

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
		return fmt.Errorf(defaultResponseErrorFmt, "couldn't read response from server")
	}

	errMsg := ""
	var errResponse = &errorResponse{}

	err = json.Unmarshal(body, errResponse)
	if err == nil {
		errMsg = errResponse.ErrorMessage
	}

	respError, exists := RespErrors[resp.StatusCode]
	if !exists {
		return fmt.Errorf(defaultResponseErrorFmt, body)
	}

	return fmt.Errorf(respError, errMsg)
}

func (c *Form3Client) Fetch(ctx context.Context, uuid string, obj api.Object) error {
	if uuid == "" {
		return fmt.Errorf(MissingOrInvalidArgumentFmt, "uuid")
	}

	endpoint, err := api.Schema.GetEndpointForObj(obj)
	if err != nil {
		return err
	}

	if strings.Contains(endpoint, "%s") {
		endpoint = fmt.Sprintf(endpoint, uuid)
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
	endpoint, err := api.Schema.GetEndpointForObj(obj)
	if err != nil {
		return err
	}

	resp, err := http.Get(fmt.Sprintf("%s/%s", c.baseURL(), endpoint))
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	// TODO: return standard errors
	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println(string(body))
		return errors.New(fmt.Sprintf("missing obj identified by"))
	}

	dataObj := api.WrapObject(obj)

	// TODO: extract client logic in a separate pkg
	parseErr := json.NewDecoder(resp.Body).Decode(&dataObj)

	return parseErr
}

func (c *Form3Client) Create(ctx context.Context) {

}

func (c *Form3Client) Delete(ctx context.Context) {

}

func NewClient(baseURL string) (Client, error) {
	return &Form3Client{BaseURL: baseURL, Version: "v1"}, nil
}
