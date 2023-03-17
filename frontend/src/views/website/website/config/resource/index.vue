<template>
    <Nginx :id="id"></Nginx>
</template>

<script lang="ts" setup>
import { computed } from 'vue';
import Nginx from './nginx/index.vue';

const props = defineProps({
    id: {
        type: Number,
        default: 0,
    },
});

const id = computed(() => {
    return props.id;
});
</script>
