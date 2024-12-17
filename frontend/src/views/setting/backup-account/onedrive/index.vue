<template>
    <div>
        <el-drawer
            v-model="drawerVisible"
            :destroy-on-close="true"
            :close-on-click-modal="false"
            :close-on-press-escape="false"
            size="50%"
        >
            <template #header>
                <DrawerHeader :header="title + $t('setting.backupAccount')" :back="handleClose" />
            </template>
            <el-form
                @submit.prevent
                ref="formRef"
                v-loading="loading"
                label-position="top"
                :model="oneDriveData.rowData"
            >
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item :label="$t('commons.table.type')" prop="type">
                            <el-tag>{{ $t('setting.' + oneDriveData.rowData!.type) }}</el-tag>
                        </el-form-item>
                        <el-form-item>
                            <el-radio-group v-model="oneDriveData.rowData!.varsJson['isCN']" @change="changeFrom">
                                <el-radio-button :value="false">{{ $t('setting.isNotCN') }}</el-radio-button>
                                <el-radio-button :value="true">{{ $t('setting.isCN') }}</el-radio-button>
                            </el-radio-group>
                            <span class="input-help">
                                {{ $t('setting.onedrive_helper') }}
                                <el-link
                                    style="font-size: 12px; margin-left: 5px"
                                    icon="Position"
                                    @click="toDoc(true)"
                                    type="primary"
                                >
                                    {{ $t('firewall.quickJump') }}
                                </el-link>
                            </span>
                        </el-form-item>
                        <el-form-item
                            :label="$t('setting.client_id')"
                            prop="varsJson.client_id"
                            :rules="Rules.requiredInput"
                        >
                            <el-input v-model.trim="oneDriveData.rowData!.varsJson['client_id']" />
                        </el-form-item>
                        <el-form-item
                            :label="$t('setting.client_secret')"
                            prop="varsJson.client_secret"
                            :rules="Rules.requiredInput"
                        >
                            <el-input v-model.trim="oneDriveData.rowData!.varsJson['client_secret']" />
                        </el-form-item>
                        <el-form-item
                            :label="$t('setting.redirect_uri')"
                            prop="varsJson.redirect_uri"
                            :rules="Rules.requiredInput"
                        >
                            <el-input v-model.trim="oneDriveData.rowData!.varsJson['redirect_uri']" />
                        </el-form-item>
                        <el-form-item :label="$t('setting.code')" prop="varsJson.code" :rules="rules.driveCode">
                            <div style="width: 100%">
                                <el-input
                                    style="width: calc(100% - 80px)"
                                    :rows="3"
                                    type="textarea"
                                    clearable
                                    v-model.trim="oneDriveData.rowData!.varsJson['code']"
                                />
                                <el-button class="append-button" @click="jumpAzure(formRef)">
                                    {{ $t('setting.loadCode') }}
                                </el-button>
                            </div>
                            <span class="input-help">
                                {{ $t('setting.codeHelper') }}
                                <el-link
                                    style="font-size: 12px; margin-left: 5px"
                                    icon="Position"
                                    @click="toDoc(false)"
                                    type="primary"
                                >
                                    {{ $t('firewall.quickJump') }}
                                </el-link>
                            </span>
                        </el-form-item>
                        <el-form-item :label="$t('setting.backupDir')" prop="backupPath">
                            <el-input clearable v-model.trim="oneDriveData.rowData!.backupPath" placeholder="/1panel" />
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
import { GlobalStore } from '@/store';

const globalStore = GlobalStore();

const loading = ref(false);
type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

const rules = reactive({
    driveCode: [{ validator: checkDriveCode, required: true, trigger: 'blur' }],
});
function checkDriveCode(rule: any, value: any, callback: any) {
    if (!value) {
        return callback(new Error(i18n.global.t('setting.codeWarning')));
    }
    const reg = /^[A-Za-z0-9_.-]+$/;
    if (!reg.test(value)) {
        return callback(new Error(i18n.global.t('setting.codeWarning')));
    }
    callback();
}

const emit = defineEmits(['search']);
const oneDriveInfo = ref();

interface DialogProps {
    title: string;
    rowData?: Backup.BackupInfo;
}
const title = ref<string>('');
const drawerVisible = ref(false);
const oneDriveData = ref<DialogProps>({
    title: '',
});
const acceptParams = async (params: DialogProps): Promise<void> => {
    oneDriveData.value = params;
    oneDriveData.value.rowData.varsJson['isCN'] = oneDriveData.value.rowData.varsJson['isCN'] || false;
    title.value = i18n.global.t('commons.button.' + oneDriveData.value.title);
    drawerVisible.value = true;
    const res = await getOneDriveInfo();
    oneDriveInfo.value = res.data;
    if (!oneDriveData.value.rowData.id) {
        oneDriveData.value.rowData.varsJson = {
            isCN: false,
            client_id: res.data.client_id,
            client_secret: res.data.client_secret,
            redirect_uri: res.data.redirect_uri,
        };
    }
};

const changeFrom = () => {
    if (oneDriveData.value.rowData.varsJson['isCN']) {
        oneDriveData.value.rowData.varsJson = {
            isCN: true,
            client_id: '',
            client_secret: '',
            redirect_uri: '',
        };
    } else {
        oneDriveData.value.rowData.varsJson = {
            isCN: false,
            client_id: oneDriveInfo.value.client_id,
            client_secret: oneDriveInfo.value.client_secret,
            redirect_uri: oneDriveInfo.value.redirect_uri,
        };
    }
};

const handleClose = () => {
    emit('search');
    drawerVisible.value = false;
};
const jumpAzure = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    const result = await formEl.validateField('varsJson.client_id', callback);
    if (!result) {
        return;
    }
    const result1 = await formEl.validateField('varsJson.redirect_uri', callback);
    if (!result1) {
        return;
    }
    let client_id = oneDriveData.value.rowData.varsJson['client_id'];
    let redirect_uri = oneDriveData.value.rowData.varsJson['redirect_uri'];
    let commonUrl = `response_type=code&client_id=${client_id}&redirect_uri=${redirect_uri}&scope=offline_access+Files.ReadWrite.All+User.Read`;
    if (!oneDriveData.value.rowData!.varsJson['isCN']) {
        window.open('https://login.microsoftonline.com/common/oauth2/v2.0/authorize?' + commonUrl, '_blank');
    } else {
        window.open('https://login.chinacloudapi.cn/common/oauth2/v2.0/authorize?' + commonUrl, '_blank');
    }
};

function callback(error: any) {
    if (error) {
        return error.message;
    } else {
        return;
    }
}

const toDoc = (isConf: boolean) => {
    let item = isConf ? '#32-onedrive' : '#33-onedrive';
    if (globalStore.isIntl) {
        item = isConf ? '#using-your-own-client-info-for-onedrive' : '#auth-code-of-onedrive';
    }
    window.open(globalStore.docsUrl + '/user_manual/settings/' + item, '_blank', 'noopener,noreferrer');
};

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        if (!oneDriveData.value.rowData) return;
        oneDriveData.value.rowData.vars = JSON.stringify(oneDriveData.value.rowData!.varsJson);
        loading.value = true;
        if (oneDriveData.value.title === 'create') {
            await addBackup(oneDriveData.value.rowData)
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
        await editBackup(oneDriveData.value.rowData)
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
