<template>
    <component :is="typeComponentMap[type]" :height="height" :option="option" />
</template>
<script lang="ts" setup>
import line from './components/Line.vue';
defineOptions({ name: 'VCharts' });
defineProps({
    type: {
        type: String,
        default: 'line',
    },
    height: {
        type: String,
        default: '200px',
    },
    option: {
        type: Object,
        required: true,
    }, // { title , xDatas, yDatas, formatStr  }
});

const typeComponentMap = { line };
</script>
