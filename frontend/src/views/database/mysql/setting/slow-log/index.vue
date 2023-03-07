<template>
    <div>
        <span style="float: left; line-height: 30px">{{ $t('database.longQueryTime') }}</span>
        <div style="margin-left: 5px; float: left">
            <el-input type="number" v-model.number="variables.long_query_time">
                <template #append>{{ $t('database.second') }}</template>
            </el-input>
        </div>
        <span style="float: left; margin-left: 20px; line-height: 30px">{{ $t('database.isOn') }}</span>
        <el-switch
            style="margin-left: 5px; float: left"
            v-model="variables.slow_query_log"
            active-value="ON"
            inactive-value="OFF"
            @change="handleSlowLogs"
        />
        <div style="margin-left: 20px; float: left">
            <el-checkbox border v-model="isWatch">{{ $t('commons.button.watch') }}</el-checkbox>
        </div>
        <el-button style="margin-left: 20px" @click="onDownload" icon="Download">
            {{ $t('file.download') }}
        </el-button>
        <codemirror
            :autofocus="true"
            :placeholder="$t('database.noData')"
            :indent-with-tab="true"
            :tabSize="4"
            style="margin-top: 10px; height: calc(100vh - 392px)"
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
import { MsgSuccess } from '@/utils/message';

const extensions = [javascript(), oneDark];
const slowLogs = ref();
const view = shallowRef();
const handleReady = (payload) => {
    view.value = payload.view;
};

const confirmDialogRef = ref();

const isWatch = ref();
let timer: NodeJS.Timer | null = null;

const mysqlName = ref();
const variables = reactive({
    slow_query_log: 'OFF',
    long_query_time: 10,
});
const oldVariables = ref();

interface DialogProps {
    mysqlName: string;
    variables: Database.MysqlVariables;
}
const acceptParams = async (params: DialogProps): Promise<void> => {
    mysqlName.value = params.mysqlName;
    variables.slow_query_log = params.variables.slow_query_log;
    variables.long_query_time = Number(params.variables.long_query_time);

    const pathRes = await loadBaseDir();
    let path = `${pathRes.data}/apps/mysql/${mysqlName.value}/data/1Panel-slow.log`;
    if (variables.slow_query_log === 'ON') {
        loadMysqlSlowlogs(path);
    }
    timer = setInterval(() => {
        if (variables.slow_query_log === 'ON' && isWatch.value) {
            loadMysqlSlowlogs(path);
        }
    }, 1000 * 5);
    oldVariables.value = { ...variables };
};
const emit = defineEmits(['loading']);

const handleSlowLogs = async () => {
    let params = {
        header: i18n.global.t('database.confChange'),
        operationInfo: i18n.global.t('database.restartNowHelper'),
        submitInputInfo: i18n.global.t('database.restartNow'),
    };
    confirmDialogRef.value!.acceptParams(params);
};

const onCancle = async () => {
    variables.slow_query_log = variables.slow_query_log === 'ON' ? 'OFF' : 'ON';
};

const onSave = async () => {
    let param = [] as Array<Database.VariablesUpdate>;
    if (variables.slow_query_log !== oldVariables.value.slow_query_log) {
        param.push({ param: 'slow_query_log', value: variables.slow_query_log });
    }
    if (variables.slow_query_log === 'ON') {
        param.push({ param: 'long_query_time', value: variables.long_query_time + '' });
        param.push({ param: 'slow_query_log_file', value: '/var/lib/mysql/1Panel-slow.log' });
    }
    emit('loading', true);
    await updateMysqlVariables(param)
        .then(() => {
            emit('loading', false);
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        })
        .catch(() => {
            emit('loading', false);
        });
};

const onDownload = async () => {
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
    slowLogs.value = res.data;
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
