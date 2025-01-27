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
                <div class="float-left">
                    <el-input type="number" v-model.number="variables.long_query_time" />
                </div>
                <el-button class="float-left ml-5" @click="changeSlowLogs">
                    {{ $t('commons.button.save') }}
                </el-button>
                <div class="float-left ml-10">
                    <el-checkbox :disabled="!currentStatus" border v-model="isWatch">
                        {{ $t('commons.button.watch') }}
                    </el-checkbox>
                </div>
                <el-button :disabled="!currentStatus" class="ml-20" @click="onDownload" icon="Download">
                    {{ $t('commons.button.download') }}
                </el-button>
            </el-form-item>
        </el-form>
        <LogPro v-model="slowLogs"></LogPro>
        <ConfirmDialog @cancel="onCancel" ref="confirmDialogRef" @confirm="onSave"></ConfirmDialog>
    </div>
</template>
<script lang="ts" setup>
import { onBeforeUnmount, reactive, ref } from 'vue';
import { Database } from '@/api/interface/database';
import ConfirmDialog from '@/components/confirm-dialog/index.vue';
import { loadDBFile, updateMysqlVariables } from '@/api/modules/database';
import { dateFormatForName, downloadWithContent } from '@/utils/util';
import i18n from '@/lang';
import { MsgError, MsgInfo, MsgSuccess } from '@/utils/message';
import LogPro from '@/components/log-pro/index.vue';

const slowLogs = ref();
const detailShow = ref();
const currentStatus = ref();

const confirmDialogRef = ref();

const isWatch = ref();
let timer: NodeJS.Timer | null = null;

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
        loadMysqlSlowlogs();
        timer = setInterval(() => {
            if (variables.slow_query_log === 'ON' && isWatch.value) {
                loadMysqlSlowlogs();
            }
        }, 1000 * 5);
    } else {
        detailShow.value = false;
    }
};
const emit = defineEmits(['loading']);

const handleSlowLogs = async () => {
    if (variables.slow_query_log === 'ON') {
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

// const getDynamicHeight = () => {
//     if (variables.slow_query_log === 'ON') {
//         if (globalStore.openMenuTabs) {
//             return `calc(100vh - 467px)`;
//         } else {
//             return `calc(100vh - 437px)`;
//         }
//     }
//     if (globalStore.openMenuTabs) {
//         return `calc(100vh - 413px)`;
//     } else {
//         return `calc(100vh - 383px)`;
//     }
// };

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

const onDownload = async () => {
    if (!slowLogs.value) {
        MsgInfo(i18n.global.t('database.noData'));
        return;
    }
    downloadWithContent(slowLogs.value, currentDB.database + '-slowlogs-' + dateFormatForName(new Date()) + '.log');
};

const loadMysqlSlowlogs = async () => {
    const res = await loadDBFile(currentDB.type + '-slow-logs', currentDB.database);
    slowLogs.value = res.data || '';
    // nextTick(() => {
    //     const state = view.value.state;
    //     view.value.dispatch({
    //         selection: { anchor: state.doc.length, head: state.doc.length },
    //         scrollIntoView: true,
    //     });
    // });
};

onBeforeUnmount(() => {
    clearInterval(Number(timer));
    timer = null;
});

defineExpose({
    acceptParams,
});
</script>
