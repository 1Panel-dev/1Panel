<template>
    <template v-for="subItem in menuList" :key="subItem.path">
        <el-sub-menu
            v-if="subItem.children && subItem.children.length > 1"
            :index="subItem.path"
            popper-class="sidebar-container-popper"
        >
            <template #title>
                <el-icon class="sub-icon">
                    <SvgIcon :iconName="(subItem.meta?.icon as string)" :className="'svg-icon'"></SvgIcon>
                </el-icon>
                <span class="sub-span">{{ $t(subItem.meta?.title as string) }}</span>
            </template>
            <SubItem :menuList="subItem.children" />
        </el-sub-menu>

        <el-menu-item v-else-if="subItem.children && subItem.children.length === 1" :index="subItem.children[0].path">
            <el-icon>
                <SvgIcon :iconName="(subItem.meta?.icon as string)" :className="'svg-icon'"></SvgIcon>
            </el-icon>
            <template #title>
                <span>{{ $t(subItem.meta?.title as string) }}</span>
            </template>
        </el-menu-item>

        <el-menu-item v-else :index="subItem.path">
            <el-icon v-if="subItem.meta?.icon">
                <SvgIcon :iconName="(subItem.meta?.icon as string)" :className="'svg-icon'"></SvgIcon>
            </el-icon>
            <template #title>
                <span v-if="subItem.meta?.icon">{{ $t(subItem.meta?.title as string) }}</span>
                <span v-else style="margin-left: 10px">{{ $t(subItem.meta?.title as string) }}</span>
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
@import '../index.scss';
</style>
