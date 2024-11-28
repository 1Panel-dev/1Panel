<template>
    <el-tabs tab-position="left" v-model="index">
        <el-tab-pane :label="$t('website.updateConfig')" name="0">
            <Config :id="id" v-if="index == '0'"></Config>
        </el-tab-pane>
        <el-tab-pane :label="$t('php.disableFunction')" name="1">
            <Function :id="id" v-if="index == '1'"></Function>
        </el-tab-pane>
        <el-tab-pane :label="$t('php.uploadMaxSize')" name="2">
            <Upload :id="id" v-if="index == '2'"></Upload>
        </el-tab-pane>
        <el-tab-pane :label="$t('runtime.version')" name="3">
            <Version :id="id" v-if="index == '3'"></Version>
        </el-tab-pane>
    </el-tabs>
</template>

<script lang="ts" setup>
import { GetRuntime } from '@/api/modules/runtime';
import { GetWebsite } from '@/api/modules/website';
import { computed, onMounted, ref } from 'vue';
import Config from './config/index.vue';
import Function from './function/index.vue';
import Upload from './upload/index.vue';
import Version from './version/index.vue';

const props = defineProps({
    id: {
        type: Number,
        default: 0,
    },
});

const id = computed(() => {
    return props.id;
});

const index = ref('0');
const configPHP = ref(false);
const installId = ref(0);
const runtimeID = ref(0);

const getWebsiteDetail = async () => {
    const res = await GetWebsite(props.id);
    if (res.data.type === 'runtime') {
        runtimeID.value = res.data.runtimeID;
        installId.value = res.data.appInstallId;
        const runRes = await GetRuntime(res.data.runtimeID);
        if (runRes.data.resource === 'appstore') {
            configPHP.value = true;
        }
    }
};

onMounted(() => {
    getWebsiteDetail();
});
</script>
