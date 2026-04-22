package policygroup

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/chaitin/chaitin-cli/products/safeline/cmd"
	"github.com/spf13/cobra"
)

// NewCommand creates the policy-group command.
func NewCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "policy-group",
		Short: "Policy group management commands",
		Long:  "Commands for managing SafeLine policy groups (detection engines).",
	}
	c.AddCommand(newListCmd())
	c.AddCommand(newGetCmd())
	c.AddCommand(newUpdateCmd())
	return c
}

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all policy groups",
		Long: `List all policy groups (detection engines).

Examples:
  safeline policy-group list`,
		RunE: func(c *cobra.Command, args []string) error {
			cl := cmd.NewClient()

			query := map[string]string{
				"count":  "20",
				"offset": "0",
				"scope":  "detect:policy",
			}
			env, err := cl.Do("GET", "/api/FilterV2API", nil, query)
			if err != nil {
				return err
			}

			// Print based on output format
			if cmd.GetOutput() == "json" {
				return cmd.PrintEnvelope(c, env)
			}

			// Table output - parse FilterV2API response structure
			var result struct {
				Total int `json:"total"`
				Items []struct {
					ID   int    `json:"id"`
					Name string `json:"name"`
				} `json:"items"`
			}
			if err := json.Unmarshal(env.Data, &result); err != nil {
				return err
			}

			headers := []string{"ID", "Name"}
			var rows [][]string
			for _, r := range result.Items {
				rows = append(rows, []string{
					fmt.Sprintf("%d", r.ID),
					r.Name,
				})
			}
			cmd.PrintTable(headers, rows)
			return nil
		},
	}
}

func newGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get policy group details",
		Long: `Get policy group details by ID.

Examples:
  safeline policy-group get 1`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			cl := cmd.NewClient()

			env, err := cl.Do("GET", "/api/PolicyGroupAPI", nil, map[string]string{"id": args[0]})
			if err != nil {
				return err
			}

			// Print based on output format
			if cmd.GetOutput() == "json" {
				return cmd.PrintEnvelope(c, env)
			}

			// Key-value output
			var result struct {
				ID                     int                    `json:"id"`
				Name                   string                 `json:"name"`
				ModulesDetectionConfig map[string]interface{} `json:"modules_detection_config"`
			}
			if err := json.Unmarshal(env.Data, &result); err != nil {
				return err
			}

			cmd.PrintKeyValue(map[string]string{
				"ID":   fmt.Sprintf("%d", result.ID),
				"Name": result.Name,
			})

			// Print enabled modules
			if len(result.ModulesDetectionConfig) > 0 {
				fmt.Println("\nModules:")
				headers := []string{"Module", "State"}
				var rows [][]string
				for name, config := range result.ModulesDetectionConfig {
					if cfg, ok := config.(map[string]interface{}); ok {
						state := fmt.Sprintf("%v", cfg["state"])
						rows = append(rows, []string{name, state})
					}
				}
				cmd.PrintTable(headers, rows)
			}
			return nil
		},
	}
}

func newUpdateCmd() *cobra.Command {
	var module, state string

	c := &cobra.Command{
		Use:   "update <id>",
		Short: "Update policy group module states",
		Long: `Update detection module states in a policy group.

Module keys:
  m_sqli             - SQL injection detection
  m_xss              - XSS detection
  m_cmd_injection    - Command injection detection
  m_file_include     - File inclusion detection
  m_file_upload      - File upload detection
  m_php_code_injection - PHP code injection detection
  m_php_unserialize  - PHP deserialization detection
  m_java             - Java detection
  m_java_unserialize - Java deserialization detection
  m_ssrf             - SSRF detection
  m_ssti             - SSTI detection
  m_csrf             - CSRF detection
  m_scanner          - Scanner detection
  m_response         - Response detection
  m_rule             - Built-in rules

Examples:
  # Enable SQL injection and XSS detection
  safeline policy-group update 1 --module m_sqli,m_xss --state enabled

  # Disable multiple modules
  safeline policy-group update 1 --module m_cmd_injection,m_ssrf --state disabled`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			if module == "" {
				return fmt.Errorf("--module is required")
			}
			if state == "" {
				return fmt.Errorf("--state is required")
			}
			if state != "enabled" && state != "disabled" {
				return fmt.Errorf("--state must be 'enabled' or 'disabled'")
			}

			cl := cmd.NewClient()

			// First get the current policy group
			env, err := cl.Do("GET", "/api/PolicyGroupAPI", nil, map[string]string{"id": args[0]})
			if err != nil {
				return err
			}

			// Parse the full policy group config
			var pg map[string]interface{}
			if err := json.Unmarshal(env.Data, &pg); err != nil {
				return fmt.Errorf("failed to parse policy group: %w", err)
			}

			// Update module states
			modules := strings.Split(module, ",")
			modulesConfig, ok := pg["modules_detection_config"].(map[string]interface{})
			if !ok {
				return fmt.Errorf("modules_detection_config not found in policy group")
			}
			for _, m := range modules {
				m = strings.TrimSpace(m)
				if m == "" {
					continue
				}
				if config, ok := modulesConfig[m].(map[string]interface{}); ok {
					config["state"] = state
				}
			}

			body, err := json.Marshal(pg)
			if err != nil {
				return fmt.Errorf("failed to marshal request: %w", err)
			}

			// Check dry-run
			if cmd.IsDryRun() {
				fmt.Fprintf(c.ErrOrStderr(), "[DRY-RUN] PUT /api/PolicyGroupAPI\n")
				fmt.Fprintf(c.ErrOrStderr(), "Body: %s\n", string(body))
				return nil
			}

			env, err = cl.Do("PUT", "/api/PolicyGroupAPI", strings.NewReader(string(body)), nil)
			if err != nil {
				return err
			}

			if cmd.GetOutput() == "json" {
				return cmd.PrintEnvelope(c, env)
			}

			fmt.Printf("Policy group %s updated successfully\n", args[0])
			return nil
		},
	}

	c.Flags().StringVar(&module, "module", "", "Comma-separated list of module keys (required)")
	c.Flags().StringVar(&state, "state", "", "State to set (enabled|disabled) (required)")
	c.MarkFlagRequired("module")
	c.MarkFlagRequired("state")

	return c
}
