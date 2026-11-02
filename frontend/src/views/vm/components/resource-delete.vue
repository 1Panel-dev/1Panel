<template>
    <DialogPro
        v-model="visible"
        :title="$t('commons.button.delete') + ' - ' + (source?.name || '')"
        size="small"
        :show-close="!loading"
        :close-on-press-escape="!loading"
    >
        <el-form label-position="top" :model="form">
            <el-form-item>
                <el-checkbox v-model="form.force" :disabled="loading">{{ $t('app.forceDelete') }}</el-checkbox>
                <span class="input-help">{{ $t('app.forceDeleteHelper') }}</span>
            </el-form-item>
            <el-form-item>
                <div class="w-full">
                    <span>{{ $t('commons.msg.operateConfirm') }}</span>
                    <strong class="resource-delete-name">{{ source?.name }}</strong>
                    <el-input v-model="form.confirmName" :placeholder="source?.name" :disabled="loading" />
                </div>
            </el-form-item>
        </el-form>
        <template #footer>
            <el-button :disabled="loading" @click="visible = false">{{ $t('commons.button.cancel') }}</el-button>
            <el-button
                v-permission
                type="primary"
                :loading="loading"
                :disabled="loading || form.confirmName !== source?.name"
                @click="submit"
            >
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DialogPro>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue';
import { VM } from '@/api/interface/vm';
import { deleteVMNetwork, deleteVMStorage } from '@/api/modules/vm';
import { MsgSuccess } from '@/utils/message';
import i18n from '@/lang';

const props = defineProps<{ resourceType: 'network' | 'storagePool' }>();
const emit = defineEmits<{ (e: 'deleted'): void }>();
const source = ref<{ id: number; name: string }>();
const visible = ref(false);
const loading = ref(false);
const form = reactive({ force: false, confirmName: '' });

const open = (row: { id: number; name: string }) => {
    if (loading.value) return;
    source.value = row;
    form.force = false;
    form.confirmName = '';
    visible.value = true;
};

const submit = async () => {
    if (loading.value || !source.value || form.confirmName !== source.value.name) return;
    const req: VM.ResourceDelete = { id: source.value.id, force: form.force };
    loading.value = true;
    try {
        await (props.resourceType === 'network' ? deleteVMNetwork(req) : deleteVMStorage(req));
        visible.value = false;
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        emit('deleted');
    } finally {
        loading.value = false;
    }
};

defineExpose({ open });
</script>

<style scoped>
.resource-delete-name {
    color: var(--el-color-danger);
}
</style>
