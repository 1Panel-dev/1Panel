<template>
    <el-drawer
        v-model="dialogVisible"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        size="30%"
    >
        <template #header>
            <DrawerHeader :header="$t('database.databaseConnInfo')" :back="handleClose" />
        </template>
        <el-form @submit.prevent v-loading="loading" ref="formRef" :model="form" label-position="top">
            <el-row type="flex" justify="center">
                <el-col :span="22">
                    <el-form-item :label="$t('database.containerConn')" v-if="form.from === 'local'">
                        <el-card class="mini-border-card">
                            <el-descriptions :column="1">
                                <el-descriptions-item :label="$t('database.connAddress')">
                                    <el-tooltip
                                        v-if="loadPgInfo(true).length > 48"
                                        :content="loadPgInfo(true)"
                                        placement="top"
                                    >
                                        {{ loadPgInfo(true).substring(0, 48) }}...
                                    </el-tooltip>
                                    <span else>
                                        {{ loadPgInfo(true) }}
                                    </span>
                                    <CopyButton :content="loadPgInfo(true)" type="icon" />
                                </el-descriptions-item>
                                <el-descriptions-item :label="$t('database.connPort')">
                                    5432
                                    <CopyButton content="5432" type="icon" />
                                </el-descriptions-item>
                            </el-descriptions>
                        </el-card>
                        <span class="input-help">
                            {{ $t('database.containerConnHelper') }}
                        </span>
                    </el-form-item>
                    <el-form-item :label="$t('database.remoteConn')">
                        <el-card class="mini-border-card">
                            <el-descriptions :column="1">
                                <el-descriptions-item :label="$t('database.connAddress')">
                                    <el-tooltip
                                        v-if="loadPgInfo(false).length > 48"
                                        :content="loadPgInfo(false)"
                                        placement="top"
                                    >
                                        {{ loadPgInfo(false).substring(0, 48) }}...
                                    </el-tooltip>
                                    <span else>
                                        {{ loadPgInfo(false) }}
                                    </span>
                                    <CopyButton :content="loadPgInfo(false)" type="icon" />
                                </el-descriptions-item>
                                <el-descriptions-item :label="$t('database.connPort')">
                                    {{ form.port }}
                                    <CopyButton :content="form.port + ''" type="icon" />
                                </el-descriptions-item>
                            </el-descriptions>
                        </el-card>
                        <span v-if="form.from === 'local'" class="input-help">
                            {{ $t('database.remoteConnHelper2') }}
                        </span>
                    </el-form-item>

                    <el-divider border-style="dashed" />
                    <div v-if="form.from === 'local'">
                        <el-form-item :label="$t('commons.login.username')" prop="username">
                            <el-input type="text" readonly disabled v-model="form.username">
                                <template #append>
                                    <el-button-group>
                                        <CopyButton :content="form.username" />
                                    </el-button-group>
                                </template>
                            </el-input>
                        </el-form-item>
                        <el-form-item
                            :label="$t('commons.login.password')"
                            :rules="Rules.paramComplexity"
                            prop="password"
                        >
                            <el-input
                                style="width: calc(100% - 205px)"
                                type="password"
                                show-password
                                clearable
                                v-model="form.password"
                            />
                            <el-button-group>
                                <CopyButton class="copy_button" :content="form.password" />
                                <el-button @click="random">
                                    {{ $t('commons.button.random') }}
                                </el-button>
                            </el-button-group>
                        </el-form-item>
                    </div>
                    <div v-if="form.from !== 'local'">
                        <el-form-item :label="$t('commons.login.username')">
                            <el-tag>{{ form.username }}</el-tag>
                            <CopyButton :content="form.username" type="icon" />
                        </el-form-item>
                        <el-form-item :label="$t('commons.login.password')">
                            <el-tag>{{ form.password }}</el-tag>
                            <CopyButton :content="form.password" type="icon" />
                        </el-form-item>
                    </div>
                </el-col>
            </el-row>
        </el-form>

        <ConfirmDialog ref="confirmDialogRef" @confirm="onSubmit" @cancel="loadPassword"></ConfirmDialog>

        <template #footer>
            <span class="dialog-footer">
                <el-button :disabled="loading" @click="dialogVisible = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button :disabled="loading || form.status !== 'Running'" type="primary" @click="onSave(formRef)">
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
    status: '',
    systemIP: '',
    password: '',
    containerName: '',
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

function loadPgInfo(isContainer: boolean) {
    if (isContainer) {
        return form.from === 'local' ? form.containerName : form.systemIP;
    } else {
        return form.from === 'local' ? form.systemIP : form.remoteIP;
    }
}

const random = async () => {
    form.password = getRandomStr(16);
};

const handleClose = () => {
    dialogVisible.value = false;
};

const loadAccess = async () => {
    if (form.from === 'local') {
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
        form.status = res.data.status;
        form.username = res.data.username || '';
        form.password = res.data.password || '';
        form.port = res.data.port || 5432;
        form.containerName = res.data.containerName || '';
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

<style lang="scss" scoped>
.copy_button {
    border-radius: 0px;
    border-left-width: 0px;
}
:deep(.el-input__wrapper) {
    border-top-right-radius: 0px;
    border-bottom-right-radius: 0px;
}
</style>
