package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	destDir := "unpacked_nodes"
	if err := os.MkdirAll(destDir, 0755); err != nil {
		fmt.Printf("Error creating directory %s: %v\n", destDir, err)
		os.Exit(1)
	}

	origDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		os.Exit(1)
	}

	if err := os.Chdir("nodes"); err != nil {
		fmt.Println("Error: 'nodes' folder not found")
		os.Exit(1)
	}

	zipFiles, err := filepath.Glob("*.zip")
	if err != nil {
		fmt.Println("Error searching for zip files:", err)
		os.Exit(1)
	}

	for _, zipFile := range zipFiles {
		folderName := strings.TrimSuffix(zipFile, filepath.Ext(zipFile))

		fmt.Printf("--- Processing: %s ---\n", zipFile)

		cmdUnzip := exec.Command("unzip", "-q", "-o", zipFile)
		_ = cmdUnzip.Run()

		sccPath := filepath.Join(folderName, "scc")
		if fi, err := os.Stat(sccPath); err == nil && fi.IsDir() {
			txzFiles, _ := filepath.Glob(filepath.Join(sccPath, "*.txz"))
			for _, txzFile := range txzFiles {
				fmt.Printf("Extracting TXZ in: %s\n", sccPath)
				cmdTar := exec.Command("tar", "-xJf", txzFile, "-C", sccPath)
				_ = cmdTar.Run()
			}
		}

		fmt.Printf("Moving %s to %s...\n", folderName, destDir)
		oldPath := folderName
		newPath := filepath.Join("..", destDir, folderName)

		if err := os.Rename(oldPath, newPath); err != nil {
			_ = exec.Command("mv", oldPath, filepath.Join("..", destDir, "/")).Run()
		}

		fmt.Printf("OK: %s processed.\n", folderName)
		fmt.Println("----------------------------")
	}

	_ = os.Chdir(origDir)
	fmt.Printf("Done! All files are in: ./%s\n", destDir)
}
