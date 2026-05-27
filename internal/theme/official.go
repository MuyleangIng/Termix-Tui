package theme

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/muyleanging/termix/internal/config"
)

const officialThemesURL = "https://github.com/JanDeDobbeleer/oh-my-posh/releases/latest/download/themes.zip"

func InstallOfficialThemes(ctx context.Context, cfg config.Config) (int, error) {
	if len(cfg.ThemeDirs) == 0 {
		return 0, nil
	}
	dir := cfg.ThemeDirs[0]
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return 0, err
	}

	tmp, err := os.CreateTemp("", "termix-themes-*.zip")
	if err != nil {
		return 0, err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	count, err := downloadOfficialThemes(ctx, tmp)
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return 0, err
	}
	if count == 0 {
		return 0, fmt.Errorf("downloaded official themes archive was empty")
	}
	return extractThemeArchive(tmpPath, dir)
}

func downloadOfficialThemes(ctx context.Context, dst io.Writer) (int64, error) {
	client := http.Client{Timeout: 90 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, officialThemesURL, nil)
	if err != nil {
		return 0, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("download official themes: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return 0, fmt.Errorf("download official themes: %s", resp.Status)
	}
	return io.Copy(dst, resp.Body)
}

func extractThemeArchive(path, dir string) (int, error) {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return 0, err
	}
	defer reader.Close()

	count := 0
	for _, file := range reader.File {
		if file.FileInfo().IsDir() || !strings.HasSuffix(strings.ToLower(file.Name), ".omp.json") {
			continue
		}
		name := filepath.Base(file.Name)
		if name == "." || name == string(filepath.Separator) {
			continue
		}
		if err := extractThemeFile(file, filepath.Join(dir, name)); err != nil {
			return count, err
		}
		count++
	}
	if count == 0 {
		return 0, fmt.Errorf("official themes archive did not contain .omp.json files")
	}
	return count, nil
}

func extractThemeFile(file *zip.File, target string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(target)
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	return err
}
