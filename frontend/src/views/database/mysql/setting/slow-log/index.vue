<template>
    <div>
        <el-form label-position="left" label-width="80px">
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
                <div style="float: left; margin-left: 20px">
                    <el-checkbox style="margin-top: 2px" :disabled="!currentStatus" border v-model="isWatch">
                        {{ $t('commons.button.watch') }}
                    </el-checkbox>
                </div>
                <el-button :disabled="!currentStatus" style="margin-left: 20px" @click="onDownload" icon="Download">
                    {{ $t('file.download') }}
                </el-button>
            </el-form-item>
        </el-form>
        <codemirror
            :autofocus="true"
            :placeholder="$t('database.noData')"
            :indent-with-tab="true"
            :tabSize="4"
            style="height: calc(100vh - 428px); width: 100%"
            :lineWrapping="true"
            :matchBrackets="true"
            theme="cobalt"
            :styleActiveLine="true"
            :extensions="extensions"
            @ready="handleReady"
            v-model="slowLogs"
            :disabled="true"
        />

        <br />
        <ConfirmDialog @cancle="onCancle" ref="confirmDialogRef" @confirm="onSave"></ConfirmDialog>
    </div>
</template>
<script lang="ts" setup>
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import { nextTick, onBeforeUnmount, reactive, ref, shallowRef } from 'vue';
import { Database } from '@/api/interface/database';
import { LoadFile } from '@/api/modules/files';
import ConfirmDialog from '@/components/confirm-dialog/index.vue';
import { updateMysqlVariables } from '@/api/modules/database';
import { dateFormatForName } from '@/utils/util';
import i18n from '@/lang';
import { loadBaseDir } from '@/api/modules/setting';
import { MsgError, MsgInfo, MsgSuccess } from '@/utils/message';

const extensions = [javascript(), oneDark];
const slowLogs = ref();
const view = shallowRef();
const handleReady = (payload) => {
    view.value = payload.view;
};
const detailShow = ref();
const currentStatus = ref();

const confirmDialogRef = ref();

const isWatch = ref();
let timer: NodeJS.Timer | null = null;

const mysqlName = ref();
const variables = reactive({
    slow_query_log: 'OFF',
    long_query_time: 10,
});

interface DialogProps {
    mysqlName: string;
    variables: Database.MysqlVariables;
}
const acceptParams = async (params: DialogProps): Promise<void> => {
    mysqlName.value = params.mysqlName;
    variables.slow_query_log = params.variables.slow_query_log;
    variables.long_query_time = Number(params.variables.long_query_time);

    if (variables.slow_query_log === 'ON') {
        currentStatus.value = true;
        detailShow.value = true;
        const pathRes = await loadBaseDir();
        let path = `${pathRes.data}/apps/mysql/${mysqlName.value}/data/1Panel-slow.log`;
        loadMysqlSlowlogs(path);
        timer = setInterval(() => {
            if (variables.slow_query_log === 'ON' && isWatch.value) {
                loadMysqlSlowlogs(path);
            }
        }, 1000 * 5);
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

const onCancle = async () => {
    variables.slow_query_log = currentStatus.value ? 'ON' : 'OFF';
    detailShow.value = currentStatus.value;
};

const onSave = async () => {
    let param = [] as Array<Database.VariablesUpdate>;
    param.push({ param: 'slow_query_log', value: variables.slow_query_log });
    if (variables.slow_query_log === 'ON') {
        param.push({ param: 'long_query_time', value: variables.long_query_time + '' });
        param.push({ param: 'slow_query_log_file', value: '/var/lib/mysql/1Panel-slow.log' });
    }
    emit('loading', true);
    await updateMysqlVariables(param)
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
    const downloadUrl = window.URL.createObjectURL(new Blob([slowLogs.value]));
    const a = document.createElement('a');
    a.style.display = 'none';
    a.href = downloadUrl;
    a.download = mysqlName.value + '-slowlogs-' + dateFormatForName(new Date()) + '.log';
    const event = new MouseEvent('click');
    a.dispatchEvent(event);
};

const loadMysqlSlowlogs = async (path: string) => {
    const res = await LoadFile({ path: path });
    slowLogs.value = res.data || '';
    nextTick(() => {
        const state = view.value.state;
        view.value.dispatch({
            selection: { anchor: state.doc.length, head: state.doc.length },
            scrollIntoView: true,
        });
    });
};

onBeforeUnmount(() => {
    clearInterval(Number(timer));
    timer = null;
});

defineExpose({
    acceptParams,
});
</script>
