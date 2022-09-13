<template>
    <div class="logo flx-center">
        <span v-show="!isCollapse">{{ panelName }}</span>
    </div>
</template>

<script setup lang="ts">
defineProps<{ isCollapse: boolean; panelName: string }>();
</script>

<style scoped lang="scss">
@import '../index.scss';
</style>
