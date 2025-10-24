<template>
    <div>
        <el-row v-loading="loading">
            <el-col :span="22" :offset="1">
                <el-form
                    v-loading="loading"
                    ref="runtimeForm"
                    label-position="top"
                    :model="containerConfig"
                    :rules="rules"
                    label-width="125px"
                    :validate-on-rule-change="false"
                >
                    <el-form-item :label="$t('app.containerName')" prop="containerName">
                        <el-input v-model.trim="containerConfig.containerName"></el-input>
                    </el-form-item>
                    <el-tabs type="border-card">
                        <el-tab-pane :label="$t('commons.table.port')">
                            <PortConfig :exposedPorts="containerConfig.exposedPorts" />
                        </el-tab-pane>
                        <el-tab-pane :label="$t('runtime.environment')">
                            <Environment :environments="containerConfig.environments" />
                        </el-tab-pane>
                        <el-tab-pane :label="$t('container.mount')">
                            <Volumes :volumes="containerConfig.volumes" />
                        </el-tab-pane>
                        <el-tab-pane :label="$t('runtime.extraHosts')">
                            <ExtraHosts :extraHosts="containerConfig.extraHosts" />
                        </el-tab-pane>
                    </el-tabs>
                    <el-form-item class="mt-2">
                        <el-button type="primary" @click="onSaveStart(runtimeForm)">
                            {{ $t('commons.button.save') }}
                        </el-button>
                    </el-form-item>
                </el-form>
            </el-col>
        </el-row>
    </div>
</template>

<script lang="ts" setup>
import PortConfig from '@/views/website/runtime/components/port/index.vue';
import Environment from '@/views/website/runtime/components/environment/index.vue';
import Volumes from '@/views/website/runtime/components/volume/index.vue';
import ExtraHosts from '@/views/website/runtime/components/extra_hosts/index.vue';

import { Runtime } from '@/api/interface/runtime';
import { getPHPContainerConfig, updatePHPContainerConfig } from '@/api/modules/runtime';
import { FormInstance } from 'element-plus';
import { ref } from 'vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { Rules } from '@/global/form-rules';

const props = defineProps<{
    id: number;
}>();
const loading = ref<boolean>(false);
const runtimeForm = ref<FormInstance>();
const containerConfig = ref<Runtime.PHPContainerConfig>({
    id: 0,
    containerName: '',
    exposedPorts: [],
    environments: [],
    volumes: [],
    extraHosts: [],
});

const rules = {
    containerName: [Rules.containerName, Rules.requiredInput],
};

const getConfig = async () => {
    const res = await getPHPContainerConfig(props.id);
    containerConfig.value.exposedPorts = res.data.exposedPorts || [];
    containerConfig.value.environments = res.data.environments || [];
    containerConfig.value.volumes = res.data.volumes || [];
    containerConfig.value.containerName = res.data.containerName || '';
    containerConfig.value.extraHosts = res.data.extraHosts || [];
};

const onSaveStart = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        const action = await ElMessageBox.confirm(
            i18n.global.t('runtime.phpConfigHelper'),
            i18n.global.t('database.confChange'),
            {
                confirmButtonText: i18n.global.t('commons.button.confirm'),
                cancelButtonText: i18n.global.t('commons.button.cancel'),
                type: 'info',
            },
        );
        if (action === 'confirm') {
            loading.value = true;
            submit();
        }
    });
};

const submit = async () => {
    try {
        await updatePHPContainerConfig({
            id: props.id,
            containerName: containerConfig.value.containerName,
            exposedPorts: containerConfig.value.exposedPorts,
            environments: containerConfig.value.environments,
            volumes: containerConfig.value.volumes,
            extraHosts: containerConfig.value.extraHosts,
        });
        MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
    } catch (error) {
    } finally {
        loading.value = false;
    }
};

onMounted(() => {
    getConfig();
});
</script>
