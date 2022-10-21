<template>
    <div class="demo-collapse">
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
                                    <td>2022/10/12 14:46:35</td>
                                </tr>
                                <tr>
                                    <td>总连接次数</td>
                                    <!-- <td>{{ mysqlStatus!.Connections }}</td> -->
                                </tr>
                                <tr>
                                    <td>发送</td>
                                    <td>{{ (mysqlStatus!.Bytes_sent / 2014).toFixed(2) }} KB</td>
                                </tr>
                                <tr>
                                    <td>接收</td>
                                    <td>{{ (mysqlStatus!.Bytes_received / 2014).toFixed(2) }} KB</td>
                                </tr>
                            </table>
                        </el-col>
                        <el-col :span="6">
                            <table style="width: 100%" class="myTable">
                                <tr>
                                    <td>每秒查询</td>
                                    <td>0</td>
                                </tr>
                                <tr>
                                    <td>每秒事务</td>
                                    <td>0</td>
                                </tr>
                                <tr>
                                    <td>File</td>
                                    <td>{{ mysqlStatus!.File }}</td>
                                </tr>
                                <tr>
                                    <td>Position</td>
                                    <td>{{ mysqlStatus!.Position }}</td>
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
                                    <td>
                                        {{
                                            (mysqlStatus!.Threads_running / mysqlStatus!.Max_used_connections).toFixed(
                                                2,
                                            )
                                        }}%
                                    </td>
                                    <td>若值过大,增加max_connections</td>
                                </tr>
                                <tr>
                                    <td>线程缓存命中率</td>
                                    <td></td>
                                    <td>若过低,增加thread_cache_size</td>
                                </tr>
                                <tr>
                                    <td>索引命中率</td>
                                    <td>
                                        {{ (1 - mysqlStatus!.Key_reads / mysqlStatus!.Key_read_requests).toFixed(2) }}%
                                    </td>
                                    <td>若过低,增加key_buffer_size</td>
                                </tr>
                                <tr>
                                    <td>Innodb 索引命中率</td>
                                    <td>
                                        {{
                                            (
                                                1 -
                                                mysqlStatus!.Innodb_buffer_pool_reads /
                                                    mysqlStatus!.Innodb_buffer_pool_read_requests
                                            ).toFixed(2)
                                        }}%
                                    </td>
                                    <td>若过低,增加innodb_buffer_pool_size</td>
                                </tr>
                                <tr>
                                    <td>查询缓存命中率</td>
                                    <td>OFF</td>
                                    <td>若过低,增加query_cache_size</td>
                                </tr>
                                <tr>
                                    <td>创建临时表到磁盘</td>
                                    <td>13.62%</td>
                                    <td>若过大,尝试增加tmp_table_size</td>
                                </tr>
                                <tr>
                                    <td>已打开的表</td>
                                    <td>{{ mysqlStatus!.Open_tables }}</td>
                                    <td colspan="20">table_open_cache配置值应大于等于此值</td>
                                </tr>
                                <tr>
                                    <td>没有使用索引的量</td>
                                    <td>{{ mysqlStatus!.Select_full_join }}</td>
                                    <td>若不为0,请检查数据表的索引是否合理</td>
                                </tr>
                                <tr>
                                    <td>没有索引的JOIN量</td>
                                    <td>0</td>
                                    <td>若不为0,请检查数据表的索引是否合理</td>
                                </tr>
                                <tr>
                                    <td>排序后的合并次数</td>
                                    <td>{{ mysqlStatus!.Sort_merge_passes }}</td>
                                    <td>若值过大,增加sort_buffer_size</td>
                                </tr>
                                <tr>
                                    <td>锁表次数</td>
                                    <td>{{ mysqlStatus!.Table_locks_waited }}</td>
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
                                        <el-input clearable v-model="form.port">
                                            <template #append>MB</template>
                                        </el-input>
                                        <span class="input-help">用于索引的缓冲区大小</span>
                                    </el-form-item>
                                    <el-form-item label="query_cache_size">
                                        <el-input clearable v-model="form.port">
                                            <template #append>MB</template>
                                        </el-input>
                                        <span class="input-help">查询缓存,不开启请设为0</span>
                                    </el-form-item>
                                    <el-form-item label="tmp_table_size">
                                        <el-input clearable v-model="form.port">
                                            <template #append>MB</template>
                                        </el-input>
                                        <span class="input-help">临时表缓存大小</span>
                                    </el-form-item>
                                    <el-form-item label="innodb_buffer_pool_size">
                                        <el-input clearable v-model="form.port">
                                            <template #append>MB</template>
                                        </el-input>
                                        <span class="input-help">Innodb缓冲区大小</span>
                                    </el-form-item>
                                    <el-form-item label="innodb_log_buffer_size">
                                        <el-input clearable v-model="form.port">
                                            <template #append>MB</template>
                                        </el-input>
                                        <span class="input-help">Innodb日志缓冲区大小</span>
                                    </el-form-item>
                                    <el-form-item label="sort_buffer_size">
                                        <el-input clearable v-model="form.port">
                                            <template #append>KB</template>
                                        </el-input>
                                        <span class="input-help">* 连接数, 每个线程排序的缓冲大小</span>
                                    </el-form-item>
                                    <el-form-item label="read_buffer_size">
                                        <el-input clearable v-model="form.port">
                                            <template #append>KB</template>
                                        </el-input>
                                        <span class="input-help">* 连接数, 读入缓冲区大小</span>
                                    </el-form-item>

                                    <el-form-item>
                                        <el-button
                                            icon="Collection"
                                            type="primary"
                                            size="default"
                                            @click="onSave(panelFormRef, 'remoteAccess', form.remoteAccess)"
                                        >
                                            {{ $t('commons.button.save') }}
                                        </el-button>
                                        <el-button
                                            icon="RefreshLeft"
                                            size="default"
                                            @click="onSave(panelFormRef, 'remoteAccess', form.remoteAccess)"
                                        >
                                            重启数据库
                                        </el-button>
                                    </el-form-item>
                                </el-col>
                                <el-col :span="2"><br /></el-col>
                                <el-col :span="6">
                                    <el-form-item label="read_rnd_buffer_size">
                                        <el-input clearable v-model="form.port">
                                            <template #append>KB</template>
                                        </el-input>
                                        <span class="input-help">* 连接数, 随机读取缓冲区大小</span>
                                    </el-form-item>
                                    <el-form-item label="join_buffer_size">
                                        <el-input clearable v-model="form.port">
                                            <template #append>KB</template>
                                        </el-input>
                                        <span class="input-help">* 连接数, 关联表缓存大小</span>
                                    </el-form-item>
                                    <el-form-item label="thread_stack">
                                        <el-input clearable v-model="form.port">
                                            <template #append>KB</template>
                                        </el-input>
                                        <span class="input-help">* 连接数, 每个线程的堆栈大小</span>
                                    </el-form-item>
                                    <el-form-item label="binlog_cache_size">
                                        <el-input clearable v-model="form.port">
                                            <template #append>KB</template>
                                        </el-input>
                                        <span class="input-help">* 连接数, 二进制日志缓存大小(4096的倍数)</span>
                                    </el-form-item>
                                    <el-form-item label="thread_cache_size">
                                        <el-input clearable v-model="form.port" />
                                        <span class="input-help">线程池大小</span>
                                    </el-form-item>
                                    <el-form-item label="table_open_cache">
                                        <el-input clearable v-model="form.port" />
                                        <span class="input-help">表缓存</span>
                                    </el-form-item>
                                    <el-form-item label="max_connections">
                                        <el-input clearable v-model="form.port" />
                                        <span class="input-help">最大连接数</span>
                                    </el-form-item>
                                </el-col>
                            </el-row>
                        </el-form>
                    </el-card>
                </el-collapse-item>
                <el-collapse-item title="日志" name="4"></el-collapse-item>
            </el-collapse>
        </el-card>
    </div>
</template>

<script lang="ts" setup>
import { FormInstance } from 'element-plus';
import { onMounted, reactive, ref } from 'vue';
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import { LoadFile } from '@/api/modules/files';
import { loadMysqlStatus, loadMysqlVariables } from '@/api/modules/database';
import { Database } from '@/api/interface/database';

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
const mysqlVariables = ref();
const mysqlStatus = ref<Database.MysqlStatus>();

const onSave = async (formEl: FormInstance | undefined, key: string, val: any) => {
    console.log(formEl, key, val);
};
const loadMysqlConf = async (path: string) => {
    const res = await LoadFile({ path: path });
    mysqlConf.value = res.data;
};
const loadVariables = async () => {
    const res = await loadMysqlVariables();
    mysqlVariables.value = res.data;
};
const loadStatus = async () => {
    const res = await loadMysqlStatus();
    mysqlStatus.value = res.data;
};

onMounted(() => {
    console.log('dasdasd');
    loadMysqlConf('/opt/1Panel/conf/mysql.conf');
    loadStatus();
    loadVariables();
});
</script>
