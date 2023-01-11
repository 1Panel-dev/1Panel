<template>
    <div>
        <RouterButton :buttons="buttons" />
        <br />
        <LayoutContent>
            <router-view></router-view>
        </LayoutContent>
    </div>
</template>

<script lang="ts" setup>
import LayoutContent from '@/layout/layout-content.vue';
import RouterButton from '@/components/router-button/index.vue';
import i18n from '@/lang';

const buttons = [
    {
        label: i18n.global.t('app.all'),
        path: '/apps/all',
    },
    {
        label: i18n.global.t('app.installed'),
        path: '/apps/installed',
    },
];
</script>
