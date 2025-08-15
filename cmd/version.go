package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// VersionCmd команда для показа версии
func VersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Показать версию miniReviewer",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("miniReviewer v0.1.0")
			fmt.Println("AI-powered code review assistant using Ollama")
		},
	}
}
