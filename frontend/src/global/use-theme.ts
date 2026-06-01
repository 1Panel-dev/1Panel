import { useGlobalStore } from '@/composables/useGlobalStore';
import { setPrimaryColor } from '@/utils/theme';

let themeListenerInitialized = false;

export const useTheme = () => {
    const { isXpackOrEE, themeConfig } = useGlobalStore();

    const switchTheme = () => {
        const rawTheme = themeConfig.value.theme;
        let itemTheme = rawTheme;
        if (itemTheme === 'auto') {
            const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
            itemTheme = prefersDark ? 'dark' : 'light';
        }
        document.documentElement.className = itemTheme === 'dark' ? 'dark' : 'light';
        if (isXpackOrEE.value && themeConfig.value.themeColor) {
            try {
                const themeColor = JSON.parse(themeConfig.value.themeColor);
                const color = itemTheme === 'dark' ? themeColor.dark : themeColor.light;

                if (color) {
                    themeConfig.value.primary = color;
                    setPrimaryColor(color);
                }
            } catch (e) {
                console.error('Failed to parse themeColor', e);
            }
        }
    };

    const ensureSystemThemeListener = () => {
        if (themeListenerInitialized || typeof window === 'undefined') {
            return;
        }

        const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
        const onSystemThemeChange = () => {
            if (themeConfig.value.theme === 'auto') {
                switchTheme();
            }
        };

        mediaQuery.addEventListener('change', onSystemThemeChange);
        themeListenerInitialized = true;
    };

    ensureSystemThemeListener();

    return {
        switchTheme,
    };
};
