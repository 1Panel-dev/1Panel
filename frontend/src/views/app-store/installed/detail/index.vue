<template>
    <el-drawer :close-on-click-modal="false" v-model="open" size="40%">
        <template #header>
            <Header :header="$t('app.param')" :back="handleClose"></Header>
        </template>
        <el-descriptions border :column="1">
            <el-descriptions-item v-for="(param, key) in params" :label="param.label" :key="key">
                {{ param.value }}
            </el-descriptions-item>
        </el-descriptions>
    </el-drawer>
</template>
<script lang="ts" setup>
import { App } from '@/api/interface/app';
import { GetAppInstallParams } from '@/api/modules/app';
import { ref } from 'vue';
import Header from '@/components/drawer-header/index.vue';

interface ParamProps {
    id: Number;
}
const paramData = ref<ParamProps>({
    id: 0,
});

let open = ref(false);
let loading = ref(false);
const params = ref<App.InstallParams[]>();

const acceptParams = (props: ParamProps) => {
    params.value = [];
    paramData.value.id = props.id;
    get();
    open.value = true;
};

const handleClose = () => {
    open.value = false;
};

const get = async () => {
    try {
        loading.value = true;
        const res = await GetAppInstallParams(Number(paramData.value.id));
        params.value = res.data;
    } catch (error) {
    } finally {
        loading.value = false;
    }
};

defineExpose({ acceptParams });
</script>
