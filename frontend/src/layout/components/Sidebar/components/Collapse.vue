<template>
    <div class="menu-collapse">
        <el-icon class="collapse-icon" :size="25" @click="menuStore.setCollapse()">
            <component :is="isCollapse ? 'expand' : 'fold'" />
        </el-icon>
    </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { MenuStore } from '@/store';
const menuStore = MenuStore();
const isCollapse = computed(() => menuStore.isCollapse);
</script>

<style scoped lang="scss">
.menu-collapse {
    display: flex;
    align-items: center;
    box-sizing: border-box;
    border-top: 1px solid var(--panel-footer-border);
    height: 48px;
}

.collapse-icon {
    color: var(--panel-main-bg-color-1);
    margin-left: 25px;
    &:hover {
        color: $primary-color;
        cursor: pointer;
    }
}
</style>
