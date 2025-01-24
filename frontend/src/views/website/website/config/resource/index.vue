<template>
    <el-tabs tab-position="left" v-model="index">
        <el-tab-pane :label="'OpenResty'" name="0">
            <Nginx :id="id" v-if="index == '0'"></Nginx>
        </el-tab-pane>
    </el-tabs>
</template>

<script lang="ts" setup>
import { GetRuntime } from '@/api/modules/runtime';
import { getWebsite } from '@/api/modules/website';
import { computed, onMounted, ref } from 'vue';
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
let configPHP = ref(false);
let installId = ref(0);

const getWebsiteDetail = async () => {
    const res = await getWebsite(props.id);
    if (res.data.type === 'runtime') {
        installId.value = res.data.appInstallId;
        const runRes = await GetRuntime(res.data.runtimeID);
        if (runRes.data.type == 'php' && runRes.data.resource === 'appstore') {
            configPHP.value = true;
        }
    }
};

onMounted(() => {
    getWebsiteDetail();
});
</script>
