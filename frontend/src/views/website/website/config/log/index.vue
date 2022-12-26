<template>
    <el-tabs tab-position="left" type="border-card" v-model="index">
        <el-tab-pane :label="$t('website.accessLog')" name="0">
            <LogFile :path="website.accessLogPath" v-if="index == '0'"></LogFile>
        </el-tab-pane>
        <el-tab-pane :label="$t('website.errLog')" name="1">
            <LogFile :path="website.errorLogPath" v-if="index == '1'"></LogFile>
        </el-tab-pane>
    </el-tabs>
</template>

<script lang="ts" setup>
import { GetWebsite } from '@/api/modules/website';
import { computed, onMounted, ref } from 'vue';
import LogFile from './log-fiile/index.vue';

const props = defineProps({
    id: {
        type: Number,
        default: 0,
    },
});
const id = computed(() => {
    return props.id;
});

let loading = ref(false);
let website = ref();
let index = ref('-1');

const getWebsite = () => {
    loading.value = true;
    GetWebsite(id.value)
        .then((res) => {
            website.value = res.data;
            index.value = '0';
        })
        .finally(() => {
            loading.value = false;
        });
};

onMounted(() => {
    getWebsite();
});
</script>
