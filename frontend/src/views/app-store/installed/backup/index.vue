<template>
    <el-dialog
        v-model="open"
        :title="$t('commons.button.backup') + ' - ' + appInstallName"
        width="40%"
        :close-on-click-modal="false"
        :before-close="handleClose"
    >
        <el-form ref="backupForm" label-position="left" v-loading="loading">
            <el-form-item
                :label="$t('setting.compressPassword')"
                style="margin-top: 10px"
                v-if="backupReq.type === 'app' || backupReq.type === 'website'"
            >
                <el-input v-model="backupReq.secret" :placeholder="$t('setting.backupRecoverMessage')" />
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
import { handleBackup } from '@/api/modules/setting';
import { MsgSuccess } from '@/utils/message';

let appInstallName = ref('');
let loading = ref(false);
let open = ref(false);
const backupForm = ref<FormInstance>();
const emit = defineEmits<{ (e: 'search'): void }>();

interface DialogProps {
    type: string;
    name: string;
    detailName: string;
    status: string;
}

let backupReq = ref({
    type: '',
    name: '',
    detailName: '',
    secret: '',
});

const handleClose = () => {
    open.value = false;
};
const acceptParams = (params: DialogProps): void => {
    appInstallName.value = params.name;
    backupReq.value = {
        type: params.type,
        name: params.name,
        detailName: params.detailName,
        secret: '',
    };
    open.value = true;
};

const submit = async () => {
    let params = {
        type: backupReq.value.type,
        name: backupReq.value.name,
        detailName: backupReq.value.detailName,
        secret: backupReq.value.secret,
    };
    loading.value = true;
    await handleBackup(params)
        .then(() => {
            loading.value = false;
            handleClose();
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            emit('search');
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
