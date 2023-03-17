<template>
    <span class="bread-crumbs-item" ref="item">
        <el-link><slot></slot></el-link>
        <i v-if="!props.right" :class="'panel p-arrow-right'"></i>
    </span>
</template>

<script setup lang="ts">
import { ref } from 'vue';

const props = defineProps({
    right: Boolean,
});

const item = ref();
defineExpose({ item });
</script>

<style lang="scss" setup>
.bread-crumbs-item {
    margin-top: 3px;
}
</style>
