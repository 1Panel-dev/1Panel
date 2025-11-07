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
        label: i18n.global.t('firewall.portRule', 2),
        path: '/hosts/firewall/port',
    },
    {
        label: i18n.global.t('firewall.forwardRule', 2),
        path: '/hosts/firewall/forward',
    },
    {
        label: i18n.global.t('firewall.ipRule', 2),
        path: '/hosts/firewall/ip',
    },
];
</script>
