<template>
    <div class="header">
        <div class="header-lf flx-center">
            <CollapseIcon id="collapseIcon"></CollapseIcon>
            <Breadcrumb
                id="breadcrumb"
                v-if="themeConfig.breadcrumb"
            ></Breadcrumb>
        </div>
        <div class="header-ri flx-center">
            <div class="header-icon">
                <!-- Language -->
                <Language id="language"></Language>
                <!-- Theme -->
                <Theme id="theme"></Theme>
            </div>
            <!-- User name -->
            <span class="username">1Panel</span>
            <!-- Avatar -->
            <Avatar></Avatar>
        </div>
    </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import CollapseIcon from './components/CollapseIcon.vue';
import Breadcrumb from './components/Breadcrumb.vue';
import Language from './components/Language.vue';
import Theme from './components/Theme.vue';
import Avatar from './components/Avatar.vue';
import { GlobalStore } from '@/store';

const globalStore = GlobalStore();
const themeConfig = computed(() => globalStore.themeConfig);
</script>

<style scoped lang="scss">
@import './index.scss';
</style>
