<template>
    <el-dropdown trigger="click" @command="handleSetLanguage">
        <span>
            <el-tooltip effect="dark" :content="$t('commons.header.language')" placement="bottom">
                <i :class="'panel p-language'" class="icon-style"></i>
            </el-tooltip>
        </span>
        <template #dropdown>
            <el-dropdown-menu>
                <el-dropdown-item :disabled="language && language === 'zh'" command="zh">{{
                    $t('commons.header.zh')
                }}</el-dropdown-item>
                <el-dropdown-item :disabled="language === 'en'" command="en">{{
                    $t('commons.header.en')
                }}</el-dropdown-item>
            </el-dropdown-menu>
        </template>
    </el-dropdown>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { GlobalStore } from '@/store';
import { getBrowserLang } from '@/utils/util';

const i18n = useI18n();
const globalStore = GlobalStore();
const language = computed((): string => globalStore.language);

// 切换语言
const handleSetLanguage = (lang: string) => {
    i18n.locale.value = lang;
    globalStore.updateLanguage(lang);
};

onMounted(() => {
    handleSetLanguage(language.value || getBrowserLang());
});
</script>

<style scoped lang="scss">
@import '../index.scss';
</style>
