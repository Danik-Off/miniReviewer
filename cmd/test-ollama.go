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
			fmt.Println("🧪 Тестирование подключения к Ollama...")
			
			host := viper.GetString("ollama.host")
			fmt.Printf("Хост: %s\n", host)
			
			// Проверяем доступность Ollama
			client := ollama.NewClient()
			if err := client.HealthCheck(); err != nil {
				fmt.Printf("❌ Не удается подключиться к Ollama: %v\n", err)
				fmt.Println("Убедитесь, что Ollama запущен: ollama serve")
				os.Exit(1)
			}
			
			fmt.Println("✅ Подключение к Ollama успешно!")
			
			// Тестируем простой запрос
			fmt.Println("🧠 Тестирую AI-запрос...")
			response, err := client.Generate("Скажи 'Привет' на русском языке")
			if err != nil {
				fmt.Printf("❌ Ошибка AI-запроса: %v\n", err)
				os.Exit(1)
			}
			
			fmt.Printf("🤖 Ответ AI: %s\n", response)
			fmt.Println("✅ Ollama работает корректно!")
		},
	}
}
