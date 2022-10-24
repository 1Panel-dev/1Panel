<template>
    <div class="demo-collapse" v-if="onSetting">
        <el-card>
            <el-collapse v-model="activeName" accordion>
                <el-collapse-item title="基础设置" name="1">
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
                <el-collapse-item title="配置修改" name="2">
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
                <el-collapse-item title="当前状态" name="3">
                    <el-row :gutter="20">
                        <el-col :span="1"><br /></el-col>
                        <el-col :span="6">
                            <table style="width: 100%" class="myTable">
                                <tr>
                                    <td>启动时间</td>
                                    <td>{{ mysqlStatus.run }}</td>
                                </tr>
                                <tr>
                                    <td>总连接次数</td>
                                    <td>{{ mysqlStatus.connections }}</td>
                                </tr>
                                <tr>
                                    <td>发送</td>
                                    <td>{{ mysqlStatus!.bytesSent }}</td>
                                </tr>
                                <tr>
                                    <td>接收</td>
                                    <td>{{ mysqlStatus!.bytesReceived }}</td>
                                </tr>
                            </table>
                        </el-col>
                        <el-col :span="6">
                            <table style="width: 100%" class="myTable">
                                <tr>
                                    <td>每秒查询</td>
                                    <td>{{ mysqlStatus!.queryPerSecond }}</td>
                                </tr>
                                <tr>
                                    <td>每秒事务</td>
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
                                    <td>活动/峰值连接数</td>
                                    <td>{{ mysqlStatus!.connInfo }}</td>
                                    <td>若值过大,增加max_connections</td>
                                </tr>
                                <tr>
                                    <td>线程缓存命中率</td>
                                    <td>{{ mysqlStatus!.threadCacheHit }}</td>
                                    <td>若过低,增加thread_cache_size</td>
                                </tr>
                                <tr>
                                    <td>索引命中率</td>
                                    <td>{{ mysqlStatus!.indexHit }}</td>
                                    <td>若过低,增加key_buffer_size</td>
                                </tr>
                                <tr>
                                    <td>Innodb 索引命中率</td>
                                    <td>{{ mysqlStatus!.innodbIndexHit }}</td>
                                    <td>若过低,增加innodb_buffer_pool_size</td>
                                </tr>
                                <tr>
                                    <td>查询缓存命中率</td>
                                    <td>{{ mysqlStatus!.cacheHit }}</td>
                                    <td>若过低,增加query_cache_size</td>
                                </tr>
                                <tr>
                                    <td>创建临时表到磁盘</td>
                                    <td>{{ mysqlStatus!.tmpTableToDB }}</td>
                                    <td>若过大,尝试增加tmp_table_size</td>
                                </tr>
                                <tr>
                                    <td>已打开的表</td>
                                    <td>{{ mysqlStatus!.openTables }}</td>
                                    <td>table_open_cache配置值应大于等于此值</td>
                                </tr>
                                <tr>
                                    <td>没有使用索引的量</td>
                                    <td>{{ mysqlStatus!.selectFullJoin }}</td>
                                    <td>若不为0,请检查数据表的索引是否合理</td>
                                </tr>
                                <tr>
                                    <td>没有索引的JOIN量</td>
                                    <td>{{ mysqlStatus!.selectRangeCheck }}</td>
                                    <td>若不为0,请检查数据表的索引是否合理</td>
                                </tr>
                                <tr>
                                    <td>排序后的合并次数</td>
                                    <td>{{ mysqlStatus!.sortMergePasses }}</td>
                                    <td>若值过大,增加sort_buffer_size</td>
                                </tr>
                                <tr>
                                    <td>锁表次数</td>
                                    <td>{{ mysqlStatus!.tableLocksWaited }}</td>
                                    <td>若值过大,请考虑增加您的数据库性能</td>
                                </tr>
                            </table>
                        </el-col>
                    </el-row>
                </el-collapse-item>
                <el-collapse-item title="性能调整" name="4">
                    <el-card>
                        <el-form :model="form" ref="panelFormRef" label-width="160px">
                            <el-row>
                                <el-col :span="1"><br /></el-col>
                                <el-col :span="6">
                                    <el-form-item label="key_buffer_size">
                                        <el-input clearable v-model="mysqlVariables.key_buffer_size">
                                            <template #append>MB</template>
                                        </el-input>
                                        <span class="input-help">用于索引的缓冲区大小</span>
                                    </el-form-item>
                                    <el-form-item label="query_cache_size">
                                        <el-input clearable v-model="mysqlVariables.query_cache_size">
                                            <template #append>MB</template>
                                        </el-input>
                                        <span class="input-help">查询缓存,不开启请设为0</span>
                                    </el-form-item>
                                    <el-form-item label="tmp_table_size">
                                        <el-input clearable v-model="mysqlVariables.tmp_table_size">
                                            <template #append>MB</template>
                                        </el-input>
                                        <span class="input-help">临时表缓存大小</span>
                                    </el-form-item>
                                    <el-form-item label="innodb_buffer_pool_size">
                                        <el-input clearable v-model="mysqlVariables.innodb_buffer_pool_size">
                                            <template #append>MB</template>
                                        </el-input>
                                        <span class="input-help">Innodb缓冲区大小</span>
                                    </el-form-item>
                                    <el-form-item label="innodb_log_buffer_size">
                                        <el-input clearable v-model="mysqlVariables.innodb_log_buffer_size">
                                            <template #append>MB</template>
                                        </el-input>
                                        <span class="input-help">Innodb日志缓冲区大小</span>
                                    </el-form-item>
                                    <el-form-item label="sort_buffer_size">
                                        <el-input clearable v-model="mysqlVariables.sort_buffer_size">
                                            <template #append>KB</template>
                                        </el-input>
                                        <span class="input-help">* 连接数, 每个线程排序的缓冲大小</span>
                                    </el-form-item>
                                    <el-form-item label="read_buffer_size">
                                        <el-input clearable v-model="mysqlVariables.read_buffer_size">
                                            <template #append>KB</template>
                                        </el-input>
                                        <span class="input-help">* 连接数, 读入缓冲区大小</span>
                                    </el-form-item>

                                    <el-form-item>
                                        <el-button icon="Collection" type="primary" size="default">
                                            {{ $t('commons.button.save') }}
                                        </el-button>
                                        <el-button icon="RefreshLeft" size="default">重启数据库</el-button>
                                    </el-form-item>
                                </el-col>
                                <el-col :span="2"><br /></el-col>
                                <el-col :span="6">
                                    <el-form-item label="read_rnd_buffer_size">
                                        <el-input clearable v-model="mysqlVariables.read_rnd_buffer_size">
                                            <template #append>KB</template>
                                        </el-input>
                                        <span class="input-help">* 连接数, 随机读取缓冲区大小</span>
                                    </el-form-item>
                                    <el-form-item label="join_buffer_size">
                                        <el-input clearable v-model="mysqlVariables.join_buffer_size">
                                            <template #append>KB</template>
                                        </el-input>
                                        <span class="input-help">* 连接数, 关联表缓存大小</span>
                                    </el-form-item>
                                    <el-form-item label="thread_stack">
                                        <el-input clearable v-model="mysqlVariables.thread_stack">
                                            <template #append>KB</template>
                                        </el-input>
                                        <span class="input-help">* 连接数, 每个线程的堆栈大小</span>
                                    </el-form-item>
                                    <el-form-item label="binlog_cache_size">
                                        <el-input clearable v-model="mysqlVariables.binlog_cache_size">
                                            <template #append>KB</template>
                                        </el-input>
                                        <span class="input-help">* 连接数, 二进制日志缓存大小(4096的倍数)</span>
                                    </el-form-item>
                                    <el-form-item label="thread_cache_size">
                                        <el-input clearable v-model="mysqlVariables.thread_cache_size" />
                                        <span class="input-help">线程池大小</span>
                                    </el-form-item>
                                    <el-form-item label="table_open_cache">
                                        <el-input clearable v-model="mysqlVariables.table_open_cache" />
                                        <span class="input-help">表缓存</span>
                                    </el-form-item>
                                    <el-form-item label="max_connections">
                                        <el-input clearable v-model="mysqlVariables.max_connections" />
                                        <span class="input-help">最大连接数</span>
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
import { FormInstance } from 'element-plus';
import { reactive, ref } from 'vue';
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import { LoadFile } from '@/api/modules/files';
import { loadMysqlStatus, loadMysqlVariables } from '@/api/modules/database';
import { Database } from '@/api/interface/database';
import { computeSize } from '@/utils/util';

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
const mysqlVariables = reactive<Database.MysqlVariables>({
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

const acceptParams = (): void => {
    onSetting.value = true;
    loadMysqlConf('/opt/1Panel/conf/mysql.conf');
    loadStatus();
    loadVariables();
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
    mysqlVariables.key_buffer_size = res.data.key_buffer_size / 1024 / 1024;
    mysqlVariables.query_cache_size = res.data.query_cache_size / 1024 / 1024;
    mysqlVariables.tmp_table_size = res.data.tmp_table_size / 1024 / 1024;
    mysqlVariables.innodb_buffer_pool_size = res.data.innodb_buffer_pool_size / 1024 / 1024;
    mysqlVariables.innodb_log_buffer_size = res.data.innodb_log_buffer_size / 1024 / 1024;

    mysqlVariables.sort_buffer_size = res.data.sort_buffer_size / 1024;
    mysqlVariables.read_buffer_size = res.data.read_buffer_size / 1024;
    mysqlVariables.read_rnd_buffer_size = res.data.read_rnd_buffer_size / 1024;
    mysqlVariables.join_buffer_size = res.data.join_buffer_size / 1024;
    mysqlVariables.thread_stack = res.data.thread_stack / 1024;
    mysqlVariables.binlog_cache_size = res.data.binlog_cache_size / 1024;
    mysqlVariables.thread_cache_size = res.data.thread_cache_size;
    mysqlVariables.table_open_cache = res.data.table_open_cache;
    mysqlVariables.max_connections = res.data.max_connections;
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
