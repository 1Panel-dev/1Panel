<template>
    <DialogPro v-model="open" :title="$t('website.delete') + ' - ' + websiteName" size="small" @close="handleClose">
        <div :key="key" v-loading="loading">
            <el-form ref="deleteForm" label-position="left">
                <el-form-item>
                    <el-checkbox v-model="deleteReq.forceDelete" :label="$t('website.forceDelete')" />
                    <span class="input-help">
                        {{ $t('website.forceDeleteHelper') }}
                    </span>
                </el-form-item>
                <el-form-item v-if="type === 'deployment'">
                    <el-checkbox v-model="deleteReq.deleteApp" :label="$t('website.deleteApp')" />
                    <span class="input-help">
                        {{ $t('website.deleteAppHelper') }}
                    </span>
                </el-form-item>

                <el-form-item>
                    <el-checkbox v-model="deleteReq.deleteBackup" :label="$t('website.deleteBackup')" />
                    <span class="input-help">
                        {{ $t('website.deleteBackupHelper') }}
                    </span>
                </el-form-item>
                <el-form-item v-if="subSites != ''">
                    <span class="input-help text-red-500">
                        {{ $t('website.deleteSubsite', [subSites]) }}
                    </span>
                </el-form-item>
                <el-form-item>
                    <span v-html="deleteHelper"></span>
                    <el-input v-model="deleteInfo" :placeholder="websiteName" />
                </el-form-item>
            </el-form>
        </div>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit()" :disabled="loading || deleteInfo != websiteName">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DialogPro>
</template>

<script lang="ts" setup>
import { deleteWebsite } from '@/api/modules/website';
import i18n from '@/lang';
import { FormInstance } from 'element-plus';
import { ref } from 'vue';
import { Website } from '@/api/interface/website';
import { MsgSuccess } from '@/utils/message';

const key = 1;
const open = ref(false);
const loading = ref(false);
const deleteReq = ref({
    id: 0,
    deleteApp: false,
    deleteBackup: false,
    forceDelete: false,
});
const type = ref('');
const em = defineEmits(['close']);
const deleteForm = ref<FormInstance>();
const deleteInfo = ref('');
const websiteName = ref('');
const deleteHelper = ref('');
const subSites = ref('');

const handleClose = () => {
    open.value = false;
    em('close', false);
};

const acceptParams = async (website: Website.WebsiteDTO) => {
    deleteReq.value = {
        id: 0,
        deleteApp: false,
        deleteBackup: false,
        forceDelete: false,
    };

    subSites.value = '';
    if (website.childSites && website.childSites.length > 0) {
        subSites.value = website.childSites.join(',');
    }
    deleteInfo.value = '';
    deleteReq.value.id = website.id;
    websiteName.value = website.primaryDomain;
    deleteHelper.value = i18n.global.t('website.deleteConfirmHelper', [website.primaryDomain]);
    type.value = website.type;
    open.value = true;
};

const submit = () => {
    loading.value = true;
    deleteWebsite(deleteReq.value)
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
