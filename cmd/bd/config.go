package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/untoldecay/BeadsLog/internal/config"
	"github.com/untoldecay/BeadsLog/internal/debug"
	"github.com/untoldecay/BeadsLog/internal/storage"
	"github.com/untoldecay/BeadsLog/internal/syncbranch"
)

var configCmd = &cobra.Command{
	Use:     "config",
	GroupID: "setup",
	Short:   "Manage configuration settings",
	Long: `Manage configuration settings for external integrations and preferences.

Configuration is stored per-project in .beads/*.db and is version-control-friendly.

Common namespaces:
  - jira.*       Jira integration settings
  - linear.*     Linear integration settings
  - github.*     GitHub integration settings
  - custom.*     Custom integration settings
  - status.*     Issue status configuration

Custom Status States:
  You can define custom status states for multi-step pipelines using the
  status.custom config key. Statuses should be comma-separated.

  Example:
    bd config set status.custom "awaiting_review,awaiting_testing,awaiting_docs"

  This enables issues to use statuses like 'awaiting_review' in addition to
  the built-in statuses (open, in_progress, blocked, deferred, closed).

Examples:
  bd config set jira.url "https://company.atlassian.net"
  bd config set jira.project "PROJ"
  bd config set status.custom "awaiting_review,awaiting_testing"
  bd config get jira.url
  bd config list
  bd config unset jira.url`,
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	Run: func(_ *cobra.Command, args []string) {
		key := args[0]
		value := args[1]

		// Check if this is a yaml-only key (startup settings like no-db, no-daemon, etc.)
		// These must be written to config.yaml, not SQLite, because they're read
		// before the database is opened. (GH#536)
		if config.IsYamlOnlyKey(key) {
			if err := config.SetYamlConfig(key, value); err != nil {
				fmt.Fprintf(os.Stderr, "Error setting config: %v\n", err)
				os.Exit(1)
			}

			if jsonOutput {
				outputJSON(map[string]interface{}{
					"key":      key,
					"value":    value,
					"location": "config.yaml",
				})
			} else {
				fmt.Printf("Set %s = %s (in config.yaml)\n", key, value)
			}
			return
		}

		// Database-stored config requires direct mode
		if err := ensureDirectMode("config set requires direct database access"); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		ctx := rootCtx

		// Special handling for sync.branch to apply validation
		if strings.TrimSpace(key) == syncbranch.ConfigKey {
			if err := syncbranch.Set(ctx, store, value); err != nil {
				fmt.Fprintf(os.Stderr, "Error setting config: %v\n", err)
				os.Exit(1)
			}
		} else {
			if err := store.SetConfig(ctx, key, value); err != nil {
				fmt.Fprintf(os.Stderr, "Error setting config: %v\n", err)
				os.Exit(1)
			}
		}

		if jsonOutput {
			outputJSON(map[string]string{
				"key":   key,
				"value": value,
			})
		} else {
			fmt.Printf("Set %s = %s\n", key, value)
		}
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]

		// Check if this is a yaml-only key (startup settings)
		// These are read from config.yaml via viper, not SQLite. (GH#536)
		if config.IsYamlOnlyKey(key) {
			value := config.GetYamlConfig(key)

			if jsonOutput {
				outputJSON(map[string]interface{}{
					"key":      key,
					"value":    value,
					"location": "config.yaml",
				})
			} else {
				if value == "" {
					fmt.Printf("%s (not set in config.yaml)\n", key)
				} else {
					fmt.Printf("%s\n", value)
				}
			}
			return
		}

		// Database-stored config requires direct mode
		if err := ensureDirectMode("config get requires direct database access"); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		ctx := rootCtx
		var value string
		var err error

		// Special handling for sync.branch to support env var override
		if strings.TrimSpace(key) == syncbranch.ConfigKey {
			value, err = syncbranch.Get(ctx, store)
		} else {
			value, err = store.GetConfig(ctx, key)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting config: %v\n", err)
			os.Exit(1)
		}

		if jsonOutput {
			outputJSON(map[string]string{
				"key":   key,
				"value": value,
			})
		} else {
			if value == "" {
				fmt.Printf("%s (not set)\n", key)
			} else {
				fmt.Printf("%s\n", value)
			}
		}
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration",
	Run: func(cmd *cobra.Command, args []string) {
		// Get all settings from Viper (config.yaml and defaults)
		allViperSettings := config.AllSettings()

		// Try to get database-stored config (if DB is accessible)
		dbConfig := make(map[string]string)
		var dbErr error
		if storeActive && store != nil { // Check if store is initialized and active
			dbConfig, dbErr = store.GetAllConfig(rootCtx)
			if dbErr != nil && dbErr != storage.ErrDBNotInitialized {
				fmt.Fprintf(os.Stderr, "Warning: failed to retrieve database config: %v\n", dbErr)
			}
		} else {
			// If store is not active or initialized, ensure direct mode error is shown if it's a DB operation.
			// For list, we're okay showing YAML config even without DB.
			debug.Logf("database not active or initialized, skipping DB config list")
		}

		if jsonOutput {
			outputJSON(map[string]interface{}{
				"yaml_settings": allViperSettings,
				"db_settings":   dbConfig,
				"db_error":      dbErr.Error(), // Only if error occurred
			})
			return
		}

		fmt.Println("\nConfiguration:")

		// Collect all unique keys from both sources for sorted display
		allKeys := make(map[string]bool)
		for k := range allViperSettings {
			allKeys[k] = true
		}
		for k := range dbConfig {
			allKeys[k] = true
		}

		sortedKeys := make([]string, 0, len(allKeys))
		for k := range allKeys {
			sortedKeys = append(sortedKeys, k)
		}
		sort.Strings(sortedKeys)

		// Display settings
		var yamlCount, dbCount int
		for _, k := range sortedKeys {
			// Check if it's a YAML-only key or explicitly set in YAML
			isYamlOnly := config.IsYamlOnlyKey(k)
			valFromViper := config.GetYamlConfig(k) // Gets effective value from Viper (which considers defaults)

			if dbVal, ok := dbConfig[k]; ok {
				// Key exists in DB
				if isYamlOnly {
					// Should not happen for IsYamlOnlyKey to be in DB config, but handle defensively
					fmt.Printf("  %s = %s (DB/YAML conflict, DB value shown)\n", k, dbVal)
				} else {
					// DB-stored config
					fmt.Printf("  %s = %s (DB)\n", k, dbVal)
				}
				dbCount++
			} else if valFromViper != "" {
				// Key is not in DB, but has a value from Viper (YAML or default)
				source := "YAML"
				if config.GetValueSource(k) == config.SourceDefault {
					source = "Default"
				}
				fmt.Printf("  %s = %s (%s)\n", k, valFromViper, source)
				yamlCount++
			}
		}

		if yamlCount == 0 && dbCount == 0 {
			fmt.Println("  No configuration set.")
		}

		// Show config.yaml overrides for DB values
		showConfigYAMLOverrides(dbConfig)
	},
}

// showConfigYAMLOverrides warns when config.yaml or env vars override database settings.
// This addresses the confusion when `bd config list` shows one value but the effective
// value used by commands is different due to higher-priority config sources.
func showConfigYAMLOverrides(dbConfig map[string]string) {
	var overrides []string

	// Check sync.branch - can be overridden by BEADS_SYNC_BRANCH env var or config.yaml sync-branch
	if dbSyncBranch, ok := dbConfig[syncbranch.ConfigKey]; ok && dbSyncBranch != "" {
		effectiveBranch := syncbranch.GetFromYAML()
		if effectiveBranch != "" && effectiveBranch != dbSyncBranch {
			overrides = append(overrides, fmt.Sprintf("  sync.branch: database has '%s' but effective value is '%s' (from config.yaml or env)", dbSyncBranch, effectiveBranch))
		}
	}

	if len(overrides) > 0 {
		fmt.Println("\n⚠️  Config overrides (higher priority sources):")
		for _, o := range overrides {
			fmt.Println(o)
		}
		fmt.Println("\nNote: config.yaml and environment variables take precedence over database config.")
	}
}

var configUnsetCmd = &cobra.Command{
	Use:   "unset <key>",
	Short: "Delete a configuration value",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Config operations work in direct mode only
		if err := ensureDirectMode("config unset requires direct database access"); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		key := args[0]

		ctx := rootCtx
		if err := store.DeleteConfig(ctx, key); err != nil {
			fmt.Fprintf(os.Stderr, "Error deleting config: %v\n", err)
			os.Exit(1)
		}

		if jsonOutput {
			outputJSON(map[string]string{
				"key": key,
			})
		} else {
			fmt.Printf("Unset %s\n", key)
		}
	},
}

func init() {
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configUnsetCmd)
	rootCmd.AddCommand(configCmd)
}
