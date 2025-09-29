<template>
    <el-config-provider :locale="i18nLocale" :button="config" size="default">
        <router-view v-if="isRouterAlive" />
    </el-config-provider>
</template>

<script setup lang="ts">
import { reactive, computed, ref, nextTick, provide } from 'vue';
import { GlobalStore } from '@/store';
import zhCn from 'element-plus/es/locale/lang/zh-cn';
import zhTw from 'element-plus/es/locale/lang/zh-tw';
import en from 'element-plus/es/locale/lang/en';
import ja from 'element-plus/es/locale/lang/ja';
import ms from 'element-plus/es/locale/lang/ms';
import ptBR from 'element-plus/es/locale/lang/pt-br';
import ru from 'element-plus/es/locale/lang/ru';
import ko from 'element-plus/es/locale/lang/ko';
import tr from 'element-plus/es/locale/lang/tr';
import esES from 'element-plus/es/locale/lang/es';
import { useTheme } from '@/global/use-theme';
useTheme();

const globalStore = GlobalStore();
const config = reactive({
    autoInsertSpace: false,
});

const i18nLocale = computed(() => {
    if (globalStore.language === 'zh') return zhCn;
    if (globalStore.language === 'zh-Hant') return zhTw;
    if (globalStore.language === 'en') return en;
    if (globalStore.language === 'ja') return ja;
    if (globalStore.language === 'ms') return ms;
    if (globalStore.language === 'ru') return ru;
    if (globalStore.language === 'pt-br') return ptBR;
    if (globalStore.language === 'ko') return ko;
    if (globalStore.language === 'tr') return tr;
    if (globalStore.language === 'es-ES') return esES;
    return zhCn;
});

const isRouterAlive = ref(true);

const reload = () => {
    isRouterAlive.value = false;
    nextTick(() => {
        isRouterAlive.value = true;
    });
};
provide('reload', reload);
</script>

<style scoped lang="scss"></style>
