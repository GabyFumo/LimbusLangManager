import {
    GetConfig,
    SelectGamePathManual,
    InstallOrUpdatePack,
    RemovePack,
    CheckRemoteVersion
} from '../wailsjs/go/main/App';

const AVAILABLE_PACKS = [
    {
        id: "fr-pack",
        name: "Traduction Française",
        author: "Équipe FR",
        repo: "Eden-Office/LimbusCompanyBusFR"
    },
    {
        id: "es-pack",
        name: "Traducción al Español",
        author: "Comunidad ES",
        repo: "Dreams-Office/LimbusCompanySpanishTranslationTeam"
    }
];

let appConfig = { game_path: "", installed_packs: {} };
let remoteVersions = {};

const pathDisplay = document.getElementById('game-path-display');
const packsList = document.getElementById('packs-list');
const statusBar = document.getElementById('status-bar');

function setStatus(message, isError = false) {
    statusBar.textContent = message;
    statusBar.style.color = isError ? '#ff6b6b' : '#8c8a87';
}

async function init() {
    try {
        appConfig = await GetConfig();
        updatePathUI();
        renderPacks();
        await fetchAllRemoteVersions();
    } catch (err) {
        setStatus(`Erreur d'initialisation : ${err}`, true);
    }
}

function updatePathUI() {
    if (appConfig.game_path) {
        pathDisplay.textContent = appConfig.game_path;
        pathDisplay.style.color = "#e8e6e3";
    } else {
        pathDisplay.textContent = "Jeu non détecté. Veuillez sélectionner le dossier manuellement.";
        pathDisplay.style.color = "#ff6b6b";
    }
}

async function fetchAllRemoteVersions() {
    setStatus("Vérification des mises à jour sur GitHub...");
    for (const pack of AVAILABLE_PACKS) {
        try {
            const latestTag = await CheckRemoteVersion(pack.repo);
            remoteVersions[pack.id] = latestTag;
        } catch (err) {
            remoteVersions[pack.id] = "Erreur API";
        }
    }
    renderPacks();
    setStatus("Prêt.");
}

function renderPacks() {
    packsList.innerHTML = '';

    AVAILABLE_PACKS.forEach(pack => {
        const installed = appConfig.installed_packs?.[pack.id];
        const localVer = installed ? installed.version : "Non installé";
        const remoteVer = remoteVersions[pack.id] || "...";

        const hasUpdate = installed && remoteVer !== "..." && remoteVer !== "Erreur API" && localVer !== remoteVer;

        let actionLabel = "Installer";
        if (installed) {
            actionLabel = hasUpdate ? `Mettre à jour (${remoteVer})` : "Réinstaller";
        }

        const card = document.createElement('div');
        card.className = 'pack-card';
        card.innerHTML = `
            <div class="pack-details">
                <h3>${pack.name}</h3>
                <div class="pack-meta">
                    <span>Dépôt : ${pack.repo}</span>
                    <span>Installé : <strong>${localVer}</strong></span>
                    <span class="${hasUpdate ? 'badge-update' : ''}">
                        GitHub : ${remoteVer} ${hasUpdate ? '(MàJ disponible !)' : ''}
                    </span>
                </div>
            </div>
            <div class="pack-actions">
                <button class="btn btn-install" data-id="${pack.id}" data-repo="${pack.repo}">
                    ${actionLabel}
                </button>
                ${installed ? `<button class="btn danger btn-remove" data-id="${pack.id}">Supprimer</button>` : ''}
            </div>
        `;
        packsList.appendChild(card);
    });

    attachPackEvents();
}

function attachPackEvents() {
    document.querySelectorAll('.btn-install').forEach(btn => {
        btn.addEventListener('click', async (e) => {
            const packId = e.target.dataset.id;
            const repo = e.target.dataset.repo;

            e.target.disabled = true;
            setStatus(`Téléchargement et extraction de ${repo} en cours...`);

            try {
                const msg = await InstallOrUpdatePack(packId, repo);
                appConfig = await GetConfig();
                renderPacks();
                setStatus(msg);
            } catch (err) {
                setStatus(`Erreur : ${err}`, true);
                e.target.disabled = false;
            }
        });
    });

    document.querySelectorAll('.btn-remove').forEach(btn => {
        btn.addEventListener('click', async (e) => {
            const packId = e.target.dataset.id;
            e.target.disabled = true;
            setStatus(`Suppression du pack ${packId}...`);

            try {
                await RemovePack(packId);
                appConfig = await GetConfig();
                renderPacks();
                setStatus(`Pack supprimé avec succès.`);
            } catch (err) {
                setStatus(`Erreur lors de la suppression : ${err}`, true);
                e.target.disabled = false;
            }
        });
    });
}

document.getElementById('btn-change-path').addEventListener('click', async () => {
    try {
        const newPath = await SelectGamePathManual();
        appConfig.game_path = newPath;
        updatePathUI();
        setStatus("Chemin du jeu mis à jour.");
    } catch (err) {
        setStatus(`${err}`, true);
    }
});

document.getElementById('btn-refresh').addEventListener('click', fetchAllRemoteVersions);

window.addEventListener('DOMContentLoaded', init);