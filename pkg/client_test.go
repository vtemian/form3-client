package pkg

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/vtemian/form3/pkg/api"
)

// Note - NOT RFC4122 compliant
func pseudoUUID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return strings.ToLower(fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])), nil
}

func loadFixture(path string, result *api.DataObject) error {
	jsonFile, err := os.Open(path)
	if err != nil {
		return err
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	return json.Unmarshal(byteValue, result)
}

func loadFixtures(kind api.Object) []api.Object {
	var files []string

	root := os.Getenv("FIXTURES_PATH")
	if root == "" {
		root = "./../fixtures/"
	}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, fmt.Sprintf("_%s_", api.Schema.TypeName(kind))) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	var objs []api.Object

	for _, file := range files {
		obj, err := api.Schema.NewDataObj(api.Schema.TypeName(kind))
		if err != nil {
			// TODO: log error
			fmt.Println("err:", err)
		}

		if err := loadFixture(file, obj); err == nil {
			objs = append(objs, obj.Data)
		} else {
			// TODO: log error
			fmt.Println("err:", err)
		}
	}

	return objs
}

var _ = Describe("Form3Client", func() {
	host := os.Getenv("TEST_API_HOST")
	if host == "" {
		host = "http://localhost:8080"
	}

	form3Client := NewClient(WithBaseURL(host))
	expectedAccounts := loadFixtures(api.Account{})

	var entries []TableEntry
	for _, fixture := range expectedAccounts {
		entries = append(entries, Entry(
			fmt.Sprintf("should fetch account %s", fixture.GetID()), fixture))
	}

	DescribeTable("fetch account",
		func(expectedAccount *api.Account) {
			account := api.NewAccount(expectedAccount.GetID(), expectedAccount.GetVersion())

			err := form3Client.Fetch(context.TODO(), account)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(account).To(BeEquivalentTo(expectedAccount))
		}, entries...)

	Describe("fetch fails", func() {
		It("should return 404 for missing account", func() {
			account := api.NewAccount("20dba636-7fac-4747-b27a-327ca12b9b27", 0)

			err := form3Client.Fetch(context.TODO(), account)
			Expect(err).Should(HaveOccurred())

			Expect(err).To(BeEquivalentTo(fmt.Errorf(RespErrors[http.StatusNotFound],
				"record 20dba636-7fac-4747-b27a-327ca12b9b27 does not exist")))
		})

		It("should return 400 for invalid uuid", func() {
			account := api.NewAccount("test", 0)

			err := form3Client.Fetch(context.TODO(), account)
			Expect(err).Should(HaveOccurred())

			Expect(err).To(BeEquivalentTo(fmt.Errorf(RespErrors[http.StatusBadRequest],
				"id is not a valid uuid")))
		})

		It("should return invalid request for missing uuid", func() {
			account := &api.Account{}

			err := form3Client.Fetch(context.TODO(), account)
			Expect(err).Should(HaveOccurred())

			Expect(err).To(BeEquivalentTo(fmt.Errorf(MissingOrInvalidArgumentFmt, "uuid")))
		})
	})

	Describe("list accounts", func() {
		It("should return a list of all accounts", func() {
			accounts := &api.AccountList{}

			err := form3Client.List(context.TODO(), accounts, nil)
			Expect(err).ShouldNot(HaveOccurred())

			sort.Slice(accounts.Items, func(i, j int) bool {
				return accounts.Items[i].GetID() < accounts.Items[j].GetID()
			})
			sort.Slice(expectedAccounts, func(i, j int) bool {
				return expectedAccounts[i].GetID() < expectedAccounts[j].GetID()
			})

			for i := range expectedAccounts {
				Expect(&accounts.Items[i]).To(BeEquivalentTo((expectedAccounts[i]).(*api.Account)))
			}
		})
		It("should return an error if the container is not valid", func() {
			accounts := &api.Account{}

			err := form3Client.List(context.TODO(), accounts, nil)
			Expect(err).Should(HaveOccurred())

			Expect(err).To(BeEquivalentTo(ErrInvalidObjectType))
		})

		It("should return only one account per page", func() {
			accounts := &api.AccountList{}

			options := &ListOptions{
				Filter: &ListFilter{
					BankID: "400305",
				},
				PageNumber: 1,
				PageSize:   1,
			}
			err := form3Client.List(context.TODO(), accounts, options)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(accounts.Items).To(HaveLen(len(expectedAccounts) - 1))
		})

		It("should return only two accounts per page", func() {
			accounts := &api.AccountList{}

			options := &ListOptions{
				PageNumber: 0,
				PageSize:   2,
			}
			err := form3Client.List(context.TODO(), accounts, options)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(accounts.Items).To(HaveLen(len(expectedAccounts)))
		})
	})

	Describe("delete account", func() {
		It("should delete an account", func() {
			account := expectedAccounts[0].(*api.Account)

			err := form3Client.Delete(context.TODO(), *account)
			Expect(err).ShouldNot(HaveOccurred())

			err = form3Client.Fetch(context.TODO(), account)
			Expect(err).Should(HaveOccurred())

			Expect(err).To(BeEquivalentTo(fmt.Errorf(RespErrors[http.StatusNotFound],
				fmt.Sprintf("record %s does not exist", account.GetID()))))
		})

		It("should return 204 for missing account", func() {
			account := api.NewAccount("20dba636-7fac-4747-b27a-327ca12b9b27", 0)

			err := form3Client.Delete(context.TODO(), account)
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("should return 400 for invalid uuid", func() {
			account := api.NewAccount("test", 0)

			err := form3Client.Delete(context.TODO(), account)
			Expect(err).Should(HaveOccurred())

			Expect(err).To(BeEquivalentTo(fmt.Errorf(RespErrors[http.StatusBadRequest],
				"id is not a valid uuid")))
		})

		It("should return invalid request for missing uuid", func() {
			account := &api.Account{}

			err := form3Client.Delete(context.TODO(), account)
			Expect(err).Should(HaveOccurred())

			Expect(err).To(BeEquivalentTo(fmt.Errorf(MissingOrInvalidArgumentFmt, "ID")))
		})

		It("should return invalid request for invalid version", func() {
			account := api.NewAccount("20dba636-7fac-4747-b27a-327ca12b9b27", -1)

			err := form3Client.Delete(context.TODO(), account)
			Expect(err).Should(HaveOccurred())

			Expect(err).To(BeEquivalentTo(fmt.Errorf(MissingOrInvalidArgumentFmt, "Version")))
		})

		It("should return missing item for missing version", func() {
			account := expectedAccounts[1].(*api.Account)
			account.Version++

			err := form3Client.Delete(context.TODO(), account)
			Expect(err).Should(HaveOccurred())

			Expect(err).To(BeEquivalentTo(fmt.Errorf(RespErrors[http.StatusNotFound], "invalid version")))
		})
	})

	Describe("create account", func() {
		It("should create a new account", func() {
			uuid, err := pseudoUUID()
			Expect(err).ShouldNot(HaveOccurred())

			account := &api.Account{
				OrganisationResource: api.OrganisationResource{
					OrganisationID: "721763e9-b2e2-4ebb-8de9-b440e3cf23a6",
					Resource: api.Resource{
						Type:    "accounts",
						ID:      uuid,
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
			Expect(err).ShouldNot(HaveOccurred())

			expectedAccount := api.NewAccount(account.GetID(), account.GetVersion())
			err = form3Client.Fetch(context.TODO(), expectedAccount)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(*account).To(Equal(*expectedAccount))
		})

		It("should return 400 for invalid body", func() {
			uuid, err := pseudoUUID()
			Expect(err).ShouldNot(HaveOccurred())

			account := &api.Account{
				OrganisationResource: api.OrganisationResource{
					OrganisationID: "721763e9-b2e2-4ebb-8de9-b440e3cf23a6",
					Resource: api.Resource{
						Type:    "accounts",
						ID:      uuid,
						Version: 0,
					},
				},
				Attributes: api.AccountAttributes{},
			}

			err = form3Client.Create(context.TODO(), account)
			Expect(err).Should(HaveOccurred())

			Expect(err.Error()).To(ContainSubstring("invalid request"))
			Expect(err.Error()).To(ContainSubstring("account_classification in body should be one of [Personal Business]"))
		})
	})
})
