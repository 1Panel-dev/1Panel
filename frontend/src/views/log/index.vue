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
        label: i18n.global.t('logs.panelLog'),
        path: '/logs/operation',
    },
    {
        label: 'SSH ' + i18n.global.t('ssh.loginLogs'),
        path: '/logs/ssh',
    },
    {
        label: i18n.global.t('logs.websiteLog'),
        path: '/logs/website',
    },
];
</script>
