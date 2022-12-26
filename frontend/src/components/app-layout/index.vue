<template>
    <Layout>
        <template #menu>
            <Menu :panelName="themeConfig.panelName"></Menu>
        </template>
        <template #footer>
            <Footer></Footer>
        </template>
    </Layout>
</template>
<script setup lang="ts">
import Layout from '@/layout/index.vue';
import Footer from './footer/index.vue';
import Menu from './menu/index.vue';
import { onMounted, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { GlobalStore } from '@/store';
import { useTheme } from '@/hooks/use-theme';
import { getSettingInfo } from '@/api/modules/setting';

const i18n = useI18n();
const globalStore = GlobalStore();
const themeConfig = computed(() => globalStore.themeConfig);
const { switchDark } = useTheme();

const loadDataFromDB = async () => {
    const res = await getSettingInfo();
    i18n.locale.value = res.data.language;
    i18n.warnHtmlMessage = false;
    globalStore.updateLanguage(res.data.language);
    globalStore.setThemeConfig({ ...themeConfig.value, theme: res.data.theme });
    globalStore.setThemeConfig({ ...themeConfig.value, panelName: res.data.panelName });
    switchDark();
};
onMounted(() => {
    loadDataFromDB();
});
</script>
