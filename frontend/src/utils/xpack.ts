import {
    getLicenseStatus,
    getMasterLicenseStatus,
    getSettingBaseInfo,
    getEnterpriseLicenseStatus,
} from '@/api/modules/setting';
import { useTheme } from '@/global/use-theme';
import {
    searchXpackSetting,
    updateXpackSettingByKey as updateXpackSettingByKeyFromExtension,
} from '@/extensions/xpack';
import { useGlobalStore } from '@/composables/useGlobalStore';
import faviconUrl from '@/assets/images/favicon.svg';

export function resetXSetting() {
    const { masterAlias, themeConfig, watermark, watermarkShow } = useGlobalStore();
    themeConfig.value.title = '';
    themeConfig.value.logo = '';
    themeConfig.value.logoWithText = '';
    themeConfig.value.favicon = '';
    watermark.value = null;
    watermarkShow.value = false;
    masterAlias.value = '';
}

async function getColoredFavicon(url: string, color: string) {
    const res = await fetch(url);
    let svgText = await res.text();
    svgText = svgText.replace(/fill=(["'])(.*?)\1/g, `fill="${color}"`);
    return `data:image/svg+xml,${encodeURIComponent(svgText)}`;
}

export async function initFavicon() {
    const { isXpackOrEE, themeConfig } = useGlobalStore();
    document.title = themeConfig.value.panelName;
    const favicon = themeConfig.value.favicon;
    const isPro = isXpackOrEE.value;
    const themeColor = themeConfig.value.primary;
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
    const res = await searchXpackSetting();
    if (!res) {
        initFavicon();
        resetXSetting();
        return;
    }
    initFavicon();
    return res;
}

const loadDataFromDB = async () => {
    const { entrance, openMenuTabs } = useGlobalStore();
    const res = await getSettingBaseInfo();
    document.title = res.data.panelName;
    entrance.value = res.data.securityEntrance;
    openMenuTabs.value = res.data.menuTabs === 'Enable';
};

export async function loadProductProFromDB() {
    const { isEnterprise, isEnterpriseLicenseLoaded, isEnterpriseLicensed, isProductPro, productProExpires } =
        useGlobalStore();
    if (!isEnterprise.value) {
        isEnterpriseLicenseLoaded.value = true;
        const res = await getLicenseStatus();
        if (!res || !res.data) {
            isProductPro.value = false;
        } else {
            isProductPro.value = res.data.status === 'Bound';
            if (isProductPro.value) {
                productProExpires.value = Number(res.data.productPro);
            }
        }
        return;
    }
    const res = await getEnterpriseLicenseStatus();
    isEnterpriseLicenseLoaded.value = true;
    if (!res || !res.data) {
        isEnterpriseLicensed.value = false;
    } else {
        isEnterpriseLicensed.value = res.data.status === 'Bound';
        isProductPro.value = isEnterpriseLicensed.value;
    }
}

export async function loadMasterProductProFromDB() {
    const { isEnterprise, isEnterpriseLicenseLoaded, isEnterpriseLicensed, isMasterProductPro } = useGlobalStore();
    if (!isEnterprise.value) {
        isEnterpriseLicenseLoaded.value = true;
        const res = await getMasterLicenseStatus();
        if (!res || !res.data) {
            isMasterProductPro.value = false;
        } else {
            isMasterProductPro.value = res.data.status === 'Bound';
        }
    } else {
        const res = await getEnterpriseLicenseStatus();
        isEnterpriseLicenseLoaded.value = true;
        if (!res || !res.data) {
            isEnterpriseLicensed.value = false;
        } else {
            isEnterpriseLicensed.value = res.data.status === 'Bound';
            isMasterProductPro.value = res.data.status === 'Bound';
        }
    }
    useTheme().switchTheme();
    initFavicon();
    loadDataFromDB();
}

export async function getXpackSettingForTheme() {
    const { masterAlias, themeConfig, watermark, watermarkShow } = useGlobalStore();
    const res2 = await searchXpackSetting();
    if (res2) {
        themeConfig.value.title = res2.data?.title;
        themeConfig.value.logo = res2.data?.logo;
        themeConfig.value.logoWithText = res2.data?.logoWithText;
        themeConfig.value.favicon = res2.data?.favicon;
        themeConfig.value.loginImage = res2.data?.loginImage;
        themeConfig.value.loginBgType = res2.data?.loginBgType;
        themeConfig.value.loginBackground = res2.data?.loginBackground;
        themeConfig.value.loginBtnLinkColor = res2.data?.loginBtnLinkColor;
        themeConfig.value.themeColor = res2.data?.themeColor;
        masterAlias.value = res2.data.masterAlias;
        if (res2.data?.theme) {
            themeConfig.value.theme = res2.data.theme;
        }
        watermarkShow.value = res2.data.watermarkShow === 'Enable';
        try {
            watermark.value = JSON.parse(res2.data.watermark);
        } catch {
            watermark.value = null;
        }
    } else {
        resetXSetting();
    }
    useTheme().switchTheme();
    initFavicon();
}

export async function updateXpackSettingByKey(key: string, value: string) {
    return updateXpackSettingByKeyFromExtension(key, value);
}
