<template>
    <component :is="typeComponentMap[type]" :height="height" :option="option" :dataZoom="dataZoom" class="v-charts" />
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
    option: {
        type: Object,
        required: true,
    }, // { title , xData, yData, formatStr  }
});

const typeComponentMap = { line, pie };
</script>
