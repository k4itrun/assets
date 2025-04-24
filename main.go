package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	maxDepth   = 4
	ignoreDirs = ".git,node_modules"
)

func main() {
	readmePath := "README.md"
	rootPath := "."

	treeContent := generateTree(rootPath, "", true, 0)
	stats := calculateStats(rootPath)

	content := fmt.Sprintf(`# 📁 All my ~/assents

Last update: %s

## 📊 Statistics
- Directories: %d
- Archives: %d
- Total size: %s

## 🌳 Directory Tree
`+"```"+`
%s
`+"```"+`

## 🛠 Contributing

We greatly appreciate any contributions to this project! Whether you want to open new issues, submit pull requests, or share suggestions for improvements, your input is invaluable. We encourage you to refer to our [Contributing Guidelines](CONTRIBUTING.md) to facilitate a seamless collaboration process.

You can also support the development of this software through a donation, helping me bring new optimal and improved projects to life.

[![ko-fi](https://ko-fi.com/img/githubbutton_sm.svg)](https://ko-fi.com/A0A11481X5)

Thank you for your interest and support! ✌️

> [!WARNING]
> Don't clone these junk files, they are really too heavy just to save anything.

## 🏷 License

This project uses the MIT license. You can find the full license details in the [LICENSE](license.md) file.

## 📱 Contact

For any inquiries or support, you can reach out via [billoneta@proto.me](mailto:billoneta@proto.me).

<sub>Generated %s</sub>
`,
		time.Now().Format("02/01/2006 15:04:05"),
		stats.dirs,
		stats.files,
		formatSize(stats.totalSize),
		treeContent,
		time.Now().Format(time.RFC1123),
	)

	if err := os.WriteFile(readmePath, []byte(content), 0644); err != nil {
		fmt.Printf("Error escribiendo README: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ README.md actualizado correctamente")
}

type dirStats struct {
	dirs      int
	files     int
	totalSize int64
}

func calculateStats(root string) dirStats {
	var stats dirStats
	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			stats.dirs++
		} else {
			stats.files++
			if info, err := d.Info(); err == nil {
				stats.totalSize += info.Size()
			}
		}
		return nil
	})
	return stats
}

func generateTree(path, prefix string, isLast bool, depth int) string {
	var sb strings.Builder

	base := filepath.Base(path)

	if depth > 0 {
		ignored := strings.Split(ignoreDirs, ",")
		for _, dir := range ignored {
			if strings.TrimSpace(dir) == base {
				return fmt.Sprintf("%s└── [IGNORADO] %s\n", prefix, base)
			}
		}
	}

	files, err := os.ReadDir(path)
	if err != nil {
		return fmt.Sprintf("%s└── ⚠️ Error de acceso\n", prefix)
	}

	var visibleFiles []fs.DirEntry
	for _, file := range files {
		if !strings.HasPrefix(file.Name(), ".") {
			visibleFiles = append(visibleFiles, file)
		}
	}

	if len(visibleFiles) == 0 {
		return fmt.Sprintf("%s└── (vacío)\n", prefix)
	}

	for i, file := range visibleFiles {
		isLastItem := i == len(visibleFiles)-1
		pointer := "├── "
		if isLastItem {
			pointer = "└── "
		}

		icon, meta := getFileMetadata(file, path)

		if file.IsDir() {
			fmt.Fprintf(&sb, "%s%s%s %s\n", prefix, pointer, icon, file.Name())
			if depth < maxDepth {
				newPrefix := prefix
				if isLastItem {
					newPrefix += "    "
				} else {
					newPrefix += "│   "
				}
				sb.WriteString(generateTree(filepath.Join(path, file.Name()), newPrefix, isLastItem, depth+1))
			} else if hasContent(filepath.Join(path, file.Name())) {
				fmt.Fprintf(&sb, "%s    └── ... (contiene más archivos)\n", prefix)
			}
		} else {
			fmt.Fprintf(&sb, "%s%s%s %s %s\n", prefix, pointer, icon, file.Name(), meta)
		}
	}

	return sb.String()
}

func hasContent(dirPath string) bool {
	files, err := os.ReadDir(dirPath)
	return err == nil && len(files) > 0
}

func getFileMetadata(file fs.DirEntry, path string) (string, string) {
	name := file.Name()
	ext := strings.ToLower(filepath.Ext(name))
	info, _ := file.Info()
	size := info.Size()

	switch {
	case file.IsDir():
		return "📂", ""

	case strings.HasPrefix(file.Type().String(), "image"):
		return "🖼️", fmt.Sprintf("(%s)", formatSize(size))
	case strings.HasPrefix(file.Type().String(), "video"):
		return "🎬", fmt.Sprintf("(%s)", formatSize(size))
	case strings.HasPrefix(file.Type().String(), "audio"):
		return "🎵", fmt.Sprintf("(%s)", formatSize(size))

	case ext == ".zip", ext == ".rar", ext == ".7z":
		return "🗜️", fmt.Sprintf("(%s)", formatSize(size))

	case ext == ".pdf":
		return "📕", fmt.Sprintf("(%s)", formatSize(size))
	case ext == ".doc", ext == ".docx":
		return "📝", fmt.Sprintf("(%s)", formatSize(size))

	case ext == ".go":
		return "🐹", fmt.Sprintf("(%s)", formatSize(size))
	case ext == ".js", ext == ".ts":
		return "📜", fmt.Sprintf("(%s)", formatSize(size))
	case ext == ".py":
		return "🐍", fmt.Sprintf("(%s)", formatSize(size))

	default:
		return "📄", fmt.Sprintf("(%s)", formatSize(size))
	}
}

func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
