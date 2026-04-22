<template>
    <DrawerPro :header="$t('commons.button.' + mode)" v-model="drawerVisiable" size="large" @close="handleClose">
        <el-alert :title="$t('aiTools.tensorRT.imageAlert')" class="common-prompt" :closable="false" type="warning" />
        <el-form ref="formRef" label-position="top" :model="tensorRTLLM" :rules="rules" v-loading="loading">
            <el-form-item :label="$t('commons.table.name')" prop="name">
                <el-input clearable v-model.trim="tensorRTLLM.name" :disabled="mode == 'edit'" />
            </el-form-item>
            <el-form-item :label="$t('app.containerName')" prop="containerName">
                <el-input v-model.trim="tensorRTLLM.containerName"></el-input>
            </el-form-item>
            <el-form-item :label="$t('container.image')" prop="image">
                <el-input v-model.trim="tensorRTLLM.image" />
            </el-form-item>
            <el-form-item :label="$t('app.version')" prop="version">
                <el-input v-model.trim="tensorRTLLM.version" />
            </el-form-item>
            <div v-if="isFxplay">
                <el-form-item :label="$t('aiTools.tensorRT.modelSpeedup')" prop="modelSpeedup">
                    <el-switch v-model="tensorRTLLM.modelSpeedup" @change="changeModelSpeedup"></el-switch>
                </el-form-item>
                <el-form-item
                    :label="$t('aiTools.tensorRT.modelType')"
                    prop="modelType"
                    v-if="tensorRTLLM.modelSpeedup"
                >
                    <el-select v-model="tensorRTLLM.modelType">
                        <el-option label="Qwen3" value="Qwen3" />
                    </el-select>
                </el-form-item>
            </div>
            <el-form-item :label="$t('aiTools.tensorRT.modelDir')" prop="modelDir">
                <el-input v-model="tensorRTLLM.modelDir">
                    <template #prepend>
                        <el-button icon="Folder" @click="modelDirRef.acceptParams({ dir: true })" />
                    </template>
                </el-input>
            </el-form-item>
            <el-form-item :label="$t('runtime.runScript')" prop="command">
                <el-input v-model="tensorRTLLM.command"></el-input>
                <span class="input-help">
                    {{ $t('aiTools.tensorRT.commandHelper') }}
                </span>
            </el-form-item>
            <NodeConfig v-model="tensorRTLLM" />
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="drawerVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button :disabled="loading" type="primary" @click="onSubmit">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
        <FileList ref="modelDirRef" @choose="getModelDir" />
    </DrawerPro>
</template>

<script lang="ts" setup>
import NodeConfig from '@/views/website/runtime//components/node-config.vue';
import DrawerPro from '@/components/drawer-pro/index.vue';
import FileList from '@/components/file-list/index.vue';

import { reactive, ref } from 'vue';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElForm, FormInstance } from 'element-plus';
import { createTensorRTLLM, updateTensorRTLLM } from '@/api/modules/ai';
import { MsgSuccess } from '@/utils/message';
import { useGlobalStore } from '@/composables/useGlobalStore';
const { isFxplay } = useGlobalStore();

const loading = ref(false);
const mode = ref('create');
const drawerVisiable = ref(false);
const newTensorRTLLM = () => {
    return {
        name: '',
        containerName: '',
        version: '1.2.0rc0',
        modelDir: '',
        image: 'nvcr.io/nvidia/tensorrt-llm/release',
        command: 'bash -c "trtllm-serve ${MODEL_PATH} --host 0.0.0.0 --port 8000"',
        exposedPorts: [],
        environments: [],
        extraHosts: [],
        volumes: [],
        modelSpeedup: false,
        modelType: 'Qwen3',
    };
};
const modelDirRef = ref();
const tensorRTLLM = ref(newTensorRTLLM());
const emit = defineEmits(['search']);

const openCreate = (): void => {
    mode.value = 'create';
    drawerVisiable.value = true;
    tensorRTLLM.value = newTensorRTLLM();
};

const openEdit = (rowData: any): void => {
    mode.value = 'edit';
    tensorRTLLM.value = { ...rowData };
    if (tensorRTLLM.value.environments == null) {
        tensorRTLLM.value.environments = [];
    }
    if (tensorRTLLM.value.volumes == null) {
        tensorRTLLM.value.volumes = [];
    }
    if (tensorRTLLM.value.exposedPorts == null) {
        tensorRTLLM.value.exposedPorts = [];
    }
    tensorRTLLM.value.command = rowData.command.slice(1, -1).replace(/^'|'$/g, '').replace(/\\"/g, '"');
    if (tensorRTLLM.value.extraHosts == null) {
        tensorRTLLM.value.extraHosts = [];
    }
    drawerVisiable.value = true;
};

const handleClose = () => {
    drawerVisiable.value = false;
};

const getModelDir = (path: string) => {
    tensorRTLLM.value.modelDir = path;
};

const changeModelSpeedup = () => {
    if (tensorRTLLM.value.modelSpeedup) {
        tensorRTLLM.value.command =
            'bash -c "${MODEL_PATH}/fusionxpark_accelerator --model_path ${MODEL_PATH} --host 0.0.0.0 --port 8000"';
    } else {
        tensorRTLLM.value.command = 'bash -c "trtllm-serve ${MODEL_PATH} --host 0.0.0.0 --port 8000"';
    }
};

const rules = reactive({
    name: [Rules.requiredInput],
    version: [Rules.requiredInput],
    modelDir: [Rules.requiredInput],
    containerName: [Rules.requiredInput],
    image: [Rules.requiredInput],
    command: [Rules.requiredInput],
    modelType: [Rules.requiredSelect],
});

const formRef = ref<FormInstance>();

const onSubmit = async () => {
    formRef.value?.validate(async (valid) => {
        if (!valid) return;
        loading.value = true;
        if (mode.value === 'edit') {
            await updateTensorRTLLM(tensorRTLLM.value)
                .then(() => {
                    loading.value = false;
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                    emit('search');
                    drawerVisiable.value = false;
                })
                .catch(() => {
                    loading.value = false;
                });
            return;
        }
        await createTensorRTLLM(tensorRTLLM.value)
            .then(() => {
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                emit('search');
                drawerVisiable.value = false;
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

watch(
    () => tensorRTLLM.value.name,
    (newVal) => {
        if (newVal && mode.value == 'create') {
            tensorRTLLM.value.containerName = newVal;
        }
    },
    { deep: true },
);

defineExpose({
    openCreate,
    openEdit,
});
</script>
