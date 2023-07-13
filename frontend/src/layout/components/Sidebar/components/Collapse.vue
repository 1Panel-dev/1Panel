<template>
    <div class="menu-collapse">
        <el-icon class="collapse-icon" :size="25" @click="menuStore.setCollapse()">
            <component :is="isCollapse ? 'expand' : 'fold'"></component>
        </el-icon>
    </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { MenuStore } from '@/store/modules/menu';
const menuStore = MenuStore();
const isCollapse = computed((): boolean => menuStore.isCollapse);
</script>

<style scoped lang="scss">
.menu-collapse {
    height: 48px;
    border-top: 1px solid #e4e7ed;
    box-sizing: border-box;
}

.collapse-icon {
    margin-top: 10px;
    margin-left: 25px;
    &:hover {
        color: $primary-color;
        cursor: pointer;
    }
}
</style>
