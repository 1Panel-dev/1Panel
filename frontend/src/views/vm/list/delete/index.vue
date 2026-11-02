<template>
    <DialogPro v-model="visible" :title="$t('commons.button.delete') + ' - ' + (source?.name || '')" size="small">
        <el-form label-position="top" :model="form">
            <el-form-item>
                <el-checkbox v-model="form.forceDelete">
                    {{ $t('xpack.vm.forceDelete') }}
                </el-checkbox>
                <span class="input-help">{{ $t('xpack.vm.forceDeleteHelper') }}</span>
            </el-form-item>
            <el-form-item>
                <div class="confirm-delete-input">
                    <div class="confirm-delete-input__tip">
                        <span>{{ $t('database.delete') }}</span>
                        <span class="confirm-delete-input__name">{{ source?.name || '' }}</span>
                        <span>{{ $t('xpack.vm.deleteNameHelper') }}</span>
                    </div>
                    <el-input v-model="form.confirmName" :placeholder="source?.name || ''" />
                </div>
            </el-form-item>
        </el-form>
        <template #footer>
            <el-button @click="visible = false">{{ $t('commons.button.cancel') }}</el-button>
            <el-button
                v-permission
                type="primary"
                :loading="loading"
                :disabled="form.confirmName !== source?.name"
                @click="submit"
            >
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DialogPro>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue';
import i18n from '@/lang';
import { newUUID } from '@/utils/id';
import { MsgSuccess } from '@/utils/message';
import { VM } from '@/api/interface/vm';
import { deleteVM } from '@/api/modules/vm';

const emit = defineEmits<{
    (e: 'task-success', taskID: string): void;
}>();

const visible = ref(false);
const loading = ref(false);
const source = ref<VM.Item>();
const form = reactive({
    forceDelete: false,
    confirmName: '',
});

const open = (row: VM.Item) => {
    source.value = row;
    form.forceDelete = false;
    form.confirmName = '';
    visible.value = true;
};

const submit = async () => {
    if (!source.value || form.confirmName !== source.value.name) {
        return;
    }
    const taskID = newUUID();
    loading.value = true;
    await deleteVM({
        taskID,
        id: source.value.id,
        force: form.forceDelete,
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

<style scoped lang="scss">
.confirm-delete-input {
    width: 100%;
}

.confirm-delete-input__tip {
    font-size: 12px;
    line-height: 20px;
    word-break: break-word;
}

.confirm-delete-input__name {
    color: var(--el-color-danger);
    font-weight: 500;
}
</style>
