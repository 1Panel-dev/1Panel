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
.logo {
    box-sizing: border-box;
    height: 55px;
    margin: 5px 0;
    img:first-child {
        width: 130px;
        object-fit: contain;
    }
    img:nth-child(2) {
        height: 45px;
        object-fit: contain;
    }
}
</style>
