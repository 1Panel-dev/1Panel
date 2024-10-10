<template>
    <el-button class="fresh-button" icon="Refresh" @click="onRefresh"></el-button>
</template>

<script setup lang="ts">
defineOptions({ name: 'TableRefresh' });

const emit = defineEmits(['search']);

const onRefresh = () => {
    emit('search');
};
</script>

<style lang="scss" scoped>
.fresh-button {
    float: right;
}
</style>
