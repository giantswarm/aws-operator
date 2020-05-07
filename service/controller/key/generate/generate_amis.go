//go:generate go run generate_amis.go
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"github.com/Masterminds/semver"
	"golang.org/x/net/html"

	"github.com/giantswarm/aws-operator/service/controller/key"
)

const (
	primaryDomain     = "flatcar-linux.net"
	chinaDomain       = "flatcar-prod-ami-import-cn-north-1.s3.cn-north-1.amazonaws.com.cn"
	minimumVersion    = "2191.5.0"
	channel           = "stable"
	arch              = "amd64-usr"
	generatedFilename = "amis.go"
	generatedPackage  = "key"
	generatedTemplate = `// NOTE: This file is generated. Do not edit.
package {{ .Package }}

import "encoding/json"

var amiJSON = []byte({{ .AMIInfoString }})
var amiInfo = map[string]map[string]string{}

func init() {
	err := json.Unmarshal(amiJSON, &amiInfo)
	if err != nil {
		panic(err)
	}
}
`
)

type sourceFileTemplateData struct {
	AMIInfoString string
	Package       string
}

func scrapeVersions(source io.Reader) ([]string, error) {
	z := html.NewTokenizer(source)
	var versions []string
	for {
		tt := z.Next()
		switch {
		case tt == html.ErrorToken:
			return versions, nil
		case tt == html.StartTagToken:
			t := z.Token()
			if t.Data != "a" {
				continue
			}
			for _, attr := range t.Attr {
				if attr.Key != "href" {
					continue
				}
				// Versions to extract look like href="./123.4.5/" or href="123.4.5"
				// so we trim off suffix and prefix if they exist and then ensure this
				// is a valid semver version.
				href := strings.TrimSuffix(attr.Val, "/")
				if strings.HasPrefix(href, "./") {
					href = strings.TrimPrefix(href, "./")
				}
				if _, err := semver.NewVersion(href); err != nil {
					break // href is invalid, no need to look at other attrs
				}
				versions = append(versions, href)
			}
		}
	}
}

func scrapeVersionAMIs(source io.Reader) (map[string]string, error) {
	body, err := ioutil.ReadAll(source)
	if err != nil {
		return nil, err
	}

	var list key.AMIInfoList
	err = json.Unmarshal(body, &list)
	if err != nil {
		return nil, err
	}

	result := map[string]string{}
	for _, region := range list.AMIs {
		result[region.Name] = region.HVM
	}

	return result, nil
}

func main() {
	var versions []string
	{
		url := fmt.Sprintf("https://%s.release.%s/%s/", channel, primaryDomain, arch)
		fmt.Println("scraping", url)
		response, err := http.Get(url)
		if err != nil {
			log.Fatal(err)
		}
		versions, err = scrapeVersions(response.Body)
	}

	mergedAMIs := map[string]map[string]string{}
	for _, version := range versions {
		if semver.MustParse(version).LessThan(semver.MustParse(minimumVersion)) {
			continue
		}
		url := fmt.Sprintf("https://%s.release.%s/%s/%s/flatcar_production_ami_all.json", channel, primaryDomain, arch, version)
		fmt.Println("scraping", url)
		response, err := http.Get(url)
		if err != nil {
			log.Fatal(err)
		}
		if response.StatusCode == 403 {
			continue
		}
		mergedAMIs[version], err = scrapeVersionAMIs(response.Body)
		if err != nil {
			log.Fatal(err)
		}
	}

	for version := range mergedAMIs {
		url := fmt.Sprintf("https://%s/%s/%s/%s.json", chinaDomain, channel, arch, version)
		fmt.Println("scraping", url)
		response, err := http.Get(url)
		if err != nil {
			log.Fatal(err)
		}
		if response.StatusCode == 403 {
			continue
		}
		chinaVersionAMIs, err := scrapeVersionAMIs(response.Body)
		if err != nil {
			log.Fatal(err)
		}
		for region, image := range chinaVersionAMIs {
			mergedAMIs[version][region] = image
		}
	}

	result, err := json.MarshalIndent(mergedAMIs, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	t, err := template.New(generatedFilename).Parse(generatedTemplate)
	if err != nil {
		log.Fatal(err)
	}

	_, thisSourceFilePath, _, _ := runtime.Caller(0)
	thisSourceFileDirectory := filepath.Dir(thisSourceFilePath)
	path := filepath.Join(thisSourceFileDirectory, "..", generatedFilename)
	file, err := os.OpenFile(path, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}

	err = t.Execute(file, sourceFileTemplateData{
		AMIInfoString: fmt.Sprintf("`%s`", result),
		Package:       generatedPackage,
	})
	if err != nil {
		log.Fatal(err)
	}
}