<template>
    <div class="header">
        <div class="header-lf flx-center">
            <CollapseIcon id="collapseIcon"></CollapseIcon>
            <Breadcrumb id="breadcrumb" v-if="themeConfig.breadcrumb"></Breadcrumb>
        </div>
    </div>
</template>
<script setup lang="ts">
import { computed } from 'vue';
import CollapseIcon from './components/collapseicon.vue';
import Breadcrumb from './components/breadcrumb.vue';
import { GlobalStore } from '@/store';

const globalStore = GlobalStore();
const themeConfig = computed(() => globalStore.themeConfig);
</script>

<style scoped lang="scss">
@import './index.scss';
</style>
