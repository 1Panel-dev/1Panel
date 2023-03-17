<template>
    <div>
        <RouterButton :buttons="buttons" />
        <LayoutContent>
            <router-view></router-view>
        </LayoutContent>
    </div>
</template>

<script lang="ts" setup>
import LayoutContent from '@/layout/layout-content.vue';
import RouterButton from '@/components/router-button/index.vue';

const buttons = [
    {
        label: 'MySQL',
        path: '/databases/mysql',
    },
    {
        label: 'Redis',
        path: '/databases/redis',
    },
];
</script>
