package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"miniReviewer/internal/analyzer"
	"miniReviewer/internal/git"
	"miniReviewer/internal/types"
)

// AnalyzeCmd команда для анализа кода
func AnalyzeCmd() *cobra.Command {
	var from, to, commit, output string
	var ignore []string

	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "AI-анализ изменений в коде",
		Long: `Анализирует изменения в git репозитории с использованием AI (Ollama).
Может анализировать коммиты, ветки или текущие изменения.`,
		Run: func(cmd *cobra.Command, args []string) {
			verbose := viper.GetBool("verbose")
			
			fmt.Println("🚀 Запуск AI-анализа...")
			fmt.Printf("Модель: %s\n", viper.GetString("ollama.default_model"))
			
			if verbose {
				fmt.Println("🔍 Подробный режим включен")
				fmt.Printf("Параметры анализа:\n")
				fmt.Printf("  - От: %s\n", from)
				fmt.Printf("  - До: %s\n", to)
				fmt.Printf("  - Коммит: %s\n", commit)
				fmt.Printf("  - Выходной файл: %s\n", output)
				fmt.Printf("  - Игнорируемые паттерны: %v\n", ignore)
			}
			
			// Проверяем, что мы в git репозитории
			if verbose {
				fmt.Println("🔍 Проверяю git репозиторий...")
			}
			
			gitClient := git.NewClient()
			if !gitClient.IsRepository() {
				fmt.Println("❌ Git репозиторий не найден. Убедитесь, что вы находитесь в git репозитории.")
				os.Exit(1)
			}
			
			if verbose {
				fmt.Println("✅ Git репозиторий найден")
			}

			// Получаем diff
			var diff string
			var err error
			
			if verbose {
				fmt.Println("📝 Получаю изменения...")
			}
			
			if commit != "" {
				fmt.Printf("Анализ коммита: %s\n", commit)
				diff, err = gitClient.GetDiff(commit, commit+"~1")
			} else if from != "" && to != "" {
				fmt.Printf("Анализ изменений от %s до %s\n", from, to)
				diff, err = gitClient.GetDiff(from, to)
			} else {
				fmt.Println("Анализ текущих изменений")
				diff, err = gitClient.GetStatus()
			}
			
			if err != nil {
				fmt.Printf("❌ Ошибка получения изменений: %v\n", err)
				os.Exit(1)
			}

			if diff == "" {
				fmt.Println("✅ Нет изменений для анализа")
				return
			}
			
			if verbose {
				fmt.Printf("📄 Размер изменений: %d символов\n", len(diff))
			}

			if len(ignore) > 0 {
				fmt.Printf("Игнорируемые паттерны: %v\n", ignore)
			}
			
			fmt.Println("📝 Анализирую код с помощью AI...")
			
			if verbose {
				fmt.Println("🧠 Запускаю AI-анализ...")
			}
			
			// Анализируем код
			codeAnalyzer := analyzer.NewCodeAnalyzer()
			result, err := codeAnalyzer.AnalyzeCode(diff, "Git changes analysis")
			if err != nil {
				fmt.Printf("❌ Ошибка AI-анализа: %v\n", err)
				os.Exit(1)
			}
			
			if verbose {
				fmt.Println("✅ AI-анализ завершен успешно")
			}

			// Выводим результат
			fmt.Printf("\n📊 Результат анализа:\n")
			fmt.Printf("Оценка: %d/100\n", result.Score)
			fmt.Printf("Найдено проблем: %d\n", len(result.Issues))
			
			if verbose {
				fmt.Printf("📈 Детальная информация:\n")
				fmt.Printf("  - Временная метка: %s\n", result.Timestamp.Format("2006-01-02 15:04:05"))
				fmt.Printf("  - Файл: %s\n", result.File)
			}
			
			for i, issue := range result.Issues {
				if verbose {
					// Подробный вывод с размышлениями модели
					fmt.Printf("\n%d. [%s] %s (строка %d):\n", i+1, strings.ToUpper(issue.Severity), issue.Type, issue.Line)
					fmt.Printf("   💬 Проблема: %s\n", issue.Message)
					if issue.Suggestion != "" {
						fmt.Printf("   💡 Предложение: %s\n", issue.Suggestion)
					}
					if issue.Reasoning != "" {
						fmt.Printf("   🧠 %s\n", issue.Reasoning)
					}
				} else {
					// Краткий вывод - только проблема и строка
					if issue.Line > 0 {
						fmt.Printf("\n%d. [%s] %s (строка %d): %s\n", i+1, strings.ToUpper(issue.Severity), issue.Type, issue.Line, issue.Message)
					} else {
						fmt.Printf("\n%d. [%s] %s: %s\n", i+1, strings.ToUpper(issue.Severity), issue.Type, issue.Message)
					}
				}
			}

			// Сохраняем в файл если указан
			if output != "" {
				if verbose {
					fmt.Printf("💾 Сохраняю результат в файл: %s\n", output)
				}
				
				if err := saveResultToFile(result, output); err != nil {
					fmt.Printf("❌ Ошибка сохранения: %v\n", err)
				} else {
					fmt.Printf("\n💾 Результат сохранен в: %s\n", output)
				}
			}
			
			fmt.Println("\n✅ Анализ завершен")
		},
	}

	cmd.Flags().StringVar(&from, "from", "", "исходная ветка/коммит")
	cmd.Flags().StringVar(&to, "to", "", "целевая ветка/коммит")
	cmd.Flags().StringVar(&commit, "commit", "", "анализ конкретного коммита")
	cmd.Flags().StringVarP(&output, "output", "o", "", "файл для вывода результата")
	cmd.Flags().StringArrayVar(&ignore, "ignore", []string{}, "паттерны для игнорирования")

	return cmd
}

// saveResultToFile сохраняет результат анализа в файл
func saveResultToFile(result *types.CodeAnalysisResult, filename string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}
