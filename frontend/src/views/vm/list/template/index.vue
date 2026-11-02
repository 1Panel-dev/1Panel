<template>
    <DialogPro v-model="visible" :title="$t('xpack.vm.convertToTemplate') + ' - ' + (source?.name || '')" size="small">
        <el-form ref="formRef" v-loading="storageLoading" label-position="top" :model="form" :rules="rules">
            <el-form-item :label="$t('xpack.vm.templateName')" prop="templateName">
                <el-input v-model.trim="form.templateName" :maxlength="vmNameMaxLength" show-word-limit />
            </el-form-item>
            <el-form-item :label="$t('xpack.vm.storagePool')">
                <el-select v-model="form.storagePool" class="full-width" filterable>
                    <el-option
                        v-for="item in storageItems"
                        :key="item.id"
                        :label="item.path ? `${item.name} (${item.path})` : item.name"
                        :value="item.name"
                        :disabled="isVMResourceUnavailable(item.status)"
                    />
                </el-select>
            </el-form-item>
            <el-form-item :label="$t('commons.table.description')">
                <el-input v-model="form.description" type="textarea" :rows="3" />
            </el-form-item>
        </el-form>
        <template #footer>
            <el-button @click="visible = false">{{ $t('commons.button.cancel') }}</el-button>
            <el-button
                v-permission
                type="primary"
                :loading="loading"
                :disabled="storageLoading || selectedStorageUnavailable"
                @click="submit"
            >
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DialogPro>
</template>

<script setup lang="ts">
import { computed, nextTick, reactive, ref } from 'vue';
import { type FormInstance, type FormRules } from 'element-plus';
import i18n from '@/lang';
import { newUUID } from '@/utils/id';
import { MsgSuccess } from '@/utils/message';
import { VM } from '@/api/interface/vm';
import { isVMResourceUnavailable } from '@/utils/vm';
import { createVMTemplate } from '@/api/modules/vm';
import { Rules } from '@/global/form-rules';

const props = defineProps<{
    storageItems: VM.StorageItem[];
    storageLoading?: boolean;
}>();

const emit = defineEmits<{
    (e: 'task-success', taskID: string): void;
}>();

const visible = ref(false);
const loading = ref(false);
const source = ref<VM.Item>();
const formRef = ref<FormInstance>();
const form = reactive({
    templateName: '',
    storagePool: 'default',
    description: '',
});
const selectedStorageUnavailable = computed(() =>
    isVMResourceUnavailable(props.storageItems.find((item) => item.name === form.storagePool)?.status),
);
const rules = reactive<FormRules>({
    templateName: [Rules.requiredInput, Rules.vmName],
});

const templateSuffix = '-tpl';
const vmNameMaxLength = 64;

const defaultTemplateName = (name: string) => {
    const maxBaseLength = vmNameMaxLength - templateSuffix.length;
    return `${name.slice(0, maxBaseLength)}${templateSuffix}`;
};

const open = (row: VM.Item) => {
    source.value = row;
    form.templateName = defaultTemplateName(row.name);
    form.storagePool = row.storagePool || props.storageItems[0]?.name || 'default';
    form.description = row.description || '';
    visible.value = true;
    nextTick(() => formRef.value?.clearValidate());
};

const submit = async () => {
    if (!source.value || !formRef.value) {
        return;
    }
    if (props.storageLoading || selectedStorageUnavailable.value) {
        return;
    }
    await formRef.value.validate();
    const taskID = newUUID();
    loading.value = true;
    await createVMTemplate({
        taskID,
        name: source.value.name,
        templateName: form.templateName,
        storagePool: form.storagePool,
        description: form.description,
    })
        .then((res) => {
            visible.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            emit('task-success', res.data?.taskID || taskID);
        })
        .finally(() => {
            loading.value = false;
        });
};

defineExpose({ open });
</script>
