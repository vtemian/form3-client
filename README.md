# form3-client

Name: Vlad Temian

Experience: ~ 1 year Go, 6 years Python

### Architecture

Since there wasn't a time limit, I thought to maybe try something different.

The client is heavily inspired by Kubernetes' client-go. It has a `Schema` where the users can register their resources.
A resource needs to implement the `Object` interface and to be registered with a specific endpoint. It then uses
 `reflect` to know which objects to instantiate and parse.

It's generic and this structure can be re-used for different resources. Some special adjustments needs to be made for
 nested endpoints. The exposed API is pretty simple:

```go
// Fetch uses an Account object and dynamically builds the final url 
account := api.NewAccount("20dba636-7fac-4747-b27a-327ca12b9b27", 0) // helper function for &Account{...}
err := form3Client.Fetch(context.TODO(), account)
```

```go
// List uses an AccountList that has an Items []Account field
accounts := &api.AccountList{}
err := form3Client.List(context.TODO(), accounts, nil)
```

```go
accounts := &api.AccountList{}

// Implement non-generic filtering, as an example
options := &ListOptions{
    Filter: &ListFilter{
        BankID: "400305",
    },
    PageNumber: 1,
    PageSize:   1,
}
err := form3Client.List(context.TODO(), accounts, options)
```

```go
// Delete, same as fetch, uses an Account object to dynamically build the final url
account := api.NewAccount("20dba636-7fac-4747-b27a-327ca12b9b27", 0)
err := form3Client.Delete(context.TODO(), account)
```

```go
// Create needs a full account object
account := &api.Account{
    OrganisationResource: api.OrganisationResource{
        OrganisationID: "721763e9-b2e2-4ebb-8de9-b440e3cf23a6",
        Resource: api.Resource{
            Type:    "accounts",
            ID:      "20dba636-7fac-4747-b27a-327ca12b9b27",
            Version: 0,
        },
    },
    Attributes: api.AccountAttributes{
        Country:               "GB",
        BaseCurrency:          "GBP",
        BankID:                "400300",
        BankIDCode:            "GBDSC",
        BIC:                   "NWBKGB22",
        AccountClassification: "Personal",
    },
}

err = form3Client.Create(context.TODO(), account)
```

The code may be harder to understand, since there are a lot of low level calls and maybe is not that Go like,
more Python like. It was a fun exercise to play with.

### Testing

I think that end-to-end tests and integration tests are more reliable and broader than unit tests. In order to ensure
idempotency and isolation, I've used a separate script that cleans the database and load some initial fixtures,
 before running the test suits. Also, there is a github action set in order to run the formatting check, linting and 
 the integration tests.