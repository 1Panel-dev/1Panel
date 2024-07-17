<template>
    <el-dialog
        v-model="dialogVisible"
        :title="$t('database.unBindRemoteDB') + ' - ' + deleteMysqlReq.database"
        width="30%"
        :close-on-click-modal="false"
    >
        <el-form ref="deleteForm" v-loading="loading" @submit.prevent>
            <el-form-item>
                <el-checkbox v-model="deleteMysqlReq.forceDelete" :label="$t('database.unBindForce')" />
                <span class="input-help">
                    {{ $t('database.unBindForceHelper') }}
                </span>
            </el-form-item>
            <el-form-item>
                <el-checkbox v-model="deleteMysqlReq.deleteBackup" :label="$t('app.deleteBackup')" />
                <span class="input-help">
                    {{ $t('database.deleteBackupHelper') }}
                </span>
            </el-form-item>

            <span style="font-size: 12px">{{ $t('database.unBindRemoteHelper') }}</span>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="dialogVisible = false" :disabled="loading">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button type="primary" @click="submit" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-dialog>
</template>
<script lang="ts" setup>
import { FormInstance } from 'element-plus';
import { ref } from 'vue';
import i18n from '@/lang';
import { deleteDatabase } from '@/api/modules/database';
import { MsgSuccess } from '@/utils/message';

let deleteMysqlReq = ref({
    id: 0,
    database: '',
    deleteBackup: false,
    forceDelete: false,
});
let dialogVisible = ref(false);
let loading = ref(false);

const deleteForm = ref<FormInstance>();

interface DialogProps {
    id: number;
    name: string;
    database: string;
}
const emit = defineEmits<{ (e: 'search'): void }>();

const acceptParams = async (prop: DialogProps) => {
    deleteMysqlReq.value = {
        id: prop.id,
        database: prop.database,
        deleteBackup: false,
        forceDelete: false,
    };
    dialogVisible.value = true;
};

const submit = async () => {
    loading.value = true;
    deleteDatabase(deleteMysqlReq.value)
        .then(() => {
            loading.value = false;
            emit('search');
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            dialogVisible.value = false;
        })
        .catch(() => {
            loading.value = false;
        });
};

defineExpose({
    acceptParams,
});
</script>
