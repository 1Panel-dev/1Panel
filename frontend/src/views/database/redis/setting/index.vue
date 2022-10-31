<template>
    <div class="demo-collapse" v-if="onSetting">
        <el-card>
            <el-collapse v-model="activeName" accordion>
                <el-collapse-item :title="$t('database.baseSetting')" name="1">
                    <el-form :model="baseInfo" ref="panelFormRef" :rules="rules" label-width="120px">
                        <el-row>
                            <el-col :span="1"><br /></el-col>
                            <el-col :span="10">
                                <el-form-item :label="$t('setting.port')" prop="port">
                                    <el-input clearable type="number" v-model.number="baseInfo.port">
                                        <template #append>
                                            <el-button
                                                @click="onSave(panelFormRef, 'port', baseInfo.port)"
                                                icon="Collection"
                                            >
                                                {{ $t('commons.button.save') }}
                                            </el-button>
                                        </template>
                                    </el-input>
                                </el-form-item>
                                <el-form-item :label="$t('setting.password')" prop="requirepass">
                                    <el-input type="password" show-password clearable v-model="baseInfo.requirepass">
                                        <template #append>
                                            <el-button
                                                @click="onSave(panelFormRef, 'password', baseInfo.requirepass)"
                                                icon="Collection"
                                            >
                                                {{ $t('commons.button.save') }}
                                            </el-button>
                                        </template>
                                    </el-input>
                                    <span class="input-help">{{ $t('database.requirepassHelper') }}</span>
                                </el-form-item>
                                <el-form-item :label="$t('database.timeout')" prop="timeout">
                                    <el-input clearable type="number" v-model.number="baseInfo.timeout">
                                        <template #append>
                                            <el-button
                                                @click="onSave(panelFormRef, 'timeout', baseInfo.timeout)"
                                                icon="Collection"
                                            >
                                                {{ $t('commons.button.save') }}
                                            </el-button>
                                        </template>
                                    </el-input>
                                    <span class="input-help">{{ $t('database.timeoutHelper') }}</span>
                                </el-form-item>
                                <el-form-item :label="$t('database.maxclients')" prop="maxclients">
                                    <el-input clearable type="number" v-model.number="baseInfo.maxclients">
                                        <template #append>
                                            <el-button
                                                @click="onSave(panelFormRef, 'maxclients', baseInfo.maxclients)"
                                                icon="Collection"
                                            >
                                                {{ $t('commons.button.save') }}
                                            </el-button>
                                        </template>
                                    </el-input>
                                </el-form-item>
                                <el-form-item :label="$t('database.databases')" prop="databases">
                                    <el-input clearable type="number" v-model.number="baseInfo.databases">
                                        <template #append>
                                            <el-button
                                                @click="onSave(panelFormRef, 'databases', baseInfo.databases)"
                                                icon="Collection"
                                            >
                                                {{ $t('commons.button.save') }}
                                            </el-button>
                                        </template>
                                    </el-input>
                                </el-form-item>
                                <el-form-item :label="$t('database.maxmemory')" prop="maxmemory">
                                    <el-input clearable type="number" v-model.number="baseInfo.maxmemory">
                                        <template #append>
                                            <el-button
                                                @click="onSave(panelFormRef, 'maxmemory', baseInfo.maxmemory)"
                                                icon="Collection"
                                            >
                                                {{ $t('commons.button.save') }}
                                            </el-button>
                                        </template>
                                    </el-input>
                                    <span class="input-help">{{ $t('database.maxmemoryHelper') }}</span>
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
                        @click="onSave(panelFormRef, 'remoteAccess', baseInfo.port)"
                    >
                        {{ $t('commons.button.save') }}
                    </el-button>
                </el-collapse-item>
                <el-collapse-item :title="$t('database.currentStatus')" name="3">
                    <el-row>
                        <el-col :span="1"><br /></el-col>
                        <el-col :span="12">
                            <table style="margin-top: 20px; width: 100%" class="myTable">
                                <tr>
                                    <td>uptime_in_days</td>
                                    <td>{{ redisStatus!.uptime_in_days }}</td>
                                    <td>{{ $t('database.uptimeInDays') }}</td>
                                </tr>
                                <tr>
                                    <td>tcp_port</td>
                                    <td>{{ redisStatus!.tcp_port }}</td>
                                    <td>{{ $t('database.tcpPort') }}</td>
                                </tr>
                                <tr>
                                    <td>connected_clients</td>
                                    <td>{{ redisStatus!.connected_clients }}</td>
                                    <td>{{ $t('database.connectedClients') }}</td>
                                </tr>
                                <tr>
                                    <td>used_memory_rss</td>
                                    <td>{{ redisStatus!.used_memory_rss }}</td>
                                    <td>{{ $t('database.usedMemoryRss') }}</td>
                                </tr>
                                <tr>
                                    <td>used_memory</td>
                                    <td>{{ redisStatus!.used_memory }}</td>
                                    <td>{{ $t('database.usedMemory') }}</td>
                                </tr>
                                <tr>
                                    <td>mem_fragmentation_ratio</td>
                                    <td>{{ redisStatus!.mem_fragmentation_ratio }}</td>
                                    <td>{{ $t('database.tmpTableToDBHelper') }}</td>
                                </tr>
                                <tr>
                                    <td>total_connections_received</td>
                                    <td>{{ redisStatus!.total_connections_received }}</td>
                                    <td>{{ $t('database.totalConnectionsReceived') }}</td>
                                </tr>
                                <tr>
                                    <td>total_commands_processed</td>
                                    <td>{{ redisStatus!.total_commands_processed }}</td>
                                    <td>{{ $t('database.totalCommandsProcessed') }}</td>
                                </tr>
                                <tr>
                                    <td>instantaneous_ops_per_sec</td>
                                    <td>{{ redisStatus!.instantaneous_ops_per_sec }}</td>
                                    <td>{{ $t('database.instantaneousOpsPerSec') }}</td>
                                </tr>
                                <tr>
                                    <td>keyspace_hits</td>
                                    <td>{{ redisStatus!.keyspace_hits }}</td>
                                    <td>{{ $t('database.keyspaceHits') }}</td>
                                </tr>
                                <tr>
                                    <td>keyspace_misses</td>
                                    <td>{{ redisStatus!.keyspace_misses }}</td>
                                    <td>{{ $t('database.keyspaceMisses') }}</td>
                                </tr>
                                <tr>
                                    <td>hit</td>
                                    <td>{{ redisStatus!.hit }}</td>
                                    <td>{{ $t('database.hit') }}</td>
                                </tr>
                                <tr>
                                    <td>latest_fork_usec</td>
                                    <td>{{ redisStatus!.latest_fork_usec }}</td>
                                    <td>{{ $t('database.latestForkUsec') }}</td>
                                </tr>
                            </table>
                        </el-col>
                    </el-row>
                </el-collapse-item>
                <el-collapse-item :title="$t('database.persistence')" name="4">
                    <el-form :model="baseInfo" ref="panelFormRef" label-width="120px">
                        <el-row>
                            <el-col :span="1"><br /></el-col>
                            <el-col :span="10">
                                <el-form-item label="appendonly" prop="appendonly">
                                    <el-switch v-model="baseInfo.appendonly"></el-switch>
                                </el-form-item>
                                <el-form-item label="appendfsync" prop="appendfsync">
                                    <el-radio-group v-model="baseInfo.appendfsync">
                                        <el-radio label="always">always</el-radio>
                                        <el-radio label="everysec">everysec</el-radio>
                                        <el-radio label="no">no</el-radio>
                                    </el-radio-group>
                                </el-form-item>
                            </el-col>
                        </el-row>
                    </el-form>
                </el-collapse-item>
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
import { loadRedisConf, loadRedisStatus, updateRedisConf } from '@/api/modules/database';
import i18n from '@/lang';
import { Rules } from '@/global/form-rules';

const extensions = [javascript(), oneDark];
const activeName = ref('1');

const baseInfo = reactive({
    port: 3306,
    requirepass: '',
    timeout: 0,
    maxclients: 0,
    databases: 0,
    maxmemory: 0,

    dir: '',
    appendonly: '',
    appendfsync: '',
    save: '',
});
const rules = reactive({
    port: [Rules.port],
    timeout: [Rules.number],
    maxclients: [Rules.number],
    databases: [Rules.number],
    maxmemory: [Rules.number],

    appendonly: [Rules.requiredSelect],
    appendfsync: [Rules.requiredSelect],
});
const panelFormRef = ref<FormInstance>();
const mysqlConf = ref();

let redisStatus = reactive({
    tcp_port: '',
    uptime_in_days: '',
    connected_clients: '',
    used_memory: '',
    used_memory_rss: '',
    used_memory_peak: '',
    mem_fragmentation_ratio: '',
    total_connections_received: '',
    total_commands_processed: '',
    instantaneous_ops_per_sec: '',
    keyspace_hits: '',
    keyspace_misses: '',
    hit: '',
    latest_fork_usec: '',
});

const onSetting = ref<boolean>(false);
const redisName = ref();
const db = ref();

interface DialogProps {
    redisName: string;
    db: number;
}
const acceptParams = (params: DialogProps): void => {
    onSetting.value = true;
    redisName.value = params.redisName;
    db.value = params.db;
    loadBaseInfo();
    loadStatus();
};
const onClose = (): void => {
    onSetting.value = false;
};

const onSave = async (formEl: FormInstance | undefined, key: string, val: any) => {
    if (!formEl) return;
    const result = await formEl.validateField(key, callback);
    if (!result) {
        return;
    }
    let changeForm = {
        redisName: redisName.value,
        db: 0,
        paramName: key,
        value: val + '',
    };
    await updateRedisConf(changeForm);
    ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
};
function callback(error: any) {
    if (error) {
        return error.message;
    } else {
        return;
    }
}

const loadBaseInfo = async () => {
    let params = {
        redisName: redisName.value,
        db: db.value,
    };
    const res = await loadRedisConf(params);
    baseInfo.timeout = Number(res.data?.timeout);
    baseInfo.maxclients = Number(res.data?.maxclients);
    baseInfo.databases = Number(res.data?.databases);
    baseInfo.requirepass = res.data?.requirepass;
    baseInfo.maxmemory = Number(res.data?.maxmemory);
    baseInfo.appendonly = res.data?.appendonly;
    baseInfo.appendfsync = res.data?.appendfsync;
    loadMysqlConf(`/opt/1Panel/data/apps/redis/${redisName.value}/conf/redis.conf`);
};

const loadMysqlConf = async (path: string) => {
    const res = await LoadFile({ path: path });
    mysqlConf.value = res.data;
};

const loadStatus = async () => {
    let params = {
        redisName: redisName.value,
        db: db.value,
    };
    const res = await loadRedisStatus(params);
    let hit = (
        (Number(res.data.keyspace_hits) / (Number(res.data.keyspace_hits) + Number(res.data.keyspace_misses))) *
        100
    ).toFixed(2);

    redisStatus.uptime_in_days = res.data.uptime_in_days;
    redisStatus.tcp_port = res.data.tcp_port;
    redisStatus.connected_clients = res.data.connected_clients;
    redisStatus.used_memory_rss = (Number(res.data.used_memory_rss) / 1024 / 1024).toFixed(2) + ' MB';
    redisStatus.used_memory = (Number(res.data.used_memory) / 1024 / 1024).toFixed(2) + ' MB';
    redisStatus.mem_fragmentation_ratio = res.data.mem_fragmentation_ratio;
    redisStatus.total_connections_received = res.data.total_connections_received;
    redisStatus.total_commands_processed = res.data.total_commands_processed;
    redisStatus.instantaneous_ops_per_sec = res.data.instantaneous_ops_per_sec;
    redisStatus.keyspace_hits = res.data.keyspace_hits;
    redisStatus.keyspace_misses = res.data.keyspace_misses;
    redisStatus.hit = hit;
    redisStatus.latest_fork_usec = res.data.latest_fork_usec;
};

defineExpose({
    acceptParams,
    onClose,
});
</script>
