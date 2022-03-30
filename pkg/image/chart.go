package image

import (
	"encoding/json"
	"fmt"
	util "github.com/bryant-rh/mcli/pkg/util"
	"io/fs"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
)

type Charts struct{}
type ChartInterface interface {
	// ListImages List all the containers images in helm charts
	ListImages(ContextMountDir string) ([]string, error)
}

//ParseChartImages
func ParseChartImages(srcPath string) ([]string, error) {
	chartsPath := filepath.Join(srcPath, copyToChart)
	if !util.IsExist(chartsPath) {
		return nil, nil
	}

	var images []string
	imageSearcher, err := NewCharts()
	if err != nil {
		return nil, err
	}

	err = filepath.Walk(chartsPath, func(path string, f fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !f.IsDir() {
			return nil
		}

		if util.IsExist(filepath.Join(path, "Chart.yaml")) && util.IsExist(filepath.Join(path, "values.yaml")) &&
			util.IsExist(filepath.Join(path, "templates")) {
			ima, err := imageSearcher.ListImages(path)
			if err != nil {
				return err
			}
			images = append(images, ima...)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return util.FormatImages(images), nil
}

// ListImages List all the containers images in helm charts
func (charts *Charts) ListImages(chartPath string) ([]string, error) {
	var list []string
	images, err := GetImageList(chartPath)
	if err != nil {
		return list, fmt.Errorf("get images failed,chart path:%s, err: %s", chartPath, err)
	}
	if len(images) != 0 {
		list = append(list, images...)
	}
	return list, nil
}

func NewCharts() (ChartInterface, error) {
	return &Charts{}, nil
}

//GetImageList
func GetImageList(chartPath string) ([]string, error) {
	var list []string
	content, err := RenderHelmChart(chartPath)
	if err != nil {
		return list, fmt.Errorf("render helm chart failed %s", err)
	}

	for _, v := range content {
		images := DecodeImages(v)
		if len(images) != 0 {
			list = append(list, images...)
		}
	}

	return list, nil
}
func Load(chartPath string) (*chart.Chart, error) {
	return loader.LoadDir(chartPath)
}

func PackageHelmChart(chartPath string) (string, error) {
	ch, err := Load(chartPath)
	if err != nil {
		return "", err
	}

	name, err := chartutil.Save(ch, ".")
	if err != nil {
		return "", err
	}

	return name, nil
}
func RenderHelmChart(chartPath string) (map[string]string, error) {
	ch, err := Load(chartPath)
	if err != nil {
		return nil, err
	}

	options := chartutil.ReleaseOptions{
		Name: "dryrun",
	}
	valuesToRender, err := chartutil.ToRenderValues(ch, nil, options, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to render values %v", err)
	}

	content, err := engine.Render(ch, valuesToRender)
	if err != nil {
		b, _ := json.Marshal(valuesToRender)
		logrus.Debugf("values is %s", b)
		return nil, fmt.Errorf("render helm chart error %s", err.Error())
	}

	return content, nil
}
