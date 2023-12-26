<template>
    <div>
        <el-drawer v-model="drawerVisible" :destroy-on-close="true" :close-on-click-modal="false" size="50%">
            <template #header>
                <DrawerHeader :header="title + $t('setting.backupAccount')" :back="handleClose" />
            </template>
            <el-form
                @submit.prevent
                ref="formRef"
                v-loading="loading"
                label-position="top"
                :model="onedriveData.rowData"
            >
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item :label="$t('commons.table.type')" prop="type" :rules="Rules.requiredSelect">
                            <el-tag>{{ $t('setting.' + onedriveData.rowData!.type) }}</el-tag>
                        </el-form-item>
                        <el-form-item>
                            <el-checkbox
                                disabled
                                v-model="onedriveData.rowData!.varsJson['isCN']"
                                :label="$t('setting.isCN')"
                            />
                        </el-form-item>
                        <el-form-item :label="$t('setting.code')" prop="varsJson.code" :rules="rules.driveCode">
                            <div style="width: 100%">
                                <el-input
                                    style="width: calc(100% - 80px)"
                                    :autosize="{ minRows: 3, maxRows: 15 }"
                                    type="textarea"
                                    clearable
                                    v-model.trim="onedriveData.rowData!.varsJson['code']"
                                />
                                <el-button class="append-button" @click="jumpAzure">
                                    {{ $t('setting.loadCode') }}
                                </el-button>
                            </div>
                            <span class="input-help">
                                {{ $t('setting.codeHelper') }}
                                <el-link
                                    style="font-size: 12px; margin-left: 5px"
                                    icon="Position"
                                    @click="toDoc()"
                                    type="primary"
                                >
                                    {{ $t('firewall.quickJump') }}
                                </el-link>
                            </span>
                        </el-form-item>
                        <el-form-item :label="$t('setting.backupDir')" prop="backupPath">
                            <el-input clearable v-model.trim="onedriveData.rowData!.backupPath" placeholder="/1panel" />
                        </el-form-item>
                    </el-col>
                </el-row>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button :disabled="loading" @click="handleClose">
                        {{ $t('commons.button.cancel') }}
                    </el-button>
                    <el-button :disabled="loading" type="primary" @click="onSubmit(formRef)">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-drawer>
    </div>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import { Backup } from '@/api/interface/backup';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { addBackup, editBackup, getOneDriveInfo } from '@/api/modules/setting';
import { MsgSuccess } from '@/utils/message';

const loading = ref(false);
type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

const rules = reactive({
    driveCode: [{ validator: checkDriveCode, required: true, trigger: 'blur' }],
});
function checkDriveCode(rule: any, value: any, callback: any) {
    const reg = /^[A-Za-z0-9_.-]+$/;
    if (!reg.test(value)) {
        return callback(new Error(i18n.global.t('setting.codeWarning')));
    }
    callback();
}

const emit = defineEmits(['search']);

interface DialogProps {
    title: string;
    rowData?: Backup.BackupInfo;
}
const title = ref<string>('');
const drawerVisible = ref(false);
const onedriveData = ref<DialogProps>({
    title: '',
});
const acceptParams = (params: DialogProps): void => {
    onedriveData.value = params;
    title.value = i18n.global.t('commons.button.' + onedriveData.value.title);
    drawerVisible.value = true;
};

const handleClose = () => {
    emit('search');
    drawerVisible.value = false;
};
const jumpAzure = async () => {
    const res = await getOneDriveInfo();
    let commonUrl = `response_type=code&client_id=${res.data}&redirect_uri=http://localhost/login/authorized&scope=offline_access+Files.ReadWrite.All+User.Read`;
    if (!onedriveData.value.rowData!.varsJson['isCN']) {
        window.open('https://login.microsoftonline.com/common/oauth2/v2.0/authorize?' + commonUrl, '_blank');
    } else {
        window.open('https://login.chinacloudapi.cn/common/oauth2/v2.0/authorize?' + commonUrl, '_blank');
    }
};

const toDoc = () => {
    window.open('https://1panel.cn/docs/user_manual/settings/', '_blank', 'noopener,noreferrer');
};

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        if (!onedriveData.value.rowData) return;
        onedriveData.value.rowData.vars = JSON.stringify(onedriveData.value.rowData!.varsJson);
        loading.value = true;
        if (onedriveData.value.title === 'create') {
            await addBackup(onedriveData.value.rowData)
                .then(() => {
                    loading.value = false;
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                    emit('search');
                    drawerVisible.value = false;
                })
                .catch(() => {
                    loading.value = false;
                });
            return;
        }
        await editBackup(onedriveData.value.rowData)
            .then(() => {
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                emit('search');
                drawerVisible.value = false;
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

<style lang="scss" scoped>
.append-button {
    width: 80px;
    background-color: var(--el-fill-color-light);
    color: var(--el-color-info);
}
</style>
