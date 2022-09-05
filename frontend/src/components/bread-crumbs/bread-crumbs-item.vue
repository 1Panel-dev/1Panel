<template>
    <div class="bread-crumbs-item">
        <el-link><slot></slot></el-link> <i v-if="!props.right" :class="'panel p-arrow-right'"></i>
    </div>
</template>

<script setup lang="ts">
const props = defineProps({
    right: Boolean,
});
</script>
