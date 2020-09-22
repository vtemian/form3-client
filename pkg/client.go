package pkg

import "context"

type Client interface {
	Fetch(ctx context.Context)
	List(ctx context.Context)
	Create(ctx context.Context)
	Delete(ctx context.Context)
}

type Form3Client struct {
	BaseURL string
}

func (c *Form3Client) Fetch(ctx context.Context) {

}

func (c *Form3Client) List(ctx context.Context) {

}

func (c *Form3Client) Create(ctx context.Context) {

}

func (c *Form3Client) Delete(ctx context.Context) {

}

func NewClient() (Client, error) {
	return &Form3Client{BaseURL: ""}, nil
}
