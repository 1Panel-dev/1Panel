<template>
    <div class="logo flx-center">
        <img :src="getLogoUrl(isCollapse)" alt="" />
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
    box-sizing: border-box;
    height: 55px;
    margin: 5px 0;
    padding: 0 15px;
    img {
        object-fit: contain;
        height: 40px;
    }
}
</style>
