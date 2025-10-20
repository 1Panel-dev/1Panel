<template>
    <DrawerPro :header="$t('commons.button.' + mode)" v-model="drawerVisiable" size="large" @close="handleClose">
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
            <el-row :gutter="20">
                <el-col :span="8">
                    <el-form-item :label="$t('commons.table.port')" prop="port">
                        <el-input v-model.number="tensorRTLLM.port" />
                    </el-form-item>
                </el-col>
                <el-col :span="6">
                    <el-form-item :label="$t('app.allowPort')" prop="hostIP">
                        <el-switch
                            v-model="tensorRTLLM.hostIP"
                            :active-value="'0.0.0.0'"
                            :inactive-value="'127.0.0.1'"
                        />
                    </el-form-item>
                </el-col>
            </el-row>

            <el-form-item :label="$t('aiTools.tensorRT.modelDir')" prop="modelDir">
                <el-input v-model="tensorRTLLM.modelDir">
                    <template #prepend>
                        <el-button icon="Folder" @click="modelDirRef.acceptParams({ dir: true })" />
                    </template>
                </el-input>
            </el-form-item>
            <el-form-item :label="$t('aiTools.model.model')" prop="model">
                <el-input v-model="tensorRTLLM.model">
                    <template #prepend>
                        <el-button
                            icon="Folder"
                            @click="modelRef.acceptParams({ path: tensorRTLLM.modelDir, dir: true })"
                        />
                    </template>
                </el-input>
            </el-form-item>
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
        <FileList ref="modelRef" @choose="getModelPath" />
    </DrawerPro>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElForm, FormInstance } from 'element-plus';
import DrawerPro from '@/components/drawer-pro/index.vue';
import FileList from '@/components/file-list/index.vue';
import { createTensorRTLLM, updateTensorRTLLM } from '@/api/modules/ai';
import { MsgSuccess } from '@/utils/message';

const loading = ref(false);
const mode = ref('create');
const drawerVisiable = ref(false);
const newTensorRTLLM = () => {
    return {
        name: '',
        containerName: '',
        port: 8000,
        version: '1.2.0rc0',
        modelDir: '',
        model: '',
        hostIP: '',
        image: 'nvcr.io/nvidia/tensorrt-llm/release',
    };
};
const modelDirRef = ref();
const modelRef = ref();
const tensorRTLLM = ref(newTensorRTLLM());
const emit = defineEmits(['search']);

const openCreate = (): void => {
    mode.value = 'create';
    drawerVisiable.value = true;
};

const openEdit = (rowData: any): void => {
    mode.value = 'edit';
    tensorRTLLM.value = { ...rowData };
    drawerVisiable.value = true;
};

const handleClose = () => {
    drawerVisiable.value = false;
};

const getModelDir = (path: string) => {
    tensorRTLLM.value.modelDir = path;
};

const getModelPath = (path: string) => {
    const modelDir = tensorRTLLM.value.modelDir;
    if (modelDir && path.startsWith(modelDir)) {
        tensorRTLLM.value.model = path.replace(modelDir, '').replace(/^[\/\\]+/, '');
    } else {
        tensorRTLLM.value.model = path;
    }
};

const rules = reactive({
    name: [Rules.requiredInput],
    port: [Rules.requiredInput],
    version: [Rules.requiredInput],
    modelDir: [Rules.requiredInput],
    model: [Rules.requiredInput],
    containerName: [Rules.requiredInput],
    image: [Rules.requiredInput],
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
