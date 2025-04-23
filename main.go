package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	readme := "README.md"
	content := fmt.Sprintf("# 📂 All my ~/assents [2022]\n\nUpdated at: %s\n", time.Now().Format(time.RFC1123))

	err := os.WriteFile(readme, []byte(content), 0644)
	if err != nil {
		fmt.Printf("Error writing to README: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("README.md updated")
}
