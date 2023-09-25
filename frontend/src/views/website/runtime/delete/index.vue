<template>
    <el-dialog
        v-model="open"
        :close-on-click-modal="false"
        :title="$t('commons.button.delete') + ' - ' + resourceName"
        width="30%"
        :before-close="handleClose"
    >
        <div :key="key" :loading="loading">
            <el-form ref="deleteForm" label-position="left">
                <el-form-item>
                    <el-checkbox v-model="deleteReq.forceDelete" :label="$t('website.forceDelete')" />
                    <span class="input-help">
                        {{ $t('website.forceDeleteHelper') }}
                    </span>
                </el-form-item>
            </el-form>
        </div>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit()" :loading="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script lang="ts" setup>
import { DeleteRuntime } from '@/api/modules/runtime';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { FormInstance } from 'element-plus';
import { ref } from 'vue';

const key = 1;
const open = ref(false);
const loading = ref(false);
const deleteReq = ref({
    id: 0,
    forceDelete: false,
});
const em = defineEmits(['close']);
const deleteForm = ref<FormInstance>();
const resourceName = ref('');

const handleClose = () => {
    open.value = false;
    em('close', false);
};

const acceptParams = async (id: number, name: string) => {
    deleteReq.value = {
        id: id,
        forceDelete: false,
    };
    resourceName.value = name;
    open.value = true;
};

const submit = () => {
    loading.value = true;
    DeleteRuntime(deleteReq.value)
        .then(() => {
            handleClose();
            MsgSuccess(i18n.global.t('commons.msg.deleteSuccess'));
        })
        .finally(() => {
            loading.value = false;
        });
};

defineExpose({
    acceptParams,
});
</script>
