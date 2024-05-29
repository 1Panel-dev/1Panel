import { GlobalStore } from '@/store';

export const useTheme = () => {
    const globalStore = GlobalStore();
    const switchTheme = () => {
        if (globalStore.themeConfig.isGold && globalStore.isProductPro) {
            const body = document.documentElement as HTMLElement;
            body.setAttribute('class', 'dark-gold');
            return;
        }
        const body = document.documentElement as HTMLElement;
        if (globalStore.themeConfig.theme === 'dark') body.setAttribute('class', 'dark');
        else body.setAttribute('class', '');
    };

    return {
        switchTheme,
    };
};
