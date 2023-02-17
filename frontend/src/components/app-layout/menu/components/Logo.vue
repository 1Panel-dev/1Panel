<template>
    <div class="logo flx-center">
        <img v-show="!isCollapse" :src="getLogoUrl(isCollapse)" alt="" />
        <img v-show="isCollapse" :src="getLogoUrl(isCollapse)" alt="" />
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
    img:first-child {
        width: 130px;
        object-fit: contain;
    }
    img:nth-child(2) {
        height: 45px;
        object-fit: contain;
    }
}
</style>
