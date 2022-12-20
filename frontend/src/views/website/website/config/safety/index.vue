<template>
    <el-tabs tab-position="left" type="border-card" v-model="index">
        <el-tab-pane :label="$t('firewall.ccDeny')" name="cc">
            <CCDeny :id="id" v-if="index == 'cc'"></CCDeny>
        </el-tab-pane>
        <el-tab-pane :label="$t('firewall.ipWhiteList')" name="ipWhiteList">
            <IpList :id="id" :rule="'ipWhiteList'" :param-key="'$ipWhiteAllow'" v-if="index == 'ipWhiteList'"></IpList>
        </el-tab-pane>
        <el-tab-pane :label="$t('firewall.ipBlockList')" name="ipBlockList">
            <IpList :id="id" :rule="'ipBlockList'" :param-key="'$ipBlockDeny'" v-if="index == 'ipBlockList'"></IpList>
        </el-tab-pane>
        <el-tab-pane :label="$t('firewall.urlWhiteList')" name="urlWhiteList">
            <ValueList
                :id="id"
                :rule="'urlWhiteList'"
                :param-key="'$urlWhiteAllow'"
                v-if="index == 'urlWhiteList'"
            ></ValueList>
        </el-tab-pane>
        <el-tab-pane :label="$t('firewall.urlBlockList')" name="urlBlockList">
            <ValueList
                :id="id"
                :rule="'urlBlockList'"
                :param-key="'$urlBlockDeny'"
                v-if="index == 'urlBlockList'"
            ></ValueList>
        </el-tab-pane>
        <el-tab-pane :label="$t('firewall.cookieBlockList')" name="cookie">
            <ValueList
                :id="id"
                :rule="'cookieBlockList'"
                :param-key="'$cookieDeny'"
                v-if="index == 'cookie'"
            ></ValueList>
        </el-tab-pane>
        <el-tab-pane :label="$t('firewall.argsCheck')" name="args">
            <ValueList :id="id" :rule="'argsCheckList'" :param-key="'$argsDeny'" v-if="index == 'args'"></ValueList>
        </el-tab-pane>
        <el-tab-pane :label="$t('firewall.postCheck')" name="post">
            <ValueList :id="id" :rule="'postCheckList'" :param-key="'$postDeny'" v-if="index == 'post'"></ValueList>
        </el-tab-pane>
        <el-tab-pane :label="$t('firewall.fileExtBlockList')" name="fileExtBlockList">
            <FileBlockList :id="id" v-if="index == 'fileExtBlockList'"></FileBlockList>
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
