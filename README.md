# LimbusLangManager

## Check first

All packs are from [LocalizeLimbusCompany](https://github.com/LocalizeLimbusCompany/LocalizeLimbusCompany).

Go support them!

## Introduction

LimbusLangManager is an open-source tool designed to simplify the management of community-created language patches for the game Limbus Company. It provides a user-friendly interface to install, manage, and switch between different language modifications without the need for manual file manipulation.

## Features

- **Automatic Game Detection**: Scans common installation paths to find Limbus Company.
- **One-Click Installation**: Downloads and installs language packs directly from GitHub repositories.
- **Version Management**: Tracks installed versions and notifies users of available updates.
- **Clean Uninstallation**: Removes all traces of a language pack.
- **Simple Interface**: A modern, intuitive UI for managing translations.

## Installation & Setup

1.  **Download the Application**:
    -   Visit the [Releases](https://github.com/GabyFumo/LimbusLangManager/releases) page.
    -   Download `LimbusLangManager.exe` for Windows.

2.  **First Launch & Configuration**:
    -   When you run the application for the first time, it will automatically search for your Limbus Company installation directory.
    -   If the game is not found, you will be prompted to select the installation folder manually.
    -   **Path**: `D:\SteamLibrary\steamapps\common\Limbus Company` (or your custom path).

## Usage Guide

### Install from Releases
1. Download from the [Releases](https://github.com/GabyFumo/LimbusLangManager/releases) page and start `LimbusLangManager.exe`

   > **Recommended**: put it in a folder, a config file will appear and will store your settings and path.

2. Click install within the list of packs to install the desired language pack(s)
3. Wait for the download and installation to complete
4. Press `LAUNCH GAME` and enjoy the game in your preferred language


### Building the Application
To create a standalone executable:
1.  Install requirements:
    - Go (1.21+)
	- NPM (Node 15+)
2.  Install Wails CLI:
	```bash
	go install github.com/wailsapp/wails/v2/cmd/wails@latest
	```
3.  Build the app
	```bash
	wails build
	```
	This will create a `LimbusLangManager.exe` in the `build` folder.
4. Use it like the Releases version, enjoy!

## Support me!

* Buy me a coffee
https://ko-fi.com/gabyfumo
