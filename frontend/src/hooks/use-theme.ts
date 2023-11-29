import { computed, onBeforeMount } from 'vue';
import { getLightColor, getDarkColor } from '@/utils/theme/tool';
import { GlobalStore } from '@/store';
import { MsgSuccess } from '@/utils/message';

export const useTheme = () => {
    const globalStore = GlobalStore();
    const themeConfig = computed(() => globalStore.themeConfig);

    const switchDark = () => {
        if (themeConfig.value.theme === 'auto') {
            const prefersDark = window.matchMedia('(prefers-color-scheme: dark)');
            themeConfig.value.theme = prefersDark.matches ? 'dark' : 'light';
        }
        const body = document.documentElement as HTMLElement;
        if (themeConfig.value.theme === 'dark') body.setAttribute('class', 'dark');
        else body.setAttribute('class', '');
    };

    const changePrimary = (val: string) => {
        if (!val) {
            val = '#409EFF';
            MsgSuccess('主题颜色已重置为 #409EFF');
        }
        globalStore.setThemeConfig({ ...themeConfig.value, primary: val });
        document.documentElement.style.setProperty(
            '--el-color-primary-dark-2',
            `${getDarkColor(themeConfig.value.primary, 0.1)}`,
        );
        document.documentElement.style.setProperty('--el-color-primary', themeConfig.value.primary);
        for (let i = 1; i <= 9; i++) {
            document.documentElement.style.setProperty(
                `--el-color-primary-light-${i}`,
                `${getLightColor(themeConfig.value.primary, i / 10)}`,
            );
        }
    };

    onBeforeMount(() => {
        switchDark();
        changePrimary(themeConfig.value.primary);
    });

    return {
        switchDark,
        changePrimary,
    };
};
