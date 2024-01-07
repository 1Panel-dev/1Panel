<template>
    <el-dialog
        v-model="dialogVisible"
        :title="$t('commons.button.delete') + ' - ' + dbName"
        width="30%"
        :close-on-click-modal="false"
    >
        <el-form ref="deleteForm" v-loading="loading" @submit.prevent>
            <el-form-item>
                <el-checkbox v-model="deleteReq.forceDelete" :label="$t('app.forceDelete')" />
                <span class="input-help">
                    {{ $t('app.forceDeleteHelper') }}
                </span>
            </el-form-item>
            <el-form-item>
                <el-checkbox v-model="deleteReq.deleteBackup" :label="$t('app.deleteBackup')" />
                <span class="input-help">
                    {{ $t('database.deleteBackupHelper') }}
                </span>
            </el-form-item>
            <el-form-item>
                <div>
                    <span style="font-size: 12px">{{ $t('database.delete') }}</span>
                    <span style="font-size: 12px; color: red; font-weight: 500">{{ dbName }}</span>
                    <span style="font-size: 12px">{{ $t('database.deleteHelper') }}</span>
                </div>
                <el-input v-model="deleteInfo" :placeholder="dbName"></el-input>
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="dialogVisible = false" :disabled="loading">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button type="primary" @click="submit" :disabled="deleteInfo != dbName || loading">
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
import { deletePostgresqlDB } from '@/api/modules/database';
import { MsgSuccess } from '@/utils/message';

let deleteReq = ref({
    id: 0,
    type: '',
    database: '',
    deleteBackup: false,
    forceDelete: false,
});
let dialogVisible = ref(false);
let loading = ref(false);
let deleteInfo = ref('');
let dbName = ref('');

const deleteForm = ref<FormInstance>();

interface DialogProps {
    id: number;
    type: string;
    name: string;
    database: string;
}
const emit = defineEmits<{ (e: 'search'): void }>();

const acceptParams = async (prop: DialogProps) => {
    deleteReq.value = {
        id: prop.id,
        type: prop.type,
        database: prop.database,
        deleteBackup: false,
        forceDelete: false,
    };
    dbName.value = prop.name;
    deleteInfo.value = '';
    dialogVisible.value = true;
};

const submit = async () => {
    loading.value = true;
    deletePostgresqlDB(deleteReq.value)
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
