import { GlobalStore } from '@/store';

export const useTheme = () => {
    const globalStore = GlobalStore();
    const switchTheme = () => {
        if (globalStore.themeConfig.isGold && globalStore.isMasterProductPro) {
            const body = document.documentElement as HTMLElement;
            body.setAttribute('class', 'dark-gold');
            return;
        }

        let itemTheme = globalStore.themeConfig.theme;
        if (globalStore.themeConfig.theme === 'auto') {
            const prefersDark = window.matchMedia('(prefers-color-scheme: dark)');
            itemTheme = prefersDark.matches ? 'dark' : 'light';
        }
        const body = document.documentElement as HTMLElement;
        if (itemTheme === 'dark') body.setAttribute('class', 'dark');
        else body.setAttribute('class', '');
    };

    return {
        switchTheme,
    };
};
