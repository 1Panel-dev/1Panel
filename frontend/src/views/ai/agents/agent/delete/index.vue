<template>
    <DialogPro
        v-model="open"
        :title="$t('commons.button.delete') + ' - ' + agentName"
        size="small"
        @close="handleClose"
    >
        <el-form ref="deleteForm" label-position="left" v-loading="loading">
            <el-form-item>
                <el-checkbox v-model="deleteReq.forceDelete" :label="$t('website.forceDelete')" />
                <span class="input-help">
                    {{ $t('website.forceDeleteHelper') }}
                </span>
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit" :loading="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DialogPro>
    <TaskLog ref="taskLogRef" @close="handleClose" />
</template>

<script setup lang="ts">
import { FormInstance } from 'element-plus';
import { ref } from 'vue';
import { deleteAgent } from '@/api/modules/ai';
import { newUUID } from '@/utils/util';
import TaskLog from '@/components/log/task/index.vue';

const open = ref(false);
const loading = ref(false);
const taskLogRef = ref();
const agentName = ref('');
const deleteForm = ref<FormInstance>();
const deleteReq = ref({
    id: 0,
    taskID: '',
    forceDelete: false,
});

const emit = defineEmits(['close']);

const acceptParams = (id: number, name: string) => {
    deleteReq.value = {
        id: id,
        taskID: newUUID(),
        forceDelete: false,
    };
    agentName.value = name;
    open.value = true;
};

const handleClose = () => {
    open.value = false;
    emit('close');
};

const submit = async () => {
    loading.value = true;
    try {
        await deleteAgent(deleteReq.value);
        handleClose();
        taskLogRef.value?.openWithTaskID(deleteReq.value.taskID);
    } finally {
        loading.value = false;
    }
};

defineExpose({
    acceptParams,
});
</script>
