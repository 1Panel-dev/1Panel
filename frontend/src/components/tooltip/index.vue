<template>
    <div class="tooltip-container">
        <el-tooltip :disabled="showTooltip">
            <template #content>
                <div :style="{ width: tootipWidth, 'word-break': 'break-all' }">{{ text }}</div>
            </template>
            <p ref="tooltipBox" class="text-box">
                <span v-if="islink" ref="tooltipItem" class="table-link">{{ text }}</span>
                <span v-else ref="tooltipItem" class="">{{ text }}</span>
            </p>
        </el-tooltip>
    </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';

defineOptions({ name: 'Tooltip' });

const showTooltip = ref();
const tooltipBox = ref();
const tooltipItem = ref();

const tootipWidth = ref();

defineProps({
    text: {
        type: String,
        default: '',
    },
    islink: {
        type: Boolean,
        default: true,
    },
});

onMounted(() => {
    const boxWidth = tooltipBox.value.offsetWidth;
    const itemWidth = tooltipItem.value.offsetWidth;
    tootipWidth.value = itemWidth > 250 ? '250px' : itemWidth + 'px';
    showTooltip.value = boxWidth > itemWidth;
});
</script>
<style scoped lang="scss">
.tooltip-container {
    width: 100%;
    .text-box {
        margin: 0;
        overflow: hidden;
        white-space: nowrap;
        text-overflow: ellipsis;
    }
}
</style>
