# miniReviewer

Консольный помощник для проведения code review с использованием AI (Ollama), написанный на Go.

## 🆕  Возможности

- **Анализ последнего коммита** - команда `./miniReviewer analyze --last`
- **Поддержка JavaScript/TypeScript** - статический анализ + AI размышления
- **Анализ конкретных файлов/папок** - флаг `--path` для всех команд
- **Два режима вывода** - краткий (только проблемы) и подробный (с AI размышлениями)
- **Автоматическое определение языка** - поддержка 8+ языков программирования
- **AI-анализ изменений в git репозитории** с помощью Ollama
- **Проверка качества кода** с использованием LLM и статического анализа
- **Анализ последнего коммита** одной командой (`--last`)
- **Генерация умных отчетов** по review в различных форматах
- **AI-анализ архитектуры** и предложения по улучшению
- **Проверка безопасности кода** с выявлением уязвимостей
- **Поддержка множества языков**: Go, JavaScript, TypeScript, Python, Java, C++, Rust, Kotlin
- **Два режима вывода**: краткий (только проблемы) и подробный (с размышлениями AI)
- **Интеграция с популярными моделями** Ollama (gemma, codellama, llama2, mistral)
- **Анализ конкретных файлов или папок** с помощью флага `--path`

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
ollama pull gemma3:
ollama pull codellama:13b
ollama pull llama2:7b
ollama pull mistral:7b
```

## Использование

### Основные команды

```bash
# AI-анализ изменений в git репозитории
./miniReviewer analyze                    # Анализ текущих изменений
./miniReviewer analyze --last             # Анализ последнего коммита
./miniReviewer analyze --commit <hash>    # Анализ конкретного коммита
./miniReviewer analyze --from main --to feature-branch  # Анализ между ветками

# Проверка качества кода
./miniReviewer quality                     # Анализ всех Go файлов в текущей директории
./miniReviewer quality --path file.js      # Анализ конкретного файла
./miniReviewer quality --path folder/      # Анализ папки
./miniReviewer quality --verbose           # Подробный вывод с размышлениями AI

# Анализ безопасности
./miniReviewer security                    # Проверка безопасности всех файлов
./miniReviewer security --path file.js     # Проверка безопасности конкретного файла
./miniReviewer security --verbose          # Подробный анализ с AI

# Анализ архитектуры
./miniReviewer architecture                # Анализ архитектуры проекта
./miniReviewer architecture --path file.go # Анализ архитектуры конкретного файла

# Генерация отчетов
./miniReviewer report --format html        # HTML отчет
./miniReviewer report --format markdown    # Markdown отчет
./miniReviewer report --format json        # JSON отчет

# Тестирование Ollama
./miniReviewer test-ollama                # Проверка подключения к Ollama

# Помощь
./miniReviewer --help                      # Общая справка
./miniReviewer quality --help              # Справка по команде quality
```

### Флаги

#### Глобальные флаги
- `--model <model>` - указать модель Ollama (по умолчанию: gemma3:latest)
- `--verbose` - подробный вывод с размышлениями AI
- `--config <file>` - указать конфигурационный файл (по умолчанию: .miniReviewer.yaml)

#### Флаги команды analyze
- `--last` - анализ последнего коммита
- `--commit <hash>` - анализ конкретного коммита
- `--from <branch>` - исходная ветка/коммит для сравнения
- `--to <branch>` - целевая ветка/коммит для сравнения
- `--output <file>` - файл для сохранения результата
- `--ignore <pattern>` - игнорировать файлы по паттерну

#### Флаги команды quality
- `--path <path>` - путь к файлу или папке для анализа (по умолчанию: текущая директория)
- `--severity <level>` - уровень важности (low, medium, high, critical)
- `--output <file>` - файл для сохранения результата
- `--ignore <pattern>` - игнорировать файлы по паттерну

#### Флаги команды security
- `--path <path>` - путь к файлу или папке для анализа
- `--check-dependencies` - проверка зависимостей на уязвимости
- `--scan-code` - сканирование кода на проблемы безопасности
- `--output <file>` - файл для сохранения результата

#### Флаги команды architecture
- `--path <path>` - путь к файлу или папке для анализа
- `--output <file>` - файл для сохранения результата

#### Флаги команды report
- `--format <format>` - формат отчета (html, markdown, json)
- `--output <file>` - файл для сохранения отчета

## Примеры использования

### Анализ последнего коммита
```bash
# Быстрый анализ последнего коммита
./miniReviewer analyze --last

# Подробный анализ с размышлениями AI
./miniReviewer analyze --last --verbose
```

### Анализ конкретного файла
```bash
# Анализ JavaScript файла
./miniReviewer quality --path src/app.js

# Подробный анализ с AI
./miniReviewer quality --path src/app.js --verbose

# Анализ Go файла
./miniReviewer quality --path internal/handler.go
```

### Анализ папки
```bash
# Анализ всей папки src/
./miniReviewer quality --path src/

# Анализ с игнорированием определенных паттернов
./miniReviewer quality --path src/ --ignore "*.test.js" --ignore "vendor/*"
```

### Анализ безопасности
```bash
# Проверка безопасности всех файлов
./miniReviewer security

# Проверка конкретного файла
./miniReviewer security --path src/auth.js --verbose
```

### Сравнение веток
```bash
# Анализ изменений между main и feature веткой
./miniReviewer analyze --from main --to feature/new-feature --verbose
```

### Генерация отчетов
```bash
# HTML отчет
./miniReviewer report --format html --output review-report.html

# Markdown отчет
./miniReviewer report --format markdown --output review-report.md
```

## Режимы вывода

miniReviewer поддерживает два режима вывода для максимальной гибкости:

### Краткий режим (по умолчанию)
Без флага `--verbose` показывает только основные проблемы:
```
⚠️  [HIGH] security (строка 2): Использование eval() - опасная функция
⚠️  [MEDIUM] quality (строка 6): Небезопасное сравнение == вместо ===
```

### Подробный режим (--verbose)
С флагом `--verbose` показывает размышления AI и детальную информацию:
```
⚠️  [HIGH] security (строка 2):
   💬 Использование eval() - опасная функция
   💡 Заменить eval() на безопасную альтернативу
   🧠 eval() может выполнить произвольный код и создать уязвимости безопасности
```

## Поддерживаемые языки

miniReviewer автоматически определяет тип файла и применяет соответствующие правила анализа:

- **Go** (.go) - AI-анализ + статические проверки
- **JavaScript** (.js) - Статический анализ + AI для размышлений
- **TypeScript** (.ts) - Статический анализ + AI для размышлений
- **Python** (.py) - AI-анализ
- **Java** (.java) - AI-анализ
- **C++** (.cpp, .cc, .cxx) - AI-анализ
- **Rust** (.rs) - AI-анализ
- **Kotlin** (.kt) - AI-анализ

### Конфигурация

Создайте файл `.miniReviewer.yaml` в корне проекта:

```yaml
# Настройки Ollama
ollama:
  host: "http://localhost:11434"
  default_model: "gemma3:latest"
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

## Дополнительные возможности

### Автоматическое определение языка
miniReviewer автоматически определяет тип файла по расширению и применяет соответствующие правила анализа. Для JavaScript и TypeScript файлов используется комбинация статического анализа и AI для получения размышлений.

### Гибкая настройка игнорирования
```bash
# Игнорирование тестовых файлов
./miniReviewer quality --ignore "*.test.js" --ignore "*.spec.js"

# Игнорирование папок
./miniReviewer quality --ignore "node_modules/*" --ignore "dist/*"
```

### Интеграция с CI/CD
miniReviewer можно легко интегрировать в CI/CD пайплайны для автоматической проверки качества кода:

```bash
# Проверка качества в CI
./miniReviewer quality --path src/ --output quality-report.json

# Анализ последнего коммита в CI
./miniReviewer analyze --last --output commit-analysis.json
```
./miniReviewer security --scan-code --model codellama:13b
```

## Структура проекта

```
miniReviewer/
├── main.go                    # Основной файл приложения
├── cmd/                       # Команды CLI
│   ├── analyze.go            # Команда AI-анализа git изменений
│   ├── quality.go            # Команда проверки качества кода
│   ├── security.go           # Команда анализа безопасности
│   ├── architecture.go       # Команда анализа архитектуры
│   ├── report.go             # Команда генерации отчетов
│   ├── test-ollama.go        # Тестирование подключения к Ollama
│   └── version.go            # Информация о версии
├── internal/                  # Внутренняя логика
│   ├── analyzer/             # Анализаторы кода
│   │   └── code.go           # AI и статические анализаторы
│   ├── git/                  # Git интеграция
│   ├── filesystem/           # Работа с файловой системой
│   ├── ollama/               # Интеграция с Ollama
│   ├── reporter/             # Генераторы отчетов
│   └── types/                # Общие типы данных
├── go.mod                    # Файл модуля Go
├── go.sum                    # Файл зависимостей
├── .miniReviewer.yaml        # Конфигурация по умолчанию
└── README.md                 # Документация
```

## Модели Ollama для анализа кода

### Рекомендуемые модели

1. **gemma3:latest** - быстрый анализ, хорошее качество
2. **codellama:13b** - высокое качество анализа, медленнее
3. **llama2:7b** - универсальная модель, хороший баланс
4. **mistral:7b** - отличное качество для европейских языков

### Загрузка моделей

```bash
# Загрузка CodeLlama (рекомендуется для анализа кода)
ollama pull gemma3:latest

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
ollama pull gemma3:latest
```

## Лицензия

MIT License
