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
            <el-tabs type="border-card">
                <el-tab-pane :label="$t('commons.table.port')">
                    <PortConfig v-model="tensorRTLLM" :mode="mode" />
                </el-tab-pane>
                <el-tab-pane :label="$t('runtime.environment')">
                    <Environment :environments="tensorRTLLM.environments" />
                </el-tab-pane>
                <el-tab-pane :label="$t('container.mount')"><Volumes :volumes="tensorRTLLM.volumes" /></el-tab-pane>
            </el-tabs>
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
import PortConfig from '@/views/website/runtime/port/index.vue';
import Environment from '@/views/website/runtime/environment/index.vue';
import Volumes from '@/views/website/runtime/volume/index.vue';
import DrawerPro from '@/components/drawer-pro/index.vue';
import FileList from '@/components/file-list/index.vue';

import { reactive, ref } from 'vue';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElForm, FormInstance } from 'element-plus';
import { createTensorRTLLM, updateTensorRTLLM } from '@/api/modules/ai';
import { MsgSuccess } from '@/utils/message';

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
        command: 'bash -c "trtllm-serve /models/ --host 0.0.0.0 --port 8000"',
        exposedPorts: [],
        environments: [],
        volumes: [],
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
    drawerVisiable.value = true;
};

const handleClose = () => {
    drawerVisiable.value = false;
};

const getModelDir = (path: string) => {
    tensorRTLLM.value.modelDir = path;
};

const rules = reactive({
    name: [Rules.requiredInput],
    version: [Rules.requiredInput],
    modelDir: [Rules.requiredInput],
    containerName: [Rules.requiredInput],
    image: [Rules.requiredInput],
    command: [Rules.requiredInput],
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
