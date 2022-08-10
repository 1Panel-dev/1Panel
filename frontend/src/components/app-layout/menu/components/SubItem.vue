<template>
    <template v-for="subItem in menuList" :key="subItem.path">
        <el-sub-menu
            v-if="subItem.children && subItem.children.length > 1"
            :index="subItem.path"
        >
            <template #title>
                <el-icon>
                    <component :is="subItem.meta?.icon"></component>
                </el-icon>
                <span>{{ subItem.meta?.title }}</span>
            </template>
            <SubItem :menuList="subItem.children" />
        </el-sub-menu>
        <el-menu-item
            v-else-if="subItem.children && subItem.children.length === 1"
            :index="subItem.children[0].path"
        >
            <el-icon>
                <component :is="subItem.meta?.icon"></component>
            </el-icon>
            <template v-if="!subItem.meta?.isLink" #title>
                <span>{{ $t(subItem.meta?.title as string) }}</span>
            </template>
        </el-menu-item>
        <el-menu-item v-else :index="subItem.path">
            <el-icon>
                <component :is="subItem.meta?.icon"></component>
            </el-icon>
            <template v-if="!subItem.meta?.isLink" #title>
                <span>{{ subItem.meta?.title }}</span>
            </template>
            <template v-else #title>
                <a class="menu-href" :href="subItem.isLink" target="_blank">{{
                    subItem.meta?.title
                }}</a>
            </template>
        </el-menu-item>
    </template>
</template>

<script setup lang="ts">
import { RouteRecordRaw } from 'vue-router';

defineProps<{ menuList: RouteRecordRaw[] }>();
</script>

<style scoped lang="scss">
@import '../index.scss';
</style>
