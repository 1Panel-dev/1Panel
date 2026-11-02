<template>
    <DrawerPro
        v-model="visible"
        :header="$t('commons.button.create')"
        size="large"
        :auto-close="!loading"
        @close="resetForm"
    >
        <el-form ref="formRef" label-position="top" :model="form" :rules="rules">
            <el-form-item :label="$t('commons.table.name')" prop="name">
                <el-input v-model.trim="form.name" />
            </el-form-item>
            <el-form-item :label="$t('commons.table.type')" prop="type">
                <el-tag type="primary">{{ $t('xpack.vm.dir') }}</el-tag>
            </el-form-item>
            <el-form-item :label="$t('xpack.vm.path')" prop="path">
                <el-input v-model.trim="form.path">
                    <template #append>
                        <el-button icon="Folder" @click="pathRef.acceptParams({ dir: true, path: form.path || '/' })" />
                    </template>
                </el-input>
            </el-form-item>
            <el-form-item>
                <el-checkbox v-model="form.autoStart">{{ $t('xpack.vm.resourceAutoStart') }}</el-checkbox>
            </el-form-item>
            <el-form-item :label="$t('commons.table.description')" prop="description">
                <el-input v-model="form.description" type="textarea" :rows="3" />
            </el-form-item>
        </el-form>
        <template #footer>
            <el-button :disabled="loading" @click="visible = false">{{ $t('commons.button.cancel') }}</el-button>
            <el-button v-permission type="primary" :disabled="loading" @click="submit">
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DrawerPro>
    <div v-if="visible && loading" v-loading="loading" class="vm-drawer-loading-mask"></div>
    <FileList ref="pathRef" @choose="loadPath" />
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue';
import { type FormInstance, type FormRules } from 'element-plus';
import FileList from '@/components/file-list/index.vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { VM } from '@/api/interface/vm';
import { createVMStorage } from '@/api/modules/vm';
import { Rules } from '@/global/form-rules';

const emit = defineEmits(['search']);
const visible = ref(false);
const loading = ref(false);
const formRef = ref<FormInstance>();
const pathRef = ref();

const form = reactive<VM.StorageCreate & { id: number }>({
    id: 0,
    name: '',
    type: 'dir',
    path: '',
    autoStart: true,
    description: '',
});

const rules = reactive<FormRules>({
    name: [Rules.requiredInput, Rules.vmNetwork],
    path: [Rules.requiredInput],
});

const acceptParams = () => {
    resetForm();
    visible.value = true;
};

const loadPath = async (path: string) => {
    form.path = path;
    await formRef.value?.validateField('path');
};

const submit = async () => {
    if (!formRef.value) return;
    await formRef.value.validate();
    loading.value = true;
    await createVMStorage({ ...form, path: form.path.trim() })
        .then(() => {
            visible.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        })
        .finally(() => {
            emit('search');
            loading.value = false;
        });
};

const resetForm = () => {
    form.id = 0;
    form.name = '';
    form.type = 'dir';
    form.path = '';
    form.autoStart = true;
    form.description = '';
    formRef.value?.clearValidate();
};

defineExpose({ acceptParams });
</script>

<style scoped lang="scss">
.vm-drawer-loading-mask {
    position: fixed;
    top: 0;
    right: 0;
    bottom: 0;
    z-index: 3000;
    width: 50%;
    background: var(--el-mask-color);
}
</style>
