package updater

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"regexp"
)

type ReleaseInfo struct {
	TagName     string
	DownloadURL string
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
		if percent > pr.lastPercent {
			pr.lastPercent = percent
			pr.onProgress(percent)
		}
	}
	return n, err
}

func GetLatestRelease(repo string) (*ReleaseInfo, error) {
	latestURL := fmt.Sprintf("https://github.com/%s/releases/latest", repo)
	req, err := http.NewRequest("GET", latestURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	finalURL := resp.Request.URL.String()
	parts := strings.Split(finalURL, "/tag/")
	if len(parts) < 2 {
		return nil, errors.New("aucune release trouvée sur ce dépôt GitHub")
	}
	tagName := parts[1]

	assetsURL := fmt.Sprintf("https://github.com/%s/releases/expanded_assets/%s", repo, tagName)
	reqAssets, err := http.NewRequest("GET", assetsURL, nil)
	if err != nil {
		return nil, err
	}
	reqAssets.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")

	respAssets, err := http.DefaultClient.Do(reqAssets)
	if err != nil {
		return nil, err
	}
	defer respAssets.Body.Close()

	bodyBytes, err := io.ReadAll(respAssets.Body)
	if err != nil {
		return nil, err
	}

	re := regexp.MustCompile(`href="(/[^"]+/releases/download/[^"]+\.zip)"`)
	match := re.FindStringSubmatch(string(bodyBytes))
	if len(match) < 2 {
		return nil, errors.New("aucun fichier .zip trouvé dans la dernière release")
	}

	downloadURL := "https://github.com" + match[1]

	return &ReleaseInfo{
		TagName:     tagName,
		DownloadURL: downloadURL,
	}, nil
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

	archiveName := getArchiveName(downloadURL)
	return unzip(tmpPath, destDir, archiveName)
}

func getArchiveName(downloadURL string) string {
	parsed, err := url.Parse(downloadURL)
	var filename string
	if err == nil {
		filename = filepath.Base(parsed.Path)
	} else {
		filename = filepath.Base(downloadURL)
	}
	if idx := strings.Index(filename, "?"); idx != -1 {
		filename = filename[:idx]
	}
	name := strings.TrimSuffix(filename, filepath.Ext(filename))
	if name == "" || name == "." || name == "/" {
		return "langpack"
	}
	return name
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

func unzip(src, dest, archiveName string) ([]string, error) {
	var extractedFiles []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	if archiveName == "" {
		base := filepath.Base(src)
		archiveName = strings.TrimSuffix(base, filepath.Ext(base))
	}

	langPrefix := findLangPrefix(r.File)
	lowerLangPrefix := strings.ToLower(langPrefix)

	// Scan items to count how many top-level directories will be extracted
	topLevelDirs := make(map[string]bool)
	for _, f := range r.File {
		cleanName := strings.ReplaceAll(f.Name, "\\", "/")
		cleanName = strings.TrimPrefix(cleanName, "/")
		lowerName := strings.ToLower(cleanName)

		if strings.HasPrefix(cleanName, "__MACOSX/") {
			continue
		}

		var relPath string
		if langPrefix != "" {
			if !strings.HasPrefix(lowerName, lowerLangPrefix) {
				continue
			}
			relPath = cleanName[len(langPrefix):]
		} else {
			relPath = cleanName
		}

		relPath = strings.TrimPrefix(relPath, "/")
		if relPath == "" {
			continue
		}

		parts := strings.Split(relPath, "/")
		topSegment := strings.TrimSpace(parts[0])
		if topSegment == "" || topSegment == "." {
			continue
		}

		isDir := len(parts) > 1 || f.FileInfo().IsDir() || strings.HasSuffix(cleanName, "/")
		if isDir {
			topLevelDirs[strings.ToLower(topSegment)] = true
		}
	}

	finalDest := dest
	if len(topLevelDirs) > 1 && archiveName != "" {
		finalDest = filepath.Join(dest, archiveName)
	}

	if err := os.MkdirAll(finalDest, os.ModePerm); err != nil {
		return nil, err
	}

	cleanDest := filepath.Clean(finalDest)
	destPrefix := cleanDest + string(os.PathSeparator)

	for _, f := range r.File {
		cleanName := strings.ReplaceAll(f.Name, "\\", "/")
		cleanName = strings.TrimPrefix(cleanName, "/")
		lowerName := strings.ToLower(cleanName)

		if strings.HasPrefix(cleanName, "__MACOSX/") {
			continue
		}

		var relPath string
		if langPrefix != "" {
			if !strings.HasPrefix(lowerName, lowerLangPrefix) {
				continue
			}
			relPath = cleanName[len(langPrefix):]
		} else {
			relPath = cleanName
		}

		relPath = strings.TrimPrefix(relPath, "/")
		if relPath == "" {
			continue
		}

		targetPath := filepath.Join(finalDest, filepath.FromSlash(relPath))

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