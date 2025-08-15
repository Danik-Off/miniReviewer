package cmd

import (
	"fmt"
	"os"

	"miniReviewer/internal/analyzer"
	"miniReviewer/internal/filesystem"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ArchitectureCmd команда для анализа архитектуры
func ArchitectureCmd() *cobra.Command {
	var path, output string

	cmd := &cobra.Command{
		Use:   "architecture",
		Short: "AI-анализ архитектуры проекта",
		Long: `Анализирует архитектуру проекта с использованием AI (Ollama).
Оценивает структуру, предлагает улучшения и выявляет проблемы.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("🏗️  Запуск анализа архитектуры...")
			fmt.Printf("Модель: %s\n", viper.GetString("ollama.default_model"))
			fmt.Printf("Путь: %s\n", path)

			// Анализируем структуру проекта
			ignorePatterns := viper.GetStringSlice("analysis.ignore_patterns")
			scanner := filesystem.NewScanner(ignorePatterns, 0)
			structure, err := scanner.AnalyzeProjectStructure(path)
			if err != nil {
				fmt.Printf("❌ Ошибка анализа структуры: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("📁 Структура проекта:\n%s\n", structure)

			// Анализируем архитектуру с помощью AI
			codeAnalyzer := analyzer.NewCodeAnalyzer()
			result, err := codeAnalyzer.AnalyzeCode(structure, "Project architecture analysis")
			if err != nil {
				fmt.Printf("❌ Ошибка AI-анализа: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("\n📊 Оценка архитектуры: %d/100\n", result.Score)
			for _, issue := range result.Issues {
				fmt.Printf("💡 %s: %s\n", issue.Type, issue.Message)
			}

			fmt.Println("✅ Анализ архитектуры завершен")
		},
	}

	cmd.Flags().StringVar(&path, "path", ".", "путь для анализа")
	cmd.Flags().StringVarP(&output, "output", "o", "", "файл для вывода результата")

	return cmd
}
