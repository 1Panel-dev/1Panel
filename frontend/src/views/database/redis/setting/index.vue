<template>
    <div v-if="settingShow">
        <el-radio-group v-model="confShowType">
            <el-radio-button label="base">基础配置</el-radio-button>
            <el-radio-button label="all">全部配置</el-radio-button>
        </el-radio-group>
        <el-form v-if="confShowType === 'base'" :model="baseInfo" ref="panelFormRef" :rules="rules" label-width="120px">
            <el-row style="margin-top: 20px">
                <el-col :span="1"><br /></el-col>
                <el-col :span="10">
                    <el-form-item :label="$t('setting.port')" prop="port">
                        <el-input clearable type="number" v-model.number="baseInfo.port" />
                    </el-form-item>
                    <el-form-item :label="$t('setting.password')" prop="requirepass">
                        <el-input type="password" show-password clearable v-model="baseInfo.requirepass" />
                        <span class="input-help">{{ $t('database.requirepassHelper') }}</span>
                    </el-form-item>
                    <el-form-item :label="$t('database.timeout')" prop="timeout">
                        <el-input clearable type="number" v-model.number="baseInfo.timeout" />
                        <span class="input-help">{{ $t('database.timeoutHelper') }}</span>
                    </el-form-item>
                    <el-form-item :label="$t('database.maxclients')" prop="maxclients">
                        <el-input clearable type="number" v-model.number="baseInfo.maxclients" />
                    </el-form-item>
                    <el-form-item :label="$t('database.databases')" prop="databases">
                        <el-input clearable type="number" v-model.number="baseInfo.databases" />
                    </el-form-item>
                    <el-form-item :label="$t('database.maxmemory')" prop="maxmemory">
                        <el-input clearable type="number" v-model.number="baseInfo.maxmemory" />
                        <span class="input-help">{{ $t('database.maxmemoryHelper') }}</span>
                    </el-form-item>
                    <el-form-item>
                        <el-button type="primary" size="default" style="width: 90px">
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
            <el-button type="primary" size="default" style="width: 90px; margin-top: 5px">
                {{ $t('commons.button.save') }}
            </el-button>
        </div>
    </div>
</template>

<script lang="ts" setup>
import { FormInstance } from 'element-plus';
import { reactive, ref } from 'vue';
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import { LoadFile } from '@/api/modules/files';
import { loadRedisConf } from '@/api/modules/database';
// import i18n from '@/lang';
import { Rules } from '@/global/form-rules';

const extensions = [javascript(), oneDark];
const confShowType = ref('base');

const baseInfo = reactive({
    name: '',
    port: 3306,
    requirepass: '',
    timeout: 0,
    maxclients: 0,
    databases: 0,
    maxmemory: 0,
});
const rules = reactive({
    port: [Rules.port],
    timeout: [Rules.number],
    maxclients: [Rules.number],
    databases: [Rules.number],
    maxmemory: [Rules.number],
});

const panelFormRef = ref<FormInstance>();
const mysqlConf = ref();

const settingShow = ref<boolean>(false);

const acceptParams = (): void => {
    settingShow.value = true;
    loadBaseInfo();
};
const onClose = (): void => {
    settingShow.value = false;
};

// const onSave = async (formEl: FormInstance | undefined, key: string) => {
//     if (!formEl) return;
//     const result = await formEl.validateField(key, callback);
//     if (!result) {
//         return;
//     }
//     // let changeForm = {
//     //     paramName: key,
//     //     value: val + '',
//     // };
//     // await updateRedisConf(changeForm);
//     ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
// };
// function callback(error: any) {
//     if (error) {
//         return error.message;
//     } else {
//         return;
//     }
// }

const loadBaseInfo = async () => {
    const res = await loadRedisConf();
    baseInfo.name = res.data?.name;
    baseInfo.timeout = Number(res.data?.timeout);
    baseInfo.maxclients = Number(res.data?.maxclients);
    baseInfo.databases = Number(res.data?.databases);
    baseInfo.requirepass = res.data?.requirepass;
    baseInfo.maxmemory = Number(res.data?.maxmemory);
    loadMysqlConf(`/opt/1Panel/data/apps/redis/${baseInfo.name}/conf/redis.conf`);
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
