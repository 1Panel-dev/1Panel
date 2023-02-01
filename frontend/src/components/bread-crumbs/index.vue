<template>
    <span class="bread-crumbs" ref="bread"><slot></slot></span>
</template>
<script lang="ts" setup>
import { ref } from 'vue';

const bread = ref();
defineExpose({ bread });
</script>
<style lang="scss">
.bread-crumbs {
    display: flex;
    &-item {
        a {
            text-decoration: none;
            transition: all 0.4s;
        }
    }
    i {
        font-size: 12px;
        margin-left: 5px;
        margin-right: 5px;
        line-height: 22px;
    }
}
</style>
