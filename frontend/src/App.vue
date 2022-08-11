<template>
    <el-config-provider :locale="i18nLocale" :button="config" :size="assemblySize">
        <router-view></router-view>
    </el-config-provider>
</template>

<script setup lang="ts">
import { reactive, computed } from 'vue';
import { GlobalStore } from '@/store';
// 配置element中英文
import zhCn from 'element-plus/es/locale/lang/zh-cn';
import en from 'element-plus/es/locale/lang/en';
// 使用主题
import { useTheme } from '@/hooks/useTheme';
useTheme();

const globalStore = GlobalStore();
// 配置element按钮文字中间是否有空格
const config = reactive({
    autoInsertSpace: false,
});

// element 语言配置
const i18nLocale = computed((): any => {
    if (globalStore.language && globalStore.language == 'zh') return zhCn;
    if (globalStore.language == 'en') return en;
    return '';
});

// 配置全局组件大小 (small/default(medium)/large)
const assemblySize = computed((): string => globalStore.assemblySize);
</script>

<style scoped lang="scss"></style>
