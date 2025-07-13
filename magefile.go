//go:build mage

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const (
	binaryName = "promptly"
	version    = "1.0.0"
)

// Build builds the binary
func Build() error {
	fmt.Println("Building", binaryName)
	return sh.Run("go", "build", "-ldflags=-s -w", "-o", binaryName, ".")
}

// Clean removes build artifacts
func Clean() error {
	fmt.Println("Cleaning...")
	mg.Deps(cleanBinary)
	return sh.Run("go", "clean")
}

func cleanBinary() error {
	return os.RemoveAll(binaryName)
}

// Test runs tests
func Test() error {
	fmt.Println("Running tests...")
	return sh.Run("go", "test", "./...")
}

// Install builds and installs the binary to /usr/local/bin
func Install() error {
	mg.Deps(Build)
	fmt.Println("Installing to /usr/local/bin")
	return sh.Copy(filepath.Join("/usr/local/bin", binaryName), binaryName)
}

// Release builds binaries for multiple platforms
func Release() error {
	fmt.Println("Building release binaries...")
	
	platforms := map[string]map[string]string{
		"linux":   {"amd64": ""},
		"darwin":  {"amd64": "", "arm64": ""},
		"windows": {"amd64": ".exe"},
	}

	for goos, arches := range platforms {
		for goarch, ext := range arches {
			env := map[string]string{
				"GOOS":   goos,
				"GOARCH": goarch,
			}
			
			output := fmt.Sprintf("%s-%s-%s%s", binaryName, goos, goarch, ext)
			fmt.Printf("Building %s...\n", output)
			
			if err := sh.RunWith(env, "go", "build", "-ldflags=-s -w", "-o", output, "."); err != nil {
				return err
			}
		}
	}
	
	return nil
}

// Deps downloads and tidies dependencies
func Deps() error {
	fmt.Println("Downloading dependencies...")
	if err := sh.Run("go", "mod", "tidy"); err != nil {
		return err
	}
	return sh.Run("go", "mod", "download")
}