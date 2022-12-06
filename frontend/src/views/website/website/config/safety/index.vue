<template>
    <el-tabs tab-position="left" type="border-card" v-model="index">
        <el-tab-pane :label="'CC 防护'" name="cc">
            <CCDeny :id="id" v-if="index == 'cc'"></CCDeny>
        </el-tab-pane>
        <el-tab-pane :label="'白名单'"></el-tab-pane>
        <el-tab-pane :label="'黑名单'"></el-tab-pane>
        <el-tab-pane :label="'并发限制'"></el-tab-pane>
        <el-tab-pane :label="'请求限制'"></el-tab-pane>
        <el-tab-pane :label="'禁止爬虫'"></el-tab-pane>
    </el-tabs>
</template>
<script lang="ts" setup>
import { computed, ref } from 'vue';
import CCDeny from './ccdeny/index.vue';

const props = defineProps({
    id: {
        type: Number,
        default: 0,
    },
});

const id = computed(() => {
    return props.id;
});

let index = ref('cc');
</script>
