package main

import (
	"fmt"
	"log"
	"os"

	"miniReviewer/cmd"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	verbose bool
	model   string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "miniReviewer",
		Short: "AI-powered code review assistant using Ollama",
		Long: `miniReviewer - это консольный помощник для проведения code review 
с использованием AI (Ollama). Он анализирует код, предлагает улучшения 
и генерирует подробные отчеты.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Инициализация конфигурации
			initConfig()
		},
	}

	// Глобальные флаги
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "конфигурационный файл (по умолчанию .miniReviewer.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "подробный вывод")
	rootCmd.PersistentFlags().StringVar(&model, "model", "gemma3n:e4b", "модель Ollama для использования")

	// Команды
	rootCmd.AddCommand(cmd.AnalyzeCmd())
	rootCmd.AddCommand(cmd.QualityCmd())
	rootCmd.AddCommand(cmd.SecurityCmd())
	rootCmd.AddCommand(cmd.ArchitectureCmd())
	rootCmd.AddCommand(cmd.ReportCmd())
	rootCmd.AddCommand(cmd.VersionCmd())
	rootCmd.AddCommand(cmd.TestOllamaCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка: %v\n", err)
		os.Exit(1)
	}
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName(".miniReviewer")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME")
	}

	// Значения по умолчанию
	viper.SetDefault("ollama.host", "http://localhost:11434")
	viper.SetDefault("ollama.default_model", "gemma3n:e4b")
	viper.SetDefault("ollama.max_tokens", 4000)
	viper.SetDefault("ollama.temperature", 0.1)
	viper.SetDefault("ollama.timeout", "300s")

	viper.SetDefault("analysis.languages", []string{"go", "python", "javascript", "typescript", "java", "c++"})
	viper.SetDefault("analysis.ignore_patterns", []string{"vendor/*", "node_modules/*", "*.min.js", "*.min.css"})
	viper.SetDefault("analysis.max_file_size", "1MB")

	viper.SetDefault("quality.max_complexity", 10)
	viper.SetDefault("quality.max_function_length", 50)
	viper.SetDefault("quality.max_file_length", 1000)
	viper.SetDefault("quality.enable_ai_suggestions", true)

	viper.SetDefault("security.enabled", true)
	viper.SetDefault("security.check_dependencies", true)
	viper.SetDefault("security.ai_vulnerability_scan", true)

	viper.SetDefault("reports.format", "html")
	viper.SetDefault("reports.include_metrics", true)
	viper.SetDefault("reports.include_ai_suggestions", true)
	viper.SetDefault("reports.include_code_examples", true)

	// Чтение конфигурации
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Printf("Ошибка чтения конфигурации: %v", err)
		}
	}

	// Применение флагов командной строки
	if model != "" {
		viper.Set("ollama.default_model", model)
	}
	if verbose {
		viper.Set("verbose", verbose)
	}
}
