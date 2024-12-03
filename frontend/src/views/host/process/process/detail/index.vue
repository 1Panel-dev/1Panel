<template>
    <el-drawer v-model="open" size="40%" :close-on-click-modal="false" :close-on-press-escape="false">
        <template #header>
            <DrawerHeader :header="$t('app.detail')" :back="handleClose" :resource="resourceName" />
        </template>
        <el-row>
            <el-col>
                <el-tabs v-model="activeName" type="card">
                    <el-tab-pane :label="$t('process.basic')" name="basic">
                        <el-descriptions :column="2" border>
                            <el-descriptions-item :label="$t('commons.table.name')" min-width="100px">
                                {{ data.name }}
                            </el-descriptions-item>
                            <el-descriptions-item :label="$t('process.status')">{{ data.status }}</el-descriptions-item>
                            <el-descriptions-item :label="$t('process.pid')">{{ data.PID }}</el-descriptions-item>
                            <el-descriptions-item :label="$t('process.ppid')">{{ data.PPID }}</el-descriptions-item>
                            <el-descriptions-item :label="$t('process.numThreads')">
                                {{ data.numThreads }}
                            </el-descriptions-item>
                            <el-descriptions-item :label="$t('process.numConnections')">
                                {{ data.numConnections }}
                            </el-descriptions-item>
                            <el-descriptions-item :label="$t('process.diskRead')">
                                {{ data.diskRead }}
                            </el-descriptions-item>
                            <el-descriptions-item :label="$t('process.diskWrite')">
                                {{ data.diskWrite }}
                            </el-descriptions-item>
                            <el-descriptions-item :label="$t('commons.table.user')">
                                {{ data.username }}
                            </el-descriptions-item>
                            <el-descriptions-item :label="$t('process.startTime')">
                                {{ data.startTime }}
                            </el-descriptions-item>
                            <el-descriptions-item :label="$t('process.cmdLine')">
                                {{ data.cmdLine }}
                            </el-descriptions-item>
                        </el-descriptions>
                    </el-tab-pane>
                    <el-tab-pane :label="$t('process.mem')" name="mem">
                        <el-descriptions :column="2" border>
                            <el-descriptions-item :label="'rss'">{{ data.rss }}</el-descriptions-item>
                            <el-descriptions-item :label="'swap'">{{ data.swap }}</el-descriptions-item>
                            <el-descriptions-item :label="'vms'">{{ data.vms }}</el-descriptions-item>
                            <el-descriptions-item :label="'hwm'">{{ data.hwm }}</el-descriptions-item>
                            <el-descriptions-item :label="'data'">{{ data.data }}</el-descriptions-item>
                            <el-descriptions-item :label="'stack'">{{ data.stack }}</el-descriptions-item>
                            <el-descriptions-item :label="'locked'">{{ data.locked }}</el-descriptions-item>
                        </el-descriptions>
                    </el-tab-pane>
                    <el-tab-pane :label="$t('process.openFiles')" name="openFiles">
                        <el-table :data="data.openFiles" border style="width: 100%">
                            <el-table-column prop="path" :label="$t('process.file')" />
                            <el-table-column prop="fd" label="fd" width="100px" />
                        </el-table>
                    </el-tab-pane>
                    <el-tab-pane :label="$t('process.env')" name="env">
                        <codemirror
                            :autofocus="true"
                            :indent-with-tab="true"
                            :tabSize="4"
                            style="height: calc(100vh - 200px)"
                            :lineWrapping="true"
                            :matchBrackets="true"
                            theme="cobalt"
                            :styleActiveLine="true"
                            :extensions="extensions"
                            v-model="envStr"
                            :disabled="true"
                        />
                    </el-tab-pane>
                    <el-tab-pane :label="$t('process.net')" name="net">
                        <el-table :data="data.connects" border style="width: 100%">
                            <el-table-column prop="localaddr" :label="$t('process.laddr')">
                                <template #default="{ row }">
                                    <span>{{ row.localaddr.ip }}</span>
                                    <span v-if="row.localaddr.port > 0">:{{ row.localaddr.port }}</span>
                                </template>
                            </el-table-column>
                            <el-table-column prop="remoteaddr" :label="$t('process.raddr')">
                                <template #default="{ row }">
                                    <span>{{ row.remoteaddr.ip }}</span>
                                    <span v-if="row.remoteaddr.port > 0">:{{ row.remoteaddr.port }}</span>
                                </template>
                            </el-table-column>
                            <el-table-column prop="status" :label="$t('process.state')" />
                        </el-table>
                    </el-tab-pane>
                </el-tabs>
            </el-col>
        </el-row>
    </el-drawer>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';

interface InfoProps {
    info: object;
}

let open = ref(false);
let data = ref();
const resourceName = ref('');
const activeName = ref('basic');
const envStr = ref('');

const extensions = [javascript(), oneDark];

const handleClose = () => {
    open.value = false;
};

const acceptParams = async (params: InfoProps): Promise<void> => {
    activeName.value = 'basic';
    data.value = params.info;
    resourceName.value = data.value.name;
    if (data.value.envs) {
        envStr.value = data.value.envs.join('\n');
    }
    open.value = true;
};

defineExpose({
    acceptParams,
});
</script>
