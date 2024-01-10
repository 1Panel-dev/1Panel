<template>
    <el-drawer v-model="dialogVisible" :destroy-on-close="true" :close-on-click-modal="false" size="30%">
        <template #header>
            <DrawerHeader :header="$t('database.databaseConnInfo')" :back="handleClose" />
        </template>
        <el-form @submit.prevent v-loading="loading" ref="formRef" :model="form" label-position="top">
            <el-row type="flex" justify="center" v-if="form.from === 'local'">
                <el-col :span="22">
                    <el-form-item :label="$t('database.containerConn')">
                        <el-tag>
                            {{ form.serviceName + ':' + form.port }}
                        </el-tag>
                        <CopyButton :content="form.serviceName + ':' + form.port" type="icon" />
                        <span class="input-help">
                            {{ $t('database.containerConnHelper') }}
                        </span>
                    </el-form-item>
                    <el-form-item :label="$t('database.remoteConn')">
                        <el-tooltip v-if="loadConnInfo(true).length > 48" :content="loadConnInfo(true)" placement="top">
                            <el-tag>{{ loadConnInfo(true).substring(0, 48) }}...</el-tag>
                        </el-tooltip>
                        <el-tag v-else>{{ loadConnInfo(true) }}</el-tag>
                        <CopyButton :content="form.systemIP + ':' + form.port" type="icon" />
                        <span class="input-help">{{ $t('database.remoteConnHelper2') }}</span>
                    </el-form-item>

                    <el-divider border-style="dashed" />
                    <el-form-item :label="$t('commons.login.username')" prop="username">
                        <el-input type="text" readonly disabled v-model="form.username">
                            <template #append>
                                <el-button-group>
                                    <CopyButton :content="form.username" />
                                </el-button-group>
                            </template>
                        </el-input>
                    </el-form-item>
                    <el-form-item :label="$t('commons.login.password')" :rules="Rules.paramComplexity" prop="password">
                        <el-input type="password" show-password clearable v-model="form.password">
                            <template #append>
                                <CopyButton :content="form.password" />
                                <el-button @click="random" class="p-ml-5">
                                    {{ $t('commons.button.random') }}
                                </el-button>
                            </template>
                        </el-input>
                    </el-form-item>
                </el-col>
            </el-row>
            <el-row type="flex" justify="center" v-if="form.from !== 'local'">
                <el-col :span="22">
                    <el-form-item :label="$t('database.remoteConn')">
                        <el-tooltip
                            v-if="loadConnInfo(false).length > 48"
                            :content="loadConnInfo(false)"
                            placement="top"
                        >
                            <el-tag>{{ loadConnInfo(false).substring(0, 48) }}...</el-tag>
                        </el-tooltip>
                        <el-tag v-else>{{ loadConnInfo(false) }}</el-tag>
                        <CopyButton :content="form.remoteIP + ':' + form.port" type="icon" />
                    </el-form-item>
                    <el-form-item :label="$t('commons.login.username')">
                        <el-tag>{{ form.username }}</el-tag>
                        <CopyButton :content="form.username" type="icon" />
                    </el-form-item>
                    <el-form-item :label="$t('commons.login.password')">
                        <el-tag>{{ form.password }}</el-tag>
                        <CopyButton :content="form.password" type="icon" />
                    </el-form-item>
                </el-col>
            </el-row>
        </el-form>

        <ConfirmDialog ref="confirmDialogRef" @confirm="onSubmit" @cancel="loadPassword"></ConfirmDialog>

        <template #footer>
            <span class="dialog-footer">
                <el-button :disabled="loading" @click="dialogVisible = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button :disabled="loading" type="primary" @click="onSave(formRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-drawer>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import { getDatabase, updatePostgresqlPassword } from '@/api/modules/database';
import ConfirmDialog from '@/components/confirm-dialog/index.vue';
import { GetAppConnInfo } from '@/api/modules/app';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { MsgSuccess } from '@/utils/message';
import { getRandomStr } from '@/utils/util';
import { getSettingInfo } from '@/api/modules/setting';

const loading = ref(false);

const dialogVisible = ref(false);
const form = reactive({
    systemIP: '',
    password: '',
    serviceName: '',
    privilege: false,
    port: 0,

    from: '',
    type: '',
    database: '',
    username: '',
    remoteIP: '',
});

const confirmDialogRef = ref();

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

interface DialogProps {
    from: string;
    type: string;
    database: string;
}

const acceptParams = (param: DialogProps): void => {
    form.password = '';
    form.from = param.from;
    form.type = param.type;
    form.database = param.database;
    loadAccess();
    loadPassword();
    dialogVisible.value = true;
};

function loadConnInfo(isLocal: boolean) {
    let ip = isLocal ? form.systemIP : form.remoteIP;
    let info = ip + ':' + form.port;
    return info;
}

const random = async () => {
    form.password = getRandomStr(16);
};

const handleClose = () => {
    dialogVisible.value = false;
};

const loadAccess = async () => {
    if (form.from === 'local') {
        // const res = await loadRemoteAccess(form.type, form.database);
        form.privilege = false;
    }
};

const loadSystemIP = async () => {
    const res = await getSettingInfo();
    form.systemIP = res.data.systemIP || i18n.global.t('database.localIP');
};

const loadPassword = async () => {
    if (form.from === 'local') {
        const res = await GetAppConnInfo(form.type, form.database);
        form.username = res.data.username || '';
        form.password = res.data.password || '';
        form.port = res.data.port || 5432;
        form.serviceName = res.data.serviceName || '';
        loadSystemIP();
        return;
    }
    const res = await getDatabase(form.database);
    form.password = res.data.password || '';
    form.port = res.data.port || 5432;
    form.username = res.data.username;
    form.password = res.data.password;
    form.remoteIP = res.data.address;
};

const onSubmit = async () => {
    let param = {
        id: 0,
        from: form.from,
        type: form.type,
        database: form.database,
        value: form.password,
    };
    loading.value = true;
    await updatePostgresqlPassword(param)
        .then(() => {
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            dialogVisible.value = false;
        })
        .catch(() => {
            loading.value = false;
        });
};

const onSave = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        let params = {
            header: i18n.global.t('database.confChange'),
            operationInfo: i18n.global.t('database.restartNowHelper'),
            submitInputInfo: i18n.global.t('database.restartNow'),
        };
        confirmDialogRef.value!.acceptParams(params);
    });
};

defineExpose({
    acceptParams,
});
</script>
