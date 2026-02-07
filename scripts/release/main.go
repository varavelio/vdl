package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Binary struct {
	OS   string
	Arch string
}

var (
	Version string = "unknown"
	Commit  string = "unknown"
	Date    string = time.Now().UTC().Format(time.RFC3339)

	Binaries []Binary = []Binary{
		{OS: "linux", Arch: "amd64"},
		{OS: "linux", Arch: "arm64"},
		{OS: "darwin", Arch: "amd64"},
		{OS: "darwin", Arch: "arm64"},
		{OS: "windows", Arch: "amd64"},
		{OS: "windows", Arch: "arm64"},
	}
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Setup and Move CWD to the repository root
	root, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("could not find project root: %w", err)
	}
	if err := os.Chdir(root); err != nil {
		return fmt.Errorf("could not chdir to project root: %w", err)
	}
	distDir := filepath.Join(root, "dist")
	fmt.Printf("Working directory: %s\n", root)

	// Initialize Version Variables
	if err := initVariables(); err != nil {
		return fmt.Errorf("failed to init variables: %w", err)
	}
	fmt.Printf("Releasing Version: %s (Commit: %s, Date: %s)\n", Version, Commit, Date)

	// Clean and Create dist/ directory
	if err := os.RemoveAll(distDir); err != nil {
		return fmt.Errorf("failed to clean dist directory: %w", err)
	}
	if err := os.MkdirAll(distDir, 0755); err != nil {
		return fmt.Errorf("failed to create dist directory: %w", err)
	}

	// Build and archive WASM
	if err := buildAndArchiveWasm(root, distDir); err != nil {
		return fmt.Errorf("failed to build and archive wasm: %w", err)
	}

	// Build and Archive CLI Binaries
	for _, bin := range Binaries {
		fmt.Printf("Building %s/%s...\n", bin.OS, bin.Arch)
		if err := buildAndArchive(root, distDir, bin); err != nil {
			return fmt.Errorf("failed to build %s/%s: %w", bin.OS, bin.Arch, err)
		}
	}

	// Generate Checksums
	fmt.Println("Generating checksums...")
	if err := generateChecksums(distDir); err != nil {
		return fmt.Errorf("failed to generate checksums: %w", err)
	}

	fmt.Println("Release build completed successfully in ./dist")
	return nil
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "Taskfile.yml")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("root not found")
		}
		dir = parent
	}
}

func initVariables() error {
	out, err := exec.Command("git", "describe", "--tags", "--abbrev=0").Output()
	if err == nil {
		Version = strings.TrimPrefix(strings.TrimSpace(strings.ToLower(string(out))), "v")
	} else {
		return fmt.Errorf("cannot get version from latest git tag: %w", err)
	}

	out, err = exec.Command("git", "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		return fmt.Errorf("git rev-parse failed: %w", err)
	}
	Commit = strings.TrimSpace(string(out))
	return nil
}

func buildAndArchiveWasm(root, distDir string) error {
	fmt.Println("Building WASM...")

	// Build vdl.wasm
	wasmPath := filepath.Join(distDir, "vdl.wasm")
	cmd := exec.Command("go", "build", "-o", wasmPath, "./cmd/vdlwasm/.")
	cmd.Dir = filepath.Join(root, "toolchain")
	cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to build wasm: %w", err)
	}

	// Copy wasm_exec.js
	// We need to locate where Go is installed to find wasm_exec.js
	// Usually GOROOT is set, or we can use `go env GOROOT`
	goRootOut, err := exec.Command("go", "env", "GOROOT").Output()
	if err != nil {
		return fmt.Errorf("failed to get GOROOT: %w", err)
	}
	goRoot := strings.TrimSpace(string(goRootOut))
	wasmExecSrc := filepath.Join(goRoot, "lib/wasm/wasm_exec.js")
	wasmExecDst := filepath.Join(distDir, "wasm_exec.js")

	if err := copyFile(wasmExecSrc, wasmExecDst); err != nil {
		return fmt.Errorf("failed to copy wasm_exec.js: %w", err)
	}

	// Prepare extra files for archive (README, LICENSE)
	extraFiles := []string{"README.md", "LICENSE"}

	filesToArchive := make(map[string]string) // src -> name in archive
	filesToArchive[wasmPath] = "vdl.wasm"
	filesToArchive[wasmExecDst] = "wasm_exec.js"

	for _, f := range extraFiles {
		src := filepath.Join(root, f)
		if _, err := os.Stat(src); err == nil {
			filesToArchive[src] = f
		}
	}

	// Create vdl_wasm.tar.gz
	archivePath := filepath.Join(distDir, "vdl_wasm.tar.gz")
	if err := createTarGz(archivePath, filesToArchive); err != nil {
		return fmt.Errorf("failed to create wasm archive: %w", err)
	}

	// Cleanup raw WASM files
	if err := os.Remove(wasmPath); err != nil {
		return fmt.Errorf("failed to remove vdl.wasm: %w", err)
	}
	if err := os.Remove(wasmExecDst); err != nil {
		return fmt.Errorf("failed to remove wasm_exec.js: %w", err)
	}

	return nil
}

func buildAndArchive(root, distDir string, bin Binary) error {
	binaryName := "vdl"
	archiveType := "tar.gz"
	if bin.OS == "windows" {
		binaryName = "vdl.exe"
		archiveType = "zip"
	}

	// Temporary output path for the raw binary
	rawBinPath := filepath.Join(distDir, binaryName)

	// LDFLAGS
	ldflags := strings.Join([]string{
		"-s -w",
		"-X github.com/varavelio/vdl/toolchain/internal/version.Version=" + Version,
		"-X github.com/varavelio/vdl/toolchain/internal/version.Commit=" + Commit,
		"-X github.com/varavelio/vdl/toolchain/internal/version.Date=" + Date,
	}, " ")

	cmd := exec.Command("go", "build", "-ldflags", ldflags, "-o", rawBinPath, "./cmd/vdl")
	cmd.Dir = filepath.Join(root, "toolchain")
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0", "GOOS="+bin.OS, "GOARCH="+bin.Arch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go build failed: %w", err)
	}

	// Prepare files for archive
	filesToArchive := make(map[string]string)
	filesToArchive[rawBinPath] = binaryName

	extraFiles := []string{"README.md", "LICENSE"}
	for _, f := range extraFiles {
		src := filepath.Join(root, f)
		if _, err := os.Stat(src); err == nil {
			filesToArchive[src] = f
		}
	}

	// Archive
	archiveName := fmt.Sprintf("vdl_%s_%s.%s", bin.OS, bin.Arch, archiveType)
	archivePath := filepath.Join(distDir, archiveName)

	if archiveType == "zip" {
		if err := createZip(archivePath, filesToArchive); err != nil {
			return err
		}
	} else {
		if err := createTarGz(archivePath, filesToArchive); err != nil {
			return err
		}
	}

	// Cleanup raw binary
	if err := os.Remove(rawBinPath); err != nil {
		return fmt.Errorf("failed to remove raw binary: %w", err)
	}

	return nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func createZip(target string, files map[string]string) error {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	for src, name := range files {
		if err := addFileToZip(archive, src, name); err != nil {
			return err
		}
	}
	return nil
}

func addFileToZip(archive *zip.Writer, src, name string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = name
	header.Method = zip.Deflate

	writer, err := archive.CreateHeader(header)
	if err != nil {
		return err
	}

	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(writer, file)
	return err
}

func createTarGz(target string, files map[string]string) error {
	file, err := os.Create(target)
	if err != nil {
		return err
	}
	defer file.Close()

	gw := gzip.NewWriter(file)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	for src, name := range files {
		if err := addFileToTar(tw, src, name); err != nil {
			return err
		}
	}
	return nil
}

func addFileToTar(tw *tar.Writer, src, name string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(info, info.Name())
	if err != nil {
		return err
	}
	header.Name = name

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	_, err = io.Copy(tw, srcFile)
	return err
}

func generateChecksums(distDir string) error {
	files, err := os.ReadDir(distDir)
	if err != nil {
		return err
	}

	checksumFile, err := os.Create(filepath.Join(distDir, "checksums.txt"))
	if err != nil {
		return err
	}
	defer checksumFile.Close()

	for _, file := range files {
		if file.IsDir() || file.Name() == "checksums.txt" {
			continue
		}

		path := filepath.Join(distDir, file.Name())
		hash, err := calculateSHA256(path)
		if err != nil {
			return fmt.Errorf("failed to hash %s: %w", file.Name(), err)
		}

		if _, err := fmt.Fprintf(checksumFile, "%s  %s\n", hash, file.Name()); err != nil {
			return err
		}
	}
	return nil
}

func calculateSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
