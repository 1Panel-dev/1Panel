<template>
    <div>
        <RouterButton :buttons="buttons" />
        <LayoutContent>
            <router-view></router-view>
        </LayoutContent>
    </div>
</template>

<script lang="ts" setup>
import i18n from '@/lang';

const buttons = [
    {
        label: i18n.global.t('container.container'),
        path: '/containers/container',
    },
    {
        label: i18n.global.t('container.compose'),
        path: '/containers/compose',
    },
    {
        label: i18n.global.t('container.image'),
        path: '/containers/image',
    },
    {
        label: i18n.global.t('container.network'),
        path: '/containers/network',
    },
    {
        label: i18n.global.t('container.volume'),
        path: '/containers/volume',
    },
    {
        label: i18n.global.t('container.repo'),
        path: '/containers/repo',
    },
    {
        label: i18n.global.t('container.composeTemplate'),
        path: '/containers/template',
    },
    {
        label: i18n.global.t('container.setting'),
        path: '/containers/setting',
    },
];
</script>
