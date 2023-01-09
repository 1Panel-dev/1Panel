<template>
    <el-page-header :content="header" @back="props.back">
        <template v-if="slots.buttons" #content>
            <span>{{ header }}</span>
            <el-divider direction="vertical" />
            <slot name="buttons"></slot>
        </template>
    </el-page-header>
</template>

<script lang="ts" setup>
import { useSlots } from 'vue';
const slots = useSlots();

const props = defineProps({
    header: String,
    back: Function,
});
</script>
