export const translations = {
    en: {
        header_sub: "// LANG MANAGER",
        launch_game: "▶ LAUNCH GAME",
        game_folder_label: "GAME FOLDER (STEAM)",
        searching_path: "Searching...",
        path_not_found: "Game not detected. Please select the folder manually.",
        browse_btn: "Browse...",
        available_packs_title: "AVAILABLE TRANSLATION PACKS",
        refresh_btn: "Refresh versions",
        status_ready: "Ready.",
        status_checking: "Checking for updates on GitHub...",
        status_downloading: "Downloading and extracting {repo}...",
        status_removing: "Removing pack {id}...",
        status_removed: "Pack successfully removed.",
        status_path_updated: "Game path updated.",
        status_launching: "Launching Limbus Company via Steam...",
        repo_label: "Repo:",
        installed_label: "Installed:",
        not_installed: "Not installed",
        update_available: "(Update available!)",
        btn_install: "Install",
        btn_reinstall: "Reinstall",
        btn_update: "Update ({ver})",
        btn_remove: "Remove",
        dl_progress: "Downloading: {pct}%",
        dl_extracting: "Extracting ZIP archive...",
        api_error: "Error"
    },
    fr: {
        header_sub: "// GESTIONNAIRE DE LANGUE",
        launch_game: "▶ LANCER LE JEU",
        game_folder_label: "DOSSIER DU JEU (STEAM)",
        searching_path: "Recherche en cours...",
        path_not_found: "Jeu non détecté. Veuillez sélectionner le dossier manuellement.",
        browse_btn: "Parcourir...",
        available_packs_title: "PACKS DE TRADUCTION DISPONIBLES",
        refresh_btn: "Rafraîchir les versions",
        status_ready: "Prêt.",
        status_checking: "Vérification des mises à jour sur GitHub...",
        status_downloading: "Téléchargement et extraction de {repo}...",
        status_removing: "Suppression du pack {id}...",
        status_removed: "Pack supprimé avec succès.",
        status_path_updated: "Chemin du jeu mis à jour.",
        status_launching: "Démarrage de Limbus Company via Steam...",
        repo_label: "Dépôt :",
        installed_label: "Installé :",
        not_installed: "Non installé",
        update_available: "(MàJ disponible !)",
        btn_install: "Installer",
        btn_reinstall: "Réinstaller",
        btn_update: "Mettre à jour ({ver})",
        btn_remove: "Supprimer",
        dl_progress: "Téléchargement : {pct}%",
        dl_extracting: "Extraction de l'archive ZIP...",
        api_error: "Erreur"
    }
};

let currentLang = "en";

export function setLang(lang) {
    currentLang = translations[lang] ? lang : "en";
}

export function getLang() {
    return currentLang;
}

export function t(key, params = {}) {
    let str = translations[currentLang]?.[key] || translations["en"][key] || key;
    for (const [k, v] of Object.entries(params)) {
        str = str.replace(`{${k}}`, v);
    }
    return str;
}