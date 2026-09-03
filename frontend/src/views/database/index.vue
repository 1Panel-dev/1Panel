<template>
    <div>
        <RouterButton :buttons="buttons" />
        <LayoutContent>
            <RouterViewCache />
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
    {
        label: 'MongoDB',
        path: '/databases/mongodb',
    },
];
</script>
