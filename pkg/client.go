package pkg

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

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
}

type HttpClient struct {
	HostUrl string
}

//func (h HttpClient) Execute(method string, url string, params string) error {
//
//}

func (c *Form3Client) Fetch(ctx context.Context, uuid string, obj api.Object) error {
	// TODO: hardcode endpoints and types for now
	resp, err := http.Get(fmt.Sprintf("%s/v1/organisation/accounts/%s", c.BaseURL, uuid))
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

	dataObj := struct {
		Data api.Object `json:"data"`
	}{
		Data: obj,
	}

	// TODO: extract client logic in a separate pkg
	parseErr := json.NewDecoder(resp.Body).Decode(&dataObj)
	fmt.Printf("%+v\n", dataObj)

	return parseErr
}

func (c *Form3Client) List(ctx context.Context) {

}

func (c *Form3Client) Create(ctx context.Context) {

}

func (c *Form3Client) Delete(ctx context.Context) {

}

func NewClient(baseURL string) (Client, error) {
	return &Form3Client{BaseURL: baseURL}, nil
}
