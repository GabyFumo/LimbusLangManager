package steam

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/sys/windows/registry"
)

const LimbusAppID = "1973530"

func FindLimbusPath() (string, error) {
	steamPath, err := getSteamInstallPath()
	if err != nil {
		return "", err
	}

	vdfPath := filepath.Join(steamPath, "steamapps", "libraryfolders.vdf")
	content, err := os.ReadFile(vdfPath)
	if err != nil {
		return "", err
	}

	libraries := parseLibraryFolders(string(content))
	for _, lib := range libraries {
		gamePath := filepath.Join(lib, "steamapps", "common", "Limbus Company")
		exePath := filepath.Join(gamePath, "LimbusCompany.exe")
		if _, err := os.Stat(exePath); err == nil {
			return gamePath, nil
		}
	}

	return "", errors.New("Limbus Company introuvable dans les bibliothèques Steam")
}

func getSteamInstallPath() (string, error) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\WOW6432Node\Valve\Steam`, registry.QUERY_VALUE)
	if err == nil {
		defer k.Close()
		path, _, err := k.GetStringValue("InstallPath")
		if err == nil && path != "" {
			return path, nil
		}
	}

	k, err = registry.OpenKey(registry.CURRENT_USER, `Software\Valve\Steam`, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer k.Close()

	path, _, err := k.GetStringValue("SteamPath")
	if err != nil {
		return "", err
	}
	return filepath.FromSlash(path), nil
}

func parseLibraryFolders(vdfContent string) []string {
	var paths []string
	re := regexp.MustCompile(`"path"\s+"([^"]+)"`)
	matches := re.FindAllStringSubmatch(vdfContent, -1)

	for _, match := range matches {
		if len(match) > 1 {
			cleanPath := strings.ReplaceAll(match[1], `\\`, `\`)
			paths = append(paths, cleanPath)
		}
	}
	return paths
}