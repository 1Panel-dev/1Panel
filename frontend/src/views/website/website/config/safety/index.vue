<template>
    <el-tabs tab-position="left" type="border-card" v-model="index">
        <el-tab-pane :label="'CC 防护'" name="cc">
            <CCDeny :id="id" v-if="index == 'cc'"></CCDeny>
        </el-tab-pane>
        <el-tab-pane :label="'IP 白名单'" name="ipWhiteList">
            <IpList :id="id" :rule="'ipWhitelist'" :param-key="'$ipWhiteAllow'" v-if="index == 'ipWhiteList'"></IpList>
        </el-tab-pane>
        <el-tab-pane :label="'IP 黑名单'" name="ipBlockList">
            <IpList :id="id" :rule="'ipBlockList'" :param-key="'$ipBlockDeny'" v-if="index == 'ipBlockList'"></IpList>
        </el-tab-pane>
        <el-tab-pane :label="'文件扩展名黑名单'" name="fileExtBlockList">
            <FileBlockList :id="id" v-if="index == 'fileExtBlockList'"></FileBlockList>
        </el-tab-pane>
        <el-tab-pane :label="'URL白名单'" name="urlWhiteList">
            <ValueList
                :id="id"
                :rule="'urlWhiteList'"
                :param-key="'$urlWhiteAllow'"
                v-if="index == 'urlWhiteList'"
            ></ValueList>
        </el-tab-pane>
        <el-tab-pane :label="'URL黑名单'" name="urlBlockList">
            <ValueList
                :id="id"
                :rule="'urlBlockList'"
                :param-key="'$urlBlockDeny'"
                v-if="index == 'urlBlockList'"
            ></ValueList>
        </el-tab-pane>
        <el-tab-pane :label="'GET参数(args)校验'" name="args">
            <ValueList :id="id" :rule="'argsCheckList'" :param-key="'$argsDeny'" v-if="index == 'args'"></ValueList>
        </el-tab-pane>
        <el-tab-pane :label="'POST参数校验'" name="post">
            <ValueList :id="id" :rule="'postCheckList'" :param-key="'$postDeny'" v-if="index == 'post'"></ValueList>
        </el-tab-pane>
        <el-tab-pane :label="'Cookie黑名单'" name="cookie">
            <ValueList
                :id="id"
                :rule="'cookieBlockList'"
                :param-key="'$cookieDeny'"
                v-if="index == 'cookie'"
            ></ValueList>
        </el-tab-pane>
    </el-tabs>
</template>
<script lang="ts" setup>
import { computed, ref } from 'vue';
import CCDeny from './ccdeny/index.vue';
import IpList from './ip-list/index.vue';
import FileBlockList from './file-block-list/index.vue';
import ValueList from './value-list/index.vue';

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
