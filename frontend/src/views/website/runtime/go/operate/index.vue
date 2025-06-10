<template>
    <DrawerPro v-model="open" size="50%">
        <template #header>
            <DrawerHeader
                :header="$t('runtime.' + mode)"
                :hideResource="mode == 'create'"
                :resource="runtime.name"
                :back="handleClose"
            />
        </template>
        <el-row v-loading="loading">
            <el-col :span="22" :offset="1">
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
                    <el-form-item :label="$t('runtime.app')" prop="appID">
                        <el-row :gutter="20">
                            <el-col :span="12">
                                <el-select
                                    v-model="runtime.appID"
                                    :disabled="mode === 'edit'"
                                    @change="changeApp(runtime.appID)"
                                    class="p-w-200"
                                >
                                    <el-option
                                        v-for="(app, index) in apps"
                                        :key="index"
                                        :label="app.name"
                                        :value="app.id"
                                    ></el-option>
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
                    <el-form-item :label="$t('tool.supervisor.dir')" prop="codeDir">
                        <el-input v-model.trim="runtime.codeDir" :disabled="mode === 'edit'">
                            <template #prepend>
                                <FileList
                                    :disabled="mode === 'edit'"
                                    :path="runtime.codeDir"
                                    @choose="getPath"
                                    :dir="true"
                                ></FileList>
                            </template>
                        </el-input>
                        <span class="input-help">
                            {{ $t('runtime.goDirHelper') }}
                        </span>
                    </el-form-item>
                    <el-form-item :label="$t('runtime.runScript')" prop="params.EXEC_SCRIPT">
                        <el-input v-model="runtime.params['EXEC_SCRIPT']"></el-input>
                        <span class="input-help">
                            {{ $t('runtime.goHelper') }}
                        </span>
                    </el-form-item>
                    <el-row :gutter="20">
                        <el-col :span="7">
                            <el-form-item :label="$t('runtime.appPort')" prop="params.GO_APP_PORT">
                                <el-input v-model.number="runtime.params['GO_APP_PORT']" />
                                <span class="input-help">{{ $t('runtime.appPortHelper') }}</span>
                            </el-form-item>
                        </el-col>
                        <el-col :span="7">
                            <el-form-item :label="$t('runtime.externalPort')" prop="port">
                                <el-input v-model.number="runtime.port" />
                                <span class="input-help">{{ $t('runtime.externalPortHelper') }}</span>
                            </el-form-item>
                        </el-col>
                        <el-col :span="4">
                            <el-form-item :label="$t('commons.button.add') + $t('commons.table.port')">
                                <el-button @click="addPort">
                                    <el-icon><Plus /></el-icon>
                                </el-button>
                            </el-form-item>
                        </el-col>
                        <el-col :span="6">
                            <el-form-item :label="$t('app.allowPort')" prop="params.HOST_IP">
                                <el-switch
                                    v-model="runtime.params['HOST_IP']"
                                    :active-value="'0.0.0.0'"
                                    :inactive-value="'127.0.0.1'"
                                />
                            </el-form-item>
                        </el-col>
                    </el-row>
                    <el-row :gutter="20" v-for="(port, index) of runtime.exposedPorts" :key="index">
                        <el-col :span="7">
                            <el-form-item
                                :prop="'exposedPorts.' + index + '.containerPort'"
                                :rules="rules.params.GO_APP_PORT"
                            >
                                <el-input v-model.number="port.containerPort" :placeholder="$t('runtime.appPort')" />
                            </el-form-item>
                        </el-col>
                        <el-col :span="7">
                            <el-form-item
                                :prop="'exposedPorts.' + index + '.hostPort'"
                                :rules="rules.params.GO_APP_PORT"
                            >
                                <el-input v-model.number="port.hostPort" :placeholder="$t('runtime.externalPort')" />
                            </el-form-item>
                        </el-col>
                        <el-col :span="4">
                            <el-form-item>
                                <el-button type="primary" @click="removePort(index)" link>
                                    {{ $t('commons.button.delete') }}
                                </el-button>
                            </el-form-item>
                        </el-col>
                    </el-row>
                    <el-form-item :label="$t('app.containerName')" prop="params.CONTAINER_NAME">
                        <el-input v-model.trim="runtime.params['CONTAINER_NAME']"></el-input>
                    </el-form-item>
                </el-form>
            </el-col>
        </el-row>
        <template #footer>
            <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
            <el-button type="primary" @click="submit(runtimeForm)" :disabled="loading">
                {{ $t('commons.button.confirm') }}
            </el-button>
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
});
let runtime = reactive<Runtime.RuntimeCreate>(initData('go'));
const rules = ref<any>({
    name: [Rules.requiredInput, Rules.appName],
    appID: [Rules.requiredSelect],
    codeDir: [Rules.requiredInput],
    port: [Rules.requiredInput, Rules.paramPort, checkNumberRange(1, 65535)],
    source: [Rules.requiredSelect],
    params: {
        APP_PORT: [Rules.requiredInput, Rules.paramPort, checkNumberRange(1, 65535)],
        HOST_IP: [Rules.requiredSelect],
        CONTAINER_NAME: [Rules.requiredInput, Rules.containerName],
        EXEC_SCRIPT: [Rules.requiredInput],
    },
});
const scripts = ref<Runtime.NodeScripts[]>([]);
const em = defineEmits(['close']);

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

const getPath = (codeDir: string) => {
    runtime.codeDir = codeDir;
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
        runtime.volumes = data.volumes || [];
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
