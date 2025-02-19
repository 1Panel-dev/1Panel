<template>
    <div>
        <el-form label-position="left" label-width="80px" @submit.prevent>
            <el-form-item :label="$t('database.isOn')">
                <el-switch
                    v-model="variables.slow_query_log"
                    active-value="ON"
                    inactive-value="OFF"
                    @change="handleSlowLogs"
                />
            </el-form-item>
            <el-form-item :label="$t('database.longQueryTime')" v-if="detailShow">
                <div style="float: left">
                    <el-input type="number" v-model.number="variables.long_query_time" />
                </div>
                <el-button style="float: left; margin-left: 10px" @click="changeSlowLogs">
                    {{ $t('commons.button.save') }}
                </el-button>
            </el-form-item>
        </el-form>
        <LogFile v-if="variables.slow_query_log === 'ON'" :config="config"></LogFile>

        <ConfirmDialog @cancel="onCancel" ref="confirmDialogRef" @confirm="onSave"></ConfirmDialog>
    </div>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { Database } from '@/api/interface/database';
import LogFile from '@/components/log-file/index.vue';
import ConfirmDialog from '@/components/confirm-dialog/index.vue';
import { updateMysqlVariables } from '@/api/modules/database';
import i18n from '@/lang';
import { MsgError, MsgSuccess } from '@/utils/message';

const detailShow = ref();
const currentStatus = ref();
const config = ref();
const confirmDialogRef = ref();

const variables = reactive({
    slow_query_log: 'OFF',
    long_query_time: 10,
});

const currentDB = reactive({
    type: '',
    database: '',
});
interface DialogProps {
    type: string;
    database: string;
    variables: Database.MysqlVariables;
}
const acceptParams = async (params: DialogProps): Promise<void> => {
    currentDB.type = params.type;
    currentDB.database = params.database;
    variables.slow_query_log = params.variables.slow_query_log;
    variables.long_query_time = Number(params.variables.long_query_time);

    if (variables.slow_query_log === 'ON') {
        currentStatus.value = true;
        detailShow.value = true;
        config.value = {
            type: params.type + '-slow-logs',
            name: params.database,
            tail: true,
        };
    } else {
        detailShow.value = false;
    }
};
const emit = defineEmits(['loading']);

const handleSlowLogs = async () => {
    if (variables.slow_query_log === 'ON') {
        config.value = {
            type: currentDB.type + '-slow-logs',
            name: currentDB.database,
            tail: true,
        };
        detailShow.value = true;
        return;
    }
    let params = {
        header: i18n.global.t('database.confChange'),
        operationInfo: i18n.global.t('database.restartNowHelper'),
        submitInputInfo: i18n.global.t('database.restartNow'),
    };
    confirmDialogRef.value!.acceptParams(params);
};

const changeSlowLogs = () => {
    if (!(variables.long_query_time > 0 && variables.long_query_time <= 600)) {
        MsgError(i18n.global.t('database.thresholdRangeHelper'));
        return;
    }
    let params = {
        header: i18n.global.t('database.confChange'),
        operationInfo: i18n.global.t('database.restartNowHelper'),
        submitInputInfo: i18n.global.t('database.restartNow'),
    };
    confirmDialogRef.value!.acceptParams(params);
};

const onCancel = async () => {
    variables.slow_query_log = currentStatus.value ? 'ON' : 'OFF';
    detailShow.value = currentStatus.value;
};

const onSave = async () => {
    let param = [] as Array<Database.VariablesUpdateHelper>;
    param.push({ param: 'slow_query_log', value: variables.slow_query_log });
    if (variables.slow_query_log === 'ON') {
        param.push({ param: 'long_query_time', value: variables.long_query_time + '' });
        param.push({ param: 'slow_query_log_file', value: '/var/lib/mysql/1Panel-slow.log' });
    }
    let params = {
        type: currentDB.type,
        database: currentDB.database,
        variables: param,
    };
    emit('loading', true);
    await updateMysqlVariables(params)
        .then(() => {
            emit('loading', false);
            currentStatus.value = variables.slow_query_log === 'ON';
            detailShow.value = variables.slow_query_log === 'ON';
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        })
        .catch(() => {
            emit('loading', false);
        });
};

defineExpose({
    acceptParams,
});
</script>
