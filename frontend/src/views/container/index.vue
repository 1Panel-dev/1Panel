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
        label: i18n.global.t('container.container', 2),
        path: '/containers/container',
    },
    {
        label: i18n.global.t('container.compose', 2),
        path: '/containers/compose',
    },
    {
        label: i18n.global.t('container.image', 2),
        path: '/containers/image',
    },
    {
        label: i18n.global.t('container.network', 2),
        path: '/containers/network',
    },
    {
        label: i18n.global.t('container.volume', 2),
        path: '/containers/volume',
    },
    {
        label: i18n.global.t('container.repo', 2),
        path: '/containers/repo',
    },
    {
        label: i18n.global.t('container.composeTemplate', 2),
        path: '/containers/template',
    },
    {
        label: i18n.global.t('container.setting', 2),
        path: '/containers/setting',
    },
];
</script>
