<template>
    <DrawerPro
        v-model="open"
        :header="$t('runtime.' + mode)"
        size="large"
        :resource="mode === 'edit' ? runtime.name : ''"
        @close="handleClose"
    >
        <el-form
            ref="runtimeForm"
            label-position="top"
            :model="runtime"
            label-width="125px"
            :rules="rules"
            :validate-on-rule-change="false"
        >
            <el-form-item :label="$t('commons.table.name')" prop="name">
                <el-input :disabled="mode === 'edit'" v-model="runtime.name"></el-input>
            </el-form-item>
            <AppConfig v-model="runtime" :mode="mode" appKey="java" />
            <DirConfig
                v-model="runtime"
                :mode="mode"
                :dirHelper="$t('runtime.javaDirHelper')"
                :scriptHelper="$t('runtime.javaScriptHelper')"
            />
            <el-form-item :label="$t('app.containerName')" prop="params.CONTAINER_NAME">
                <el-input v-model.trim="runtime.params['CONTAINER_NAME']"></el-input>
            </el-form-item>
            <el-form-item :label="$t('website.remark')" prop="remark">
                <el-input type="textarea" :rows="1" clearable v-model="runtime.remark" />
            </el-form-item>
            <el-tabs type="border-card">
                <el-tab-pane :label="$t('commons.table.port')">
                    <PortConfig v-model="runtime" :mode="mode" />
                </el-tab-pane>
                <el-tab-pane :label="$t('runtime.environment')">
                    <Environment :environments="runtime.environments" />
                </el-tab-pane>
                <el-tab-pane :label="$t('container.mount')"><Volumes :volumes="runtime.volumes" /></el-tab-pane>
            </el-tabs>
        </el-form>
        <template #footer>
            <span>
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit(runtimeForm)" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { App } from '@/api/interface/app';
import { Runtime } from '@/api/interface/runtime';
import { CreateRuntime, GetRuntime, UpdateRuntime } from '@/api/modules/runtime';
import { Rules, checkNumberRange } from '@/global/form-rules';
import i18n from '@/lang';
import { MsgError, MsgSuccess } from '@/utils/message';
import { FormInstance } from 'element-plus';
import { reactive, ref, watch } from 'vue';
import PortConfig from '@/views/website/runtime/port/index.vue';
import Environment from '@/views/website/runtime/environment/index.vue';
import Volumes from '@/views/website/runtime/volume/index.vue';
import AppConfig from '@/views/website/runtime/app/index.vue';
import DirConfig from '@/views/website/runtime/dir/index.vue';

interface OperateRrops {
    id?: number;
    mode: string;
    type: string;
}

const open = ref(false);
const runtimeForm = ref<FormInstance>();
const loading = ref(false);
const mode = ref('create');
const editParams = ref<App.InstallParams[]>();
const initData = (type: string) => ({
    name: '',
    appDetailID: undefined,
    image: '',
    params: {
        HOST_IP: '0.0.0.0',
    },
    type: type,
    resource: 'appstore',
    rebuild: false,
    codeDir: '/',
    port: 8080,
    exposedPorts: [],
    environments: [],
    volumes: [],
    remark: '',
});
let runtime = reactive<Runtime.RuntimeCreate>(initData('java'));
const rules = ref<any>({
    name: [Rules.requiredInput, Rules.appName],
    appID: [Rules.requiredSelect],
    codeDir: [Rules.requiredInput],
    port: [Rules.requiredInput, Rules.paramPort, checkNumberRange(1, 65535)],
    source: [Rules.requiredSelect],
    params: {
        HOST_IP: [Rules.requiredSelect],
        CONTAINER_NAME: [Rules.requiredInput, Rules.containerName],
        EXEC_SCRIPT: [Rules.requiredInput],
    },
});
const scripts = ref<Runtime.NodeScripts[]>([]);
const em = defineEmits(['close']);

watch(
    () => runtime.params['APP_PORT'],
    (newVal) => {
        if (newVal && mode.value == 'create') {
            runtime.port = newVal;
        }
    },
    { deep: true },
);

watch(
    () => runtime.name,
    (newVal) => {
        if (newVal && mode.value == 'create') {
            runtime.params['CONTAINER_NAME'] = newVal;
        }
    },
    { deep: true },
);

const handleClose = () => {
    open.value = false;
    em('close', false);
    runtimeForm.value?.resetFields();
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        if (runtime.exposedPorts && runtime.exposedPorts.length > 0) {
            const containerPortMap = new Map();
            const hostPortMap = new Map();
            for (const port of runtime.exposedPorts) {
                if (containerPortMap[port.containerPort]) {
                    MsgError(i18n.global.t('runtime.portError'));
                    return;
                }
                if (hostPortMap[port.hostPort]) {
                    MsgError(i18n.global.t('runtime.portError'));
                    return;
                }
                hostPortMap[port.hostPort] = true;
                containerPortMap[port.containerPort] = true;
            }
        }

        if (mode.value == 'create') {
            loading.value = true;
            CreateRuntime(runtime)
                .then(() => {
                    MsgSuccess(i18n.global.t('commons.msg.createSuccess'));
                    handleClose();
                })
                .finally(() => {
                    loading.value = false;
                });
        } else {
            loading.value = true;
            UpdateRuntime(runtime)
                .then(() => {
                    MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
                    handleClose();
                })
                .finally(() => {
                    loading.value = false;
                });
        }
    });
};

const getRuntime = async (id: number) => {
    try {
        const res = await GetRuntime(id);
        const data = res.data;
        Object.assign(runtime, {
            id: data.id,
            name: data.name,
            appDetailID: data.appDetailID,
            image: data.image,
            type: data.type,
            resource: data.resource,
            appID: data.appID,
            version: data.version,
            rebuild: true,
            source: data.source,
            params: data.params,
            codeDir: data.codeDir,
            port: data.port,
            remark: data.remark,
        });
        runtime.exposedPorts = data.exposedPorts || [];
        runtime.environments = data.environments || [];
        editParams.value = data.appParams;
        open.value = true;
    } catch (error) {}
};

const acceptParams = async (props: OperateRrops) => {
    mode.value = props.mode;
    scripts.value = [];
    if (props.mode === 'create') {
        Object.assign(runtime, initData(props.type));
        open.value = true;
    } else {
        getRuntime(props.id);
    }
};

defineExpose({
    acceptParams,
});
</script>
