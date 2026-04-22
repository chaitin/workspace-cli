package ipgroup

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/chaitin/chaitin-cli/products/safeline/cmd"
	"github.com/spf13/cobra"
)

// NewCommand creates the ip-group command.
func NewCommand() *cobra.Command {
	c := &cobra.Command{
		Use:     "ip-group",
		Short:   "IP group management commands",
		Long:    "Commands for managing SafeLine IP groups.",
		Aliases: []string{"ipgroup"},
	}
	c.AddCommand(newListCmd())
	c.AddCommand(newGetCmd())
	c.AddCommand(newCreateCmd())
	c.AddCommand(newDeleteCmd())
	c.AddCommand(newAddIPCmd())
	c.AddCommand(newRemoveIPCmd())
	return c
}

func newListCmd() *cobra.Command {
	var name string
	var count, offset int

	c := &cobra.Command{
		Use:   "list",
		Short: "List all IP groups",
		Long: `List all IP groups.

Examples:
  safeline ip-group list
  safeline ip-group list --name "office"
  safeline ip-group list --count 50`,
		RunE: func(c *cobra.Command, args []string) error {
			cl := cmd.NewClient()

			query := map[string]string{
				"scope":  "detect:asset:ip_group",
				"count":  fmt.Sprintf("%d", count),
				"offset": fmt.Sprintf("%d", offset),
			}
			if name != "" {
				query["name__like"] = name
			}

			env, err := cl.Do("GET", "/api/FilterV2API", nil, query)
			if err != nil {
				return err
			}

			// Print based on output format
			if cmd.GetOutput() == "json" {
				return cmd.PrintEnvelope(c, env)
			}

			// Table output - API returns {"items": [...], "total": N}
			var resp struct {
				Items []struct {
					ID      int      `json:"id"`
					Name    string   `json:"name"`
					Comment string   `json:"comment"`
					Cidrs   []string `json:"cidrs"`
				} `json:"items"`
				Total int `json:"total"`
			}
			if err := json.Unmarshal(env.Data, &resp); err != nil {
				return err
			}

			// Sort by ID descending (newest first)
			sort.Slice(resp.Items, func(i, j int) bool {
				return resp.Items[i].ID > resp.Items[j].ID
			})

			headers := []string{"ID", "Name", "Comment", "CIDRs"}
			var rows [][]string
			for _, r := range resp.Items {
				rows = append(rows, []string{
					fmt.Sprintf("%d", r.ID),
					r.Name,
					r.Comment,
					strings.Join(r.Cidrs, ", "),
				})
			}
			cmd.PrintTable(headers, rows)
			return nil
		},
	}

	c.Flags().StringVar(&name, "name", "", "Filter by name")
	c.Flags().IntVar(&count, "count", 20, "Number of items to return")
	c.Flags().IntVar(&offset, "offset", 0, "Offset for pagination")

	return c
}

func newGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get IP group details",
		Long: `Get IP group details by ID.

Examples:
  safeline ip-group get 1`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			cl := cmd.NewClient()

			query := map[string]string{
				"scope":     "detect:asset:ip_group",
				"id__exact": args[0],
			}

			env, err := cl.Do("GET", "/api/FilterV2API", nil, query)
			if err != nil {
				return err
			}

			// Print based on output format
			if cmd.GetOutput() == "json" {
				return cmd.PrintEnvelope(c, env)
			}

			// Key-value output - API returns array for id__exact query
			var results []struct {
				ID      int      `json:"id"`
				Name    string   `json:"name"`
				Comment string   `json:"comment"`
				Cidrs   []string `json:"cidrs"`
			}
			if err := json.Unmarshal(env.Data, &results); err != nil {
				return err
			}

			if len(results) == 0 {
				return fmt.Errorf("IP group not found: %s", args[0])
			}

			r := results[0]
			cmd.PrintKeyValue(map[string]string{
				"ID":      fmt.Sprintf("%d", r.ID),
				"Name":    r.Name,
				"Comment": r.Comment,
				"CIDRs":   strings.Join(r.Cidrs, ", "),
			})
			return nil
		},
	}
}

func newCreateCmd() *cobra.Command {
	var name, ips, comment string

	c := &cobra.Command{
		Use:   "create",
		Short: "Create a new IP group",
		Long: `Create a new IP group.

Examples:
  safeline ip-group create --name "Office" --ips "192.168.1.0/24,10.0.0.1"
  safeline ip-group create --name "DC" --ips "172.16.0.0/16" --comment "Data center"`,
		RunE: func(c *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("--name is required")
			}
			if ips == "" {
				return fmt.Errorf("--ips is required")
			}

			// Parse IPs
			ipList := strings.Split(ips, ",")

			req := map[string]interface{}{
				"name":     name,
				"original": ipList,
			}
			if comment != "" {
				req["comment"] = comment
			}

			body, err := json.Marshal(req)
			if err != nil {
				return fmt.Errorf("failed to marshal request: %w", err)
			}

			// Check dry-run
			if cmd.IsDryRun() {
				fmt.Fprintf(c.ErrOrStderr(), "[DRY-RUN] POST /api/IPGroupAPI\n")
				fmt.Fprintf(c.ErrOrStderr(), "Body: %s\n", string(body))
				return nil
			}

			cl := cmd.NewClient()
			env, err := cl.Do("POST", "/api/IPGroupAPI", strings.NewReader(string(body)), nil)
			if err != nil {
				return err
			}

			// Print based on output format
			if cmd.GetOutput() == "json" {
				return cmd.PrintEnvelope(c, env)
			}

			// Key-value output - create returns full details
			var result struct {
				ID      int      `json:"id"`
				Name    string   `json:"name"`
				Comment string   `json:"comment"`
				Cidrs   []string `json:"cidrs"`
			}
			if err := json.Unmarshal(env.Data, &result); err != nil {
				return err
			}
			cmd.PrintKeyValue(map[string]string{
				"ID":      fmt.Sprintf("%d", result.ID),
				"Name":    result.Name,
				"Comment": result.Comment,
				"CIDRs":   strings.Join(result.Cidrs, ", "),
			})
			return nil
		},
	}

	c.Flags().StringVar(&name, "name", "", "IP group name (required)")
	c.Flags().StringVar(&ips, "ips", "", "Comma-separated list of IPs/CIDRs (required)")
	c.Flags().StringVar(&comment, "comment", "", "Comment")

	c.MarkFlagRequired("name")
	c.MarkFlagRequired("ips")

	return c
}

func newDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id> [<id>...]",
		Short: "Delete IP groups",
		Long: `Delete one or more IP groups by ID.

Examples:
  safeline ip-group delete 1
  safeline ip-group delete 1 2 3`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			// Convert args to integers
			ids := make([]int, 0, len(args))
			for _, arg := range args {
				var id int
				if _, err := fmt.Sscanf(arg, "%d", &id); err != nil {
					return fmt.Errorf("invalid ID: %s", arg)
				}
				ids = append(ids, id)
			}

			req := map[string]interface{}{
				"id__in": ids,
			}

			body, err := json.Marshal(req)
			if err != nil {
				return fmt.Errorf("failed to marshal request: %w", err)
			}

			// Check dry-run
			if cmd.IsDryRun() {
				fmt.Fprintf(c.ErrOrStderr(), "[DRY-RUN] DELETE /api/IPGroupAPI\n")
				fmt.Fprintf(c.ErrOrStderr(), "Body: %s\n", string(body))
				return nil
			}

			cl := cmd.NewClient()
			env, err := cl.Do("DELETE", "/api/IPGroupAPI", strings.NewReader(string(body)), nil)
			if err != nil {
				return err
			}

			// Print based on output format
			if cmd.GetOutput() == "json" {
				return cmd.PrintEnvelope(c, env)
			}

			fmt.Printf("Deleted IP group(s): %s\n", strings.Join(args, ", "))
			return nil
		},
	}
}

func newAddIPCmd() *cobra.Command {
	var ips string

	c := &cobra.Command{
		Use:   "add-ip <id>",
		Short: "Add IPs to an IP group",
		Long: `Add IPs to an existing IP group.

Examples:
  safeline ip-group add-ip 1 --ips "192.168.3.0/24,10.0.1.0/24"`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			if ips == "" {
				return fmt.Errorf("--ips is required")
			}

			// Parse IPs
			ipList := strings.Split(ips, ",")

			req := map[string]interface{}{
				"id":      json.Number(args[0]),
				"targets": ipList,
			}

			body, err := json.Marshal(req)
			if err != nil {
				return fmt.Errorf("failed to marshal request: %w", err)
			}

			// Check dry-run
			if cmd.IsDryRun() {
				fmt.Fprintf(c.ErrOrStderr(), "[DRY-RUN] POST /api/EditIPGroupItem\n")
				fmt.Fprintf(c.ErrOrStderr(), "Body: %s\n", string(body))
				return nil
			}

			cl := cmd.NewClient()
			env, err := cl.Do("POST", "/api/EditIPGroupItem", strings.NewReader(string(body)), nil)
			if err != nil {
				return err
			}

			// Print based on output format
			if cmd.GetOutput() == "json" {
				return cmd.PrintEnvelope(c, env)
			}

			fmt.Println("IPs added successfully")
			return nil
		},
	}

	c.Flags().StringVar(&ips, "ips", "", "Comma-separated list of IPs/CIDRs to add (required)")
	c.MarkFlagRequired("ips")

	return c
}

func newRemoveIPCmd() *cobra.Command {
	var ips string

	c := &cobra.Command{
		Use:   "remove-ip <id>",
		Short: "Remove IPs from an IP group",
		Long: `Remove IPs from an existing IP group.

Examples:
  safeline ip-group remove-ip 1 --ips "192.168.3.0/24"`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			if ips == "" {
				return fmt.Errorf("--ips is required")
			}

			// Parse IPs
			ipList := strings.Split(ips, ",")

			req := map[string]interface{}{
				"id":      json.Number(args[0]),
				"targets": ipList,
			}

			body, err := json.Marshal(req)
			if err != nil {
				return fmt.Errorf("failed to marshal request: %w", err)
			}

			// Check dry-run
			if cmd.IsDryRun() {
				fmt.Fprintf(c.ErrOrStderr(), "[DRY-RUN] DELETE /api/EditIPGroupItem\n")
				fmt.Fprintf(c.ErrOrStderr(), "Body: %s\n", string(body))
				return nil
			}

			cl := cmd.NewClient()
			env, err := cl.Do("DELETE", "/api/EditIPGroupItem", strings.NewReader(string(body)), nil)
			if err != nil {
				return err
			}

			// Print based on output format
			if cmd.GetOutput() == "json" {
				return cmd.PrintEnvelope(c, env)
			}

			fmt.Println("IPs removed successfully")
			return nil
		},
	}

	c.Flags().StringVar(&ips, "ips", "", "Comma-separated list of IPs/CIDRs to remove (required)")
	c.MarkFlagRequired("ips")

	return c
}
