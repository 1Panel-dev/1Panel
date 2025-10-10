<template>
    <LogFile :config="logConfig"></LogFile>
</template>

<script lang="ts" setup>
import LogFile from '@/components/log/file/index.vue';

const props = defineProps({
    id: Number,
});

const logConfig = reactive({
    id: props.id,
    type: 'php-fpm-slow-logs',
});
</script>
