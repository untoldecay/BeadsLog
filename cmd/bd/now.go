package main

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var nowCmd = &cobra.Command{
	Use:   "now",
	Short: "Show current system time in devlog format",
	Long:  `Displays the current time formatted for devlog headers and index entries (YYYY-MM-DD HH:MM).`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(time.Now().Format("2006-01-02 15:04"))
	},
}

func init() {
	rootCmd.AddCommand(nowCmd)
}
