<template>
    <div class="logo">
        <img :src="getLogoUrl(isCollapse)" alt="logo" />
    </div>
</template>

<script setup lang="ts">
import { GlobalStore } from '@/store';

defineProps<{ isCollapse: boolean }>();
const globalStore = GlobalStore();

const getLogoUrl = (isCollapse: boolean) => {
    const theme = globalStore.$state.themeConfig.theme || 'light';
    if (isCollapse) {
        return new URL(`../../../../assets/images/1panel-logo-${theme}.png`, import.meta.url).href;
    } else {
        return new URL(`../../../../assets/images/1panel-menu-${theme}.png`, import.meta.url).href;
    }
};
</script>

<style scoped lang="scss">
.logo {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 55px;
    img {
        object-fit: contain;
        height: 40px;
    }
}
</style>
