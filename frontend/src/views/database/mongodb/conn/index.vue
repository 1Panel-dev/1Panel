<template>
    <DrawerPro v-model="dialogVisible" :header="$t('database.databaseConnInfo')" @close="handleClose" size="small">
        <el-form @submit.prevent v-loading="loading" ref="formRef" :rules="rules" :model="form" label-position="top">
            <el-form-item v-if="form.from === 'local'">
                <template #label>
                    <div class="conn-label">
                        <span>{{ $t('database.containerConn') }}</span>
                        <el-button link @click="copyConnURL(true)" icon="DocumentCopy">
                            {{ $t('database.copyConnURL') }}
                        </el-button>
                    </div>
                </template>
                <el-card class="mini-border-card">
                    <el-descriptions :column="1">
                        <el-descriptions-item :label="$t('database.connAddress')">
                            <el-tooltip
                                v-if="loadMongoInfo(true).length > 48"
                                :content="loadMongoInfo(true)"
                                placement="top"
                            >
                                {{ loadMongoInfo(true).substring(0, 48) }}...
                            </el-tooltip>
                            <span v-else>
                                {{ loadMongoInfo(true) }}
                            </span>
                            <CopyButton :content="loadMongoInfo(true)" />
                        </el-descriptions-item>
                        <el-descriptions-item :label="$t('commons.table.port')">
                            27017
                            <CopyButton content="27017" />
                        </el-descriptions-item>
                    </el-descriptions>
                </el-card>
                <span class="input-help">
                    {{ $t('database.containerConnHelper') }}
                </span>
            </el-form-item>

            <el-form-item>
                <template #label>
                    <div class="conn-label">
                        <span>{{ $t('database.remoteConn') }}</span>
                        <el-button link @click="copyConnURL(false)" icon="DocumentCopy">
                            {{ $t('database.copyConnURL') }}
                        </el-button>
                    </div>
                </template>
                <el-card class="mini-border-card">
                    <el-descriptions :column="1">
                        <el-descriptions-item :label="$t('database.connAddress')">
                            <el-tooltip
                                v-if="loadMongoInfo(false).length > 48"
                                :content="loadMongoInfo(false)"
                                placement="top"
                            >
                                {{ loadMongoInfo(false).substring(0, 48) }}...
                            </el-tooltip>
                            <span v-else>
                                {{ loadMongoInfo(false) }}
                            </span>
                            <CopyButton :content="loadMongoInfo(false)" />
                        </el-descriptions-item>
                        <el-descriptions-item :label="$t('commons.table.port')">
                            {{ form.port }}
                            <CopyButton :content="form.port + ''" />
                        </el-descriptions-item>
                    </el-descriptions>
                </el-card>
                <span v-if="form.from === 'local'" class="input-help">
                    {{ $t('database.remoteConnHelper2') }}
                </span>
            </el-form-item>

            <el-divider border-style="dashed" />

            <div v-if="form.from === 'local'">
                <el-form-item :label="$t('commons.login.username')">
                    <el-input type="text" style="width: calc(100% - 60px)" readonly disabled v-model="form.username">
                        <template #append>
                            <el-button-group>
                                <CopyButton :content="form.username" :isIcon="false" />
                            </el-button-group>
                        </template>
                    </el-input>
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
                        <CopyButton :isIcon="false" class="copy_button" :content="form.password" />
                        <el-button @click="random">
                            {{ $t('commons.button.random') }}
                        </el-button>
                    </el-button-group>
                    <span class="input-help">{{ $t('commons.rule.illegalChar') }}</span>
                </el-form-item>
            </div>
            <div v-else>
                <el-form-item :label="$t('commons.login.username')">
                    <el-tag>{{ form.username || '-' }}</el-tag>
                    <CopyButton v-if="form.username" :content="form.username" />
                </el-form-item>
                <el-form-item :label="$t('commons.login.password')">
                    <el-tag>{{ form.password || '-' }}</el-tag>
                    <CopyButton v-if="form.password" :content="form.password" />
                </el-form-item>
            </div>
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
    </DrawerPro>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { Rules } from '@/global/form-rules';
import { ElForm } from 'element-plus';
import { getDatabase, updateMongodbRootPassword } from '@/api/modules/database';
import { getAppConnInfo } from '@/api/modules/app';
import { MsgSuccess } from '@/utils/message';
import { getRandomStr } from '@/utils/id';
import { getAgentSettingInfo } from '@/api/modules/setting';
import { copyText } from '@/utils/clipboard';
import i18n from '@/lang';
import { GlobalStore } from '@/store';

const globalStore = GlobalStore();
const loading = ref(false);
const dialogVisible = ref(false);

const form = reactive({
    from: '',
    status: '',
    systemIP: '',
    password: '',
    serviceName: '',
    containerName: '',
    remoteIP: '',
    port: 0,
    type: '',
    database: '',
    username: '',
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

const acceptParams = async (params: DialogProps): Promise<void> => {
    form.password = '';
    form.from = params.from;
    form.type = params.type;
    form.database = params.database;
    await loadConnInfo();
    dialogVisible.value = true;
};

const handleClose = () => {
    dialogVisible.value = false;
};

const random = () => {
    form.password = getRandomStr(16);
};

const loadSystemIP = async () => {
    const res = await getAgentSettingInfo();
    form.systemIP = res.data.systemIP || globalStore.currentNodeAddr || i18n.global.t('database.localIP');
};

const loadConnInfo = async () => {
    loading.value = true;
    try {
        await loadSystemIP();
        if (form.from === 'local') {
            const res = await getAppConnInfo(form.type, form.database);
            form.status = res.data.status || '';
            form.username = res.data.username || '';
            form.password = res.data.password || '';
            form.port = res.data.port || 27017;
            form.serviceName = res.data.serviceName || '';
            form.containerName = res.data.containerName || '';
            form.remoteIP = form.systemIP;
            return;
        }
        const res = await getDatabase(form.database);
        form.status = '';
        form.username = res.data.username || '';
        form.password = res.data.password || '';
        form.port = res.data.port || 27017;
        form.serviceName = '';
        form.containerName = '';
        form.remoteIP = res.data.address || '';
    } finally {
        loading.value = false;
    }
};

const loadMongoInfo = (isContainer: boolean) => {
    if (isContainer) {
        return form.from === 'local' ? form.containerName : form.systemIP;
    }
    return form.from === 'local' ? form.systemIP : form.remoteIP;
};

const copyConnURL = (isContainer: boolean) => {
    const host = loadMongoInfo(isContainer);
    const port = isContainer && form.from === 'local' ? 27017 : form.port;
    const user = form.username || '<username>';
    const encodedPassword = encodeURIComponent(form.password || '<password>');
    copyText(`mongodb://${user}:${encodedPassword}@${host}:${port}/admin?authSource=admin`);
};

const onSave = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        ElMessageBox.confirm(
            i18n.global.t('database.changeConnHelper', [i18n.global.t('database.rootPassword')]),
            i18n.global.t('commons.msg.infoTitle'),
            {
                confirmButtonText: i18n.global.t('commons.button.confirm'),
                cancelButtonText: i18n.global.t('commons.button.cancel'),
            },
        ).then(async () => {
            loading.value = true;
            await updateMongodbRootPassword({
                id: 0,
                from: form.from,
                type: form.type,
                database: form.database,
                value: form.password,
            })
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

defineExpose({
    acceptParams,
});
</script>

<style scoped lang="scss">
.conn-label {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;
    gap: 12px;
}

.copy_button {
    border-radius: 0px;
    border-left-width: 0px;
}

:deep(.el-input__wrapper) {
    border-top-right-radius: 0px;
    border-bottom-right-radius: 0px;
}

:deep(.el-form-item__label) {
    width: 100%;
}
</style>
