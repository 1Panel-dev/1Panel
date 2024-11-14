<template>
    <div class="logo" style="cursor: pointer" @click="goHome">
        <template v-if="isCollapse">
            <img
                v-if="globalStore.themeConfig.logo"
                :src="`/api/v1/images/logo?t=${Date.now()}`"
                style="cursor: pointer"
                alt="logo"
            />
            <MenuLogo v-else />
        </template>
        <template v-else>
            <img
                v-if="globalStore.themeConfig.logoWithText"
                :src="`/api/v1/images/logoWithText?t=${Date.now()}`"
                style="cursor: pointer"
                alt="logo"
            />
            <PrimaryLogo v-else />
        </template>
    </div>
</template>

<script setup lang="ts">
import router from '@/routers';
import { GlobalStore } from '@/store';
import PrimaryLogo from '@/assets/images/1panel-logo.svg?component';
import MenuLogo from '@/assets/images/1panel-menu-logo.svg?component';

defineProps<{ isCollapse: boolean }>();

const globalStore = GlobalStore();

const goHome = () => {
    router.push({ name: 'home' });
};
</script>

<style scoped lang="scss">
.logo {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 55px;
    z-index: 1;
    img {
        object-fit: contain;
        width: 95%;
        height: 45px;
    }
}
</style>
