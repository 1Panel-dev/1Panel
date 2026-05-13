import { useGlobalStore } from '@/composables/useGlobalStore';
import { getXpackSetting } from '@/utils/xpack';

export const useLogo = async () => {
    const { themeConfig, watermark, watermarkShow } = useGlobalStore();
    const res = await getXpackSetting();
    if (res) {
        localStorage.setItem('1p-favicon', res.data.logo);
        themeConfig.value.title = res.data.title;
        themeConfig.value.logo = res.data.logo;
        themeConfig.value.logoWithText = res.data.logoWithText;
        themeConfig.value.loginImage = res.data?.loginImage;
        themeConfig.value.loginBgType = res.data?.loginBgType;
        themeConfig.value.loginBackground = res.data?.loginBackground;
        themeConfig.value.loginBtnLinkColor = res.data?.loginBtnLinkColor;
        themeConfig.value.favicon = res.data.favicon;
        watermarkShow.value = res.data.watermarkShow === 'Enable';
        try {
            watermark.value = JSON.parse(res.data.watermark);
        } catch {
            watermark.value = null;
        }
    }

    const link = (document.querySelector("link[rel*='icon']") || document.createElement('link')) as HTMLLinkElement;
    link.type = 'image/x-icon';
    link.rel = 'shortcut icon';
    link.href = themeConfig.value.favicon ? `/api/v2/images/favicon?t=${Date.now()}` : '/public/favicon.png';
    document.getElementsByTagName('head')[0].appendChild(link);
};
