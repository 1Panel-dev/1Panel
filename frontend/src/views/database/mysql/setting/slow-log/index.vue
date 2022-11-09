<template>
    <div>
        <span style="float: left">{{ $t('database.isOn') }}</span>
        <el-switch
            style="margin-left: 20px; float: left"
            v-model="variables.slow_query_log"
            active-value="ON"
            inactive-value="OFF"
        />
        <span style="margin-left: 30px; float: left">{{ $t('database.longQueryTime') }}</span>
        <div style="margin-left: 5px; float: left">
            <el-input type="number" v-model.number="variables.long_query_time">
                <template #append>{{ $t('database.second') }}</template>
            </el-input>
        </div>
        <el-button style="margin-left: 20px" @click="onSaveStart">{{ $t('commons.button.confirm') }}</el-button>
        <div v-if="variables.slow_query_log === 'ON'">
            <codemirror
                :autofocus="true"
                placeholder="None data"
                :indent-with-tab="true"
                :tabSize="4"
                style="margin-top: 10px; max-height: 500px"
                :lineWrapping="true"
                :matchBrackets="true"
                theme="cobalt"
                :styleActiveLine="true"
                :extensions="extensions"
                v-model="slowLogs"
                :readOnly="true"
            />
        </div>

        <ConfirmDialog ref="confirmDialogRef" @confirm="onSave"></ConfirmDialog>
    </div>
</template>
<script lang="ts" setup>
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import { reactive, ref } from 'vue';
import { Database } from '@/api/interface/database';
import { LoadFile } from '@/api/modules/files';
import ConfirmDialog from '@/components/confirm-dialog/index.vue';
import { updateMysqlVariables } from '@/api/modules/database';
import { ElMessage } from 'element-plus';
import i18n from '@/lang';

const extensions = [javascript(), oneDark];
const slowLogs = ref();

const confirmDialogRef = ref();

const mysqlName = ref();
const mysqlKey = ref();
const variables = reactive({
    slow_query_log: 'OFF',
    long_query_time: 10,
});
const oldVariables = ref();

interface DialogProps {
    mysqlName: string;
    mysqlKey: string;
    variables: Database.MysqlVariables;
}
const acceptParams = (params: DialogProps): void => {
    mysqlName.value = params.mysqlName;
    mysqlKey.value = params.mysqlKey;
    variables.slow_query_log = params.variables.slow_query_log;
    variables.long_query_time = Number(params.variables.long_query_time);

    if (variables.slow_query_log === 'ON') {
        let path = `/opt/1Panel/data/apps/${mysqlKey.value}/${mysqlName.value}/data/1Panel-slow.log`;
        loadMysqlSlowlogs(path);
    }
    oldVariables.value = { ...variables };
};

const onSaveStart = async () => {
    let params = {
        header: i18n.global.t('database.confChange'),
        operationInfo: i18n.global.t('database.restartNowHelper'),
        submitInputInfo: i18n.global.t('database.restartNow'),
    };
    confirmDialogRef.value!.acceptParams(params);
};

const onSave = async () => {
    let param = [] as Array<Database.VariablesUpdate>;
    if (variables.slow_query_log !== oldVariables.value.slow_query_log) {
        param.push({ param: 'slow_query_log', value: variables.slow_query_log });
    }
    if (variables.slow_query_log === 'ON') {
        param.push({ param: 'long_query_time', value: variables.long_query_time });
        param.push({ param: 'slow_query_log_file', value: '/var/lib/mysql/1Panel-slow.log' });
    }
    await updateMysqlVariables(mysqlName.value, param);
    ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
};

const loadMysqlSlowlogs = async (path: string) => {
    const res = await LoadFile({ path: path });
    slowLogs.value = res.data;
};

defineExpose({
    acceptParams,
});
</script>
