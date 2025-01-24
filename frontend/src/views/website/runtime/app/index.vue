<template>
    <el-form-item :label="$t('runtime.app')" prop="appID">
        <el-row :gutter="20">
            <el-col :span="12">
                <el-select
                    v-model="runtime.appID"
                    :disabled="mode === 'edit'"
                    @change="changeApp(runtime.appID)"
                    class="p-w-200"
                >
                    <el-option v-for="(app, index) in apps" :key="index" :label="app.name" :value="app.id"></el-option>
                </el-select>
            </el-col>
            <el-col :span="12">
                <el-select
                    v-model="runtime.version"
                    :disabled="mode === 'edit'"
                    @change="changeVersion()"
                    class="p-w-200"
                >
                    <el-option
                        v-for="(version, index) in appVersions"
                        :key="index"
                        :label="version"
                        :value="version"
                    ></el-option>
                </el-select>
            </el-col>
        </el-row>
    </el-form-item>
</template>

<script setup lang="ts">
import { App } from '@/api/interface/app';
import { getAppByKey, getAppDetail, searchApp } from '@/api/modules/app';
import { useVModel } from '@vueuse/core';
import { defineProps } from 'vue';

const props = defineProps({
    mode: {
        type: String,
        required: true,
    },
    appKey: {
        type: String,
        required: true,
    },
    modelValue: {
        type: Object,
        required: true,
    },
});
const apps = ref<App.App[]>([]);
const appVersions = ref<string[]>([]);
const emit = defineEmits(['update:modelValue']);
const runtime = useVModel(props, 'modelValue', emit);
const appReq = reactive({
    type: props.appKey,
    page: 1,
    pageSize: 20,
    resource: 'remote',
});

const changeApp = (appID: number) => {
    for (const app of apps.value) {
        if (app.id === appID) {
            getApp(app.key, props.mode);
            break;
        }
    }
};

const changeVersion = async () => {
    try {
        const res = await getAppDetail(runtime.value.appID, runtime.value.version, 'runtime');
        runtime.value.appDetailID = res.data.id;
    } catch (error) {}
};

const getApp = async (appkey: string, mode: string) => {
    try {
        const res = await getAppByKey(appkey);
        appVersions.value = res.data.versions || [];
        if (res.data.versions.length > 0) {
            if (mode === 'create') {
                runtime.value.version = res.data.versions[0];
                changeVersion();
            }
        }
    } catch (error) {}
};

const searchAppList = async (appID: number) => {
    try {
        const res = await searchApp(appReq);
        apps.value = res.data.items || [];
        if (res.data && res.data.items && res.data.items.length > 0) {
            if (appID == null) {
                runtime.value.appID = res.data.items[0].id;
                getApp(res.data.items[0].key, props.mode);
            } else {
                res.data.items.forEach((item) => {
                    if (item.id === appID) {
                        getApp(item.key, props.mode);
                    }
                });
            }
        }
    } catch (error) {}
};

onMounted(() => {
    if (props.mode === 'create') {
        searchAppList(null);
    } else {
        searchAppList(runtime.value.appID);
    }
});
</script>
