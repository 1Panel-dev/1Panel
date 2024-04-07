<template>
    <div class="logo">
        <img :src="getLogoUrl(isCollapse)" style="cursor: pointer" alt="logo" @click="goHome" />
    </div>
</template>

<script setup lang="ts">
import router from '@/routers';
import { GlobalStore } from '@/store';

defineProps<{ isCollapse: boolean }>();
const globalStore = GlobalStore();

const goHome = () => {
    router.push({ name: 'home' });
};

const getLogoUrl = (isCollapse: boolean) => {
    if (isCollapse) {
        if (globalStore.themeConfig.logo) {
            return '/api/v1/images/logo';
        } else {
            return new URL(`../../../../assets/images/1panel-logo-light.png`, import.meta.url).href;
        }
    } else {
        if (globalStore.themeConfig.logoWithText) {
            return '/api/v1/images/logoWithText';
        } else {
            return new URL(`../../../../assets/images/1panel-menu-light.png`, import.meta.url).href;
        }
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
        width: 95%;
        height: 45px;
    }
}
</style>
