<template>
    <template v-for="subItem in menuList" :key="subItem.name">
        <el-sub-menu v-if="subItem?.children?.length > 1" :index="subItem.path" popper-class="sidebar-container-popper">
            <template #title>
                <el-icon>
                    <SvgIcon :iconName="(subItem.meta?.icon as string)" />
                </el-icon>
                <span>{{ $t(subItem.meta?.title as string, 2) }}</span>
            </template>
            <SubItem :menuList="subItem.children" :level="level + 1" />
        </el-sub-menu>

        <el-menu-item v-else-if="subItem?.children?.length === 1" :index="subItem.children[0].path">
            <el-icon>
                <SvgIcon :iconName="(subItem.meta?.icon as string)" />
            </el-icon>
            <template #title>
                <span>{{ $t(subItem.meta?.title as string, 2) }}</span>
            </template>
        </el-menu-item>

        <el-menu-item v-else-if="subItem.path === '/xpack/upage'" :index="''" @click="goUpage">
            <el-icon v-if="subItem.meta?.icon && level === 0">
                <SvgIcon :iconName="(subItem.meta?.icon as string)" />
            </el-icon>
            <template #title>
                <span v-if="subItem.meta?.icon && level === 0">{{ $t(subItem.meta?.title as string, 2) }}</span>
                <span v-else style="margin-left: 10px">{{ $t(subItem.meta?.title as string, 2) }}</span>
            </template>
        </el-menu-item>

        <el-menu-item v-else :index="subItem.path">
            <el-icon v-if="subItem.meta?.icon && level === 0">
                <SvgIcon :iconName="(subItem.meta?.icon as string)" />
            </el-icon>
            <template #title>
                <span v-if="subItem.meta?.icon && level === 0">{{ $t(subItem.meta?.title as string, 2) }}</span>
                <span v-else style="margin-left: 10px">{{ $t(subItem.meta?.title as string, 2) }}</span>
            </template>
        </el-menu-item>
    </template>
</template>

<script setup lang="ts">
import { RouteRecordRaw } from 'vue-router';
import SvgIcon from '@/components/svg-icon/svg-icon.vue';

defineProps<{ menuList: RouteRecordRaw[]; level?: number }>();

const goUpage = () => {
    window.open('https://www.lxware.cn/upage', '_blank', 'noopener,noreferrer');
};
</script>

<style scoped lang="scss">
@use '../index';
</style>
