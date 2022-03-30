package image

import (
	"bufio"
	"fmt"
	"github.com/bryant-rh/mcli/pkg/util"
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

type Manifests struct{}

type ManifestsInterface interface {
	// ListImages List all the containers images in helm charts
	ListImages(ContextMountDir string) ([]string, error)
}

//parseYamlImages
func ParseYamlImages(srcPath string) ([]string, error) {
	manifestsPath := filepath.Join(srcPath, copyToManifests)
	if !util.IsExist(manifestsPath) {
		return nil, nil
	}
	var images []string

	imageSearcher, err := NewManifests()
	if err != nil {
		return nil, err
	}

	err = filepath.Walk(manifestsPath, func(path string, f fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if f.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(f.Name()))
		if ext != ".yaml" && ext != ".yml" && ext != ".tmpl" {
			return nil
		}

		ima, err := imageSearcher.ListImages(path)

		if err != nil {
			return err
		}
		images = append(images, ima...)
		return nil
	})

	if err != nil {
		return nil, err
	}
	return util.FormatImages(images), nil
}

//NewManifests
func NewManifests() (ManifestsInterface, error) {
	return &Manifests{}, nil
}

// ListImages List all the containers images in manifest files
func (manifests *Manifests) ListImages(yamlFile string) ([]string, error) {
	var list []string

	yamlBytes, err := ioutil.ReadFile(filepath.Clean(yamlFile))
	if err != nil {
		return nil, fmt.Errorf("read file failed %s", err)
	}

	images := DecodeImages(string(yamlBytes))
	if len(images) != 0 {
		list = append(list, images...)
	}

	if err != nil {
		return list, fmt.Errorf("filepath walk failed %s", err)
	}

	return list, nil
}

// DecodeImages decode image from yaml content
func DecodeImages(body string) []string {
	var list []string

	reader := strings.NewReader(body)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		l := decodeLine(scanner.Text())
		if l != "" {
			list = append(list, l)
		}
	}
	if err := scanner.Err(); err != nil {
		logrus.Errorf(err.Error())
		return list
	}

	return list
}

func decodeLine(line string) string {
	l := strings.Replace(line, `"`, "", -1)
	ss := strings.SplitN(l, ":", 2)
	if len(ss) != 2 {
		return ""
	}
	if !strings.HasSuffix(ss[0], "image") || strings.Contains(ss[0], "#") {
		return ""
	}

	return strings.Replace(ss[1], " ", "", -1)
}
