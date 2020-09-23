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
	Fetch(ctx context.Context, uuid string, obj api.Object) error
	List(ctx context.Context)
	Create(ctx context.Context)
	Delete(ctx context.Context)
}

type Form3Client struct {
	BaseURL string
	Version string
}

type HttpClient struct {
	HostUrl string
}

func (c *Form3Client) baseURL() string {
	return fmt.Sprintf("%s/%s", c.BaseURL, c.Version)
}

//func (h HttpClient) Execute(method string, url string, params string) error {
//
//}

func (c *Form3Client) Fetch(ctx context.Context, uuid string, obj api.Object) error {
	// TODO: hardcode endpoints and types for now
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

	// TODO: return standard errors
	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println(string(body))
		return errors.New(fmt.Sprintf("missing obj identified by %s", uuid))
	}

	dataObj := api.WrapObject(obj)

	// TODO: extract client logic in a separate pkg
	parseErr := json.NewDecoder(resp.Body).Decode(&dataObj)

	return parseErr
}

func (c *Form3Client) List(ctx context.Context) {

}

func (c *Form3Client) Create(ctx context.Context) {

}

func (c *Form3Client) Delete(ctx context.Context) {

}

func NewClient(baseURL string) (Client, error) {
	return &Form3Client{BaseURL: baseURL, Version: "v1"}, nil
}
