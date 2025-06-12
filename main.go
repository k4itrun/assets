package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/k4itrun/assets/config"
)

type DirStats struct {
	dirs      int
	files     int
	totalSize int64
}

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		logFatal("getting current working directory", err)
	}

	readmePath := filepath.Join(cwd, "README.md")
	content, err := os.ReadFile(readmePath)
	if err != nil {
		logFatal("reading README.md", err)
	}

	tree := buildDirectoryTree(cwd, "", 0)
	stats := calculateStats(cwd)

	updated := replaceSection(string(content), "STATS", formatStats(stats))
	updated = replaceSection(updated, "TREE", formatTree(tree))

	if err := os.WriteFile(readmePath, []byte(updated), 0644); err != nil {
		logFatal("writing updated README.md", err)
	}

	fmt.Println("Done!!")
}

func buildDirectoryTree(path, prefix string, depth int) string {
	var sb strings.Builder
	base := filepath.Base(path)

	if depth > 0 && isIgnored(base) {
		return fmt.Sprintf("%s└── [IGNORED] %s\n", prefix, base)
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Sprintf("%s└── ⚠️ Access error\n", prefix)
	}

	visible := filterVisible(entries)
	if len(visible) == 0 {
		return fmt.Sprintf("%s└── (empty)\n", prefix)
	}

	for i, entry := range visible {
		isLast := i == len(visible)-1
		pointer := "├── "
		if isLast {
			pointer = "└── "
		}

		icon, meta := getFileMetadata(entry)
		fmt.Fprintf(&sb, "%s%s%s %s %s\n", prefix, pointer, icon, entry.Name(), meta)

		if entry.IsDir() {
			if depth < config.Default.MaxDepth {
				nextPrefix := prefix + ternary(isLast, "    ", "│   ")
				sb.WriteString(buildDirectoryTree(filepath.Join(path, entry.Name()), nextPrefix, depth+1))
			} else if hasContent(filepath.Join(path, entry.Name())) {
				fmt.Fprintf(&sb, "%s    └── ... (contains more files)\n", prefix)
			}
		}
	}
	return sb.String()
}

func calculateStats(root string) DirStats {
	var stats DirStats
	filepath.WalkDir(root, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
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

func replaceSection(original, tag, replacement string) string {
	startTag := fmt.Sprintf("<!--%s:start-->", tag)
	endTag := fmt.Sprintf("<!--%s:end-->", tag)

	start := strings.Index(original, startTag)
	end := strings.Index(original, endTag)

	if start == -1 || end == -1 || start > end {
		return original
	}

	return original[:start] + replacement + original[end+len(endTag):]
}

func formatStats(stats DirStats) string {
	return fmt.Sprintf("<!--STATS:start-->\n\n| Directories | Archives | Total Size |\n| ----------- | -------- | ---------- |\n| `%d`       | `%d`    | `%s` |\n\n<!--STATS:end-->", stats.dirs, stats.files, formatSize(stats.totalSize))
}

func formatTree(tree string) string {
	return fmt.Sprintf("<!--TREE:start-->\n\n```\n%s\n```\n\n<!--TREE:end-->", tree)
}

func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	exp := 0
	div := unit
	for n := size / unit; n >= unit; n /= unit {
		exp++
		div *= unit
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

func filterVisible(entries []fs.DirEntry) []fs.DirEntry {
	var visible []fs.DirEntry
	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), ".") {
			visible = append(visible, e)
		}
	}
	return visible
}

func getFileMetadata(file fs.DirEntry) (string, string) {
	ext := strings.ToLower(filepath.Ext(file.Name()))
	info, _ := file.Info()
	size := info.Size()

	switch {
	case file.IsDir():
		return "📂", ""
	case ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif":
		return "🖼️", fmt.Sprintf("(%s)", formatSize(size))
	case ext == ".mp4" || ext == ".ogg" || ext == ".avi" || ext == ".mkv":
		return "🎬", fmt.Sprintf("(%s)", formatSize(size))
	case ext == ".mp3" || ext == ".wav" || ext == ".flac":
		return "🎵", fmt.Sprintf("(%s)", formatSize(size))
	case ext == ".zip" || ext == ".rar" || ext == ".7z" || ext == ".tar" || ext == ".tgz" || ext == ".gz":
		return "💼", fmt.Sprintf("(%s)", formatSize(size))
	case ext == ".pdf":
		return "📕", fmt.Sprintf("(%s)", formatSize(size))
	case ext == ".doc" || ext == ".docx":
		return "📝", fmt.Sprintf("(%s)", formatSize(size))
	case ext == ".go":
		return "🐹", fmt.Sprintf("(%s)", formatSize(size))
	case ext == ".js" || ext == ".ts":
		return "📜", fmt.Sprintf("(%s)", formatSize(size))
	case ext == ".py":
		return "🐍", fmt.Sprintf("(%s)", formatSize(size))
	default:
		return "📄", fmt.Sprintf("(%s)", formatSize(size))
	}
}

func isIgnored(name string) bool {
	for _, ignored := range config.Default.IgnoreDirs {
		if strings.TrimSpace(ignored) == name {
			return true
		}
	}
	return false
}

func hasContent(dirPath string) bool {
	files, err := os.ReadDir(dirPath)
	return err == nil && len(files) > 0
}

func logFatal(context string, err error) {
	fmt.Printf("Error %s: %v\n", context, err)
	os.Exit(1)
}

func ternary(condition bool, a, b string) string {
	if condition {
		return a
	}
	return b
}
