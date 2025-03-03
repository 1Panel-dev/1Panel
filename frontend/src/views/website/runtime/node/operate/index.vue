<template>
    <DrawerPro
        v-model="open"
        :header="$t('runtime.' + mode)"
        :resource="mode === 'edit' ? runtime.name : ''"
        size="large"
        @close="handleClose"
    >
        <el-form
            v-loading="loading"
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
            <AppConfig v-model="runtime" :mode="mode" appKey="node" />
            <DirConfig v-model="runtime" :mode="mode" appKey="node" :scriptHelper="$t('runtime.customScriptHelper')" />
            <el-form-item :label="$t('app.containerName')" prop="params.CONTAINER_NAME">
                <el-input v-model.trim="runtime.params['CONTAINER_NAME']"></el-input>
            </el-form-item>
            <el-form-item :label="$t('runtime.packageManager')" prop="params.PACKAGE_MANAGER">
                <el-select v-model="runtime.params['PACKAGE_MANAGER']">
                    <el-option label="npm" value="npm"></el-option>
                    <el-option label="yarn" value="yarn"></el-option>
                    <el-option v-if="hasPnpm" label="pnpm" value="pnpm"></el-option>
                </el-select>
            </el-form-item>
            <el-form-item :label="$t('runtime.imageSource')" prop="source">
                <el-select v-model="runtime.source" filterable allow-create default-first-option>
                    <el-option
                        v-for="(source, index) in imageSources"
                        :key="index"
                        :label="source.label + ' [' + source.value + ']'"
                        :value="source.value"
                    ></el-option>
                </el-select>
                <span class="input-help">
                    {{ $t('runtime.phpsourceHelper') }}
                </span>
            </el-form-item>
            <PortConfig v-model="runtime" :mode="mode" />
            <Environment :environments="runtime.environments" />
            <Volumes :volumes="runtime.volumes" />
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
import { computed, reactive, ref, watch } from 'vue';
import PortConfig from '@/views/website/runtime/port/index.vue';
import Environment from '@/views/website/runtime/environment/index.vue';
import AppConfig from '@/views/website/runtime/app/index.vue';
import Volumes from '@/views/website/runtime/volume/index.vue';
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
        PACKAGE_MANAGER: 'npm',
        HOST_IP: '0.0.0.0',
        CUSTOM_SCRIPT: '0',
    },
    type: type,
    resource: 'appstore',
    rebuild: false,
    codeDir: '/',
    port: 4004,
    source: 'https://registry.npmjs.org/',
    exposedPorts: [],
    environments: [],
    volumes: [],
});
let runtime = reactive<Runtime.RuntimeCreate>(initData('node'));
const rules = ref<any>({
    name: [Rules.requiredInput, Rules.appName],
    appID: [Rules.requiredSelect],
    codeDir: [Rules.requiredInput],
    port: [Rules.requiredInput, Rules.paramPort, checkNumberRange(1, 65535)],
    source: [Rules.requiredSelect],
    params: {
        PACKAGE_MANAGER: [Rules.requiredSelect],
        HOST_IP: [Rules.requiredSelect],
        EXEC_SCRIPT: [Rules.requiredSelect],
        CONTAINER_NAME: [Rules.requiredInput, Rules.containerName],
    },
});
const em = defineEmits(['close']);

const hasPnpm = computed(() => {
    if (runtime.version == undefined) {
        return false;
    }
    return parseFloat(runtime.version) > 18;
});

const imageSources = [
    {
        label: i18n.global.t('commons.table.default'),
        value: 'https://registry.npmjs.org/',
    },
    {
        label: i18n.global.t('runtime.taobao'),
        value: 'https://registry.npmmirror.com',
    },
    {
        label: i18n.global.t('runtime.tencent'),
        value: 'https://mirrors.cloud.tencent.com/npm/',
    },
];

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
        });
        runtime.exposedPorts = data.exposedPorts || [];
        runtime.environments = data.environments || [];
        editParams.value = data.appParams;
        if (data.params['CUSTOM_SCRIPT'] == undefined || data.params['CUSTOM_SCRIPT'] == '0') {
            data.params['CUSTOM_SCRIPT'] = '0';
        }
        open.value = true;
    } catch (error) {}
};

const acceptParams = async (props: OperateRrops) => {
    mode.value = props.mode;
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
