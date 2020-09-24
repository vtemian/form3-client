package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/vtemian/form3/pkg/api"
)

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
	form3Client, _ := NewClient("http://localhost:8080")
	expectedAccounts := loadFixtures(api.Account{})

	var entries []TableEntry
	for _, fixture := range expectedAccounts {
		entries = append(entries, Entry(
			fmt.Sprintf("should fetch account %s", fixture.GetID()), fixture.GetID(), fixture))
	}

	DescribeTable("fetch account",
		func(uuid string, expectedAccount *api.Account) {
			account := &api.Account{}

			err := form3Client.Fetch(context.TODO(), uuid, account)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(account).To(BeEquivalentTo(expectedAccount))
		}, entries...)

	Describe("fetch fails", func() {
		It("should return 404 for missing account", func() {
			account := &api.Account{}

			err := form3Client.Fetch(context.TODO(), "20dba636-7fac-4747-b27a-327ca12b9b27", account)
			Expect(err).Should(HaveOccurred())

			Expect(err).To(BeEquivalentTo(fmt.Errorf(RespErrors[http.StatusNotFound],
				"record 20dba636-7fac-4747-b27a-327ca12b9b27 does not exist")))
		})

		It("should return 400 for invalid uuid", func() {
			account := &api.Account{}

			err := form3Client.Fetch(context.TODO(), "test", account)
			Expect(err).Should(HaveOccurred())

			Expect(err).To(BeEquivalentTo(fmt.Errorf(RespErrors[http.StatusBadRequest],
				"id is not a valid uuid")))
		})

		It("should return invalid request for missing uuid", func() {
			account := &api.Account{}

			err := form3Client.Fetch(context.TODO(), "", account)
			Expect(err).Should(HaveOccurred())

			Expect(err).To(BeEquivalentTo(fmt.Errorf(MissingOrInvalidArgumentFmt, "uuid")))
		})
	})

	Describe("list accounts", func() {
		It("should return a list of all accounts", func() {
			accounts := &api.AccountList{}

			fmt.Println("running list tests")

			err := form3Client.List(context.TODO(), accounts)
			Expect(err).ShouldNot(HaveOccurred())

			for i := range expectedAccounts {
				Expect(&accounts.Items[i]).To(BeEquivalentTo((expectedAccounts[i]).(*api.Account)))
			}
		})
	})
})
