<template>
    <el-dialog
        v-model="open"
        :title="$t('commons.button.recover')"
        width="40%"
        :close-on-click-modal="false"
        :before-close="handleClose"
    >
        <el-form ref="recoverForm" label-position="left" v-loading="loading">
            <el-form-item :label="$t('setting.compressPassword')" style="margin-top: 10px">
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
import { snapshotRecover } from '@/api/modules/setting';

let loading = ref(false);
let open = ref(false);
const recoverForm = ref<FormInstance>();
const emit = defineEmits<{ (e: 'search'): void; (e: 'close'): void }>();

interface DialogProps {
    id: number;
    isNew: boolean;
    reDownload: boolean;
}

let recoverReq = ref({
    id: 0,
    isNew: true,
    reDownload: true,
    secret: '',
});

const handleClose = () => {
    open.value = false;
};
const acceptParams = (params: DialogProps): void => {
    recoverReq.value = {
        id: params.id,
        isNew: params.isNew,
        reDownload: params.reDownload,
        secret: '',
    };
    open.value = true;
};

const submit = async () => {
    loading.value = true;
    await snapshotRecover({
        id: recoverReq.value.id,
        isNew: recoverReq.value.isNew,
        reDownload: recoverReq.value.reDownload,
        secret: recoverReq.value.secret,
    })
        .then(() => {
            emit('search');
            loading.value = false;
            handleClose();
            emit('close');
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
