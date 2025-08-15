# miniReviewer

Консольный помощник для проведения code review с использованием AI (Ollama), написанный на Go.

## Возможности

- AI-анализ изменений в git репозитории с помощью Ollama
- Проверка качества кода с использованием LLM
- Генерация умных отчетов по review
- Анализ архитектуры и предложения по улучшению
- Проверка безопасности кода
- Поддержка различных языков программирования
- Интеграция с популярными моделями Ollama (codellama, llama2, mistral)

## Требования

- Go 1.19 или выше
- Git репозиторий для анализа
- Ollama (установленный и запущенный)
- Доступ к интернету для загрузки моделей

## Установка

### 1. Установка Ollama

```bash
# macOS
curl -fsSL https://ollama.ai/install.sh | sh

# Linux
curl -fsSL https://ollama.ai/install.sh | sh

# Windows
# Скачайте с https://ollama.ai/download
```

### 2. Установка miniReviewer

```bash
git clone <repository-url>
cd miniReviewer
go build -o miniReviewer
```

### 3. Загрузка модели для анализа кода

```bash
# Рекомендуемые модели для анализа кода
ollama pull codellama:7b
ollama pull codellama:13b
ollama pull llama2:7b
ollama pull mistral:7b
```

## Использование

### Основные команды

```bash
# AI-анализ изменений в текущей ветке
./miniReviewer analyze

# Анализ конкретного коммита с AI
./miniReviewer analyze --commit <hash> --model codellama:7b

# Анализ изменений между ветками
./miniReviewer analyze --from main --to feature-branch --verbose

# AI-проверка качества кода
./miniReviewer quality --model codellama:13b

# Генерация AI-отчета
./miniReviewer report --output report.html --model mistral:7b

# AI-анализ архитектуры
./miniReviewer architecture --model codellama:7b

# AI-проверка безопасности
./miniReviewer security --model codellama:7b

# AI-анализ стиля кода
./miniReviewer style --model llama2:7b

# Проверка покрытия тестами
./miniReviewer coverage

# Помощь
./miniReviewer --help
```

### Флаги

- `--model <model>` - указать модель Ollama (по умолчанию: codellama:7b)
- `--verbose` - подробный вывод
- `--output <file>` - указать файл для вывода
- `--config <file>` - указать конфигурационный файл
- `--ignore <pattern>` - игнорировать файлы по паттерну
- `--severity <level>` - уровень важности (low, medium, high, critical)
- `--max-tokens <number>` - максимальное количество токенов для AI-ответа
- `--temperature <float>` - температура для AI-генерации (0.0-1.0)

### Конфигурация

Создайте файл `.miniReviewer.yaml` в корне проекта:

```yaml
# Настройки Ollama
ollama:
  host: "http://localhost:11434"
  default_model: "codellama:7b"
  max_tokens: 4000
  temperature: 0.1
  timeout: 300s

# Настройки анализа
analysis:
  languages: ["go", "python", "javascript", "typescript", "java", "c++"]
  ignore_patterns: ["vendor/*", "node_modules/*", "*.min.js", "*.min.css"]
  max_file_size: "1MB"
  
# Настройки качества
quality:
  max_complexity: 10
  max_function_length: 50
  max_file_length: 1000
  enable_ai_suggestions: true
  
# Настройки безопасности
security:
  enabled: true
  check_dependencies: true
  ai_vulnerability_scan: true
  
# Настройки отчетов
reports:
  format: "html"
  include_metrics: true
  include_ai_suggestions: true
  include_code_examples: true
```

## Примеры использования

### AI-анализ pull request

```bash
# Анализ изменений с использованием codellama
./miniReviewer analyze --from main --to feature/new-feature --model codellama:7b --verbose

# Генерация HTML отчета с AI-рекомендациями
./miniReviewer analyze --from main --to feature/new-feature --output pr-review.html --model codellama:13b
```

### AI-проверка качества кода

```bash
# Проверка текущих изменений с AI
./miniReviewer quality --model codellama:7b --severity high

# Проверка с игнорированием определенных файлов
./miniReviewer quality --ignore "*.test.go" --ignore "vendor/*" --model mistral:7b
```

### AI-анализ архитектуры

```bash
# Анализ архитектуры проекта
./miniReviewer architecture --model codellama:13b --verbose

# Анализ конкретной директории
./miniReviewer architecture --path ./internal --model codellama:7b
```

### AI-анализ безопасности

```bash
# Проверка зависимостей на уязвимости
./miniReviewer security --check-dependencies --model codellama:7b

# Сканирование кода на проблемы безопасности
./miniReviewer security --scan-code --model codellama:13b
```

## Структура проекта

```
miniReviewer/
├── main.go                    # Основной файл приложения
├── cmd/                       # Команды CLI
│   ├── analyze.go            # Команда AI-анализа
│   ├── quality.go            # Команда проверки качества
│   ├── security.go           # Команда безопасности
│   ├── architecture.go       # Команда анализа архитектуры
│   └── report.go             # Команда генерации отчетов
├── internal/                  # Внутренняя логика
│   ├── analyzer/             # Анализаторы кода
│   │   ├── ai/               # AI-анализаторы (Ollama)
│   │   ├── static/           # Статические анализаторы
│   │   └── git/              # Git-анализаторы
│   ├── ollama/               # Интеграция с Ollama
│   ├── reporter/             # Генераторы отчетов
│   └── config/               # Конфигурация
├── pkg/                      # Публичные пакеты
├── go.mod                    # Файл модуля Go
├── go.sum                    # Файл зависимостей
├── .miniReviewer.yaml        # Конфигурация по умолчанию
└── README.md                 # Документация
```

## Модели Ollama для анализа кода

### Рекомендуемые модели

1. **codellama:7b** - быстрый анализ, хорошее качество
2. **codellama:13b** - высокое качество анализа, медленнее
3. **llama2:7b** - универсальная модель, хороший баланс
4. **mistral:7b** - отличное качество для европейских языков

### Загрузка моделей

```bash
# Загрузка CodeLlama (рекомендуется для анализа кода)
ollama pull codellama:7b

# Загрузка Llama2
ollama pull llama2:7b

# Загрузка Mistral
ollama pull mistral:7b
```

## Разработка

```bash
# Запуск тестов
go test ./...

# Запуск с race detector
go test -race ./...

# Проверка покрытия
go test -cover ./...

# Линтинг
golangci-lint run

# Запуск Ollama (для тестирования)
ollama serve
```

## Troubleshooting

### Ollama не отвечает

```bash
# Проверка статуса Ollama
ollama list

# Перезапуск Ollama
ollama serve

# Проверка порта
lsof -i :11434
```

### Модель не найдена

```bash
# Список доступных моделей
ollama list

# Загрузка модели
ollama pull codellama:7b
```

## Лицензия

MIT License
