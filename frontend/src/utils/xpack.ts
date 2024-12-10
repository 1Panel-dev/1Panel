import { getLicenseStatus, getMasterLicenseStatus, getSettingInfo } from '@/api/modules/setting';
import { useTheme } from '@/global/use-theme';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();
const { switchTheme } = useTheme();

export function resetXSetting() {
    globalStore.themeConfig.title = '';
    globalStore.themeConfig.logo = '';
    globalStore.themeConfig.logoWithText = '';
    globalStore.themeConfig.favicon = '';
    globalStore.themeConfig.isGold = false;
}

export function initFavicon() {
    document.title = globalStore.themeConfig.panelName;
    let favicon = globalStore.themeConfig.favicon;
    const link = (document.querySelector("link[rel*='icon']") || document.createElement('link')) as HTMLLinkElement;
    link.type = 'image/x-icon';
    link.rel = 'shortcut icon';
    if (globalStore.isDarkGoldTheme) {
        let goldLink = new URL(`../assets/images/favicon-gold.png`, import.meta.url).href;
        link.href = favicon ? '/api/v2/images/favicon' : goldLink;
    } else {
        link.href = favicon ? '/api/v2/images/favicon' : '/public/favicon.png';
    }
    document.getElementsByTagName('head')[0].appendChild(link);
}

export async function getXpackSetting() {
    let searchXSetting;
    const xpackModules = import.meta.glob('../xpack/api/modules/setting.ts', { eager: true });
    if (xpackModules['../xpack/api/modules/setting.ts']) {
        searchXSetting = xpackModules['../xpack/api/modules/setting.ts']['searchXSetting'] || {};
        const res = await searchXSetting();
        if (!res) {
            initFavicon();
            resetXSetting();
            return;
        }
        initFavicon();
        return res;
    }
}

const loadDataFromDB = async () => {
    const res = await getSettingInfo();
    document.title = res.data.panelName;
    globalStore.entrance = res.data.securityEntrance;
    globalStore.setOpenMenuTabs(res.data.menuTabs === 'enable');
};

export async function loadProductProFromDB() {
    const res = await getLicenseStatus();
    if (!res || !res.data) {
        globalStore.isProductPro = false;
    } else {
        globalStore.isProductPro = res.data.status === 'bound';
        if (globalStore.isProductPro) {
            globalStore.productProExpires = Number(res.data.productPro);
        }
    }
    loadDataFromDB();
}

export async function loadMasterProductProFromDB() {
    const res = await getMasterLicenseStatus();
    if (!res || !res.data) {
        globalStore.isMasterProductPro = false;
    } else {
        globalStore.isMasterProductPro = res.data.status === 'bound';
    }
    switchTheme();
    initFavicon();
}

export async function getXpackSettingForTheme() {
    const res = await getMasterLicenseStatus();
    if (!res.data) {
        globalStore.isMasterProductPro = false;
        resetXSetting();
        switchTheme();
        initFavicon();
        return;
    }
    globalStore.isMasterProductPro = res.data.status === 'bound';
    if (!globalStore.isMasterProductPro) {
        globalStore.isMasterProductPro = false;
        resetXSetting();
        switchTheme();
        initFavicon();
        return;
    }

    let searchXSetting;
    const xpackModules = import.meta.glob('../xpack/api/modules/setting.ts', { eager: true });
    if (xpackModules['../xpack/api/modules/setting.ts']) {
        searchXSetting = xpackModules['../xpack/api/modules/setting.ts']['searchXSetting'] || {};
        const res2 = await searchXSetting();
        if (res2) {
            globalStore.themeConfig.title = res2.data?.title;
            globalStore.themeConfig.logo = res2.data?.logo;
            globalStore.themeConfig.logoWithText = res2.data?.logoWithText;
            globalStore.themeConfig.favicon = res2.data?.favicon;
            globalStore.themeConfig.isGold = res2.data?.theme === 'dark-gold';
        } else {
            resetXSetting();
        }
    }
    switchTheme();
    initFavicon();
}

export async function updateXpackSettingByKey(key: string, value: string) {
    let updateXSettingByKey;
    const xpackModules = import.meta.glob('../xpack/api/modules/setting.ts', { eager: true });
    if (xpackModules['../xpack/api/modules/setting.ts']) {
        updateXSettingByKey = xpackModules['../xpack/api/modules/setting.ts']['updateXSettingByKey'] || {};
        return updateXSettingByKey(key, value);
    }
}
