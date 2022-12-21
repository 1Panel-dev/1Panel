<template>
    <el-dialog
        v-model="open"
        :title="$t('commons.button.delete') + ' - ' + appInstallName"
        width="30%"
        :close-on-click-modal="false"
        :before-close="handleClose"
    >
        <el-form ref="deleteForm" label-position="left">
            <el-form-item>
                <el-checkbox v-model="deleteReq.forceDelete" :label="$t('app.forceDelete')" />
            </el-form-item>
            <div class="helper">
                <span class="input-help">
                    {{ $t('app.forceDeleteHelper') }}
                </span>
            </div>
            <el-form-item>
                <el-checkbox v-model="deleteReq.deleteBackup" :label="$t('app.deleteBackup')" />
            </el-form-item>
            <div class="helper">
                <span class="input-help">
                    {{ $t('app.deleteBackupHelper') }}
                </span>
            </div>
            <br />
            <span v-html="deleteHelper"></span>
            <el-form-item>
                <el-input v-model="deleteInfo" :placeholder="appInstallName" />
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :loading="loading">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button type="primary" @click="submit" :loading="loading" :disabled="deleteInfo != appInstallName">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-dialog>
</template>
<script lang="ts" setup>
import { ElMessage, FormInstance } from 'element-plus';
import { ref } from 'vue';
import { App } from '@/api/interface/app';
import { InstalledOp } from '@/api/modules/app';
import i18n from '@/lang';

let deleteReq = ref({
    operate: 'delete',
    installId: 0,
    deleteBackup: false,
    forceDelete: false,
});
let open = ref(false);
let loading = ref(false);
let deleteHelper = ref('');
let deleteInfo = ref('');
let appInstallName = ref('');

const deleteForm = ref<FormInstance>();
const em = defineEmits(['close']);

const handleClose = () => {
    open.value = false;
    em('close', open);
};

const acceptParams = async (app: App.AppInstalled) => {
    deleteReq.value = {
        operate: 'delete',
        installId: 0,
        deleteBackup: false,
        forceDelete: false,
    };
    deleteInfo.value = '';
    deleteReq.value.installId = app.id;
    deleteHelper.value = i18n.global.t('website.deleteConfirmHelper', [app.name]);
    appInstallName.value = app.name;
    open.value = true;
};

const submit = async () => {
    loading.value = true;
    InstalledOp(deleteReq.value)
        .then(() => {
            handleClose();
            ElMessage.success(i18n.global.t('commons.msg.deleteSuccess'));
        })
        .finally(() => {
            loading.value = false;
        });
};

defineExpose({
    acceptParams,
});
</script>

<style lang="scss">
.helper {
    margin-top: -20px;
}
</style>
