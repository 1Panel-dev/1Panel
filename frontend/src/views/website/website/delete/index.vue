<template>
    <el-dialog
        v-model="open"
        :close-on-click-modal="false"
        :title="$t('website.delete') + ' - ' + websiteName"
        width="30%"
        :before-close="handleClose"
    >
        <div :key="key" v-loading="loading">
            <el-form ref="deleteForm" label-position="left">
                <el-form-item>
                    <el-checkbox v-model="deleteReq.forceDelete" :label="$t('website.forceDelete')" />
                    <span class="input-help">
                        {{ $t('website.forceDeleteHelper') }}
                    </span>
                </el-form-item>
                <el-form-item v-if="type === 'deployment' || runtimeApp">
                    <el-checkbox
                        v-model="deleteReq.deleteApp"
                        :disabled="runtimeApp"
                        :label="$t('website.deleteApp')"
                    />
                    <span class="input-help">
                        {{ $t('website.deleteAppHelper') }}
                    </span>
                    <span class="input-help" style="color: red" v-if="runtimeApp">
                        {{ $t('website.deleteRuntimeHelper') }}
                    </span>
                </el-form-item>
                <el-form-item>
                    <el-checkbox v-model="deleteReq.deleteBackup" :label="$t('website.deleteBackup')" />
                    <span class="input-help">
                        {{ $t('website.deleteBackupHelper') }}
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
    </el-dialog>
</template>

<script lang="ts" setup>
import { DeleteWebsite } from '@/api/modules/website';
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
const runtimeApp = ref(false);

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
    runtimeApp.value = false;
    if (website.type === 'runtime' && website.appInstallId > 0 && website.runtimeType == 'php') {
        runtimeApp.value = true;
        deleteReq.value.deleteApp = true;
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
    DeleteWebsite(deleteReq.value)
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
