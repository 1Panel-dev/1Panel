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
        label: i18n.global.t('setting.panel'),
        path: '/settings/panel',
    },
    {
        label: i18n.global.t('setting.safe'),
        path: '/settings/safe',
    },
    {
        label: i18n.global.t('setting.backupAccount'),
        path: '/settings/backupaccount',
    },
    {
        label: i18n.global.t('setting.snapshot'),
        path: '/settings/snapshot',
    },
    {
        label: i18n.global.t('setting.about'),
        path: '/settings/about',
    },
];
</script>
