import { GlobalStore } from '@/store';
import { searchXSetting } from '@/xpack/frontend/api/modules/setting';
import { computed } from 'vue';

export const useLogo = async () => {
    const globalStore = GlobalStore();
    const themeConfig = computed(() => globalStore.themeConfig);

    if (!themeConfig.value.logo) {
        const res = await searchXSetting();
        globalStore.setThemeConfig({
            ...themeConfig.value,
            title: res.data.title,
            logo: res.data.logo,
            logoWithText: res.data.logoWithText,
            favicon: res.data.favicon,
        });
    }
    let link = document.querySelector("link[rel*='icon']") as HTMLLinkElement | null;
    if (link) {
        if (globalStore.themeConfig.favicon) {
            link.href = globalStore.themeConfig.favicon;
        } else {
            link.href = '/public/favicon.png';
        }
    } else {
        const newLink = document.createElement('link');
        newLink.rel = 'icon';
        if (globalStore.themeConfig.favicon) {
            newLink.href = globalStore.themeConfig.favicon;
        } else {
            newLink.href = '/public/favicon.png';
        }
        document.head.appendChild(newLink);
    }
};
