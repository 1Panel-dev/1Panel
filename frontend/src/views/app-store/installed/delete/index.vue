<template>
    <el-dialog
        v-model="open"
        :title="$t('commons.button.delete') + ' - ' + appInstallName"
        width="40%"
        :close-on-click-modal="false"
        :before-close="handleClose"
    >
        <el-form ref="deleteForm" label-position="left" v-loading="loading">
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
            <el-form-item v-if="appType === 'website'">
                <el-checkbox v-model="deleteReq.deleteDB" :label="$t('app.deleteDB')" />
                <span class="input-help">
                    {{ $t('app.deleteDBHelper') }}
                </span>
            </el-form-item>
            <el-form-item>
                <span v-html="deleteHelper"></span>
                <el-input v-model="deleteInfo" :placeholder="appInstallName" />
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button type="primary" @click="submit" :disabled="loading || deleteInfo !== appInstallName">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-dialog>
</template>
<script lang="ts" setup>
import { FormInstance } from 'element-plus';
import { onBeforeUnmount, ref } from 'vue';
import { App } from '@/api/interface/app';
import { InstalledOp } from '@/api/modules/app';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import bus from '../../bus';

let deleteReq = ref({
    operate: 'delete',
    installId: 0,
    deleteBackup: false,
    forceDelete: false,
    deleteDB: true,
});
let open = ref(false);
let loading = ref(false);
let deleteHelper = ref('');
let deleteInfo = ref('');
let appInstallName = ref('');
let appType = ref('');

const deleteForm = ref<FormInstance>();
const em = defineEmits(['close']);

const handleClose = () => {
    open.value = false;
    em('close', open);
};

const acceptParams = async (app: App.AppInstallDto) => {
    deleteReq.value = {
        operate: 'delete',
        installId: 0,
        deleteBackup: false,
        forceDelete: false,
        deleteDB: true,
    };
    deleteInfo.value = '';
    deleteReq.value.installId = app.id;
    appType.value = app.appType;
    deleteHelper.value = i18n.global.t('website.deleteConfirmHelper', [app.name]);
    appInstallName.value = app.name;
    open.value = true;
};

const submit = async () => {
    loading.value = true;
    InstalledOp(deleteReq.value)
        .then(() => {
            handleClose();
            MsgSuccess(i18n.global.t('commons.msg.deleteSuccess'));
            bus.emit('update', true);
        })
        .finally(() => {
            loading.value = false;
        });
};

onBeforeUnmount(() => {
    bus.off('update');
});

defineExpose({
    acceptParams,
});
</script>
