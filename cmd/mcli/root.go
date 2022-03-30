package mcli

import (
	"crypto/tls"
	"net/http"
	"os"

	"github.com/docker/cli/cli/config"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/logs"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/spf13/cobra"
)

const (
	use   = "mcli"
	short = "mcli is a tool for pull container images from imagelist and manifests amd charts"
)

var (
	// set values via build flags
	version string
	options []crane.Option
)

// versionString returns the version prefixed by 'v'
// or an empty string if no version has been populated by goreleaser.
// In this case, the --version flag will not be added by cobra.
func versionString() string {
	if len(version) == 0 {
		return ""
	}
	return "v" + version
}

// New returns a top-level command for crane. This is mostly exposed
// to share code with gcrane.
func New() *cobra.Command {
	verbose := false
	insecure := false
	platform := &platformValue{}

	root := &cobra.Command{
		Use:               use,
		Short:             short,
		Version:           versionString(),
		RunE:              func(cmd *cobra.Command, _ []string) error { return cmd.Usage() },
		DisableAutoGenTag: true,
		SilenceUsage:      true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			options = append(options, crane.WithContext(cmd.Context()))
			// TODO(jonjohnsonjr): crane.Verbose option?
			if verbose {
				logs.Debug.SetOutput(os.Stderr)
			}
			if insecure {
				options = append(options, crane.Insecure)
			}
			// if Version != "" {
			// 	binary := "mcli"
			// 	if len(os.Args[0]) != 0 {
			// 		binary = filepath.Base(os.Args[0])
			// 	}
			// 	options = append(options, crane.WithUserAgent(fmt.Sprintf("%s/%s", binary, Version)))
			// }

			options = append(options, crane.WithPlatform(platform.platform))

			transport := remote.DefaultTransport.Clone()
			transport.TLSClientConfig = &tls.Config{
				InsecureSkipVerify: insecure, //nolint: gosec
			}

			var rt http.RoundTripper = transport
			// Add any http headers if they are set in the config file.
			cf, err := config.Load(os.Getenv("DOCKER_CONFIG"))
			if err != nil {
				logs.Debug.Printf("failed to read config file: %v", err)
			} else if len(cf.HTTPHeaders) != 0 {
				rt = &headerTransport{
					inner:       rt,
					httpHeaders: cf.HTTPHeaders,
				}
			}

			options = append(options, crane.WithTransport(rt))
		},
	}

	commands := []*cobra.Command{
		NewCmdAuth("mcli", "auth"),
		NewCmdPull(&options),
	}

	root.AddCommand(commands...)

	root.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable debug logs")
	root.PersistentFlags().BoolVar(&insecure, "insecure", false, "Allow image references to be fetched without TLS")
	root.PersistentFlags().Var(platform, "platform", "Specifies the platform in the form os/arch[/variant][:osversion] (e.g. linux/amd64).")

	return root
}

// headerTransport sets headers on outgoing requests.
type headerTransport struct {
	httpHeaders map[string]string
	inner       http.RoundTripper
}

// RoundTrip implements http.RoundTripper.
func (ht *headerTransport) RoundTrip(in *http.Request) (*http.Response, error) {
	for k, v := range ht.httpHeaders {
		if http.CanonicalHeaderKey(k) == "User-Agent" {
			// Docker sets this, which is annoying, since we're not docker.
			// We might want to revisit completely ignoring this.
			continue
		}
		in.Header.Set(k, v)
	}
	return ht.inner.RoundTrip(in)
}
