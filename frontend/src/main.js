import {
    GetConfig,
    SelectGamePathManual,
    InstallOrUpdatePack,
    RemovePack,
    CheckRemoteVersion,
    LaunchGame
} from '../wailsjs/go/main/App';
import { EventsOn } from '../wailsjs/runtime/runtime';
import { t, setLang, getLang } from './i18n';

const AVAILABLE_PACKS = [
    {
        id: "fr-pack",
        name: {
            fr: "Traduction Française",
            en: "French Translation"
        },
        author: "Eden-Office",
        repo: "Eden-Office/LimbusCompanyBusFR"
    },
    {
        id: "es-pack",
        name: {
            fr: "Traduction Espagnole",
            en: "Spanish Translation"
        },
        author: "Dreams-Office",
        repo: "Dreams-Office/LimbusCompanySpanishTranslationTeam"
    },
    {
        id: "ru-pack",
        name: {
            fr: "Traduction Russe",
            en: "Russian Translation"
        },
        author: "Crescent-Corporation",
        repo: "Crescent-Corporation/LimbusCompanyBusRUS"
    },
    {
        id: 'th-pack',
        name: {
            fr: "Traduction Thaïlandaise",
            en: "Thai Translation"
        },
        author: "stlinx",
        repo: "stlinx/LocalizeLimbusTH",
    },
    {
        id: 'cn-pack',
        name: {
            fr: "Traduction Chinoise Simplifiée",
            en: "Simplified Chinese Translation"
        },
        author: 'LocalizeLimbusCompany',
        repo: 'LocalizeLimbusCompany/LocalizeLimbusCompany',
    },
    {
        id: 'tc-pack',
        name: {
            fr: "Traduction Chinoise Traditionnelle",
            en: "Traditional Chinese Translation"
        },
        author: 'LimbusTraditionalMandarin',
        repo: 'LimbusTraditionalMandarin/storyline',
    }
];

let appConfig = { game_path: "", installed_packs: {} };
let remoteVersions = {};

const pathDisplay = document.getElementById('game-path-display');
const packsList = document.getElementById('packs-list');
const statusBar = document.getElementById('status-bar');
const langSelect = document.getElementById('lang-select');

function setStatus(message, isError = false) {
    statusBar.textContent = message;
    statusBar.style.color = isError ? '#ff6b6b' : '#8c8a87';
}

function applyTranslations() {
    document.querySelectorAll('[data-i18n]').forEach(el => {
        const key = el.getAttribute('data-i18n');
        el.textContent = t(key);
    });
    updatePathUI();
    renderPacks();
    setStatus(t("status_ready"));
}

async function init() {
    try {
        appConfig = await GetConfig();
        const initialLang = appConfig.app_lang || "en";
        setLang(initialLang);
        langSelect.value = initialLang;
        applyTranslations();

        EventsOn("download:progress", (data) => {
            const wrapper = document.getElementById(`progress-wrap-${data.packId}`);
            const fill = document.getElementById(`progress-fill-${data.packId}`);
            const label = document.getElementById(`progress-label-${data.packId}`);
            if (!wrapper || !fill || !label) return;

            if (data.stage === "downloading") {
                wrapper.classList.add('active');
                fill.style.width = `${data.percent}%`;
                label.textContent = data.percent === 100
                    ? t("dl_extracting")
                    : t("dl_progress", { pct: data.percent });
            } else if (data.stage === "done") {
                wrapper.classList.remove('active');
                fill.style.width = '0%';
            }
        });

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
        pathDisplay.textContent = t("path_not_found");
        pathDisplay.style.color = "#ff6b6b";
    }
}

async function fetchAllRemoteVersions() {
    setStatus(t("status_checking"));
    for (const pack of AVAILABLE_PACKS) {
        try {
            const latestTag = await CheckRemoteVersion(pack.repo);
            remoteVersions[pack.id] = latestTag;
        } catch (err) {
            remoteVersions[pack.id] = "ERR";
        }
    }
    renderPacks();
    setStatus(t("status_ready"));
}

function renderPacks() {
    packsList.innerHTML = '';
    const lang = getLang();

    AVAILABLE_PACKS.forEach(pack => {
        const installed = appConfig.installed_packs?.[pack.id];
        const localVer = installed ? installed.version : t("not_installed");
        const rawRemote = remoteVersions[pack.id];
        const remoteVer = !rawRemote ? "..." : (rawRemote === "ERR" ? t("api_error") : rawRemote);

        const hasUpdate = installed && rawRemote && rawRemote !== "ERR" && installed.version !== rawRemote;

        let actionLabel = t("btn_install");
        if (installed) {
            actionLabel = hasUpdate ? t("btn_update", { ver: remoteVer }) : t("btn_reinstall");
        }

        const packTitle = pack.name[lang] || pack.name["en"];

        const card = document.createElement('div');
        card.className = 'pack-card';
        card.innerHTML = `
            <div class="pack-row">
                <div class="pack-details">
                    <h3>${packTitle}</h3>
                    <div class="pack-meta">
                        <span>${t("repo_label")} ${pack.repo}</span>
                        <span>${t("installed_label")} <strong>${localVer}</strong></span>
                        <span class="${hasUpdate ? 'badge-update' : ''}">
                            GitHub : ${remoteVer} ${hasUpdate ? t("update_available") : ''}
                        </span>
                    </div>
                </div>
                <div class="pack-actions">
                    <button class="btn btn-install" data-id="${pack.id}" data-repo="${pack.repo}">
                        ${actionLabel}
                    </button>
                    ${installed ? `<button class="btn danger btn-remove" data-id="${pack.id}">${t("btn_remove")}</button>` : ''}
                </div>
            </div>
            <div id="progress-wrap-${pack.id}" class="progress-wrapper">
                <div class="progress-track">
                    <div id="progress-fill-${pack.id}" class="progress-fill"></div>
                </div>
                <span id="progress-label-${pack.id}" class="progress-label">0%</span>
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
            setStatus(t("status_downloading", { repo }));

            try {
                const msg = await InstallOrUpdatePack(packId, repo);
                appConfig = await GetConfig();
                renderPacks();
                setStatus(msg);
            } catch (err) {
                setStatus(`${err}`, true);
                e.target.disabled = false;
            }
        });
    });

    document.querySelectorAll('.btn-remove').forEach(btn => {
        btn.addEventListener('click', async (e) => {
            const packId = e.target.dataset.id;
            e.target.disabled = true;
            setStatus(t("status_removing", { id: packId }));

            try {
                await RemovePack(packId);
                appConfig = await GetConfig();
                renderPacks();
                setStatus(t("status_removed"));
            } catch (err) {
                setStatus(`${err}`, true);
                e.target.disabled = false;
            }
        });
    });
}

langSelect.addEventListener('change', async (e) => {
    const newLang = e.target.value;
    setLang(newLang);
    applyTranslations();
    await SetAppLanguage(newLang);
});

document.getElementById('btn-change-path').addEventListener('click', async () => {
    try {
        const newPath = await SelectGamePathManual();
        appConfig.game_path = newPath;
        updatePathUI();
        setStatus(t("status_path_updated"));
    } catch (err) {
        setStatus(`${err}`, true);
    }
});

document.getElementById('btn-refresh').addEventListener('click', fetchAllRemoteVersions);

document.getElementById('btn-launch-game').addEventListener('click', async () => {
    setStatus(t("status_launching"));
    await LaunchGame();
});

window.addEventListener('DOMContentLoaded', init);