<template>
    <div class="logo" style="cursor: pointer" @click="goHome">
        <template v-if="isCollapse">
            <img
                v-if="globalStore.themeConfig.logo && !logoLoadFailed"
                :src="`/api/v2/images/logo?t=${Date.now()}`"
                style="cursor: pointer"
                alt="logo"
                @error="logoLoadFailed = true"
            />
            <MenuLogo v-else />
        </template>
        <template v-else>
            <img
                v-if="globalStore.themeConfig.logoWithText && !logoWithTextLoadFailed"
                :src="`/api/v2/images/logoWithText?t=${Date.now()}`"
                style="cursor: pointer"
                alt="logo"
                @error="logoWithTextLoadFailed = true"
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
import { ref } from 'vue';

defineProps<{ isCollapse: boolean }>();

const logoLoadFailed = ref(false);
const logoWithTextLoadFailed = ref(false);
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
    height: 49px;
    z-index: 1;
    img {
        object-fit: contain;
        width: 95%;
        height: 45px;
    }
}
</style>
