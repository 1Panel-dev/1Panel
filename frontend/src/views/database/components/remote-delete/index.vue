<template>
    <DialogPro v-model="open" :title="$t('database.unBindRemoteDB') + ' - ' + deleteReq.database">
        <el-form v-loading="loading" @submit.prevent>
            <el-form-item>
                <el-checkbox v-model="deleteReq.forceDelete" :label="$t('database.unBindForce')" />
                <span class="input-help">
                    {{ $t('database.unBindForceHelper') }}
                </span>
            </el-form-item>
            <el-form-item>
                <el-checkbox v-model="deleteReq.deleteBackup" :label="$t('app.deleteBackup')" />
                <span class="input-help">
                    {{ $t('database.deleteBackupHelper') }}
                </span>
            </el-form-item>

            <span style="font-size: 12px">{{ $t('database.unBindRemoteHelper') }}</span>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="open = false" :disabled="loading">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button type="primary" @click="submit" :loading="loading" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DialogPro>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import i18n from '@/lang';
import { deleteDatabase } from '@/api/modules/database';
import { MsgSuccess } from '@/utils/message';

const deleteReq = ref({
    id: 0,
    database: '',
    deleteBackup: false,
    forceDelete: false,
});
const open = ref(false);
const loading = ref(false);

interface DialogProps {
    id: number;
    database: string;
}

const emit = defineEmits<{ (e: 'search'): void }>();

const acceptParams = (prop: DialogProps) => {
    deleteReq.value = {
        id: prop.id,
        database: prop.database,
        deleteBackup: false,
        forceDelete: false,
    };
    open.value = true;
};

const submit = async () => {
    if (loading.value) return;
    loading.value = true;
    try {
        await deleteDatabase(deleteReq.value);
        emit('search');
        MsgSuccess(i18n.global.t('commons.msg.deleteSuccess'));
        open.value = false;
    } catch {
        // The request interceptor reports failures; allow retry without losing the selection.
    } finally {
        loading.value = false;
    }
};

defineExpose({
    acceptParams,
});
</script>
