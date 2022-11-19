<template>
    <el-tabs tab-position="left" type="border-card" v-model="index">
        <el-tab-pane :label="'Nginx'">
            <Nginx :id="id"></Nginx>
        </el-tab-pane>
    </el-tabs>
</template>

<script lang="ts" setup>
import { computed, ref } from 'vue';
import Nginx from './nginx/index.vue';

const props = defineProps({
    id: {
        type: Number,
        default: 0,
    },
});

const id = computed(() => {
    return props.id;
});

let index = ref('0');
</script>
