package mcli

import (
	"fmt"
	"github.com/bryant-rh/mcli/pkg/image"
	"github.com/bryant-rh/mcli/pkg/util"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/logs"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/cache"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/spf13/cobra"
)

// NewCmdPull creates a new cobra.Command for the pull subcommand.
func NewCmdPull(options *[]crane.Option) *cobra.Command {
	var cachePath, format string

	cmd := &cobra.Command{
		Use:   "pull IMAGE TARBALL",
		Short: "Pull container images from imagelist or manifests or charts",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			o := crane.GetOptions(*options...)
			imageMap := map[string]v1.Image{}
			indexMap := map[string]v1.ImageIndex{}
			//srcList, path := args[:len(args)-1], args[len(args)-1]
			srcDist, pathDist := args[0], args[1]
			var path string
			var pathname string
			var gzip_pathname string
			var platform string

			//获取文件的目录名
			pathDir := filepath.Dir(pathDist)
			//获取文件名称带后缀
			fileNameWithSuffix := filepath.Base(pathDist)
			//获取文件的后缀(文件类型)
			fileType := filepath.Ext(pathDist)
			//获取文件名称(不带后缀)
			fileNameOnly := strings.TrimSuffix(fileNameWithSuffix, fileType)

			if o.Platform != nil {
				//获取平台名称
				platform = fmt.Sprintf("%s-%s%s", o.Platform.OS, o.Platform.Architecture, o.Platform.Variant)
				//拼接platform
				pathname = fmt.Sprintf("%s-%s%s", fileNameOnly, platform, fileType)
				gzip_pathname = fmt.Sprintf("%s-%s.tar.gz", fileNameOnly, platform)
			} else {
				pathname = pathDist
				gzip_pathname = fmt.Sprintf("%s.tar.gz", fileNameOnly)
			}
			path = filepath.Join(pathDir, pathname)

			srcList, err := getImageList(srcDist)
			if err != nil {
				return fmt.Errorf("getImageList is Failed from %s, err: %w", srcDist, err)
			}
			fmt.Println("Pull imageList:")
			fmt.Println(srcList)
			for _, src := range srcList {
				//o := crane.GetOptions(*options...)
				logs.Debug.Printf("Pull image: %s", src)
				ref, err := name.ParseReference(src, o.Name...)
				if err != nil {
					return fmt.Errorf("parsing reference %q: %w", src, err)
				}

				rmt, err := remote.Get(ref, o.Remote...)
				if err != nil {
					return err
				}

				if format == "oci" && rmt.MediaType.IsIndex() && o.Platform == nil {
					idx, err := rmt.ImageIndex()
					if err != nil {
						return err
					}
					indexMap[src] = idx
					continue
				}

				img, err := rmt.Image()
				if err != nil {
					return err
				}
				if cachePath != "" {
					img = cache.Image(img, cache.NewFilesystemCache(cachePath))
				}
				imageMap[src] = img
			}

			switch format {
			case "tarball":
				if err := crane.MultiSave(imageMap, path); err != nil {
					return fmt.Errorf("saving tarball %s: %w", path, err)
				}
			case "legacy":
				if err := crane.MultiSaveLegacy(imageMap, path); err != nil {
					return fmt.Errorf("saving legacy tarball %s: %w", path, err)
				}
			case "oci":
				if err := crane.MultiSaveOCI(imageMap, path); err != nil {
					return fmt.Errorf("saving oci image layout %s: %w", path, err)
				}

				// crane.MultiSaveOCI doesn't support index, so just append these at the end.
				p, err := layout.FromPath(path)
				if err != nil {
					return err
				}
				for ref, idx := range indexMap {
					anns := map[string]string{
						"dev.ggcr.image.name": ref,
					}
					if err := p.AppendIndex(idx, layout.WithAnnotations(anns)); err != nil {
						return err
					}
				}
			default:
				return fmt.Errorf("unexpected --format: %q (valid values are: tarball, legacy, and oci)", format)
			}

			logs.Debug.Printf("Compress File: %s", gzip_pathname)

			err = util.CompressFile(gzip_pathname, path)
			if err != nil {
				return fmt.Errorf("CompressFile %s is Failed! err: %w", gzip_pathname, err)
			}
			err = os.Remove(path) //删除文件test.txt
			if err != nil {
				return fmt.Errorf("file %s remove Error! err: %w", path, err)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&cachePath, "cache_path", "c", "", "Path to cache image layers")
	cmd.Flags().StringVar(&format, "format", "tarball", fmt.Sprintf("Format in which to save images (%q, %q, or %q)", "tarball", "legacy", "oci"))

	return cmd
}

//getImageList
func getImageList(src string) ([]string, error) {
	var imageList []string

	yamlimage, err := image.ParseYamlImages(src)
	if err != nil {
		return nil, err
	}
	imageList = append(imageList, yamlimage...)

	rawimage, err := image.ParseRawImageList(src)
	if err != nil {
		return nil, err
	}
	imageList = append(imageList, rawimage...)

	chartimage, err := image.ParseChartImages(src)
	if err != nil {
		return nil, err
	}
	imageList = append(imageList, chartimage...)
	return imageList, nil
}
