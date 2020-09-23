package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

func loadFixtures(kind string) []*api.DataObject {
	var files []string

	root := os.Getenv("FIXTURES_PATH")
	if root == "" {
		root = "./../fixtures/"
	}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, fmt.Sprintf("_%s_", kind)) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	var objs []*api.DataObject

	for _, file := range files {
		obj := &api.DataObject{ Data: api.Account{} }
		if err := loadFixture(file, obj); err == nil {
			objs = append(objs, obj)
		} else {
			fmt.Println(err)
		}
	}
	fmt.Println("##############", root)
	fmt.Printf("%v\n", objs)

	return objs
}

var _ = Describe("Form3Client", func() {
	form3Client, _ := NewClient("http://localhost:8080")
	var entries []TableEntry

	for _, fixture := range loadFixtures("account") {
		entries = append(entries, Entry(
			fmt.Sprintf("should fetch account %s", fixture.Data.GetID()), fixture.Data.GetID(), fixture.Data))
	}

	DescribeTable("fetch account",
		func(uuid string, expectedAccount api.Account) {
			account := &api.Account{}

			err := form3Client.Fetch(context.TODO(), "test", account)
			Expect(err).To(BeNil())

			Expect(account).To(BeEquivalentTo(expectedAccount))
		}, entries...)
})
