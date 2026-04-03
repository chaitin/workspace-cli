package safeline

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/chaitin/workspace-cli/config"
	cmdpkg "github.com/chaitin/workspace-cli/products/safeline/cmd"
	"github.com/chaitin/workspace-cli/products/safeline/cmd/acl"
	"github.com/chaitin/workspace-cli/products/safeline/cmd/ipgroup"
	"github.com/chaitin/workspace-cli/products/safeline/cmd/log"
	"github.com/chaitin/workspace-cli/products/safeline/cmd/network"
	"github.com/chaitin/workspace-cli/products/safeline/cmd/policygroup"
	"github.com/chaitin/workspace-cli/products/safeline/cmd/policyrule"
	"github.com/chaitin/workspace-cli/products/safeline/cmd/site"
	"github.com/chaitin/workspace-cli/products/safeline/cmd/stats"
	"github.com/chaitin/workspace-cli/products/safeline/cmd/system"
	"github.com/chaitin/workspace-cli/products/safeline/pkg/auth"
	"github.com/chaitin/workspace-cli/products/safeline/pkg/client"
	"github.com/spf13/cobra"
)

var (
	url      string
	apiKey   string
	indent   bool
	insecure bool
	dryRun   bool
)

type runtimeConfig struct {
	URL    string `yaml:"url"`
	APIKey string `yaml:"api_key"`
}

// NewCommand creates the safeline product command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "safeline",
		Short: "SafeLine WAF API CLI",
		Long:  "Command-line interface for SafeLine Skyview WAF APIs.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Name() == "help" || cmd.Name() == "completion" || cmd.Name() == "version" {
				return nil
			}
			// Check parents for help/completion
			for p := cmd.Parent(); p != nil; p = p.Parent() {
				if p.Name() == "help" || p.Name() == "completion" {
					return nil
				}
			}

			// Load from environment if not set via flags
			if url == "" {
				url = os.Getenv("SKYVIEW_URL")
			}
			if apiKey == "" {
				apiKey = os.Getenv("SKYVIEW_API_KEY")
			}

			if url == "" {
				return fmt.Errorf("URL is required (use --url or set SKYVIEW_URL)")
			}

			// Sync flags to cmd package for subcommands
			output := "table"
			if indent {
				output = "json"
			}
			cmdpkg.SetFlags(url, apiKey, output, insecure, dryRun)

			return nil
		},
	}

	cmd.PersistentFlags().StringVar(&url, "url", "", "Skyview URL (or SKYVIEW_URL)")
	cmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "API key (or SKYVIEW_API_KEY)")
	cmd.PersistentFlags().BoolVar(&indent, "indent", false, "Pretty-print JSON output")
	cmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without making changes")
	cmd.PersistentFlags().BoolVar(&insecure, "insecure", true, "Skip TLS certificate verification")

	return cmd
}

// RegisterModules adds all API module commands to the safeline root command.
func RegisterModules(cmd *cobra.Command) {
	// Register new user-friendly commands
	cmd.AddCommand(stats.NewCommand())
	cmd.AddCommand(site.NewCommand())
	cmd.AddCommand(ipgroup.NewCommand())
	cmd.AddCommand(acl.NewCommand())
	cmd.AddCommand(policygroup.NewCommand())
	cmd.AddCommand(policyrule.NewCommand())
	cmd.AddCommand(system.NewCommand())
	cmd.AddCommand(log.NewCommand())
	cmd.AddCommand(network.NewCommand())

}

// NewClient creates an authenticated Skyview client from command flags.
func NewClient() *client.Client {
	transport := &auth.Transport{
		Token: apiKey,
	}

	httpClient := &http.Client{
		Transport: transport,
	}

	if insecure {
		httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		// Wrap auth transport
		httpClient.Transport = &auth.Transport{
			Base:  httpClient.Transport,
			Token: apiKey,
		}
	}

	return client.New(url, httpClient)
}

// PrintJSON prints JSON data to stdout.
func PrintJSON(data []byte) error {
	if !indent {
		_, err := fmt.Fprintln(os.Stdout, string(data))
		return err
	}
	var pretty interface{}
	if err := json.Unmarshal(data, &pretty); err != nil {
		_, err := fmt.Fprintln(os.Stdout, string(data))
		return err
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(pretty)
}

// PrintEnvelope prints the data from an API response envelope.
func PrintEnvelope(env *client.Envelope) error {
	if env.IsWarning() {
		fmt.Fprintf(os.Stderr, "WARNING: %s\n", env.GetWarningText())
	}
	return PrintJSON(env.Data)
}

// ReadInput reads JSON input from file or stdin.
func ReadInput(file string) (io.Reader, error) {
	if file == "" || file == "-" {
		return os.Stdin, nil
	}
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", file, err)
	}
	return f, nil
}

// ApplyRuntimeConfig applies configuration from config.yaml to command flags.
// If a flag is not explicitly set by the user, it uses the value from config file.
func ApplyRuntimeConfig(cmd *cobra.Command, cfg config.Raw) {
	productCfg, err := config.DecodeProduct[runtimeConfig](cfg, "safeline")
	if err != nil {
		return
	}

	if flag := cmd.Flags().Lookup("url"); flag != nil && !flag.Changed {
		url = productCfg.URL
	}

	if flag := cmd.Flags().Lookup("api-key"); flag != nil && !flag.Changed {
		apiKey = productCfg.APIKey
	}
}
