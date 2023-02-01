<template>
    <div v-show="settingShow">
        <LayoutContent :title="'Redis' + $t('database.setting')" :reload="true">
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
                    {{ $t('database.status') }}
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
                    {{ $t('database.portSetting') }}
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
                        placeholder="None data"
                        :indent-with-tab="true"
                        :tabSize="4"
                        style="margin-top: 10px; height: calc(100vh - 380px)"
                        :lineWrapping="true"
                        :matchBrackets="true"
                        theme="cobalt"
                        :styleActiveLine="true"
                        :extensions="extensions"
                        v-model="redisConf"
                        :readOnly="true"
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
                    <el-form :model="form" ref="formRef" :rules="rules" label-width="120px">
                        <el-row style="margin-top: 20px">
                            <el-col :span="1"><br /></el-col>
                            <el-col :span="10">
                                <el-form-item :label="$t('database.timeout')" prop="timeout">
                                    <el-input clearable type="number" v-model.number="form.timeout" />
                                    <span class="input-help">{{ $t('database.timeoutHelper') }}</span>
                                </el-form-item>
                                <el-form-item :label="$t('database.maxclients')" prop="maxclients">
                                    <el-input clearable type="number" v-model.number="form.maxclients" />
                                </el-form-item>
                                <el-form-item :label="$t('database.maxmemory')" prop="maxmemory">
                                    <el-input clearable type="number" v-model.number="form.maxmemory" />
                                    <span class="input-help">{{ $t('database.maxmemoryHelper') }}</span>
                                </el-form-item>
                                <el-form-item>
                                    <el-button type="primary" @click="onSubmtiForm(formRef)">
                                        {{ $t('commons.button.save') }}
                                    </el-button>
                                </el-form-item>
                            </el-col>
                        </el-row>
                    </el-form>
                </div>
                <div v-if="activeName === 'port'">
                    <el-form :model="form" ref="portRef" label-width="120px">
                        <el-row>
                            <el-col :span="1"><br /></el-col>
                            <el-col :span="10">
                                <el-form-item :label="$t('setting.port')" prop="port" :rules="Rules.port">
                                    <el-input clearable type="number" v-model.number="form.port">
                                        <template #append>
                                            <el-button @click="onSavePort(portRef)" icon="Collection">
                                                {{ $t('commons.button.save') }}
                                            </el-button>
                                        </template>
                                    </el-input>
                                </el-form-item>
                            </el-col>
                        </el-row>
                    </el-form>
                </div>
                <Persistence v-show="activeName === 'persistence'" ref="persistenceRef" />
            </template>
        </LayoutContent>

        <ConfirmDialog ref="confirmDialogRef" @confirm="submtiFile"></ConfirmDialog>
        <ConfirmDialog ref="confirmFileRef" @confirm="submtiFile"></ConfirmDialog>
        <ConfirmDialog ref="confirmFormRef" @confirm="submtiForm"></ConfirmDialog>
        <ConfirmDialog ref="confirmPortRef" @confirm="onChangePort(formRef)"></ConfirmDialog>
    </div>
</template>

<script lang="ts" setup>
import { ElMessage, FormInstance } from 'element-plus';
import { reactive, ref } from 'vue';
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import LayoutContent from '@/layout/layout-content.vue';
import { oneDark } from '@codemirror/theme-one-dark';
import { LoadFile } from '@/api/modules/files';
import ConfirmDialog from '@/components/confirm-dialog/index.vue';
import Status from '@/views/database/redis/setting/status/index.vue';
import Persistence from '@/views/database/redis/setting/persistence/index.vue';
import { loadRedisConf, updateRedisConf, updateRedisConfByFile } from '@/api/modules/database';
import i18n from '@/lang';
import { Rules } from '@/global/form-rules';
import { ChangePort, GetAppDefaultConfig } from '@/api/modules/app';
import { loadBaseDir } from '@/api/modules/setting';

const extensions = [javascript(), oneDark];

const loading = ref(false);

const form = reactive({
    name: '',
    port: 6379,
    timeout: 0,
    maxclients: 0,
    maxmemory: 0,
});
const rules = reactive({
    port: [Rules.port],
    timeout: [Rules.number],
    maxclients: [Rules.number],
    maxmemory: [Rules.number],
});

const activeName = ref('conf');
const statusRef = ref();
const persistenceRef = ref();

const useOld = ref(false);

const redisStatus = ref();
const redisName = ref();

const formRef = ref<FormInstance>();
const redisConf = ref();
const confirmDialogRef = ref();

const settingShow = ref<boolean>(false);

interface DialogProps {
    redisName: string;
    status: string;
}

const changeTab = (val: string) => {
    activeName.value = val;
    console.log(activeName.value);
};

const acceptParams = (prop: DialogProps): void => {
    redisStatus.value = prop.status;
    redisName.value = prop.redisName;
    settingShow.value = true;
    loadConfFile();
    if (redisStatus.value === 'Running') {
        statusRef.value!.acceptParams({ status: prop.status });
        persistenceRef.value!.acceptParams({ status: prop.status });
        loadform();
    }
};
const onClose = (): void => {
    settingShow.value = false;
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
            ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
        })
        .finally(() => {
            loading.value = false;
        });
};

const confirmFormRef = ref();
const onSubmtiForm = async (formEl: FormInstance | undefined) => {
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
const submtiForm = async () => {
    let param = {
        timeout: form.timeout + '',
        maxclients: form.maxclients + '',
        maxmemory: form.maxmemory + '',
    };
    loading.value = true;
    await updateRedisConf(param)
        .then(() => {
            loadform();
            loading.value = false;
            ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
        })
        .finally(() => {
            loading.value = false;
        });
};

const getDefaultConfig = async () => {
    loading.value = true;
    const res = await GetAppDefaultConfig('redis');
    redisConf.value = res.data;
    useOld.value = true;
    loading.value = false;
};

const onSaveFile = async () => {
    let params = {
        header: i18n.global.t('database.confChange'),
        operationInfo: i18n.global.t('database.restartNowHelper1'),
        submitInputInfo: i18n.global.t('database.restartNow'),
    };
    confirmDialogRef.value!.acceptParams(params);
};
const submtiFile = async () => {
    let param = {
        file: redisConf.value,
        restartNow: true,
    };
    loading.value = true;
    await updateRedisConfByFile(param)
        .then(() => {
            loading.value = false;
            ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
        })
        .catch(() => {
            loading.value = true;
        });
};

const loadform = async () => {
    const res = await loadRedisConf();
    form.name = res.data?.name;
    form.timeout = Number(res.data?.timeout);
    form.maxclients = Number(res.data?.maxclients);
    form.maxmemory = Number(res.data?.maxmemory);
    form.port = Number(res.data?.port);
};

const loadConfFile = async () => {
    const pathRes = await loadBaseDir();
    let path = `${pathRes.data}/apps/redis/${redisName.value}/conf/redis.conf`;
    const res = await LoadFile({ path: path });
    redisConf.value = res.data;
};

defineExpose({
    acceptParams,
    onClose,
});
</script>
