<template>
    <div class="demo-collapse" v-if="onSetting">
        <el-card>
            <el-collapse v-model="activeName" accordion>
                <el-collapse-item :title="$t('database.baseSetting')" name="1">
                    <el-form :model="form" ref="panelFormRef" label-width="120px">
                        <el-row>
                            <el-col :span="1"><br /></el-col>
                            <el-col :span="10">
                                <el-form-item :label="$t('setting.port')" prop="port">
                                    <el-input clearable v-model="form.port">
                                        <template #append>
                                            <el-button
                                                @click="onSave(panelFormRef, 'port', form.port)"
                                                icon="Collection"
                                            >
                                                {{ $t('commons.button.save') }}
                                            </el-button>
                                        </template>
                                    </el-input>
                                </el-form-item>
                                <el-form-item :label="$t('setting.password')" prop="password">
                                    <el-input clearable v-model="form.port">
                                        <template #append>
                                            <el-button
                                                @click="onSave(panelFormRef, 'password', form.password)"
                                                icon="Collection"
                                            >
                                                {{ $t('commons.button.save') }}
                                            </el-button>
                                        </template>
                                    </el-input>
                                </el-form-item>
                                <el-form-item :label="$t('database.remoteAccess')" prop="remoteAccess">
                                    <el-input clearable v-model="form.port">
                                        <template #append>
                                            <el-button
                                                @click="onSave(panelFormRef, 'remoteAccess', form.remoteAccess)"
                                                icon="Collection"
                                            >
                                                {{ $t('commons.button.save') }}
                                            </el-button>
                                        </template>
                                    </el-input>
                                </el-form-item>
                            </el-col>
                        </el-row>
                    </el-form>
                </el-collapse-item>
                <el-collapse-item :title="$t('database.confChange')" name="2">
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
                        v-model="mysqlConf"
                        :readOnly="true"
                    />
                    <el-button
                        type="primary"
                        style="width: 120px; margin-top: 10px"
                        @click="onSave(panelFormRef, 'remoteAccess', form.remoteAccess)"
                    >
                        {{ $t('commons.button.save') }}
                    </el-button>
                </el-collapse-item>
                <el-collapse-item :title="$t('database.currentStatus')" name="3">
                    <el-row :gutter="20">
                        <el-col :span="1"><br /></el-col>
                        <el-col :span="6">
                            <table style="width: 100%" class="myTable">
                                <tr>
                                    <td>{{ $t('database.runTime') }}</td>
                                    <td>{{ mysqlStatus.run }}</td>
                                </tr>
                                <tr>
                                    <td>{{ $t('database.connections') }}</td>
                                    <td>{{ mysqlStatus.connections }}</td>
                                </tr>
                                <tr>
                                    <td>{{ $t('database.bytesSent') }}</td>
                                    <td>{{ mysqlStatus!.bytesSent }}</td>
                                </tr>
                                <tr>
                                    <td>{{ $t('database.bytesReceived') }}</td>
                                    <td>{{ mysqlStatus!.bytesReceived }}</td>
                                </tr>
                            </table>
                        </el-col>
                        <el-col :span="6">
                            <table style="width: 100%" class="myTable">
                                <tr>
                                    <td>{{ $t('database.queryPerSecond') }}</td>
                                    <td>{{ mysqlStatus!.queryPerSecond }}</td>
                                </tr>
                                <tr>
                                    <td>{{ $t('database.queryPerSecond') }}</td>
                                    <td>{{ mysqlStatus!.txPerSecond }}</td>
                                </tr>
                                <tr>
                                    <td>File</td>
                                    <td>{{ mysqlStatus!.file }}</td>
                                </tr>
                                <tr>
                                    <td>Position</td>
                                    <td>{{ mysqlStatus!.position }}</td>
                                </tr>
                            </table>
                        </el-col>
                    </el-row>
                    <el-row>
                        <el-col :span="1"><br /></el-col>
                        <el-col :span="12">
                            <table style="margin-top: 20px; width: 100%" class="myTable">
                                <tr>
                                    <td>{{ $t('database.queryPerSecond') }}</td>
                                    <td>{{ mysqlStatus!.connInfo }}</td>
                                    <td>{{ $t('database.connInfoHelper') }}</td>
                                </tr>
                                <tr>
                                    <td>{{ $t('database.threadCacheHit') }}</td>
                                    <td>{{ mysqlStatus!.threadCacheHit }}</td>
                                    <td>{{ $t('database.threadCacheHitHelper') }}</td>
                                </tr>
                                <tr>
                                    <td>{{ $t('database.indexHit') }}</td>
                                    <td>{{ mysqlStatus!.indexHit }}</td>
                                    <td>{{ $t('database.indexHitHelper') }}</td>
                                </tr>
                                <tr>
                                    <td>{{ $t('database.innodbIndexHit') }}</td>
                                    <td>{{ mysqlStatus!.innodbIndexHit }}</td>
                                    <td>{{ $t('database.innodbIndexHitHelper') }}</td>
                                </tr>
                                <tr>
                                    <td>{{ $t('database.cacheHit') }}</td>
                                    <td>{{ mysqlStatus!.cacheHit }}</td>
                                    <td>{{ $t('database.cacheHitHelper') }}</td>
                                </tr>
                                <tr>
                                    <td>{{ $t('database.tmpTableToDB') }}</td>
                                    <td>{{ mysqlStatus!.tmpTableToDB }}</td>
                                    <td>{{ $t('database.tmpTableToDBHelper') }}</td>
                                </tr>
                                <tr>
                                    <td>{{ $t('database.openTables') }}</td>
                                    <td>{{ mysqlStatus!.openTables }}</td>
                                    <td>{{ $t('database.openTablesHelper') }}</td>
                                </tr>
                                <tr>
                                    <td>{{ $t('database.selectFullJoin') }}</td>
                                    <td>{{ mysqlStatus!.selectFullJoin }}</td>
                                    <td>{{ $t('database.selectFullJoinHelper') }}</td>
                                </tr>
                                <tr>
                                    <td>{{ $t('database.selectRangeCheck') }}</td>
                                    <td>{{ mysqlStatus!.selectRangeCheck }}</td>
                                    <td>{{ $t('database.selectRangeCheckHelper') }}</td>
                                </tr>
                                <tr>
                                    <td>{{ $t('database.sortMergePasses') }}</td>
                                    <td>{{ mysqlStatus!.sortMergePasses }}</td>
                                    <td>{{ $t('database.sortMergePassesHelper') }}</td>
                                </tr>
                                <tr>
                                    <td>{{ $t('database.tableLocksWaited') }}</td>
                                    <td>{{ mysqlStatus!.tableLocksWaited }}</td>
                                    <td>{{ $t('database.tableLocksWaitedHelper') }}</td>
                                </tr>
                            </table>
                        </el-col>
                    </el-row>
                </el-collapse-item>
                <el-collapse-item :title="$t('database.performanceTuning')" name="4">
                    <el-card>
                        <el-form
                            :model="mysqlVariables"
                            :rules="variablesRules"
                            ref="variableFormRef"
                            label-width="160px"
                        >
                            <el-row>
                                <el-col :span="1"><br /></el-col>
                                <el-col :span="9">
                                    <el-form-item :label="$t('database.optimizationScheme')">
                                        <el-select @change="changePlan" clearable v-model="plan">
                                            <el-option
                                                v-for="item in planOptions"
                                                :key="item.id"
                                                :label="item.title"
                                                :value="item.id"
                                            />
                                        </el-select>
                                    </el-form-item>
                                </el-col>
                            </el-row>
                            <el-row>
                                <el-col :span="1"><br /></el-col>
                                <el-col :span="9">
                                    <el-form-item label="key_buffer_size" prop="key_buffer_size">
                                        <el-input clearable v-model.number="mysqlVariables.key_buffer_size">
                                            <template #append>MB</template>
                                        </el-input>
                                        <span class="input-help">{{ $t('database.keyBufferSizeHelper') }}</span>
                                    </el-form-item>
                                    <el-form-item label="query_cache_size" prop="query_cache_size">
                                        <el-input clearable v-model.number="mysqlVariables.query_cache_size">
                                            <template #append>MB</template>
                                        </el-input>
                                        <span class="input-help">{{ $t('database.queryCacheSizeHelper') }}</span>
                                    </el-form-item>
                                    <el-form-item label="tmp_table_size" prop="tmp_table_size">
                                        <el-input clearable v-model.number="mysqlVariables.tmp_table_size">
                                            <template #append>MB</template>
                                        </el-input>
                                        <span class="input-help">{{ $t('database.tmpTableSizeHelper') }}</span>
                                    </el-form-item>
                                    <el-form-item label="innodb_buffer_pool_size" prop="innodb_buffer_pool_size">
                                        <el-input clearable v-model.number="mysqlVariables.innodb_buffer_pool_size">
                                            <template #append>MB</template>
                                        </el-input>
                                        <span class="input-help">{{ $t('database.innodbBufferPoolSizeHelper') }}</span>
                                    </el-form-item>
                                    <el-form-item label="innodb_log_buffer_size" prop="innodb_log_buffer_size">
                                        <el-input clearable v-model.number="mysqlVariables.innodb_log_buffer_size">
                                            <template #append>MB</template>
                                        </el-input>
                                        <span class="input-help">{{ $t('database.innodbLogBufferSizeHelper') }}</span>
                                    </el-form-item>
                                    <el-form-item label="sort_buffer_size" prop="sort_buffer_size">
                                        <el-input clearable v-model.number="mysqlVariables.sort_buffer_size">
                                            <template #append>KB</template>
                                        </el-input>
                                        <span class="input-help">{{ $t('database.sortBufferSizeHelper') }}</span>
                                    </el-form-item>
                                    <el-form-item label="read_buffer_size" prop="read_buffer_size">
                                        <el-input clearable v-model.number="mysqlVariables.read_buffer_size">
                                            <template #append>KB</template>
                                        </el-input>
                                        <span class="input-help">{{ $t('database.readBufferSizeHelper') }}</span>
                                    </el-form-item>

                                    <el-form-item>
                                        <el-button
                                            icon="Collection"
                                            @click="onSaveVariables(variableFormRef)"
                                            type="primary"
                                            size="default"
                                        >
                                            {{ $t('commons.button.save') }}
                                        </el-button>
                                        <el-button icon="RefreshLeft" size="default">
                                            {{ $t('database.restart') }}
                                        </el-button>
                                    </el-form-item>
                                </el-col>
                                <el-col :span="2"><br /></el-col>
                                <el-col :span="9">
                                    <el-form-item label="read_rnd_buffer_size" prop="read_rnd_buffer_size">
                                        <el-input clearable v-model.number="mysqlVariables.read_rnd_buffer_size">
                                            <template #append>KB</template>
                                        </el-input>
                                        <span class="input-help">{{ $t('database.readRndBufferSizeHelper') }}</span>
                                    </el-form-item>
                                    <el-form-item label="join_buffer_size" prop="join_buffer_size">
                                        <el-input clearable v-model.number="mysqlVariables.join_buffer_size">
                                            <template #append>KB</template>
                                        </el-input>
                                        <span class="input-help">{{ $t('database.joinBufferSizeHelper') }}</span>
                                    </el-form-item>
                                    <el-form-item label="thread_stack" prop="thread_stack">
                                        <el-input clearable v-model.number="mysqlVariables.thread_stack">
                                            <template #append>KB</template>
                                        </el-input>
                                        <span class="input-help">{{ $t('database.threadStackelper') }}</span>
                                    </el-form-item>
                                    <el-form-item label="binlog_cache_size" prop="binlog_cache_size">
                                        <el-input clearable v-model.number="mysqlVariables.binlog_cache_size">
                                            <template #append>KB</template>
                                        </el-input>
                                        <span class="input-help">{{ $t('database.binlogCacheSizeHelper') }}</span>
                                    </el-form-item>
                                    <el-form-item label="thread_cache_size" prop="thread_cache_size">
                                        <el-input clearable v-model.number="mysqlVariables.thread_cache_size" />
                                        <span class="input-help">{{ $t('database.threadCacheSizeHelper') }}</span>
                                    </el-form-item>
                                    <el-form-item label="table_open_cache" prop="table_open_cache">
                                        <el-input clearable v-model.number="mysqlVariables.table_open_cache" />
                                        <span class="input-help">{{ $t('database.tableOpenCacheHelper') }}</span>
                                    </el-form-item>
                                    <el-form-item label="max_connections" prop="max_connections">
                                        <el-input clearable v-model.number="mysqlVariables.max_connections" />
                                        <span class="input-help">{{ $t('database.maxConnectionsHelper') }}</span>
                                    </el-form-item>
                                </el-col>
                            </el-row>
                        </el-form>
                    </el-card>
                </el-collapse-item>
                <el-collapse-item title="日志" name="5"></el-collapse-item>
            </el-collapse>
        </el-card>
    </div>
</template>

<script lang="ts" setup>
import { ElMessage, FormInstance } from 'element-plus';
import { reactive, ref } from 'vue';
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import { LoadFile } from '@/api/modules/files';
import { planOptions } from './helper';
import { loadMysqlStatus, loadMysqlVariables, updateMysqlVariables } from '@/api/modules/database';
import { computeSize } from '@/utils/util';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';

const extensions = [javascript(), oneDark];
const activeName = ref('1');

const form = reactive({
    port: '',
    password: '',
    remoteAccess: '',
    sessionTimeout: 0,
    localTime: '',
    panelName: '',
    theme: '',
    language: '',
    complexityVerification: '',
});
const panelFormRef = ref<FormInstance>();
const mysqlConf = ref();

const plan = ref();

const variableFormRef = ref<FormInstance>();
let mysqlVariables = reactive({
    version: '',
    key_buffer_size: 0,
    query_cache_size: 0,
    tmp_table_size: 0,
    innodb_buffer_pool_size: 0,
    innodb_log_buffer_size: 0,
    sort_buffer_size: 0,
    read_buffer_size: 0,
    read_rnd_buffer_size: 0,
    join_buffer_size: 0,
    thread_stack: 0,
    binlog_cache_size: 0,
    thread_cache_size: 0,
    table_open_cache: 0,
    max_connections: 0,
});
const variablesRules = reactive({
    key_buffer_size: [Rules.number],
    query_cache_size: [Rules.number],
    tmp_table_size: [Rules.number],
    innodb_buffer_pool_size: [Rules.number],
    innodb_log_buffer_size: [Rules.number],
    sort_buffer_size: [Rules.number],
    read_buffer_size: [Rules.number],
    read_rnd_buffer_size: [Rules.number],
    join_buffer_size: [Rules.number],
    thread_stack: [Rules.number],
    binlog_cache_size: [Rules.number],
    thread_cache_size: [Rules.number],
    table_open_cache: [Rules.number],
    max_connections: [Rules.number],
});
let mysqlStatus = reactive({
    run: 0,
    connections: 0,
    bytesSent: '',
    bytesReceived: '',

    queryPerSecond: '',
    txPerSecond: '',
    file: '',
    position: 0,

    connInfo: '',
    threadCacheHit: '',
    indexHit: '',
    innodbIndexHit: '',
    cacheHit: '',
    tmpTableToDB: '',
    openTables: 0,
    selectFullJoin: 0,
    selectRangeCheck: 0,
    sortMergePasses: 0,
    tableLocksWaited: 0,
});

const onSetting = ref<boolean>(false);
const paramVersion = ref();

interface DialogProps {
    version: string;
}
const acceptParams = (params: DialogProps): void => {
    onSetting.value = true;
    loadMysqlConf('/opt/1Panel/conf/mysql.conf');
    loadStatus();
    loadVariables();
    paramVersion.value = params.version;
};
const onClose = (): void => {
    onSetting.value = false;
};

const onSave = async (formEl: FormInstance | undefined, key: string, val: any) => {
    console.log(formEl, key, val);
};

const loadMysqlConf = async (path: string) => {
    const res = await LoadFile({ path: path });
    mysqlConf.value = res.data;
};

const loadVariables = async () => {
    const res = await loadMysqlVariables();
    mysqlVariables.key_buffer_size = Number(res.data.key_buffer_size) / 1024 / 1024;
    mysqlVariables.query_cache_size = Number(res.data.query_cache_size) / 1024 / 1024;
    mysqlVariables.tmp_table_size = Number(res.data.tmp_table_size) / 1024 / 1024;
    mysqlVariables.innodb_buffer_pool_size = Number(res.data.innodb_buffer_pool_size) / 1024 / 1024;
    mysqlVariables.innodb_log_buffer_size = Number(res.data.innodb_log_buffer_size) / 1024 / 1024;

    mysqlVariables.sort_buffer_size = Number(res.data.sort_buffer_size) / 1024;
    mysqlVariables.read_buffer_size = Number(res.data.read_buffer_size) / 1024;
    mysqlVariables.read_rnd_buffer_size = Number(res.data.read_rnd_buffer_size) / 1024;
    mysqlVariables.join_buffer_size = Number(res.data.join_buffer_size) / 1024;
    mysqlVariables.thread_stack = Number(res.data.thread_stack) / 1024;
    mysqlVariables.binlog_cache_size = Number(res.data.binlog_cache_size) / 1024;
    mysqlVariables.thread_cache_size = Number(res.data.thread_cache_size);
    mysqlVariables.table_open_cache = Number(res.data.table_open_cache);
    mysqlVariables.max_connections = Number(res.data.max_connections);
};

const changePlan = async () => {
    for (const item of planOptions) {
        if (item.id === plan.value) {
            mysqlVariables.key_buffer_size = item.data.key_buffer_size;
            mysqlVariables.query_cache_size = item.data.query_cache_size;
            mysqlVariables.tmp_table_size = item.data.tmp_table_size;
            mysqlVariables.innodb_buffer_pool_size = item.data.innodb_buffer_pool_size;

            mysqlVariables.sort_buffer_size = item.data.sort_buffer_size;
            mysqlVariables.read_buffer_size = item.data.read_buffer_size;
            mysqlVariables.read_rnd_buffer_size = item.data.read_rnd_buffer_size;
            mysqlVariables.join_buffer_size = item.data.join_buffer_size;
            mysqlVariables.thread_stack = item.data.thread_stack;
            mysqlVariables.binlog_cache_size = item.data.binlog_cache_size;
            mysqlVariables.thread_cache_size = item.data.thread_cache_size;
            mysqlVariables.table_open_cache = item.data.table_open_cache;
            mysqlVariables.max_connections = item.data.max_connections;
            break;
        }
    }
};

const onSaveVariables = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        let itemForm = {
            version: paramVersion.value,
            key_buffer_size: mysqlVariables.key_buffer_size * 1024 * 1024,
            query_cache_size: mysqlVariables.query_cache_size * 1024 * 1024,
            tmp_table_size: mysqlVariables.tmp_table_size * 1024 * 1024,
            innodb_buffer_pool_size: mysqlVariables.innodb_buffer_pool_size * 1024 * 1024,
            innodb_log_buffer_size: mysqlVariables.innodb_log_buffer_size * 1024 * 1024,

            sort_buffer_size: mysqlVariables.sort_buffer_size * 1024,
            read_buffer_size: mysqlVariables.read_buffer_size * 1024,
            read_rnd_buffer_size: mysqlVariables.read_rnd_buffer_size * 1024,
            join_buffer_size: mysqlVariables.join_buffer_size * 1024,
            thread_stack: mysqlVariables.thread_stack * 1024,
            binlog_cache_size: mysqlVariables.binlog_cache_size * 1024,
            thread_cache_size: mysqlVariables.thread_cache_size,
            table_open_cache: mysqlVariables.table_open_cache,
            max_connections: mysqlVariables.max_connections,
        };
        await updateMysqlVariables(itemForm);
        ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
    });
};

const loadStatus = async () => {
    const res = await loadMysqlStatus();
    let queryPerSecond = res.data.Questions / res.data.Uptime;
    let txPerSecond = (res.data!.Com_commit + res.data.Com_rollback) / res.data.Uptime;

    let threadCacheHit = (1 - res.data.Threads_created / res.data.Connections) * 100;
    let cacheHit = (res.data.Qcache_hits / (res.data.Qcache_hits + res.data.Qcache_inserts)) * 100;
    let indexHit = (1 - res.data.Key_reads / res.data.Key_read_requests) * 100;
    let innodbIndexHit = (1 - res.data.Innodb_buffer_pool_reads / res.data.Innodb_buffer_pool_read_requests) * 100;
    let tmpTableToDB = (res.data.Created_tmp_disk_tables / res.data.Created_tmp_tables) * 100;

    mysqlStatus.run = res.data.Run;
    mysqlStatus.connections = res.data.Connections;
    mysqlStatus.bytesSent = res.data.Bytes_sent ? computeSize(res.data.Bytes_sent) : '0';
    mysqlStatus.bytesReceived = res.data.Bytes_received ? computeSize(res.data.Bytes_received) : '0';

    mysqlStatus.queryPerSecond = isNaN(queryPerSecond) || queryPerSecond === 0 ? '0' : queryPerSecond.toFixed(2);
    mysqlStatus.txPerSecond = isNaN(txPerSecond) || txPerSecond === 0 ? '0' : txPerSecond.toFixed(2);
    mysqlStatus.file = res.data.File;
    mysqlStatus.position = res.data.Position;

    mysqlStatus.connInfo = res.data.Threads_running + '/' + res.data.Max_used_connections;
    mysqlStatus.threadCacheHit = isNaN(threadCacheHit) || threadCacheHit === 0 ? '0' : threadCacheHit.toFixed(2) + '%';
    mysqlStatus.indexHit = isNaN(indexHit) || indexHit === 0 ? '0' : indexHit.toFixed(2) + '%';
    mysqlStatus.innodbIndexHit = isNaN(innodbIndexHit) || innodbIndexHit === 0 ? '0' : innodbIndexHit.toFixed(2) + '%';
    mysqlStatus.cacheHit = isNaN(cacheHit) || cacheHit === 0 ? 'OFF' : cacheHit.toFixed(2) + '%';
    mysqlStatus.tmpTableToDB = isNaN(tmpTableToDB) || tmpTableToDB === 0 ? '0' : tmpTableToDB.toFixed(2) + '%';
    mysqlStatus.openTables = res.data.Open_tables;
    mysqlStatus.selectFullJoin = res.data.Select_full_join;
    mysqlStatus.selectRangeCheck = res.data.Select_range_check;
    mysqlStatus.sortMergePasses = res.data.Sort_merge_passes;
    mysqlStatus.tableLocksWaited = res.data.Table_locks_waited;
};

defineExpose({
    acceptParams,
    onClose,
});
</script>
