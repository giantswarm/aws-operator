//go:generate go run generate_amis.go
package main

import (
	"encoding/json"
	"fmt"
	"github.com/Masterminds/semver"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"golang.org/x/net/html"

	"github.com/giantswarm/aws-operator/service/controller/key"
)

const (
	flatcarDomain = "flatcar-linux.net"
	flatcarMinimum = "2345.3.1"
	coreosDomain = "core-os.net"
	coreosMinimum = "2135.4.0"
	channel = "stable"
	arch = "amd64-usr"
	generatedFilename = "amis.go"
	generatedPackage = "key"
	generatedTemplate = `package {{ .Package }}

import "encoding/json"

var amiJSON = []byte({{ .AMIInfoString }})
var amiInfo = map[string]map[string]string{}

func init() {
	err := json.Unmarshal(amiJSON, &amiInfo)
	if err != nil {
		panic(err)
	}
}
`)

type sourceFileTemplateData struct {
	AMIInfoString string
	Package       string
}

func scrapeVersions(source string) ([]string, error) {
	url := fmt.Sprintf("https://%s.release.%s/%s/", channel, source, arch)
	fmt.Println("scraping", url)
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	z := html.NewTokenizer(response.Body)
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

func scrapeVersionAMIs(source, vendor string) (map[string]map[string]string, error) {
	versions, err := scrapeVersions(source)
	if err != nil {
		return nil, err
	}

	result := map[string]map[string]string{}
	for _, version := range versions {
		url := fmt.Sprintf("https://%s.release.%s/%s/%s/%s_production_ami_all.json", channel, source, arch, version, vendor)
		fmt.Println("scraping", url)
		response, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		if response.StatusCode == 403 {
			continue
		}

		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}

		amis := key.AMIInfoList{}
		err = json.Unmarshal(body, &amis)
		if err != nil {
			return nil, err
		}

		result[version] = map[string]string{}
		for _, region := range amis.AMIs {
			result[version][region.Name] = region.HVM
		}
	}

	return result, nil
}

func main() {
	coreosAMIs, err := scrapeVersionAMIs(coreosDomain, "coreos")
	flatcarAMIs, err := scrapeVersionAMIs(flatcarDomain, "flatcar")
	mergedAMIs := map[string]map[string]string{}

	for version, regionAMIs := range coreosAMIs {
		if semver.MustParse(version).LessThan(semver.MustParse(coreosMinimum)) {
			continue
		}
		mergedAMIs[version] = regionAMIs
	}

	for version, regionAMIs := range flatcarAMIs {
		if semver.MustParse(version).LessThan(semver.MustParse(flatcarMinimum)) {
			continue
		}
		mergedAMIs[version] = regionAMIs
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
