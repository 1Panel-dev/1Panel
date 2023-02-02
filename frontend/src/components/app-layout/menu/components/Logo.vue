<template>
    <div class="logo flx-center">
        <img v-show="!isCollapse" src="@/assets/images/1panel-menu.png" />
        <img v-show="isCollapse" src="@/assets/images/1panel-small.png" />
    </div>
</template>

<script setup lang="ts">
defineProps<{ isCollapse: boolean }>();
</script>

<style scoped lang="scss">
@import '../index.scss';
</style>
