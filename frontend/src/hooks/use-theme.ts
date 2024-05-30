import { GlobalStore } from '@/store';

export const useTheme = () => {
    const globalStore = GlobalStore();
    const switchTheme = () => {
        if (globalStore.themeConfig.isGold && globalStore.isProductPro) {
            const body = document.documentElement as HTMLElement;
            body.setAttribute('class', 'dark-gold');
            return;
        }
        if (globalStore.themeConfig.theme === 'auto') {
            const prefersDark = window.matchMedia('(prefers-color-scheme: dark)');
            globalStore.themeConfig.theme = prefersDark.matches ? 'dark' : 'light';
        }
        const body = document.documentElement as HTMLElement;
        if (globalStore.themeConfig.theme === 'dark') body.setAttribute('class', 'dark');
        else body.setAttribute('class', '');
    };

    return {
        switchTheme,
    };
};
