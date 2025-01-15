import { GlobalStore } from '@/store';
import { setPrimaryColor } from '@/utils/theme';

export const useTheme = () => {
    const switchTheme = () => {
        const globalStore = GlobalStore();
        const themeConfig = globalStore.themeConfig;
        let itemTheme = themeConfig.theme;
        if (itemTheme === 'auto') {
            const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
            itemTheme = prefersDark ? 'dark' : 'light';
        }
        document.documentElement.className = itemTheme === 'dark' ? 'dark' : 'light';
        if (globalStore.isProductPro && themeConfig.themeColor) {
            try {
                const themeColor = JSON.parse(themeConfig.themeColor);
                const color = itemTheme === 'dark' ? themeColor.dark : themeColor.light;

                if (color) {
                    themeConfig.primary = color;
                    setPrimaryColor(color);
                }
            } catch (e) {
                console.error('Failed to parse themeColor', e);
            }
        }
    };

    return {
        switchTheme,
    };
};
