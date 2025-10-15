import { getLicenseStatus, getMasterLicenseStatus, getSettingInfo } from '@/api/modules/setting';
import { useTheme } from '@/global/use-theme';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();
const { switchTheme } = useTheme();
import faviconUrl from '@/assets/images/favicon.svg';

export function resetXSetting() {
    globalStore.themeConfig.title = '';
    globalStore.themeConfig.logo = '';
    globalStore.themeConfig.logoWithText = '';
    globalStore.themeConfig.favicon = '';
    globalStore.watermark = null;
    globalStore.masterAlias = '';
}

async function getColoredFavicon(url: string, color: string) {
    const res = await fetch(url);
    let svgText = await res.text();
    svgText = svgText.replace(/fill=(["'])(.*?)\1/g, `fill="${color}"`);
    return `data:image/svg+xml,${encodeURIComponent(svgText)}`;
}

export async function initFavicon() {
    document.title = globalStore.themeConfig.panelName;
    const favicon = globalStore.themeConfig.favicon;
    const isPro = globalStore.isMasterProductPro;
    const themeColor = globalStore.themeConfig.primary;
    const customFaviconUrl = `/api/v2/images/favicon?t=${Date.now()}`;
    const fallbackSvg = isPro ? await getColoredFavicon(faviconUrl, themeColor) : '/public/favicon.png';
    const setLink = (href: string) => {
        let link = document.querySelector("link[rel*='icon']") as HTMLLinkElement;
        if (!link) {
            link = document.createElement('link');
            link.rel = 'shortcut icon';
            link.type = 'image/x-icon';
            document.head.appendChild(link);
        }
        link.href = href;
    };

    if (favicon) {
        const testImg = new Image();
        testImg.onload = () => setLink(customFaviconUrl);
        testImg.onerror = () => setLink(fallbackSvg);
        testImg.src = customFaviconUrl;
    } else {
        setLink(fallbackSvg);
    }
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
    globalStore.setOpenMenuTabs(res.data.menuTabs === 'Enable');
};

export async function loadProductProFromDB() {
    const res = await getLicenseStatus();
    if (!res || !res.data) {
        globalStore.isProductPro = false;
    } else {
        globalStore.isProductPro = res.data.status === 'Bound';
        if (globalStore.isProductPro) {
            globalStore.productProExpires = Number(res.data.productPro);
        }
    }
}

export async function loadMasterProductProFromDB() {
    const res = await getMasterLicenseStatus();
    if (!res || !res.data) {
        globalStore.isMasterProductPro = false;
    } else {
        globalStore.isMasterProductPro = res.data.status === 'Bound';
    }
    switchTheme();
    initFavicon();
    loadDataFromDB();
}

export async function getXpackSettingForTheme() {
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
            globalStore.themeConfig.loginImage = res2.data?.loginImage;
            globalStore.themeConfig.loginBgType = res2.data?.loginBgType;
            globalStore.themeConfig.loginBackground = res2.data?.loginBackground;
            globalStore.themeConfig.loginBtnLinkColor = res2.data?.loginBtnLinkColor;
            globalStore.themeConfig.themeColor = res2.data?.themeColor;
            globalStore.masterAlias = res2.data.masterAlias;
            if (res2.data?.theme) {
                globalStore.themeConfig.theme = res2.data.theme;
            }
            try {
                globalStore.watermark = JSON.parse(res2.data.watermark);
            } catch {
                globalStore.watermark = null;
            }
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
