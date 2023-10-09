<template>
    <div>
        <el-form :model="mysqlVariables" :rules="variablesRules" ref="variableFormRef" label-position="top">
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
                    <el-form-item label="join_buffer_size" prop="join_buffer_size">
                        <el-input clearable v-model.number="mysqlVariables.join_buffer_size">
                            <template #append>KB</template>
                        </el-input>
                        <span class="input-help">{{ $t('database.joinBufferSizeHelper') }}</span>
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
                        <el-button @click="onSaveStart(variableFormRef)" type="primary">
                            {{ $t('commons.button.save') }}
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
                    <el-form-item v-if="showCacheSize()" label="query_cache_size" prop="query_cache_size">
                        <el-input clearable v-model.number="mysqlVariables.query_cache_size">
                            <template #append>MB</template>
                        </el-input>
                        <span class="input-help">{{ $t('database.queryCacheSizeHelper') }}</span>
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

        <ConfirmDialog ref="confirmDialogRef" @confirm="onSaveVariables"></ConfirmDialog>
    </div>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { checkNumberRange, Rules } from '@/global/form-rules';
import { FormInstance } from 'element-plus';
import { Database } from '@/api/interface/database';
import ConfirmDialog from '@/components/confirm-dialog/index.vue';
import { updateMysqlVariables } from '@/api/modules/database';
import i18n from '@/lang';
import { planOptions } from './../helper';
import { MsgSuccess } from '@/utils/message';

const plan = ref();
const confirmDialogRef = ref();

const variableFormRef = ref<FormInstance>();
const oldVariables = ref<Database.MysqlVariables>();
let mysqlVariables = reactive({
    mysqlName: '',
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

    slow_query_log: '',
    long_query_time: 0,
});
const variablesRules = reactive({
    key_buffer_size: [Rules.number, checkNumberRange(1, 102400)],
    query_cache_size: [Rules.number, checkNumberRange(0, 102400)],
    tmp_table_size: [Rules.number, checkNumberRange(1, 102400)],
    innodb_buffer_pool_size: [Rules.number, checkNumberRange(1, 102400)],
    innodb_log_buffer_size: [Rules.number, checkNumberRange(1, 102400)],
    sort_buffer_size: [Rules.number, checkNumberRange(1, 102400)],
    read_buffer_size: [Rules.number, checkNumberRange(1, 102400)],
    read_rnd_buffer_size: [Rules.number, checkNumberRange(1, 102400)],
    join_buffer_size: [Rules.number, checkNumberRange(1, 102400)],
    thread_stack: [Rules.number, checkNumberRange(1, 102400)],
    binlog_cache_size: [Rules.number, checkNumberRange(1, 102400)],
    thread_cache_size: [Rules.number, checkNumberRange(1, 10000)],
    table_open_cache: [Rules.number, checkNumberRange(1, 10000)],
    max_connections: [Rules.number, checkNumberRange(1, 10000)],

    slow_query_log: [Rules.requiredSelect],
    long_query_time: [Rules.number, checkNumberRange(1, 102400)],
});

const currentDB = reactive({
    type: '',
    database: '',
    version: '',
});
interface DialogProps {
    type: string;
    database: string;
    version: string;
    variables: Database.MysqlVariables;
}
const acceptParams = (params: DialogProps): void => {
    currentDB.type = params.type;
    currentDB.database = params.database;
    currentDB.version = params.version;
    mysqlVariables.key_buffer_size = Number(params.variables.key_buffer_size) / 1024 / 1024;
    mysqlVariables.query_cache_size = Number(params.variables.query_cache_size) / 1024 / 1024;
    mysqlVariables.tmp_table_size = Number(params.variables.tmp_table_size) / 1024 / 1024;
    mysqlVariables.innodb_buffer_pool_size = Number(params.variables.innodb_buffer_pool_size) / 1024 / 1024;
    mysqlVariables.innodb_log_buffer_size = Number(params.variables.innodb_log_buffer_size) / 1024 / 1024;

    mysqlVariables.sort_buffer_size = Number(params.variables.sort_buffer_size) / 1024;
    mysqlVariables.read_buffer_size = Number(params.variables.read_buffer_size) / 1024;
    mysqlVariables.read_rnd_buffer_size = Number(params.variables.read_rnd_buffer_size) / 1024;
    mysqlVariables.join_buffer_size = Number(params.variables.join_buffer_size) / 1024;
    mysqlVariables.thread_stack = Number(params.variables.thread_stack) / 1024;
    mysqlVariables.binlog_cache_size = Number(params.variables.binlog_cache_size) / 1024;
    mysqlVariables.thread_cache_size = Number(params.variables.thread_cache_size);
    mysqlVariables.table_open_cache = Number(params.variables.table_open_cache);
    mysqlVariables.max_connections = Number(params.variables.max_connections);
    oldVariables.value = { ...mysqlVariables };
};
const emit = defineEmits(['loading']);

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

const onSaveStart = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        let params = {
            header: i18n.global.t('database.confChange'),
            operationInfo: i18n.global.t('database.restartNowHelper'),
            submitInputInfo: i18n.global.t('database.restartNow'),
        };
        confirmDialogRef.value!.acceptParams(params);
    });
};

const onSaveVariables = async () => {
    let param = [] as Array<Database.VariablesUpdateHelper>;
    if (oldVariables.value?.key_buffer_size !== mysqlVariables.key_buffer_size) {
        param.push({ param: 'key_buffer_size', value: mysqlVariables.key_buffer_size * 1024 * 1024 });
    }
    if (oldVariables.value?.query_cache_size !== mysqlVariables.query_cache_size) {
        param.push({ param: 'query_cache_size', value: mysqlVariables.query_cache_size * 1024 * 1024 });
    }
    if (oldVariables.value?.tmp_table_size !== mysqlVariables.tmp_table_size) {
        param.push({ param: 'tmp_table_size', value: mysqlVariables.tmp_table_size * 1024 * 1024 });
    }
    if (oldVariables.value?.innodb_buffer_pool_size !== mysqlVariables.innodb_buffer_pool_size) {
        param.push({
            param: 'innodb_buffer_pool_size',
            value: mysqlVariables.innodb_buffer_pool_size * 1024 * 1024,
        });
    }
    if (oldVariables.value?.innodb_log_buffer_size !== mysqlVariables.innodb_log_buffer_size) {
        param.push({ param: 'innodb_log_buffer_size', value: mysqlVariables.innodb_log_buffer_size * 1024 * 1024 });
    }

    if (oldVariables.value?.sort_buffer_size !== mysqlVariables.sort_buffer_size) {
        param.push({ param: 'sort_buffer_size', value: mysqlVariables.sort_buffer_size * 1024 });
    }
    if (oldVariables.value?.read_buffer_size !== mysqlVariables.read_buffer_size) {
        param.push({ param: 'read_buffer_size', value: mysqlVariables.read_buffer_size * 1024 });
    }
    if (oldVariables.value?.read_rnd_buffer_size !== mysqlVariables.read_rnd_buffer_size) {
        param.push({ param: 'read_rnd_buffer_size', value: mysqlVariables.read_rnd_buffer_size * 1024 });
    }
    if (oldVariables.value?.join_buffer_size !== mysqlVariables.join_buffer_size) {
        param.push({ param: 'join_buffer_size', value: mysqlVariables.join_buffer_size * 1024 });
    }
    if (oldVariables.value?.thread_stack !== mysqlVariables.thread_stack) {
        param.push({ param: 'thread_stack', value: mysqlVariables.thread_stack * 1024 });
    }
    if (oldVariables.value?.binlog_cache_size !== mysqlVariables.binlog_cache_size) {
        param.push({ param: 'binlog_cache_size', value: mysqlVariables.binlog_cache_size * 1024 });
    }
    if (oldVariables.value?.thread_cache_size !== mysqlVariables.thread_cache_size) {
        param.push({ param: 'thread_cache_size', value: mysqlVariables.thread_cache_size });
    }
    if (oldVariables.value?.table_open_cache !== mysqlVariables.table_open_cache) {
        param.push({ param: 'table_open_cache', value: mysqlVariables.table_open_cache });
    }
    if (oldVariables.value?.max_connections !== mysqlVariables.max_connections) {
        param.push({ param: 'max_connections', value: mysqlVariables.max_connections });
    }
    emit('loading', true);
    let params = {
        type: currentDB.type,
        database: currentDB.database,
        variables: param,
    };
    await updateMysqlVariables(params)
        .then(() => {
            emit('loading', false);
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        })
        .catch(() => {
            emit('loading', false);
        });
};

const showCacheSize = () => {
    return currentDB.version?.startsWith('5.7');
};
defineExpose({
    acceptParams,
});
</script>
