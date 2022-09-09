<template>
    <div>
        <el-tooltip effect="dark" :content="$t('commons.header.theme')" placement="bottom">
            <i :class="'panel p-theme'" class="icon-style" @click="openDrawer"></i>
        </el-tooltip>
        <el-drawer v-model="drawerVisible" :title="$t('commons.header.theme')" size="300px">
            <el-divider class="divider" content-position="center">
                <el-icon><ColdDrink /></el-icon>
                {{ $t('commons.header.globalTheme') }}
            </el-divider>
            <div class="theme-item">
                <span>{{ $t('commons.header.themeColor') }}</span>
                <el-color-picker v-model="themeConfig.primary" :predefine="colorList" @change="changePrimary">
                </el-color-picker>
            </div>
            <div class="theme-item">
                <span>{{ $t('commons.header.darkTheme') }}</span>
                <SwitchDark></SwitchDark>
            </div>
            <br />
        </el-drawer>
    </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import { useTheme } from '@/hooks/use-theme';
import SwitchDark from '@/components/switch-dark/index.vue';
import { GlobalStore } from '@/store';

// 预定义主题颜色
const colorList = [
    '#409EFF',
    '#DAA96E',
    '#0C819F',
    '#009688',
    '#27ae60',
    '#ff5c93',
    '#e74c3c',
    '#fd726d',
    '#f39c12',
    '#9b59b6',
];

// 主题初始化
const globalStore = GlobalStore();
const themeConfig = computed(() => globalStore.themeConfig);

const { changePrimary } = useTheme();

// 打开主题设置
const drawerVisible = ref(false);
const openDrawer = () => {
    drawerVisible.value = true;
};
</script>

<style scoped lang="scss">
@import '../index.scss';
</style>
