package image

import (
	"fmt"
	util "github.com/bryant-rh/mcli/pkg/util"
	"path/filepath"
)

//parseRawImageList
func ParseRawImageList(srcPath string) ([]string, error) {
	//imageListFilePath := filepath.Join(srcPath, copyToManifests, copyToImageList)
	imageListFilePath := filepath.Join(srcPath, copyToImageList)
	if !util.IsExist(imageListFilePath) {
		return nil, nil
	}

	images, err := util.ReadLines(imageListFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content %s:%v", imageListFilePath, err)
	}
	return util.FormatImages(images), nil
}
