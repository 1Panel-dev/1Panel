<template>
    <div class="collapse">
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
@import '../index.scss';

.collapse {
    height: 47px;
    border-top: 1px solid #e4e7ed;
}

.collapse-icon {
    margin-top: 10px;
    margin-left: 25px;
}

.collapse-icon:hover {
    color: $primary-color;
    cursor: pointer;
}
</style>
