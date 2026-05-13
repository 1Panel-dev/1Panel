import { getCurrentScope, onScopeDispose } from 'vue';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { setPrimaryColor } from '@/utils/theme';

export const useTheme = () => {
    const switchTheme = () => {
        const { isXpackOrEE, themeConfig } = useGlobalStore();
        let itemTheme = themeConfig.value.theme;
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

    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    const onSystemThemeChange = () => {
        const { themeConfig } = useGlobalStore();

        if (themeConfig.value.theme === 'auto') {
            switchTheme();
        }
    };
    mediaQuery.addEventListener('change', onSystemThemeChange);
    if (getCurrentScope()) {
        onScopeDispose(() => {
            mediaQuery.removeEventListener('change', onSystemThemeChange);
        });
    }

    return {
        switchTheme,
    };
};
