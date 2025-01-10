<template>
    <div>
        <el-drawer
            v-model="drawerVisible"
            :destroy-on-close="true"
            :close-on-click-modal="false"
            :close-on-press-escape="false"
            size="30%"
        >
            <template #header>
                <DrawerHeader :header="$t('xpack.theme.customColor')" :back="handleClose" />
            </template>
            <el-form ref="formRef" label-position="top" :model="form" @submit.prevent v-loading="loading">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item :label="$t('setting.light')" prop="light">
                            <div class="flex flex-wrap justify-between items-center sm:items-start">
                                <div class="flex gap-1">
                                    <template v-for="colorConfig in lightColors" :key="colorConfig.color">
                                        <el-tooltip :content="$t(colorConfig.label)" placement="top">
                                            <el-button
                                                :color="colorConfig.color"
                                                :class="form.light === colorConfig.color ? 'selected-white' : ''"
                                                circle
                                                dark
                                                @click="changeLightColor(colorConfig.color)"
                                            />
                                        </el-tooltip>
                                    </template>
                                    <el-color-picker
                                        v-model="form.light"
                                        class="ml-4"
                                        :predefine="lightPredefineColors"
                                        show-alpha
                                        @change="changeThemeColor('light', form.light)"
                                    />
                                </div>
                            </div>
                        </el-form-item>

                        <el-form-item :label="$t('setting.dark')" prop="dark">
                            <div class="flex flex-wrap justify-between items-center sm:items-start">
                                <div class="flex flex-wrap justify-between items-center mt-4 sm:mt-0">
                                    <div class="flex gap-1">
                                        <template v-for="colorConfig in darkColors" :key="colorConfig.color">
                                            <el-tooltip :content="$t(colorConfig.label)" placement="top">
                                                <el-button
                                                    :color="colorConfig.color"
                                                    :class="form.dark === colorConfig.color ? 'selected-white' : ''"
                                                    circle
                                                    dark
                                                    @click="changeDarkColor(colorConfig.color)"
                                                />
                                            </el-tooltip>
                                        </template>
                                        <el-color-picker
                                            v-model="form.dark"
                                            class="ml-4"
                                            :predefine="darkPredefineColors"
                                            show-alpha
                                            @change="changeThemeColor('dark', form.dark)"
                                        />
                                    </div>
                                </div>
                            </div>
                        </el-form-item>
                    </el-col>
                </el-row>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="onReSet">{{ $t('xpack.theme.setDefault') }}</el-button>
                    <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button :disabled="loading" type="primary" @click="onSave(formRef)">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-drawer>
    </div>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { FormInstance } from 'element-plus';
import { initFavicon, updateXpackSettingByKey } from '@/utils/xpack';
import { setPrimaryColor } from '@/utils/theme';
import { GlobalStore } from '@/store';

const emit = defineEmits<(e: 'search') => void>();
const drawerVisible = ref();
const loading = ref();

interface DialogProps {
    themeColor: {
        light: string;
        dark: string;
        themePredefineColors: {
            light: string[];
            dark: string[];
        };
    };
    theme: '';
}

interface ThemeColor {
    light: string;
    dark: string;
    themePredefineColors: {
        light: string[];
        dark: string[];
    };
}

const form = reactive({
    themeColor: {} as ThemeColor,
    theme: '',
    light: '',
    dark: '',
});
const formRef = ref<FormInstance>();

const globalStore = GlobalStore();

const STORAGE_KEY = 'theme-predefine-colors';
const lightPredefineColors = ref([
    '#005eeb',
    '#238636',
    '#3D8EFF',
    '#F0BE96',
    '#00ced1',
    '#c71585',
    '#ff4500',
    '#ff8c00',
    '#ffd700',
    '#333539',
]);
const darkPredefineColors = ref([
    '#238636',
    '#3D8EFF',
    '#005eeb',
    '#F0BE96',
    '#00ced1',
    '#c71585',
    '#ff4500',
    '#ff8c00',
    '#ffd700',
    '#333539',
]);

const defaultDarkColors = [
    { color: '#3D8EFF', label: 'xpack.theme.classicBlue' },
    { color: '#F0BE96', label: 'xpack.theme.lingXiaGold' },
    { color: '#238636', label: 'xpack.theme.freshGreen' },
];

const defaultLightColors = [
    { color: '#005eeb', label: 'xpack.theme.classicBlue' },
    { color: '#238636', label: 'xpack.theme.freshGreen' },
];

let darkColors = [...defaultDarkColors];

let lightColors = [...defaultLightColors];

const addColorToTheme = (
    colors: { color: string; label: string }[],
    newColor: { color: string; label: string },
    baseCount = 3,
    maxNewColors = 2,
): { color: string; label: string }[] => {
    const updatedColors = [...colors];
    const extraCount = Math.max(0, colors.length - baseCount);
    if (extraCount >= maxNewColors) {
        updatedColors.splice(baseCount, 1);
    }
    updatedColors.push(newColor);
    return updatedColors;
};

const defaultColors = {
    light: lightPredefineColors.value,
    dark: darkPredefineColors.value,
};

const initThemeColors = () => {
    try {
        const storedColors = localStorage.getItem(STORAGE_KEY);
        themeColors.value = storedColors ? JSON.parse(storedColors) : { ...defaultColors };
    } catch (error) {
        console.error('Failed to parse theme colors from localStorage:', error);
        themeColors.value = { ...defaultColors };
    } finally {
        localStorage.setItem(STORAGE_KEY, JSON.stringify(themeColors.value));
    }
};

const themeColors = ref({ ...defaultColors });

const acceptParams = (params: DialogProps): void => {
    form.themeColor = params.themeColor;
    form.theme = params.theme;
    form.dark = form.themeColor.dark;
    form.light = form.themeColor.light;
    if (form.themeColor.themePredefineColors) {
        localStorage.setItem(STORAGE_KEY, JSON.stringify(form.themeColor.themePredefineColors));
    }
    initThemeColors();
    lightPredefineColors.value = themeColors.value['light'];
    darkPredefineColors.value = themeColors.value['dark'];
    lightColors = defaultLightColors;
    lightPredefineColors.value.slice(0, 2).forEach((color) => {
        const exists = lightColors.some((item) => item.color === color);
        if (!exists) {
            lightColors.push({
                color,
                label: `xpack.theme.customColor`,
            });
        }
    });
    darkColors = defaultDarkColors;
    darkPredefineColors.value.slice(0, 2).forEach((color) => {
        const exists = darkColors.some((item) => item.color === color);
        if (!exists) {
            darkColors.push({
                color,
                label: `xpack.theme.customColor`,
            });
        }
    });
    drawerVisible.value = true;
};

const updateThemeColors = (theme: 'light' | 'dark', newColors: string[]) => {
    themeColors.value[theme] = newColors;
    localStorage.setItem(STORAGE_KEY, JSON.stringify(themeColors.value));
    lightPredefineColors.value = themeColors.value['light'];
    darkPredefineColors.value = themeColors.value['dark'];
};

const addAndRemoveColor = (theme: 'light' | 'dark', newColor: string) => {
    const colors = [...themeColors.value[theme]];
    colors.unshift(newColor);
    colors.pop();
    updateThemeColors(theme, colors);
};

const changeThemeColor = (theme: 'light' | 'dark', color: string) => {
    if (theme === 'light') {
        form.light = color;
        const newLightColor = { color: color, label: 'xpack.theme.customColor' };
        lightColors = addColorToTheme(lightColors, newLightColor, 2);
    } else {
        form.dark = color;
        const newDarkColor = { color: color, label: 'xpack.theme.customColor' };
        darkColors = addColorToTheme(darkColors, newDarkColor);
    }
    addAndRemoveColor(theme, color);
};

const changeLightColor = (color: string) => {
    form.light = color;
};

const changeDarkColor = (color: string) => {
    form.dark = color;
};

const onSave = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    ElMessageBox.confirm(i18n.global.t('xpack.theme.setHelper'), i18n.global.t('commons.button.save'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        await formEl.validate(async (valid) => {
            if (!valid) return;
            form.themeColor = { light: form.light, dark: form.dark, themePredefineColors: themeColors.value };
            if (globalStore.isProductPro) {
                await updateXpackSettingByKey('ThemeColor', JSON.stringify(form.themeColor));
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                globalStore.themeConfig.themeColor = JSON.stringify(form.themeColor);
                loading.value = false;
                let color: string;
                if (form.theme === 'auto') {
                    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)');
                    color = prefersDark.matches ? form.dark : form.light;
                } else {
                    color = form.theme === 'dark' ? form.dark : form.light;
                }
                globalStore.themeConfig.primary = color;
                setPrimaryColor(color);
                initFavicon();
                drawerVisible.value = false;
                emit('search');
            }
        });
    });
};

const onReSet = async () => {
    ElMessageBox.confirm(i18n.global.t('xpack.theme.setDefaultHelper'), i18n.global.t('xpack.theme.setDefault'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        form.themeColor = { light: '#005eeb', dark: '#F0BE96', themePredefineColors: themeColors.value };
        if (globalStore.isProductPro) {
            await updateXpackSettingByKey('ThemeColor', JSON.stringify(form.themeColor));
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            loading.value = false;
            globalStore.themeConfig.themeColor = JSON.stringify(form.themeColor);
            let color: string;
            if (form.theme === 'auto') {
                const prefersDark = window.matchMedia('(prefers-color-scheme: dark)');
                color = prefersDark.matches ? '#F0BE96' : '#005eeb';
            } else {
                color = form.theme === 'dark' ? '#F0BE96' : '#005eeb';
            }
            globalStore.themeConfig.primary = color;
            setPrimaryColor(color);
            initFavicon();
            drawerVisible.value = false;
        }
    });
};

const handleClose = () => {
    drawerVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>
<style lang="scss" scoped>
.selected-white {
    box-shadow: inset 0 0 0 1px white;
}
</style>
