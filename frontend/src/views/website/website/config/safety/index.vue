<template>
    <el-tabs tab-position="left" type="border-card" v-model="index">
        <el-tab-pane :label="'CC 防护'" name="cc">
            <CCDeny :id="id" v-if="index == 'cc'"></CCDeny>
        </el-tab-pane>
        <el-tab-pane :label="'IP 白名单'" name="ipWhiteList">
            <IpList :id="id" :rule="'ipWhitelist'" v-if="index == 'ipWhiteList'"></IpList>
        </el-tab-pane>
        <el-tab-pane :label="'IP 黑名单'" name="ipBlockList">
            <IpList :id="id" :rule="'ipBlockList'" v-if="index == 'ipBlockList'"></IpList>
        </el-tab-pane>
        <el-tab-pane :label="'文件扩展名黑名单'" name="FileBlockList">
            <FileBlockList :id="id" v-if="index == 'FileBlockList'"></FileBlockList>
        </el-tab-pane>
        <el-tab-pane :label="'URL黑名单'" name="urlBlockList">
            <UrlList :id="id" :rule="'url'" v-if="index == 'urlBlockList'"></UrlList>
        </el-tab-pane>
        <el-tab-pane :label="'URL白名单'" name="urlWhiteList">
            <UrlList :id="id" :rule="'whiteurl'" v-if="index == 'urlWhiteList'"></UrlList>
        </el-tab-pane>
        <el-tab-pane :label="'禁止爬虫'"></el-tab-pane>
    </el-tabs>
</template>
<script lang="ts" setup>
import { computed, ref } from 'vue';
import CCDeny from './ccdeny/index.vue';
import IpList from './ip-list/index.vue';
import FileBlockList from './file-block-list/index.vue';
import UrlList from './url-list/index.vue';

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
