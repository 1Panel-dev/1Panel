import { GlobalStore } from '@/store';
import { getXpackSetting } from '@/utils/xpack';

export const useLogo = async () => {
    const globalStore = GlobalStore();
    const res = await getXpackSetting();
    if (res) {
        localStorage.setItem('1p-favicon', res.data.logo);
        globalStore.themeConfig.title = res.data.title;
        globalStore.themeConfig.logo = res.data.logo;
        globalStore.themeConfig.logoWithText = res.data.logoWithText;
        globalStore.themeConfig.favicon = res.data.favicon;
    }

    const link = (document.querySelector("link[rel*='icon']") || document.createElement('link')) as HTMLLinkElement;
    link.type = 'image/x-icon';
    link.rel = 'shortcut icon';
    link.href = globalStore.themeConfig.favicon ? `/api/v1/images/favicon?t=${Date.now()}` : '/public/favicon.png';
    document.getElementsByTagName('head')[0].appendChild(link);
};
