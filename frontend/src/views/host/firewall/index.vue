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
        label: i18n.global.t('firewall.portRule'),
        path: '/hosts/firewall/port',
    },
    {
        label: i18n.global.t('firewall.ipRule'),
        path: '/hosts/firewall/ip',
    },
];
</script>
