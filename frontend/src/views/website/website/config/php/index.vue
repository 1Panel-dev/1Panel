<template>
    <el-tabs tab-position="left" v-model="index">
        <el-tab-pane :label="$t('website.updateConfig')" name="0">
            <Config :id="id"></Config>
        </el-tab-pane>
    </el-tabs>
</template>

<script lang="ts" setup>
import { GetRuntime } from '@/api/modules/runtime';
import { GetWebsite } from '@/api/modules/website';
import { computed, onMounted, ref } from 'vue';
import Config from './config/index.vue';

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
