<template>
    <el-tabs tab-position="left" type="border-card" v-model="index">
        <el-tab-pane :label="'CC 防护'" name="cc">
            <CCDeny :id="id" v-if="index == 'cc'"></CCDeny>
        </el-tab-pane>
        <el-tab-pane :label="'IP白名单'" name="ipWhiteList">
            <IpWhiteList :id="id" v-if="index == 'ipWhiteList'"></IpWhiteList>
        </el-tab-pane>
        <el-tab-pane :label="'IP 黑名单'" name="ipBlockList">
            <IpBlockList :id="id" v-if="index == 'ipBlockList'"></IpBlockList>
        </el-tab-pane>
        <el-tab-pane :label="'并发限制'"></el-tab-pane>
        <el-tab-pane :label="'请求限制'"></el-tab-pane>
        <el-tab-pane :label="'禁止爬虫'"></el-tab-pane>
    </el-tabs>
</template>
<script lang="ts" setup>
import { computed, ref } from 'vue';
import CCDeny from './ccdeny/index.vue';
import IpWhiteList from './ip-white-list/index.vue';
import IpBlockList from './ip-block-list/index.vue';

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
