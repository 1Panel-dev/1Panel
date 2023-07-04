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
import RouterButton from '@/components/router-button/index.vue';

const buttons = [
    {
        label: i18n.global.t('menu.config'),
        path: '/hosts/ssh/ssh',
    },
    {
        label: i18n.global.t('ssh.session'),
        path: '/hosts/ssh/session',
    },
    {
        label: i18n.global.t('ssh.loginLogs'),
        path: '/hosts/ssh/log',
    },
];
</script>
