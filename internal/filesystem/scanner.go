package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Scanner сканер файловой системы
type Scanner struct {
	ignorePatterns []string
	maxFileSize    int64
}

// NewScanner создает новый сканер файловой системы
func NewScanner(ignorePatterns []string, maxFileSize int64) *Scanner {
	return &Scanner{
		ignorePatterns: ignorePatterns,
		maxFileSize:    maxFileSize,
	}
}

// FindGoFiles находит все Go файлы в директории
func (s *Scanner) FindGoFiles(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			// Проверяем размер файла
			if s.maxFileSize > 0 && info.Size() > s.maxFileSize {
				return nil
			}
			// Проверяем паттерны игнорирования
			if !s.shouldIgnoreFile(path) {
				files = append(files, path)
			}
		}
		return nil
	})
	return files, err
}

// FindFilesByExtension находит файлы по расширению
func (s *Scanner) FindFilesByExtension(root, extension string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, extension) {
			// Проверяем размер файла
			if s.maxFileSize > 0 && info.Size() > s.maxFileSize {
				return nil
			}
			// Проверяем паттерны игнорирования
			if !s.shouldIgnoreFile(path) {
				files = append(files, path)
			}
		}
		return nil
	})
	return files, err
}

// FindSupportedFiles находит все поддерживаемые файлы в директории
func (s *Scanner) FindSupportedFiles(root string) ([]string, error) {
	var files []string
	supportedExtensions := []string{".go", ".js", ".ts", ".py", ".java", ".cpp", ".rs", ".kt"}
	
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			for _, supportedExt := range supportedExtensions {
				if ext == supportedExt {
					// Проверяем размер файла
					if s.maxFileSize > 0 && info.Size() > s.maxFileSize {
						return nil
					}
					// Проверяем паттерны игнорирования
					if !s.shouldIgnoreFile(path) {
						files = append(files, path)
					}
					break
				}
			}
		}
		return nil
	})
	return files, err
}

// shouldIgnoreFile проверяет, должен ли файл быть проигнорирован
func (s *Scanner) shouldIgnoreFile(file string) bool {
	for _, pattern := range s.ignorePatterns {
		if strings.Contains(file, pattern) {
			return true
		}
	}
	return false
}

// AnalyzeProjectStructure анализирует структуру проекта
func (s *Scanner) AnalyzeProjectStructure(root string) (string, error) {
	var structure strings.Builder
	structure.WriteString("Структура проекта:\n")

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(root, path)
		if relPath == "." {
			return nil
		}

		depth := strings.Count(relPath, string(os.PathSeparator))
		indent := strings.Repeat("  ", depth)

		if info.IsDir() {
			structure.WriteString(fmt.Sprintf("%s📁 %s/\n", indent, filepath.Base(path)))
		} else {
			structure.WriteString(fmt.Sprintf("%s📄 %s\n", indent, filepath.Base(path)))
		}

		return nil
	})

	return structure.String(), err
}

// GetFileInfo получает информацию о файле
func (s *Scanner) GetFileInfo(filepath string) (*FileInfo, error) {
	info, err := os.Stat(filepath)
	if err != nil {
		return nil, err
	}

	return &FileInfo{
		Name:    info.Name(),
		Size:    info.Size(),
		Mode:    info.Mode(),
		ModTime: info.ModTime(),
		IsDir:   info.IsDir(),
	}, nil
}

// FileInfo информация о файле
type FileInfo struct {
	Name    string      `json:"name"`
	Size    int64       `json:"size"`
	Mode    os.FileMode `json:"mode"`
	ModTime time.Time   `json:"mod_time"`
	IsDir   bool        `json:"is_dir"`
}
