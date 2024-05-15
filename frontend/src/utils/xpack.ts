import { GlobalStore } from '@/store';
const globalStore = GlobalStore();

export function resetXSetting() {
    globalStore.themeConfig.title = '';
    globalStore.themeConfig.logo = '';
    globalStore.themeConfig.logoWithText = '';
    globalStore.themeConfig.favicon = '';
}

export function initFavicon() {
    let favicon = globalStore.themeConfig.favicon;
    const link = (document.querySelector("link[rel*='icon']") || document.createElement('link')) as HTMLLinkElement;
    link.type = 'image/x-icon';
    link.rel = 'shortcut icon';
    link.href = favicon ? '/api/v1/images/favicon' : '/public/favicon.png';
    document.getElementsByTagName('head')[0].appendChild(link);
}

export async function getXpackSetting() {
    const searchXSettingGlob = import.meta.glob('xpack/api/modules/setting.ts');
    const module = await searchXSettingGlob?.['../xpack/api/modules/setting.ts']?.();
    const res = module?.searchXSetting();
    if (!res) {
        resetXSetting();
        return;
    }
    return res;
}

export async function updateXpackSetting(fromData: FormData) {
    const searchXSettingGlob = import.meta.glob('xpack/api/modules/setting.ts');
    const module = await searchXSettingGlob?.['../xpack/api/modules/setting.ts']?.();
    return module?.updateXSetting(fromData);
}
