<template>
    <div class="main-div" :style="{ '--main-height': mainHeight + 'px' }">
        <slot></slot>
    </div>
</template>
<script lang="ts" setup>
import { ref, onMounted, onUnmounted, computed } from 'vue';
const props = defineProps({
    heightDiff: {
        type: Number,
        default: 0,
    },
});
const windowHeight = ref(window.innerHeight);
const mainHeight = computed(() => windowHeight.value - props.heightDiff);

const updateHeight = () => {
    windowHeight.value = window.innerHeight;
};

onMounted(() => {
    window.addEventListener('resize', updateHeight);
});

onUnmounted(() => {
    window.removeEventListener('resize', updateHeight);
});
defineOptions({ name: 'MainDiv' });
</script>
<style scoped>
.main-div {
    height: var(--main-height);
    overflow-y: auto;
    overflow-x: hidden;
}
</style>
