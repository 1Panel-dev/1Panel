<template>
    <svg :class="svgClass" aria-hidden="true">
        <use :xlink:href="iconClassName" :fill="color" />
    </svg>
</template>
<script setup lang="ts">
import { computed } from 'vue';
const props = defineProps({
    iconName: {
        type: String,
        required: true,
    },
    className: {
        type: String,
        default: '',
    },
    color: {
        type: String,
        default: '#409eff',
    },
});
// 图标在 iconfont 中的名字
const iconClassName = computed(() => {
    return `#${props.iconName}`;
});
// 给图标添加上类名
const svgClass = computed(() => {
    if (props.className) {
        return `svg-icon ${props.className}`;
    }
    return 'svg-icon';
});
</script>
<style scoped>
.svg-icon {
    width: 1.5em;
    height: 1.5em;
    position: relative;
    fill: currentColor;
    vertical-align: -2px;
    padding-top: 0.3em;
    padding-bottom: 0.3em;
    padding-left: 0.3em;
    padding-right: 0.3em;
}
</style>
