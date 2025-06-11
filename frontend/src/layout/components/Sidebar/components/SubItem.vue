<template>
    <template v-for="subItem in menuList" :key="subItem.path">
        <el-sub-menu v-if="subItem?.children?.length > 1" :index="subItem.path" popper-class="sidebar-container-popper">
            <template #title>
                <el-icon>
                    <SvgIcon :iconName="(subItem.meta?.icon as string)" />
                </el-icon>
                <span>{{ $t(subItem.meta?.title as string, 2) }}</span>
            </template>
            <SubItem :menuList="subItem.children" />
        </el-sub-menu>

        <el-menu-item v-else-if="subItem?.children?.length === 1" :index="subItem.children[0].path">
            <el-icon>
                <SvgIcon :iconName="(subItem.meta?.icon as string)" />
            </el-icon>
            <template #title>
                <span>{{ $t(subItem.meta?.title as string, 2) }}</span>
            </template>
        </el-menu-item>

        <el-menu-item v-else :index="subItem.path">
            <el-icon v-if="subItem.meta?.icon">
                <SvgIcon :iconName="(subItem.meta?.icon as string)" />
            </el-icon>
            <template #title>
                <span v-if="subItem.meta?.icon">{{ $t(subItem.meta?.title as string, 2) }}</span>
                <span v-else style="margin-left: 10px">{{ $t(subItem.meta?.title as string, 2) }}</span>
            </template>
        </el-menu-item>
    </template>
</template>

<script setup lang="ts">
import { RouteRecordRaw } from 'vue-router';
import SvgIcon from '@/components/svg-icon/svg-icon.vue';

defineProps<{ menuList: RouteRecordRaw[] }>();
</script>

<style scoped lang="scss">
@use '../index';
</style>
