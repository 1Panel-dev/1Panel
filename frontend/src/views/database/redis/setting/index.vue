<template>
    <div v-show="settingShow" v-loading="loading">
        <LayoutContent :title="database + ' ' + $t('commons.button.set')" :reload="true">
            <template #buttons>
                <el-button type="primary" :plain="activeName !== 'conf'" @click="changeTab('conf')">
                    {{ $t('database.confChange') }}
                </el-button>
                <el-button
                    type="primary"
                    :disabled="redisStatus !== 'Running'"
                    :plain="activeName !== 'status'"
                    @click="changeTab('status')"
                >
                    {{ $t('database.currentStatus') }}
                </el-button>
                <el-button
                    type="primary"
                    :disabled="redisStatus !== 'Running'"
                    :plain="activeName !== 'tuning'"
                    @click="changeTab('tuning')"
                >
                    {{ $t('database.performanceTuning') }}
                </el-button>
                <el-button type="primary" :plain="activeName !== 'port'" @click="changeTab('port')">
                    {{ $t('commons.table.port') }}
                </el-button>
                <el-button
                    type="primary"
                    :disabled="redisStatus !== 'Running'"
                    :plain="activeName !== 'persistence'"
                    @click="changeTab('persistence')"
                >
                    {{ $t('database.persistence') }}
                </el-button>
            </template>
            <template #main>
                <div v-if="activeName === 'conf'">
                    <codemirror
                        :autofocus="true"
                        :placeholder="$t('commons.msg.noneData')"
                        :indent-with-tab="true"
                        :tabSize="4"
                        :style="{ height: `calc(100vh - ${loadHeight()})`, 'margin-top': '10px' }"
                        :lineWrapping="true"
                        :matchBrackets="true"
                        theme="cobalt"
                        :styleActiveLine="true"
                        :extensions="extensions"
                        @ready="handleReady"
                        v-model="redisConf"
                    />
                    <el-button style="margin-top: 10px" @click="getDefaultConfig()">
                        {{ $t('app.defaultConfig') }}
                    </el-button>
                    <el-button type="primary" @click="onSaveFile" style="margin-top: 10px">
                        {{ $t('commons.button.save') }}
                    </el-button>
                    <el-row>
                        <el-col :span="8">
                            <el-alert
                                v-if="useOld"
                                style="margin-top: 10px"
                                :title="$t('app.defaultConfigHelper')"
                                type="info"
                                :closable="false"
                            ></el-alert>
                        </el-col>
                    </el-row>
                </div>
                <Status v-show="activeName === 'status'" ref="statusRef" />
                <div v-if="activeName === 'tuning'">
                    <el-form :model="form" ref="formRef" :rules="rules" label-position="top">
                        <el-row style="margin-top: 20px">
                            <el-col :span="1"><br /></el-col>
                            <el-col :span="10">
                                <el-form-item :label="$t('database.timeout')" prop="timeout">
                                    <el-input clearable type="number" v-model.number="form.timeout">
                                        <template #append>{{ $t('commons.units.second') }}</template>
                                    </el-input>
                                    <span class="input-help">{{ $t('database.timeoutHelper') }}</span>
                                </el-form-item>
                                <el-form-item :label="$t('database.maxclients')" prop="maxclients">
                                    <el-input clearable type="number" v-model.number="form.maxclients" />
                                </el-form-item>
                                <el-form-item :label="$t('database.maxmemory')" prop="maxmemory">
                                    <el-input clearable type="number" v-model.number="form.maxmemory">
                                        <template #append>mb</template>
                                    </el-input>

                                    <span class="input-help">{{ $t('database.maxmemoryHelper') }}</span>
                                </el-form-item>
                                <el-form-item>
                                    <el-button type="primary" @click="onSubmitForm(formRef)">
                                        {{ $t('commons.button.save') }}
                                    </el-button>
                                </el-form-item>
                            </el-col>
                        </el-row>
                    </el-form>
                </div>
                <div v-if="activeName === 'port'">
                    <el-form :model="form" ref="portRef" label-position="top">
                        <el-row>
                            <el-col :span="1"><br /></el-col>
                            <el-col :span="10">
                                <el-form-item :label="$t('commons.table.port')" prop="port" :rules="Rules.port">
                                    <el-input clearable type="number" v-model.number="form.port" />
                                </el-form-item>
                                <el-form-item>
                                    <el-button @click="onSavePort(portRef)" icon="Collection">
                                        {{ $t('commons.button.save') }}
                                    </el-button>
                                </el-form-item>
                            </el-col>
                        </el-row>
                    </el-form>
                </div>
                <Persistence @loading="changeLoading" v-show="activeName === 'persistence'" ref="persistenceRef" />
            </template>
        </LayoutContent>

        <ConfirmDialog ref="confirmDialogRef" @confirm="submitFile"></ConfirmDialog>
        <ConfirmDialog ref="confirmFileRef" @confirm="submitFile"></ConfirmDialog>
        <ConfirmDialog ref="confirmFormRef" @confirm="submitForm"></ConfirmDialog>
        <ConfirmDialog ref="confirmPortRef" @confirm="onChangePort(portRef)"></ConfirmDialog>
    </div>
</template>

<script lang="ts" setup>
import { FormInstance } from 'element-plus';
import { nextTick, reactive, ref, shallowRef } from 'vue';
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import ConfirmDialog from '@/components/confirm-dialog/index.vue';
import Status from '@/views/database/redis/setting/status/index.vue';
import Persistence from '@/views/database/redis/setting/persistence/index.vue';
import { loadDBFile, loadRedisConf, updateRedisConf, updateDBFile } from '@/api/modules/database';
import i18n from '@/lang';
import { checkNumberRange, Rules } from '@/global/form-rules';
import { ChangePort, GetAppDefaultConfig } from '@/api/modules/app';
import { MsgSuccess } from '@/utils/message';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();

const extensions = [javascript(), oneDark];

const loading = ref(false);

const view = shallowRef();
const handleReady = (payload) => {
    view.value = payload.view;
};

const loadHeight = () => {
    return globalStore.openMenuTabs ? '410px' : '380px';
};

const form = reactive({
    name: '',
    port: 6379,
    timeout: 0,
    maxclients: 0,
    maxmemory: 0,
});
const rules = reactive({
    port: [Rules.port],
    timeout: [Rules.number, checkNumberRange(0, 9999999)],
    maxclients: [Rules.number, checkNumberRange(1, 65504)],
    maxmemory: [Rules.number, checkNumberRange(0, 999999)],
});

const activeName = ref('conf');
const statusRef = ref();
const persistenceRef = ref();

const useOld = ref(false);

const redisStatus = ref();
const database = ref();

const formRef = ref<FormInstance>();
const redisConf = ref();
const confirmDialogRef = ref();

const settingShow = ref<boolean>(false);

interface DialogProps {
    database: string;
    status: string;
}

const changeTab = (val: string) => {
    activeName.value = val;
    switch (val) {
        case 'conf':
            loadConfFile();
            break;
        case 'persistence':
            persistenceRef.value!.acceptParams({ status: redisStatus.value, database: database.value });
            break;
        case 'tuning':
        case 'port':
            loadForm();
            break;
        case 'status':
            statusRef.value!.acceptParams({ status: redisStatus.value, database: database.value });
            break;
    }
};

const changeLoading = (status: boolean) => {
    loading.value = status;
};

const acceptParams = (prop: DialogProps): void => {
    redisStatus.value = prop.status;
    database.value = prop.database;
    settingShow.value = true;
    loadConfFile();
};

const portRef = ref();
const confirmPortRef = ref();
const onSavePort = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    const result = await formEl.validateField('port', callback);
    if (!result) {
        return;
    }
    let params = {
        header: i18n.global.t('database.confChange'),
        operationInfo: i18n.global.t('database.restartNowHelper'),
        submitInputInfo: i18n.global.t('database.restartNow'),
    };
    confirmPortRef.value!.acceptParams(params);
    return;
};
function callback(error: any) {
    if (error) {
        return error.message;
    } else {
        return;
    }
}
const onChangePort = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    const result = await formEl.validateField('port', callback);
    if (!result) {
        return;
    }
    let params = {
        key: 'redis',
        name: form.name,
        port: form.port,
    };
    loading.value = true;
    await ChangePort(params)
        .then(() => {
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        })
        .finally(() => {
            loading.value = false;
        });
};

const confirmFormRef = ref();
const onSubmitForm = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        let params = {
            header: i18n.global.t('database.confChange'),
            operationInfo: i18n.global.t('database.restartNowHelper'),
            submitInputInfo: i18n.global.t('database.restartNow'),
        };
        confirmFormRef.value!.acceptParams(params);
    });
};
const submitForm = async () => {
    let param = {
        database: database.value,
        timeout: form.timeout + '',
        maxclients: form.maxclients + '',
        maxmemory: form.maxmemory + 'mb',
    };
    loading.value = true;
    await updateRedisConf(param)
        .then(() => {
            loadForm();
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        })
        .finally(() => {
            loading.value = false;
        });
};

const getDefaultConfig = async () => {
    loading.value = true;
    await GetAppDefaultConfig('redis', '')
        .then((res) => {
            redisConf.value = res.data;
            useOld.value = true;
            loading.value = false;
        })
        .catch(() => {
            loading.value = false;
        });
};

const onSaveFile = async () => {
    let params = {
        header: i18n.global.t('database.confChange'),
        operationInfo: i18n.global.t('database.restartNowHelper1'),
        submitInputInfo: i18n.global.t('database.restartNow'),
    };
    confirmDialogRef.value!.acceptParams(params);
};
const submitFile = async () => {
    let param = {
        type: 'redis',
        database: database.value,
        file: redisConf.value,
    };
    loading.value = true;
    await updateDBFile(param)
        .then(() => {
            useOld.value = false;
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        })
        .catch(() => {
            loading.value = false;
        });
};

const loadForm = async () => {
    const res = await loadRedisConf(database.value);
    form.name = res.data?.name;
    form.timeout = Number(res.data?.timeout);
    form.maxclients = Number(res.data?.maxclients);
    form.maxmemory = Number(res.data?.maxmemory.replaceAll('mb', '')) / 1048576;
    form.port = Number(res.data?.port);
};

const loadConfFile = async () => {
    useOld.value = false;
    loading.value = true;
    await loadDBFile('redis-conf', database.value)
        .then((res) => {
            loading.value = false;
            redisConf.value = res.data;
            nextTick(() => {
                const state = view.value.state;
                view.value.dispatch({
                    selection: { anchor: state.doc.length, head: state.doc.length },
                    scrollIntoView: true,
                });
            });
        })
        .catch(() => {
            loading.value = false;
        });
};

defineExpose({
    acceptParams,
});
</script>
