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
import LayoutContent from '@/layout/layout-content.vue';
import RouterButton from '@/components/router-button/index.vue';

const buttons = [
    {
        label: i18n.global.t('monitor.monitor'),
        path: '/hosts/monitor/monitor',
    },
    {
        label: i18n.global.t('monitor.setting'),
        path: '/hosts/monitor/setting',
    },
];
</script>
