<template>
    <el-icon :size="25" class="collapse-icon" @click="menuStore.setCollapse()">
        <component :is="isCollapse ? 'expand' : 'fold'"></component>
    </el-icon>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { MenuStore } from '@/store/modules/menu';
const menuStore = MenuStore();
const isCollapse = computed((): boolean => menuStore.isCollapse);
</script>

<style scoped lang="scss">
// @import '../index.scss';

.collapse-icon {
    cursor: pointer;
}
</style>
