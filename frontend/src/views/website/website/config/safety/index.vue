<template>
    <el-tabs tab-position="left" v-model="index">
        <el-tab-pane :label="$t('firewall.ccDeny')" name="cc">
            <CCDeny :id="id" v-if="index == 'cc'"></CCDeny>
        </el-tab-pane>
        <el-tab-pane :label="$t('firewall.ipWhiteList')" name="ipWhiteList">
            <SimpleList
                :id="id"
                :rule="'ip_white'"
                :param-key="'$ipWhiteAllow'"
                v-if="index == 'ipWhiteList'"
            ></SimpleList>
        </el-tab-pane>
        <el-tab-pane :label="$t('firewall.ipBlockList')" name="ipBlockList">
            <SimpleList
                :id="id"
                :rule="'ip_block'"
                :param-key="'$ipBlockDeny'"
                v-if="index == 'ipBlockList'"
            ></SimpleList>
        </el-tab-pane>
        <el-tab-pane :label="$t('firewall.urlWhiteList')" name="urlWhiteList">
            <SimpleList
                :id="id"
                :rule="'url_white'"
                :param-key="'$urlWhiteAllow'"
                v-if="index == 'urlWhiteList'"
            ></SimpleList>
        </el-tab-pane>
        <el-tab-pane :label="$t('firewall.urlBlockList')" name="urlBlockList">
            <SimpleList
                :id="id"
                :rule="'url_block'"
                :param-key="'$urlBlockDeny'"
                v-if="index == 'urlBlockList'"
            ></SimpleList>
        </el-tab-pane>
        <el-tab-pane :label="$t('firewall.cookieBlockList')" name="cookie">
            <ValueList :id="id" :rule="'cookie_block'" :param-key="'$cookieDeny'" v-if="index == 'cookie'"></ValueList>
        </el-tab-pane>
        <el-tab-pane :label="$t('firewall.argsCheck')" name="args">
            <ValueList :id="id" :rule="'args_check'" :param-key="'$argsDeny'" v-if="index == 'args'"></ValueList>
        </el-tab-pane>
        <el-tab-pane :label="$t('firewall.postCheck')" name="post">
            <ValueList :id="id" :rule="'post_check'" :param-key="'$postDeny'" v-if="index == 'post'"></ValueList>
        </el-tab-pane>
        <el-tab-pane :label="$t('firewall.userAgent')" name="ua">
            <ValueList :id="id" :rule="'user_agent'" :param-key="'$userAgent'" v-if="index == 'ua'"></ValueList>
        </el-tab-pane>
        <el-tab-pane :label="$t('firewall.fileExtBlockList')" name="fileExtBlockList">
            <FileBlockList :id="id" v-if="index == 'fileExtBlockList'"></FileBlockList>
        </el-tab-pane>
    </el-tabs>
</template>
<script lang="ts" setup>
import { computed, ref } from 'vue';
import CCDeny from './ccdeny/index.vue';
import SimpleList from './simple-list/index.vue';
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
