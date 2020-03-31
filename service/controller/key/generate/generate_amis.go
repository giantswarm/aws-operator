//go:generate go run generate_amis.go
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"golang.org/x/net/html"

	"github.com/giantswarm/aws-operator/service/controller/key"
)

const keyPackage = "key"
const filename = "amis.go"
const sourceFileTemplate = `package {{ .Package }}

import "encoding/json"

var amiJSON = []byte({{ .AMIInfoString }})
var amiInfo = map[string]AMIInfoList{}

func init() {
	err := json.Unmarshal(amiJSON, &amiInfo)
	if err != nil {
		panic(err)
	}
}
`

type sourceFileTemplateData struct {
	AMIInfoString string
	Package       string
}

func main() {
	url := "https://stable.release.core-os.net/amd64-usr/"
	response, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	z := html.NewTokenizer(response.Body)

	var versions []string
	end := false
	for !end {
		tt := z.Next()
		switch {
		case tt == html.ErrorToken:
			end = true
		case tt == html.StartTagToken:
			t := z.Token()
			if t.Data == "a" {
				versions = append(versions, strings.TrimSuffix(t.Attr[0].Val, "/"))
			}
		}
	}

	versionAMIs := map[string]key.AMIInfoList{}
	for _, version := range versions {
		url := fmt.Sprintf("https://stable.release.core-os.net/amd64-usr/%s/coreos_production_ami_all.json", version)
		response, err := http.Get(url)
		if err != nil {
			log.Fatal(err)
		}
		if response.StatusCode == 403 {
			continue
		}
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Fatal(err)
		}
		amis := key.AMIInfoList{}
		err = json.Unmarshal(body, &amis)
		if err != nil {
			log.Fatal(err)
		}
		versionAMIs[version] = amis
	}

	result, err := json.MarshalIndent(versionAMIs, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	t, err := template.New(filename).Parse(sourceFileTemplate)
	if err != nil {
		log.Fatal(err)
	}

	path := filepath.Join("..", filename)
	file, err := os.OpenFile(path, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
	}

	err = t.Execute(file, sourceFileTemplateData{
		AMIInfoString: fmt.Sprintf("`%s`", result),
		Package:       keyPackage,
	})
	if err != nil {
		log.Fatal(err)
	}
}
