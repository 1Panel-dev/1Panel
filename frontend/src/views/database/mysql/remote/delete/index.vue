<template>
    <el-dialog
        v-model="dialogVisible"
        :title="$t('commons.button.delete') + ' - ' + deleteMysqlReq.database"
        width="30%"
        :close-on-click-modal="false"
    >
        <el-form ref="deleteForm" v-loading="loading" @submit.prevent>
            <el-form-item>
                <el-checkbox v-model="deleteMysqlReq.forceDelete" :label="$t('app.forceDelete')" />
                <span class="input-help">
                    {{ $t('app.forceDeleteHelper') }}
                </span>
            </el-form-item>
            <el-form-item>
                <el-checkbox v-model="deleteMysqlReq.deleteBackup" :label="$t('app.deleteBackup')" />
                <span class="input-help">
                    {{ $t('database.deleteBackupHelper') }}
                </span>
            </el-form-item>
            <el-form-item>
                <div>
                    <span style="font-size: 12px">{{ $t('database.delete') }}</span>
                    <span style="font-size: 12px; color: red; font-weight: 500">{{ deleteMysqlReq.database }}</span>
                    <span style="font-size: 12px">{{ $t('database.deleteHelper') }}</span>
                </div>
                <el-input v-model="delMysqlInfo" :placeholder="deleteMysqlReq.database"></el-input>
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="dialogVisible = false" :disabled="loading">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button
                    type="primary"
                    @click="submit"
                    :disabled="delMysqlInfo != deleteMysqlReq.database || loading"
                >
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
let delMysqlInfo = ref('');

const deleteForm = ref<FormInstance>();

interface DialogProps {
    id: number;
    name: string;
    database: string;
}
const emit = defineEmits<{ (e: 'search'): void }>();

const acceptParams = async (prop: DialogProps) => {
    delMysqlInfo.value = '';
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
            MsgSuccess(i18n.global.t('commons.msg.deleteSuccess'));
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
