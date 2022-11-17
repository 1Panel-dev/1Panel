<template>
    <el-dialog v-model="open" :title="$t('website.delete')" width="40%" :before-close="handleClose">
        <div style="text-align: center">
            <el-checkbox v-model="deleteReq.deleteApp" :label="$t('website.deleteApp')" />
            <el-checkbox v-model="deleteReq.deleteBackup" :label="$t('website.deleteBackup')" />
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
import { DeleteWebsite } from '@/api/modules/website';
import i18n from '@/lang';
import { ElMessage } from 'element-plus';
import { reactive, ref } from 'vue';

// interface DeleteProps {
//     id: number;
// }
// const deleteData = ref<DeleteProps>({
//     id: 0,
// });

let open = ref(false);
let loading = ref(false);
let deleteReq = reactive({
    id: 0,
    deleteApp: false,
    deleteBackup: false,
});
const em = defineEmits(['close']);

const handleClose = () => {
    open.value = false;
    em('close', false);
};

const acceptParams = async (id: number) => {
    deleteReq.id = id;
    open.value = true;
};

const submit = () => {
    loading.value = true;
    DeleteWebsite(deleteReq)
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
