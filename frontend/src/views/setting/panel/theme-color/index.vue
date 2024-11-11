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
                                    <el-tooltip :content="$t('xpack.theme.classicBlue')" placement="top">
                                        <el-button
                                            color="#005eeb"
                                            circle
                                            dark
                                            :class="form.light === '#005eeb' ? 'selected-white' : ''"
                                            @click="changeLightColor('#005eeb')"
                                        />
                                    </el-tooltip>
                                    <el-tooltip :content="$t('xpack.theme.freshGreen')" placement="top">
                                        <el-button
                                            color="#238636"
                                            :class="form.light === '#238636' ? 'selected-white' : ''"
                                            circle
                                            dark
                                            @click="changeLightColor('#238636')"
                                        />
                                    </el-tooltip>
                                    <el-color-picker
                                        v-model="form.light"
                                        class="ml-4"
                                        :predefine="predefineColors"
                                        show-alpha
                                    />
                                </div>
                            </div>
                        </el-form-item>

                        <el-form-item :label="$t('setting.dark')" prop="dark">
                            <div class="flex flex-wrap justify-between items-center sm:items-start">
                                <div class="flex flex-wrap justify-between items-center mt-4 sm:mt-0">
                                    <div class="flex gap-1">
                                        <el-tooltip :content="$t('xpack.theme.classicBlue')" placement="top">
                                            <el-button
                                                color="#3D8EFF"
                                                circle
                                                dark
                                                :class="form.dark === '#3D8EFF' ? 'selected-white' : ''"
                                                @click="changeDarkColor('#3D8EFF')"
                                            />
                                        </el-tooltip>
                                        <el-tooltip :content="$t('xpack.theme.lingXiaGold')" placement="top">
                                            <el-button
                                                color="#F0BE96"
                                                :class="form.dark === '#F0BE96' ? 'selected-white' : ''"
                                                circle
                                                dark
                                                @click="changeDarkColor('#F0BE96')"
                                            />
                                        </el-tooltip>
                                        <el-tooltip :content="$t('xpack.theme.freshGreen')" placement="top">
                                            <el-button
                                                color="#238636"
                                                :class="form.dark === '#238636' ? 'selected-white' : ''"
                                                circle
                                                dark
                                                @click="changeDarkColor('#238636')"
                                            />
                                        </el-tooltip>
                                        <el-color-picker
                                            v-model="form.dark"
                                            class="ml-4"
                                            :predefine="predefineColors"
                                            show-alpha
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

const emit = defineEmits<{ (e: 'search'): void }>();
const drawerVisible = ref();
const loading = ref();

interface DialogProps {
    themeColor: { light: string; dark: string };
    theme: '';
}

interface ThemeColor {
    light: string;
    dark: string;
}

const form = reactive({
    themeColor: {} as ThemeColor,
    theme: '',
    light: '',
    dark: '',
});
const formRef = ref<FormInstance>();

const predefineColors = ref([
    '#005eeb',
    '#3D8EFF',
    '#F0BE96',
    '#238636',
    '#00ced1',
    '#c71585',
    '#ff4500',
    '#ff8c00',
    '#ffd700',
]);

const globalStore = GlobalStore();

const acceptParams = (params: DialogProps): void => {
    form.themeColor = params.themeColor;
    form.theme = params.theme;
    form.dark = form.themeColor.dark;
    form.light = form.themeColor.light;
    drawerVisible.value = true;
};

const changeLightColor = (color: string) => {
    form.light = color;
};

const changeDarkColor = (color: string) => {
    form.dark = color;
};

const onSave = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate(async (valid) => {
        if (!valid) return;
        form.themeColor = { light: form.light, dark: form.dark };
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
            return;
        }
    });
};

const onReSet = async () => {
    form.themeColor = { light: '#005eeb', dark: '#F0BE96' };
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
        return;
    }
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
