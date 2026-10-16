package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"LimbusLangManager/internal/config"
	"LimbusLangManager/internal/steam"
	"LimbusLangManager/internal/updater"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx context.Context
	cfg *config.AppConfig
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	cfg, _ := config.Load()
	a.cfg = cfg

	if a.cfg.GamePath == "" {
		if path, err := steam.FindLimbusPath(); err == nil {
			a.cfg.GamePath = path
			a.cfg.Save()
		}
	}
}

func (a *App) GetConfig() *config.AppConfig {
	return a.cfg
}

func (a *App) CheckRemoteVersion(repo string) (string, error) {
	release, err := updater.GetLatestRelease(repo)
	if err != nil {
		return "", err
	}
	return release.TagName, nil
}

func (a *App) SelectGamePathManual() (string, error) {
	selection, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Sélectionnez le dossier de Limbus Company",
	})
	if err != nil || selection == "" {
		return "", errors.New("sélection annulée")
	}

	if _, err := os.Stat(filepath.Join(selection, "LimbusCompany.exe")); err != nil {
		return "", errors.New("LimbusCompany.exe introuvable dans ce dossier")
	}

	a.cfg.GamePath = selection
	a.cfg.Save()
	return selection, nil
}

func (a *App) InstallOrUpdatePack(packID, repo string) (string, error) {
	if a.cfg.GamePath == "" {
		return "", errors.New("le chemin du jeu n'est pas configuré")
	}

	release, err := updater.GetLatestRelease(repo)
	if err != nil {
		return "", err
	}

	if current, exists := a.cfg.InstalledPacks[packID]; exists && current.Version == release.TagName {
		return "Déjà à jour (" + release.TagName + ")", nil
	}

	zipURL := release.DownloadURL
	if zipURL == "" {
		return "", errors.New("aucun fichier .zip trouvé dans cette release GitHub")
	}

	destDir := filepath.Join(a.cfg.GamePath, "LimbusCompany_Data", "Lang")
	files, err := updater.DownloadAndExtract(zipURL, destDir, func(percent int) {
		runtime.EventsEmit(a.ctx, "download:progress", map[string]interface{}{
			"packId":  packID,
			"percent": percent,
			"stage":   "downloading",
		})
	})
	if err != nil {
		return "", err
	}

	runtime.EventsEmit(a.ctx, "download:progress", map[string]interface{}{
		"packId":  packID,
		"percent": 100,
		"stage":   "done",
	})

	a.cfg.InstalledPacks[packID] = config.PackData{
		Repo:           repo,
		Version:        release.TagName,
		ExtractedFiles: files,
	}
	a.cfg.Save()

	return "Installé avec succès : " + release.TagName, nil
}

func (a *App) RemovePack(packID string) error {
	pack, exists := a.cfg.InstalledPacks[packID]
	if !exists {
		return errors.New("ce pack n'est pas installé")
	}

	for _, file := range pack.ExtractedFiles {
		os.Remove(file)
	}

	delete(a.cfg.InstalledPacks, packID)
	a.cfg.Save()
	return a.RemoveEmptyFolders()
}

func (a *App) RemoveEmptyFolders() error {
	if a.cfg == nil || a.cfg.GamePath == "" {
		return nil
	}
	langDir := filepath.Join(a.cfg.GamePath, "LimbusCompany_Data", "Lang")
	if _, err := os.Stat(langDir); os.IsNotExist(err) {
		return nil
	}

	return a.RemoveEmptySubdirs(langDir)
}

func (a *App) RemoveEmptySubdirs(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			subPath := filepath.Join(dir, entry.Name())
			if err := a.RemoveEmptySubdirs(subPath); err != nil {
				return err
			}
			subEntries, err := os.ReadDir(subPath)
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}
				return err
			}
			if len(subEntries) == 0 {
				if err := os.Remove(subPath); err != nil && !os.IsNotExist(err) {
					return err
				}
			}
		}
	}
	return nil
}

func (a *App) LaunchGame() {
	runtime.BrowserOpenURL(a.ctx, "steam://rungameid/1973530")
}