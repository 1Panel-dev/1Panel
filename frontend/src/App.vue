<template>
    <el-config-provider :locale="i18nLocale" :button="config" size="default">
        <router-view v-if="isRouterAlive" />
    </el-config-provider>
</template>

<script setup lang="ts">
import { reactive, computed, ref, nextTick, provide } from 'vue';
import { useGlobalStore } from '@/composables/useGlobalStore';
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
import fa from 'element-plus/es/locale/lang/fa';
import { useTheme } from '@/global/use-theme';
useTheme();

const { language } = useGlobalStore();
const config = reactive({
    autoInsertSpace: false,
});

const i18nLocale = computed(() => {
    if (language.value === 'zh') return zhCn;
    if (language.value === 'zh-Hant') return zhTw;
    if (language.value === 'en') return en;
    if (language.value === 'ja') return ja;
    if (language.value === 'ms') return ms;
    if (language.value === 'ru') return ru;
    if (language.value === 'pt-BR') return ptBR;
    if (language.value === 'ko') return ko;
    if (language.value === 'tr') return tr;
    if (language.value === 'es-ES') return esES;
    if (language.value === 'fa') return fa;
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
