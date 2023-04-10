<template>
    <el-tabs tab-position="left" v-model="index">
        <el-tab-pane :label="'OpenResty'" name="0">
            <Nginx :id="id" v-if="index == '0'"></Nginx>
        </el-tab-pane>
        <el-tab-pane :label="'FPM'" name="1" v-if="configPHP">
            <PHP :id="id" v-if="index == '1'" :installId="installId" :type="'fpm'"></PHP>
        </el-tab-pane>
        <el-tab-pane :label="'PHP'" name="2" v-if="configPHP">
            <PHP :id="id" v-if="index == '2'" :installId="installId" :type="'php'"></PHP>
        </el-tab-pane>
    </el-tabs>
</template>

<script lang="ts" setup>
import { GetRuntime } from '@/api/modules/runtime';
import { GetWebsite } from '@/api/modules/website';
import { computed, onMounted, ref } from 'vue';
import Nginx from './nginx/index.vue';
import PHP from './php-fpm/index.vue';

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
let configPHP = ref(false);
let installId = ref(0);

const getWebsiteDetail = async () => {
    const res = await GetWebsite(props.id);
    if (res.data.type === 'runtime') {
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
