<template>
    <el-dialog
        v-model="dialogVisiable"
        :title="$t('commons.button.delete') + ' - ' + dbName"
        width="30%"
        :close-on-click-modal="false"
    >
        <el-form ref="deleteForm">
            <el-form-item>
                <el-checkbox v-model="deleteReq.forceDelete" :label="$t('app.forceDelete')" />
                <span class="input-help">
                    {{ $t('app.forceDeleteHelper') }}
                </span>
            </el-form-item>
            <el-form-item>
                <el-checkbox v-model="deleteReq.deleteBackup" :label="$t('app.deleteBackup')" />
                <span class="input-help">
                    {{ $t('app.deleteBackupHelper') }}
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
                <el-button @click="dialogVisiable = false" :loading="loading">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button type="primary" @click="submit" :loading="loading" :disabled="deleteInfo != dbName">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-dialog>
</template>
<script lang="ts" setup>
import { ElMessage, FormInstance } from 'element-plus';
import { ref } from 'vue';
import i18n from '@/lang';
import { deleteMysqlDB } from '@/api/modules/database';

let deleteReq = ref({
    id: 0,
    deleteBackup: false,
    forceDelete: false,
});
let dialogVisiable = ref(false);
let loading = ref(false);
let deleteInfo = ref('');
let dbName = ref('');

const deleteForm = ref<FormInstance>();

interface DialogProps {
    id: number;
    name: string;
}
const emit = defineEmits<{ (e: 'search'): void }>();

const acceptParams = async (prop: DialogProps) => {
    deleteReq.value = {
        id: prop.id,
        deleteBackup: false,
        forceDelete: false,
    };
    dbName.value = prop.name;
    deleteInfo.value = '';
    dialogVisiable.value = true;
};

const submit = async () => {
    loading.value = true;
    deleteMysqlDB(deleteReq.value)
        .then(() => {
            loading.value = false;
            emit('search');
            ElMessage.success(i18n.global.t('commons.msg.deleteSuccess'));
            dialogVisiable.value = false;
        })
        .catch(() => {
            loading.value = false;
        });
};

defineExpose({
    acceptParams,
});
</script>
