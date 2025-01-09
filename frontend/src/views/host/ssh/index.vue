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
        label: i18n.global.t('ssh.setting', 2),
        path: '/hosts/ssh/ssh',
    },
    {
        label: i18n.global.t('ssh.session', 2),
        path: '/hosts/ssh/session',
    },
    {
        label: i18n.global.t('ssh.loginLogs', 2),
        path: '/hosts/ssh/log',
    },
];
</script>
