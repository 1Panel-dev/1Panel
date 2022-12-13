<template>
    <el-dialog
        v-model="open"
        :close-on-click-modal="false"
        :title="$t('website.delete')"
        width="40%"
        :before-close="handleClose"
    >
        <div style="text-align: center" :key="key">
            <el-checkbox v-model="deleteReq.forceDelete" :label="$t('website.forceDelete')" />
            <el-checkbox v-if="type === 'deployment'" v-model="deleteReq.deleteApp" :label="$t('website.deleteApp')" />
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
import { ref } from 'vue';
import { Website } from '@/api/interface/website';

let key = 1;
let open = ref(false);
let loading = ref(false);
let deleteReq = ref({
    id: 0,
    deleteApp: false,
    deleteBackup: false,
    forceDelete: false,
});
let type = ref('');
const em = defineEmits(['close']);

const handleClose = () => {
    open.value = false;
    em('close', false);
};

const acceptParams = async (website: Website.Website) => {
    deleteReq.value = {
        id: 0,
        deleteApp: false,
        deleteBackup: false,
        forceDelete: false,
    };
    deleteReq.value.id = website.id;
    type.value = website.type;
    open.value = true;
};

const submit = () => {
    loading.value = true;
    DeleteWebsite(deleteReq.value)
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
