<template>
    <el-dialog
        v-model="open"
        :title="$t('commons.button.recover')"
        width="40%"
        :close-on-click-modal="false"
        :before-close="handleClose"
    >
        <el-form ref="recoverForm" label-position="top" v-loading="loading">
            <div style="margin-left: 20px; line-height: 32px" v-if="recoverReq.isNew">
                {{ $t('setting.recoverHelper', [recoverReq.name]) }}
                <div>
                    <el-button style="margin-top: -4px" type="warning" link icon="WarningFilled" />
                    {{ $t('setting.recoverHelper1') }}
                </div>
                <div>
                    <el-button
                        style="margin-top: -4px"
                        :type="isSizeOk() ? 'success' : 'danger'"
                        link
                        :icon="isSizeOk() ? 'CircleCheckFilled' : 'CircleCloseFilled'"
                    />
                    {{ $t('setting.recoverHelper2', [computeSize(recoverReq.size), computeSize(recoverReq.freeSize)]) }}
                </div>
                <div>
                    <el-button
                        style="margin-top: -4px"
                        :type="isArchOk() ? 'success' : 'danger'"
                        link
                        :icon="isArchOk() ? 'CircleCheckFilled' : 'CircleCloseFilled'"
                    />
                    {{ $t('setting.recoverHelper3', [recoverReq.arch]) }}
                </div>
            </div>
            <el-form-item v-if="!recoverReq.isNew" :label="$t('setting.recoverFailed')">
                <span>{{ recoverReq.message }}</span>
            </el-form-item>
            <el-form-item v-if="!recoverReq.isNew" :label="$t('setting.snapshotLabel')">
                <el-checkbox v-model="recoverReq.reDownload">{{ $t('setting.reDownload') }}</el-checkbox>
            </el-form-item>
            <el-form-item :label="$t('setting.compressPassword')">
                <el-input v-model="recoverReq.secret" :placeholder="$t('setting.backupRecoverMessage')" />
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button @click="onRollback" v-if="!recoverReq.isNew" :disabled="loading">
                    {{ $t('setting.rollback') }}
                </el-button>
                <el-button type="primary" @click="submit" v-if="!recoverReq.isNew" :disabled="loading">
                    {{ $t('commons.button.retry') }}
                </el-button>
                <el-button type="primary" @click="submit" v-if="recoverReq.isNew" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { FormInstance } from 'element-plus';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { snapshotRecover, snapshotRollback } from '@/api/modules/setting';
import { computeSize, newUUID } from '@/utils/util';

let loading = ref(false);
let open = ref(false);
const recoverForm = ref<FormInstance>();
const emit = defineEmits<{ (e: 'search'): void }>();

interface DialogProps {
    id: number;
    isNew: boolean;
    name: string;
    reDownload: boolean;
    arch: string;
    size: number;
    freeSize: number;
    status: string;
    message: string;
}

let recoverReq = ref({
    id: 0,
    isNew: true,
    name: '',
    reDownload: true,
    secret: '',
    arch: '',
    size: 0,
    freeSize: 0,
    status: '',
    message: '',
});

const handleClose = () => {
    open.value = false;
};
const acceptParams = (params: DialogProps): void => {
    recoverReq.value = {
        id: params.id,
        isNew: params.isNew,
        name: params.name,
        reDownload: params.reDownload,
        secret: '',
        arch: params.arch,
        size: params.size,
        freeSize: params.freeSize,
        status: params.status,
        message: params.message,
    };
    open.value = true;
};

const isSizeOk = () => {
    if (recoverReq.value.size === 0 || recoverReq.value.freeSize === 0) {
        return false;
    }
    return recoverReq.value.size * 2 < recoverReq.value.freeSize;
};
const isArchOk = () => {
    if (recoverReq.value.arch.length === 0) {
        return false;
    }
    return recoverReq.value.name.indexOf(recoverReq.value.arch) !== -1;
};

const submit = async () => {
    loading.value = true;
    await snapshotRecover({
        id: recoverReq.value.id,
        taskID: newUUID(),
        isNew: recoverReq.value.isNew,
        reDownload: recoverReq.value.reDownload,
        secret: recoverReq.value.secret,
    })
        .then(() => {
            emit('search');
            loading.value = false;
            handleClose();
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        })
        .catch(() => {
            loading.value = false;
        });
};

const onRollback = async (row: any) => {
    ElMessageBox.confirm(i18n.global.t('setting.rollbackHelper'), i18n.global.t('setting.rollback'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        loading.value = true;
        await snapshotRollback({ id: row.id, taskID: newUUID(), isNew: false, reDownload: false, secret: '' })
            .then(() => {
                emit('search');
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

defineExpose({
    acceptParams,
});
</script>
<style scoped lang="scss"></style>
