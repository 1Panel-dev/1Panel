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
        <el-form @submit.prevent v-loading="loading" ref="formRef" :model="form" label-position="top" :rules="rules">
            <el-row type="flex" justify="center">
                <el-col :span="22">
                    <el-form-item :label="$t('database.containerConn')" v-if="form.from === 'local'">
                        <el-card class="mini-border-card">
                            <el-descriptions :column="1">
                                <el-descriptions-item :label="$t('database.connAddress')">
                                    <el-tooltip
                                        v-if="loadRedisInfo(true).length > 48"
                                        :content="loadRedisInfo(true)"
                                        placement="top"
                                    >
                                        {{ loadRedisInfo(true).substring(0, 48) }}...
                                    </el-tooltip>
                                    <span else>
                                        {{ loadRedisInfo(true) }}
                                    </span>
                                    <CopyButton :content="loadRedisInfo(true)" type="icon" />
                                </el-descriptions-item>
                                <el-descriptions-item :label="$t('database.connPort')">
                                    6379
                                    <CopyButton content="6379" type="icon" />
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
                                        v-if="loadRedisInfo(false).length > 48"
                                        :content="loadRedisInfo(false)"
                                        placement="top"
                                    >
                                        {{ loadRedisInfo(false).substring(0, 48) }}...
                                    </el-tooltip>
                                    <span else>
                                        {{ loadRedisInfo(false) }}
                                    </span>
                                    <CopyButton :content="loadRedisInfo(false)" type="icon" />
                                </el-descriptions-item>
                                <el-descriptions-item :label="$t('database.connPort')">
                                    {{ form.port }}
                                    <CopyButton :content="form.port + ''" type="icon" />
                                </el-descriptions-item>
                            </el-descriptions>
                        </el-card>
                        <span class="input-help">
                            {{ $t('database.remoteConnHelper2') }}
                        </span>
                    </el-form-item>

                    <el-divider border-style="dashed" />
                    <el-form-item :label="$t('commons.login.password')" v-if="form.from === 'local'" prop="password">
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

                    <div v-if="form.from !== 'local'">
                        <el-form-item :label="$t('commons.login.password')">
                            <el-tag>{{ form.password }}</el-tag>
                            <CopyButton :content="form.password" type="icon" />
                        </el-form-item>
                    </div>
                </el-col>
            </el-row>
        </el-form>

        <ConfirmDialog ref="confirmDialogRef" @confirm="onSubmit"></ConfirmDialog>

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
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import { changeRedisPassword, getDatabase } from '@/api/modules/database';
import ConfirmDialog from '@/components/confirm-dialog/index.vue';
import { GetAppConnInfo } from '@/api/modules/app';
import { MsgSuccess } from '@/utils/message';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { getRandomStr } from '@/utils/util';
import { getSettingInfo } from '@/api/modules/setting';

const loading = ref(false);

const dialogVisible = ref(false);
const form = reactive({
    status: '',
    systemIP: '',
    password: '',
    serviceName: '',
    containerName: '',
    port: 0,

    from: '',
    database: '',
    remoteIP: '',
});
const rules = reactive({
    password: [{ validator: checkPassword, trigger: 'blur' }],
});

function checkPassword(rule: any, value: any, callback: any) {
    if (form.password !== '') {
        const reg = /^[a-zA-Z0-9]{1}[a-zA-Z0-9.%@!~_-]{4,126}[a-zA-Z0-9]{1}$/;
        if (!reg.test(value) && value !== '') {
            callback(new Error(i18n.global.t('commons.rule.paramComplexity', ['.%@!~_-'])));
        } else {
            callback();
        }
    }
    callback();
}

const confirmDialogRef = ref();

const emit = defineEmits(['checkExist', 'closeTerminal']);

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

interface DialogProps {
    from: string;
    database: string;
}
const acceptParams = (params: DialogProps): void => {
    form.password = '';
    form.from = params.from;
    form.database = params.database;
    loadPassword();
    dialogVisible.value = true;
};
const handleClose = () => {
    dialogVisible.value = false;
};

const random = async () => {
    form.password = getRandomStr(16);
};

const loadPassword = async () => {
    if (form.from === 'local') {
        const res = await GetAppConnInfo('redis', form.database);
        form.status = res.data.status;
        form.password = res.data.password || '';
        form.port = res.data.port || 3306;
        form.serviceName = res.data.serviceName || '';
        form.containerName = res.data.containerName || '';
        loadSystemIP();
        return;
    }
    const res = await getDatabase(form.database);
    form.password = res.data.password || '';
    form.port = res.data.port || 3306;
    form.password = res.data.password;
    form.remoteIP = res.data.address;
};

const loadSystemIP = async () => {
    const res = await getSettingInfo();
    form.systemIP = res.data.systemIP || i18n.global.t('database.localIP');
};

function loadRedisInfo(isContainer: boolean) {
    if (isContainer) {
        return form.from === 'local' ? form.containerName : form.systemIP;
    } else {
        return form.from === 'local' ? form.systemIP : form.remoteIP;
    }
}

const onSubmit = async () => {
    loading.value = true;
    emit('closeTerminal');
    await changeRedisPassword(form.database, form.password)
        .then(() => {
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            dialogVisible.value = false;
            emit('checkExist');
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
