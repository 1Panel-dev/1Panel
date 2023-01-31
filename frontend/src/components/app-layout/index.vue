<template>
    <Layout v-loading="loading" :element-loading-text="loadinText" fullscreen>
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
import { onMounted, computed, ref, watch, onBeforeUnmount } from 'vue';
import { useI18n } from 'vue-i18n';
import { GlobalStore } from '@/store';
import { useTheme } from '@/hooks/use-theme';
import { getSettingInfo, getSystemAvailable } from '@/api/modules/setting';

const i18n = useI18n();
const loading = ref(false);
const loadinText = ref();
const globalStore = GlobalStore();
const themeConfig = computed(() => globalStore.themeConfig);
const { switchDark } = useTheme();

let timer: NodeJS.Timer | null = null;

watch(
    () => globalStore.isLoading,
    () => {
        if (globalStore.isLoading) {
            loadStatus();
        } else {
            loading.value = globalStore.isLoading;
        }
    },
);

const loadDataFromDB = async () => {
    const res = await getSettingInfo();
    document.title = res.data.panelName;
    i18n.locale.value = res.data.language;
    i18n.warnHtmlMessage = false;
    globalStore.updateLanguage(res.data.language);
    globalStore.setThemeConfig({ ...themeConfig.value, theme: res.data.theme });
    globalStore.setThemeConfig({ ...themeConfig.value, panelName: res.data.panelName });
    switchDark();
};

const loadStatus = async () => {
    loading.value = globalStore.isLoading;
    loadinText.value = globalStore.loadingText;
    if (loading.value) {
        timer = setInterval(async () => {
            await getSystemAvailable()
                .then((res) => {
                    if (res) {
                        clearInterval(Number(timer));
                        timer = null;
                    }
                })
                .catch(() => {
                    clearInterval(Number(timer));
                    timer = null;
                });
        }, 1000 * 5);
    }
};

onBeforeUnmount(() => {
    clearInterval(Number(timer));
    timer = null;
});
onMounted(() => {
    loadStatus();
    loadDataFromDB();
});
</script>
