<template>
    <el-dialog
        v-model="open"
        :title="$t('commons.button.recover') + ' - ' + appInstallName"
        width="40%"
        :close-on-click-modal="false"
        :before-close="handleClose"
    >
        <el-form ref="recoverForm" label-position="left" v-loading="loading">
            <el-form-item
                :label="$t('setting.compressPassword')"
                style="margin-top: 10px"
                v-if="recoverReq.type === 'app' || recoverReq.type === 'website'"
            >
                <el-input v-model="recoverReq.secret" :placeholder="$t('setting.backupRecoverMessage')" />
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button type="primary" @click="submit" :disabled="loading">
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
import { handleRecover, handleRecoverByUpload } from '@/api/modules/setting';

let appInstallName = ref('');
let loading = ref(false);
let open = ref(false);
const recoverForm = ref<FormInstance>();

interface DialogProps {
    source: string;
    type: string;
    name: string;
    detailName: string;
    file: string;
    recoverType: string;
}

let recoverReq = ref({
    source: '',
    type: '',
    name: '',
    detailName: '',
    file: '',
    secret: '',
    recoverType: '',
});

const handleClose = () => {
    open.value = false;
};
const acceptParams = (params: DialogProps): void => {
    appInstallName.value = params.name;
    recoverReq.value = {
        source: params.source,
        type: params.type,
        name: params.name,
        detailName: params.detailName,
        file: params.file,
        secret: '',
        recoverType: params.recoverType,
    };
    open.value = true;
};

const submit = async () => {
    let params = {
        source: recoverReq.value.source,
        type: recoverReq.value.type,
        name: recoverReq.value.name,
        detailName: recoverReq.value.detailName,
        file: recoverReq.value.file,
        secret: recoverReq.value.secret,
    };
    loading.value = true;
    if (recoverReq.value.recoverType === 'upload') {
        await handleRecoverByUpload(params)
            .then(() => {
                loading.value = false;
                handleClose();
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            })
            .catch(() => {
                loading.value = false;
            });
        return false;
    }
    await handleRecover(params)
        .then(() => {
            loading.value = false;
            handleClose();
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        })
        .catch(() => {
            loading.value = false;
        });
};

defineExpose({
    acceptParams,
});
</script>
<style scoped lang="scss"></style>
