<template>
    <div v-if="settingShow">
        <el-card style="margin-top: 5px">
            <el-radio-group v-model="confShowType">
                <el-radio-button label="base">{{ $t('database.baseConf') }}</el-radio-button>
                <el-radio-button label="all">{{ $t('database.allConf') }}</el-radio-button>
            </el-radio-group>
            <el-form v-if="confShowType === 'base'" :model="form" ref="formRef" :rules="rules" label-width="120px">
                <el-row style="margin-top: 20px">
                    <el-col :span="1"><br /></el-col>
                    <el-col :span="10">
                        <el-form-item :label="$t('setting.port')" prop="port" :rules="Rules.port">
                            <el-input clearable type="number" v-model.number="form.port">
                                <template #append>
                                    <el-button @click="onChangePort(formRef)" icon="Collection">
                                        {{ $t('commons.button.save') }}
                                    </el-button>
                                </template>
                            </el-input>
                            <span class="input-help">{{ $t('database.portHelper') }}</span>
                        </el-form-item>
                        <el-form-item :label="$t('setting.password')" prop="requirepass">
                            <el-input type="password" show-password clearable v-model="form.requirepass" />
                            <span class="input-help">{{ $t('database.requirepassHelper') }}</span>
                        </el-form-item>
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
                            <el-button type="primary" size="default" @click="onSave(formRef)" style="width: 90px">
                                {{ $t('commons.button.save') }}
                            </el-button>
                        </el-form-item>
                    </el-col>
                </el-row>
            </el-form>
            <div v-if="confShowType === 'all'">
                <codemirror
                    :autofocus="true"
                    placeholder="None data"
                    :indent-with-tab="true"
                    :tabSize="4"
                    style="margin-top: 10px; height: calc(100vh - 280px)"
                    :lineWrapping="true"
                    :matchBrackets="true"
                    theme="cobalt"
                    :styleActiveLine="true"
                    :extensions="extensions"
                    v-model="mysqlConf"
                    :readOnly="true"
                />
                <el-button type="primary" size="default" @click="onSaveFile" style="width: 90px; margin-top: 5px">
                    {{ $t('commons.button.save') }}
                </el-button>
            </div>
        </el-card>

        <ConfirmDialog ref="confirmDialogRef" @confirm="onSubmitSave"></ConfirmDialog>
    </div>
</template>

<script lang="ts" setup>
import { ElMessage, FormInstance } from 'element-plus';
import { reactive, ref } from 'vue';
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import { LoadFile } from '@/api/modules/files';
import ConfirmDialog from '@/components/confirm-dialog/index.vue';
import { loadRedisConf, updateRedisConf, updateRedisConfByFile } from '@/api/modules/database';
import i18n from '@/lang';
import { Rules } from '@/global/form-rules';
import { ChangePort } from '@/api/modules/app';

const extensions = [javascript(), oneDark];
const confShowType = ref('base');

const restartNow = ref(false);
const form = reactive({
    name: '',
    port: 6379,
    requirepass: '',
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

const formRef = ref<FormInstance>();
const mysqlConf = ref();
const confirmDialogRef = ref();

const settingShow = ref<boolean>(false);

const acceptParams = (): void => {
    settingShow.value = true;
    loadform();
};
const onClose = (): void => {
    settingShow.value = false;
};

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
    await ChangePort(params);
    ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
    return;
};
function callback(error: any) {
    if (error) {
        return error.message;
    } else {
        return;
    }
}

const onSave = async (formEl: FormInstance | undefined) => {
    if (confShowType.value === 'all') {
        onSaveFile();
    }
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        let param = {
            timeout: form.timeout + '',
            maxclients: form.maxclients + '',
            requirepass: form.requirepass,
            maxmemory: form.maxmemory + '',
        };
        await updateRedisConf(param);
        ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
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

const onSubmitSave = async () => {
    let param = {
        file: mysqlConf.value,
        restartNow: restartNow.value,
    };
    await updateRedisConfByFile(param);
    restartNow.value = false;
    ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
};

const loadform = async () => {
    const res = await loadRedisConf();
    form.name = res.data?.name;
    form.timeout = Number(res.data?.timeout);
    form.maxclients = Number(res.data?.maxclients);
    form.requirepass = res.data?.requirepass;
    form.maxmemory = Number(res.data?.maxmemory);
    form.port = Number(res.data?.port);
    loadMysqlConf(`/opt/1Panel/data/apps/redis/${form.name}/conf/redis.conf`);
};

const loadMysqlConf = async (path: string) => {
    const res = await LoadFile({ path: path });
    mysqlConf.value = res.data;
};

defineExpose({
    acceptParams,
    onClose,
});
</script>
