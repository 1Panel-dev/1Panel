<template>
    <DialogPro v-model="open" :title="$t('commons.button.uninstall') + ' - ' + appInstallName" @close="handleClose">
        <el-form ref="deleteForm" label-position="left" v-loading="loading">
            <el-form-item>
                <el-checkbox v-model="deleteReq.forceDelete" :label="$t('app.forceUninstall')" />
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
                <el-checkbox v-model="deleteReq.deleteImage" :label="$t('app.deleteImage')" />
                <span class="input-help">
                    {{ $t('app.deleteImageHelper') }}
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
    </DialogPro>
    <TaskLog ref="taskLogRef" @close="handleClose" />
</template>
<script lang="ts" setup>
import { FormInstance } from 'element-plus';
import { onBeforeUnmount, ref } from 'vue';
import { App } from '@/api/interface/app';
import { installedOp } from '@/api/modules/app';
import i18n from '@/lang';
import bus from '@/global/bus';
import TaskLog from '@/components/log/task/index.vue';
import { v4 as uuidv4 } from 'uuid';

const deleteReq = ref({
    operate: 'delete',
    installId: 0,
    deleteBackup: false,
    forceDelete: false,
    deleteDB: true,
    deleteImage: false,
    taskID: '',
});
const open = ref(false);
const loading = ref(false);
const deleteHelper = ref('');
const deleteInfo = ref('');
const appInstallName = ref('');
const appType = ref('');
const taskLogRef = ref();

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
        deleteImage: false,
        taskID: uuidv4(),
    };
    deleteInfo.value = '';
    deleteReq.value.installId = app.id;
    appType.value = app.appType;
    deleteHelper.value = i18n.global.t('website.deleteConfirmHelper', [app.name]);
    appInstallName.value = app.name;
    open.value = true;
};

const submit = async () => {
    installedOp(deleteReq.value).then(() => {
        handleClose();
        taskLogRef.value.openWithTaskID(deleteReq.value.taskID);
        bus.emit('update', true);
    });
};

onBeforeUnmount(() => {
    bus.off('update');
});

defineExpose({
    acceptParams,
});
</script>
