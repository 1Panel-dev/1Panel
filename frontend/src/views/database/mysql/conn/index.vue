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
        <el-form @submit.prevent v-loading="loading" ref="formRef" :rules="rules" :model="form" label-position="top">
            <el-row type="flex" justify="center">
                <el-col :span="22">
                    <el-form-item :label="$t('database.containerConn')" v-if="form.from === 'local'">
                        <el-card class="mini-border-card">
                            <el-descriptions :column="1">
                                <el-descriptions-item :label="$t('database.connAddress')">
                                    <el-tooltip
                                        v-if="loadMysqlInfo(true).length > 48"
                                        :content="loadMysqlInfo(true)"
                                        placement="top"
                                    >
                                        {{ loadMysqlInfo(true).substring(0, 48) }}...
                                    </el-tooltip>
                                    <span else>
                                        {{ loadMysqlInfo(true) }}
                                    </span>
                                    <CopyButton :content="loadMysqlInfo(true)" type="icon" />
                                </el-descriptions-item>
                                <el-descriptions-item :label="$t('database.connPort')">
                                    3306
                                    <CopyButton content="3306" type="icon" />
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
                                        v-if="loadMysqlInfo(false).length > 48"
                                        :content="loadMysqlInfo(false)"
                                        placement="top"
                                    >
                                        {{ loadMysqlInfo(false).substring(0, 48) }}...
                                    </el-tooltip>
                                    <span else>
                                        {{ loadMysqlInfo(false) }}
                                    </span>
                                    <CopyButton :content="loadMysqlInfo(false)" type="icon" />
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
                        <el-form-item :label="$t('database.remoteAccess')" prop="privilege">
                            <el-switch
                                v-model="form.privilege"
                                :disabled="form.status !== 'Running'"
                                @change="onSaveAccess"
                            />
                            <span class="input-help">{{ $t('database.remoteConnHelper') }}</span>
                        </el-form-item>
                        <el-form-item :label="$t('database.rootPassword')" prop="password">
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
                            <span class="input-help">{{ $t('commons.rule.illegalChar') }}</span>
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
import { getDatabase, loadRemoteAccess, updateMysqlAccess, updateMysqlPassword } from '@/api/modules/database';
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
    serviceName: '',
    containerName: '',
    oldPrivilege: false,
    privilege: false,
    port: 0,

    from: '',
    type: '',
    database: '',
    username: '',
    remoteIP: '',
});

const rules = reactive({
    password: [Rules.requiredInput, Rules.noSpace, Rules.illegal],
});

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

function loadMysqlInfo(isContainer: boolean) {
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
        const res = await loadRemoteAccess(form.type, form.database);
        form.privilege = res.data;
        form.oldPrivilege = res.data;
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
    form.username = res.data.username;
    form.password = res.data.password;
    form.remoteIP = res.data.address;
};

const onSave = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        ElMessageBox.confirm(
            i18n.global.t('database.changeConnHelper', [i18n.global.t('commons.login.password')]),
            i18n.global.t('commons.msg.infoTitle'),
            {
                confirmButtonText: i18n.global.t('commons.button.confirm'),
                cancelButtonText: i18n.global.t('commons.button.cancel'),
            },
        ).then(async () => {
            let param = {
                id: 0,
                from: form.from,
                type: form.type,
                database: form.database,
                value: form.password,
            };
            loading.value = true;
            await updateMysqlPassword(param)
                .then(() => {
                    loading.value = false;
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                    dialogVisible.value = false;
                })
                .catch(() => {
                    loading.value = false;
                });
        });
    });
};

const onSaveAccess = async () => {
    ElMessageBox.confirm(
        i18n.global.t('database.changeConnHelper', [i18n.global.t('database.remoteAccess')]),
        i18n.global.t('commons.msg.infoTitle'),
        {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
        },
    )
        .then(async () => {
            let param = {
                id: 0,
                from: form.from,
                type: form.type,
                database: form.database,
                value: form.privilege ? '%' : 'localhost',
            };
            loading.value = true;
            await updateMysqlAccess(param)
                .then(() => {
                    loading.value = false;
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                    dialogVisible.value = false;
                })
                .catch(() => {
                    loading.value = false;
                });
        })
        .catch(() => {
            form.privilege = form.oldPrivilege;
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
