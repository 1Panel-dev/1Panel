<template>
    <el-switch
        v-model="themeConfig.isDark"
        @change="onAddDarkChange"
        inline-prompt
        active-color="#0a0a0a"
        inactive-color="#dcdfe6"
        :active-icon="Sunny"
        :inactive-icon="Moon"
    />
</template>

<script setup lang="ts" name="switchDark">
import { computed } from 'vue';
import { GlobalStore } from '@/store';
import { Sunny, Moon } from '@element-plus/icons-vue';
import { useTheme } from '@/hooks/useTheme';
const globalStore = GlobalStore();

const { switchDark } = useTheme();

const themeConfig = computed(() => globalStore.themeConfig);

const onAddDarkChange = () => {
    switchDark();
};
</script>
