package internal

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/t4Linux/t4gfm/src/internal/ui/processbar"
)

const (
	compressFormatZip            = "zip"
	compressFormatTarGz          = "tar.gz"
	compressFormatTarXz          = "tar.xz"
	compressFormatTarZs          = "tar.zst"
	compressFormatTarGzEncrypted = "tar.gz.gpg"

	compressLevelFast     = "fast"
	compressLevelBalanced = "balanced"
	compressLevelBest     = "best"
)

var compressExcludePatterns = []string{
	".git",
	"node_modules",
	"*.tmp",
	"*.cache",
	".DS_Store",
}

func zipSources(sources []string, target string, processBar *processbar.Model) error {
	var err error

	totalFiles := 0
	for _, src := range sources {
		if _, err = os.Stat(src); os.IsNotExist(err) {
			return fmt.Errorf("source path does not exist: %s", src)
		}
		count, e := countFiles(src)
		if e != nil {
			slog.Error("Error while zip file count files ", "error", e)
		}
		totalFiles += count
	}
	p, err := processBar.SendAddProcessMsg(filepath.Base(target), processbar.OpCompress, totalFiles, true)
	if err != nil {
		return fmt.Errorf("cannot spawn process : %w", err)
	}
	_, err = os.Stat(target)
	if err == nil {
		p.ErrorMsg = "File already exists"
		p.State = processbar.Cancelled
		p.DoneTime = time.Now()
		pSendErr := processBar.SendUpdateProcessMsg(p, true)
		if pSendErr != nil {
			slog.Error("Error sending process update", "error", pSendErr)
		}
		return errors.New("file already exists")
	}

	f, err := os.Create(target)
	if err != nil {
		return err
	}
	defer f.Close()
	writer := zip.NewWriter(f)
	defer writer.Close()

	zipSourcesCore(sources, processBar, &p, writer)

	if p.State != processbar.Failed {
		// TODO: User p.SetSuccessful(), p.SetFailed()
		p.State = processbar.Successful
		p.Done = totalFiles
	}
	p.DoneTime = time.Now()
	pSendErr := processBar.SendUpdateProcessMsg(p, true)
	if pSendErr != nil {
		slog.Error("Error sending process update", "error", pSendErr)
	}
	return nil
}

func zipSourcesCore(sources []string, processBar *processbar.Model,
	p *processbar.Process, writer *zip.Writer) {
	for _, src := range sources {
		srcParentDir := filepath.Dir(src)
		err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
			p.CurrentFile = filepath.Base(path)
			if err != nil {
				return err
			}
			relPath, err := filepath.Rel(srcParentDir, path)
			if err != nil {
				return err
			}

			err = writeZipFile(path, relPath, info, writer)
			if err != nil {
				return err
			}

			p.Done++
			processBar.TrySendingUpdateProcessMsg(*p)
			return nil
		})
		if err != nil {
			slog.Error("Error while zip file", "error", err)
			p.State = processbar.Failed
			break
		}
	}
}

func writeZipFile(path string, relPath string, info os.FileInfo, writer *zip.Writer) error {
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Method = zip.Deflate
	header.Name = relPath
	if info.IsDir() {
		header.Name += "/"
	}
	headerWriter, err := writer.CreateHeader(header)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return nil
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(headerWriter, file)
	if err != nil {
		return err
	}
	return nil
}

func getZipArchiveName(base string) (string, error) {
	zipName := strings.TrimSuffix(base, filepath.Ext(base)) + ".zip"
	zipName, err := renameIfDuplicate(zipName)
	return zipName, err
}

func getTarArchiveName(base string, format string) (string, error) {
	var ext string
	switch format {
	case compressFormatTarGz:
		ext = ".tar.gz"
	case compressFormatTarXz:
		ext = ".tar.xz"
	case compressFormatTarZs:
		ext = ".tar.zst"
	case compressFormatTarGzEncrypted:
		ext = ".tar.gz.gpg"
	default:
		return "", fmt.Errorf("unsupported tar format: %s", format)
	}
	tarName := strings.TrimSuffix(base, filepath.Ext(base)) + ext
	tarName, err := renameIfDuplicate(tarName)
	return tarName, err
}

func encryptFileWithGPG(inputPath string, outputPath string, passphrase string) error {
	if _, err := exec.LookPath("gpg"); err != nil {
		return fmt.Errorf("gpg command not found in PATH")
	}
	args := []string{
		"--batch",
		"--yes",
		"--pinentry-mode", "loopback",
		"--passphrase", passphrase,
		"--symmetric",
		"--cipher-algo", "AES256",
		"--output", outputPath,
		inputPath,
	}
	cmd := exec.Command("gpg", args...)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func decryptFileWithGPG(inputPath string, outputPath string, passphrase string) error {
	if _, err := exec.LookPath("gpg"); err != nil {
		return fmt.Errorf("gpg command not found in PATH")
	}
	args := []string{
		"--batch",
		"--yes",
		"--pinentry-mode", "loopback",
		"--passphrase", passphrase,
		"--decrypt",
		"--output", outputPath,
		inputPath,
	}
	cmd := exec.Command("gpg", args...)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func tarSources(
	sources []string,
	baseDir string,
	target string,
	format string,
	level string,
	verbose bool,
	excludeCommon bool,
	processBar *processbar.Model,
) error {
	if _, err := exec.LookPath("tar"); err != nil {
		return fmt.Errorf("tar command not found in PATH")
	}

	p, err := processBar.SendAddProcessMsg(filepath.Base(target), processbar.OpCompress, 1, true)
	if err != nil {
		return fmt.Errorf("cannot spawn process : %w", err)
	}

	_, err = os.Stat(target)
	if err == nil {
		p.ErrorMsg = "File already exists"
		p.State = processbar.Cancelled
		p.DoneTime = time.Now()
		_ = processBar.SendUpdateProcessMsg(p, true)
		return errors.New("file already exists")
	}

	relSources := make([]string, 0, len(sources))
	for _, src := range sources {
		rel, relErr := filepath.Rel(baseDir, src)
		if relErr != nil {
			return relErr
		}
		relSources = append(relSources, rel)
	}

	tarFlags, err := tarFlagsByFormat(format)
	if err != nil {
		return err
	}
	args := append([]string{}, tarFlags...)
	args = append(args, target, "-C", baseDir)
	if verbose {
		args = append(args, "-v")
	}
	if excludeCommon {
		for _, pattern := range compressExcludePatterns {
			args = append(args, "--exclude", pattern)
		}
	}
	args = append(args, relSources...)
	cmd := exec.Command("tar", args...)
	cmd.Env = append(os.Environ(), compressionLevelEnv(level)...)
	if runErr := cmd.Run(); runErr != nil {
		p.State = processbar.Failed
		p.ErrorMsg = runErr.Error()
		p.DoneTime = time.Now()
		_ = processBar.SendUpdateProcessMsg(p, true)
		return runErr
	}

	p.State = processbar.Successful
	p.Done = p.Total
	p.DoneTime = time.Now()
	_ = processBar.SendUpdateProcessMsg(p, true)
	return nil
}

func tarFlagsByFormat(format string) ([]string, error) {
	switch format {
	case compressFormatTarGz:
		return []string{"-czf"}, nil
	case compressFormatTarXz:
		return []string{"-cJf"}, nil
	case compressFormatTarZs:
		return []string{"--zstd", "-cf"}, nil
	default:
		return nil, fmt.Errorf("unsupported tar format: %s", format)
	}
}

func compressionLevelEnv(level string) []string {
	value := "6"
	switch level {
	case compressLevelFast:
		value = "1"
	case compressLevelBest:
		value = "9"
	}
	return []string{"GZIP=-" + value, "XZ_OPT=-" + value, "ZSTD_CLEVEL=" + value}
}
