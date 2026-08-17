<template>
    <RouterButton :buttons="buttons" />
</template>

<script lang="ts" setup>
import i18n from '@/lang';

const buttons = [
    {
        label: i18n.global.t('firewall.systemFirewall'),
        path: '/hosts/firewall/rules',
    },
    {
        label: i18n.global.t('firewall.dockerGuard'),
        path: '/hosts/firewall/docker',
    },
    {
        label: i18n.global.t('firewall.forwardRule', 2),
        path: '/hosts/firewall/forward',
    },
    {
        label: i18n.global.t('commons.button.set'),
        path: '/hosts/firewall/setting',
    },
];
</script>
