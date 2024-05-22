import { onBeforeMount, watch } from 'vue';
import { GlobalStore } from '@/store';
import { storeToRefs } from 'pinia';

export const useTheme = () => {
    const { themeConfig } = storeToRefs(GlobalStore());
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)');

    /**
     * This method is only executed when loading or manually switching for the first time.
     */
    const switchTheme = () => {
        if (themeConfig.value.theme === 'auto') {
            themeConfig.value.theme = prefersDark.matches ? 'dark' : 'light';
            if (prefersDark.addEventListener) {
                prefersDark.addEventListener('change', switchAccordingUserProxyTheme);
            } else if (prefersDark.addListener) {
                prefersDark.addListener(switchAccordingUserProxyTheme);
            }
        } else {
            prefersDark.removeEventListener('change', switchAccordingUserProxyTheme);
            prefersDark.removeListener(switchAccordingUserProxyTheme);
        }
        updateTheme(themeConfig.value.theme);
    };

    const switchAccordingUserProxyTheme = (event: MediaQueryListEvent) => {
        const preferTheme = event.matches ? 'dark' : 'light';

        themeConfig.value.theme = preferTheme;
        updateTheme(preferTheme);
    };

    const updateTheme = (theme: string) => {
        const body = document.documentElement as HTMLElement;
        body.setAttribute('class', theme);
    };

    onBeforeMount(() => {
        updateTheme(themeConfig.value.theme);
    });

    /**
     * Called internally by the system for automatically switching themes
     */
    const autoSwitchTheme = () => {
        let preferTheme = themeConfig.value.theme;
        if (themeConfig.value.theme === 'auto') {
            preferTheme = prefersDark.matches ? 'dark' : 'light';
        }
        updateTheme(preferTheme);
    };

    watch(themeConfig, () => {
        autoSwitchTheme();
    });

    return {
        autoSwitchTheme,
        switchTheme,
    };
};
