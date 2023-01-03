<template>
    <el-dialog
        v-model="open"
        :title="$t('website.renewSSL')"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        width="30%"
        :before-close="handleClose"
    >
        <div style="text-align: center">
            <span>{{ $t('ssl.renewConfirm') }}</span>
        </div>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit()" :disabled="loading" :loading="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script lang="ts" setup>
import { RenewSSL } from '@/api/modules/website';
import i18n from '@/lang';
import { ElMessage } from 'element-plus';
import { reactive, ref } from 'vue';

let open = ref(false);
let loading = ref(false);
let renewReq = reactive({
    SSLId: 0,
});
const em = defineEmits(['close']);

const handleClose = () => {
    open.value = false;
    em('close', false);
};

const acceptParams = async (id: number) => {
    renewReq.SSLId = id;
    open.value = true;
};

const submit = () => {
    loading.value = true;
    RenewSSL(renewReq)
        .then(() => {
            handleClose();
            ElMessage.success(i18n.global.t('commons.msg.renewSuccess'));
        })
        .finally(() => {
            loading.value = false;
        });
};

defineExpose({
    acceptParams,
});
</script>
