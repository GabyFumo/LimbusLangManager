package updater

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type ReleaseInfo struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

type progressReader struct {
	reader      io.Reader
	total       int64
	downloaded  int64
	lastPercent int
	onProgress  func(percent int)
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	pr.downloaded += int64(n)

	if pr.total > 0 && pr.onProgress != nil {
		percent := int(float64(pr.downloaded) / float64(pr.total) * 100)
		// On n'envoie l'événement que si le palier de 1% a avancé
		if percent > pr.lastPercent {
			pr.lastPercent = percent
			pr.onProgress(percent)
		}
	}
	return n, err
}

func GetLatestRelease(repo string) (*ReleaseInfo, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "LimbusLangManager-App")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erreur API GitHub: statut %d", resp.StatusCode)
	}

	var info ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}
	return &info, nil
}

func DownloadAndExtract(downloadURL, destDir string, onProgress func(int)) ([]string, error) {
	resp, err := http.Get(downloadURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erreur lors du téléchargement : statut HTTP %d", resp.StatusCode)
	}

	tmpFile, err := os.CreateTemp("", "langpack-*.zip")
	if err != nil {
		return nil, err
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	pr := &progressReader{
		reader:      resp.Body,
		total:       resp.ContentLength,
		lastPercent: -1,
		onProgress:  onProgress,
	}

	if _, err := io.Copy(tmpFile, pr); err != nil {
		tmpFile.Close()
		return nil, err
	}
	tmpFile.Close()

	return unzip(tmpPath, destDir)
}

func findLangPrefix(files []*zip.File) string {
	candidateCounts := make(map[string]int)

	for _, f := range files {
		cleanName := strings.ReplaceAll(f.Name, "\\", "/")
		cleanName = strings.TrimPrefix(cleanName, "/")
		if strings.HasPrefix(cleanName, "__MACOSX/") {
			continue
		}

		parts := strings.Split(cleanName, "/")
		for i, part := range parts {
			if strings.EqualFold(part, "lang") {
				isDir := (i < len(parts)-1) || f.FileInfo().IsDir() || strings.HasSuffix(f.Name, "/") || strings.HasSuffix(f.Name, "\\")
				if isDir {
					prefix := strings.Join(parts[:i+1], "/") + "/"
					candidateCounts[prefix]++
					break
				}
			}
		}
	}

	if len(candidateCounts) == 0 {
		return ""
	}

	bestPrefix := ""
	bestDepth := 999999
	bestCount := -1

	for prefix, count := range candidateCounts {
		depth := strings.Count(prefix, "/")
		if depth < bestDepth || (depth == bestDepth && count > bestCount) {
			bestPrefix = prefix
			bestDepth = depth
			bestCount = count
		}
	}

	return bestPrefix
}

func unzip(src, dest string) ([]string, error) {
	var extractedFiles []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	langPrefix := findLangPrefix(r.File)
	if langPrefix == "" {
		return nil, errors.New("dossier 'Lang' introuvable dans l'archive")
	}

	if err := os.MkdirAll(dest, os.ModePerm); err != nil {
		return nil, err
	}

	cleanDest := filepath.Clean(dest)
	destPrefix := cleanDest + string(os.PathSeparator)
	lowerLangPrefix := strings.ToLower(langPrefix)

	for _, f := range r.File {
		cleanName := strings.ReplaceAll(f.Name, "\\", "/")
		cleanName = strings.TrimPrefix(cleanName, "/")
		lowerName := strings.ToLower(cleanName)

		if strings.HasPrefix(cleanName, "__MACOSX/") || !strings.HasPrefix(lowerName, lowerLangPrefix) {
			continue
		}

		relPath := cleanName[len(langPrefix):]
		relPath = strings.TrimPrefix(relPath, "/")
		if relPath == "" {
			continue
		}

		targetPath := filepath.Join(dest, filepath.FromSlash(relPath))

		cleanTarget := filepath.Clean(targetPath)
		if !strings.HasPrefix(cleanTarget, destPrefix) && cleanTarget != cleanDest {
			return nil, errors.New("chemin de fichier illégal dans l'archive")
		}

		if f.FileInfo().IsDir() || strings.HasSuffix(cleanName, "/") {
			if err := os.MkdirAll(targetPath, os.ModePerm); err != nil {
				return nil, err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), os.ModePerm); err != nil {
			return nil, err
		}

		mode := f.Mode()
		if mode&0200 == 0 {
			mode |= 0644
		}

		outFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
		if err != nil {
			return nil, err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return nil, err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return nil, err
		}
		extractedFiles = append(extractedFiles, targetPath)
	}

	return extractedFiles, nil
}