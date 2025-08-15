package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"miniReviewer/internal/ollama"
)

// TestOllamaCmd команда для тестирования подключения к Ollama
func TestOllamaCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "test-ollama",
		Short: "Тест подключения к Ollama",
		Run: func(cmd *cobra.Command, args []string) {
			verbose := viper.GetBool("verbose")
			
			fmt.Println("🧪 Тестирование подключения к Ollama...")
			
			host := viper.GetString("ollama.host")
			fmt.Printf("Хост: %s\n", host)
			
			if verbose {
				fmt.Println("🔍 Подробный режим включен")
				fmt.Printf("Настройки Ollama:\n")
				fmt.Printf("  - Модель по умолчанию: %s\n", viper.GetString("ollama.default_model"))
				fmt.Printf("  - Максимальные токены: %d\n", viper.GetInt("ollama.max_tokens"))
				fmt.Printf("  - Температура: %.2f\n", viper.GetFloat64("ollama.temperature"))
				fmt.Printf("  - Таймаут: %s\n", viper.GetString("ollama.timeout"))
			}
			
			// Проверяем доступность Ollama
			if verbose {
				fmt.Println("🔍 Проверяю доступность Ollama...")
			}
			
			client := ollama.NewClient()
			if err := client.HealthCheck(); err != nil {
				fmt.Printf("❌ Не удается подключиться к Ollama: %v\n", err)
				fmt.Println("Убедитесь, что Ollama запущен: ollama serve")
				os.Exit(1)
			}
			
			fmt.Println("✅ Подключение к Ollama успешно!")
			
			if verbose {
				fmt.Println("📋 Получаю список доступных моделей...")
				models, err := client.GetModels()
				if err != nil {
					fmt.Printf("⚠️  Ошибка получения моделей: %v\n", err)
				} else {
					fmt.Printf("📚 Доступные модели (%d):\n", len(models))
					for i, model := range models {
						fmt.Printf("  %d. %s\n", i+1, model)
					}
				}
			}
			
			// Тестируем простой запрос
			fmt.Println("🧠 Тестирую AI-запрос...")
			
			if verbose {
				fmt.Println("📝 Отправляю тестовый запрос...")
			}
			
			response, err := client.Generate("Скажи 'Привет' на русском языке")
			if err != nil {
				fmt.Printf("❌ Ошибка AI-запроса: %v\n", err)
				os.Exit(1)
			}
			
			if verbose {
				fmt.Printf("📨 Получен ответ (размер: %d символов)\n", len(response))
			}
			
			fmt.Printf("🤖 Ответ AI: %s\n", response)
			fmt.Println("✅ Ollama работает корректно!")
		},
	}
}
