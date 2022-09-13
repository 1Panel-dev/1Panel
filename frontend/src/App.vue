<template>
    <el-config-provider :locale="i18nLocale" :button="config" :size="assemblySize">
        <router-view></router-view>
    </el-config-provider>
</template>

<script setup lang="ts">
import { reactive, computed } from 'vue';
import { GlobalStore } from '@/store';
import zhCn from 'element-plus/es/locale/lang/zh-cn';
import en from 'element-plus/es/locale/lang/en';
import { useTheme } from '@/hooks/use-theme';
useTheme();

const globalStore = GlobalStore();
const config = reactive({
    autoInsertSpace: false,
});

const i18nLocale = computed((): any => {
    if (globalStore.language && globalStore.language == 'zh') return zhCn;
    if (globalStore.language == 'en') return en;
    return '';
});

const assemblySize = computed((): string => globalStore.assemblySize);
</script>

<style scoped lang="scss"></style>
