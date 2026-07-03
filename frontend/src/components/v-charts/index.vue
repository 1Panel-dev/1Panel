<template>
    <component
        :is="typeComponentMap[type]"
        :height="height"
        :option="option"
        :dataZoom="dataZoom"
        :renderer="renderer"
        class="v-charts"
    />
</template>
<script lang="ts" setup>
import line from './components/Line.vue';
import pie from './components/Pie.vue';
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
    dataZoom: Boolean,
    renderer: {
        type: String,
        default: undefined,
    },
    option: {
        type: Object,
        required: true,
    },
});

const typeComponentMap = { line, pie };
</script>
