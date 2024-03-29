<script setup lang="ts">
import { MenuStore } from '@/store';
const menuStore = MenuStore();
</script>
<template>
    <div class="mobile-header">
        <svg-icon class="caidan" iconName="p-caidan" @click="menuStore.setCollapse()"></svg-icon>
    </div>
</template>

<style lang="scss" scoped>
.mobile-header {
    background: rgba(255, 255, 255, 0.65);
    box-shadow: 0 2px 10px 0 rgba(0, 0, 0, 0.2);
    padding: 10px;
    text-align: left;
    z-index: 999;
    .caidan {
        margin-left: 10px;
        font-size: 10px;
        cursor: pointer;
    }
}
</style>
