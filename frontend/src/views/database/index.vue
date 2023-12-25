<template>
    <div>
        <RouterButton :buttons="buttons" />
        <LayoutContent>
            <router-view></router-view>
        </LayoutContent>
    </div>
</template>

<script lang="ts" setup>
const buttons = [
    {
        label: 'MySQL',
        path: '/databases/mysql',
    },
    {
        label: 'PostgreSQL',
        path: '/databases/postgresql',
    },
    {
        label: 'Redis',
        path: '/databases/redis',
    },
];
</script>
