<template>
    <div>
        <el-card style="margin-top: 20px">
            <ComplexTable :pagination-config="paginationConfig" v-model:selects="selects" :data="data" @search="search">
                <template #toolbar>
                    <el-button-group>
                        <el-button :disabled="checkStatus('start')" @click="onOperate('start')">
                            {{ $t('container.start') }}
                        </el-button>
                        <el-button :disabled="checkStatus('stop')" @click="onOperate('stop')">
                            {{ $t('container.stop') }}
                        </el-button>
                        <el-button :disabled="checkStatus('reStart')" @click="onOperate('reStart')">
                            {{ $t('container.reStart') }}
                        </el-button>
                        <el-button :disabled="checkStatus('kill')" @click="onOperate('kill')">
                            {{ $t('container.kill') }}
                        </el-button>
                        <el-button :disabled="checkStatus('pause')" @click="onOperate('pause')">
                            {{ $t('container.pause') }}
                        </el-button>
                        <el-button :disabled="checkStatus('unPause')" @click="onOperate('unPause')">
                            {{ $t('container.unPause') }}
                        </el-button>
                        <el-button :disabled="checkStatus('remove')" @click="onOperate('remove')">
                            {{ $t('container.remove') }}
                        </el-button>
                    </el-button-group>
                    <el-button icon="Plus" style="margin-left: 10px" @click="onCreate()">
                        {{ $t('commons.button.create') }}
                    </el-button>
                </template>
                <el-table-column type="selection" fix />
                <el-table-column
                    :label="$t('commons.table.name')"
                    show-overflow-tooltip
                    min-width="100"
                    prop="name"
                    fix
                >
                    <template #default="{ row }">
                        <el-link @click="onInspect(row.containerID)" type="primary">{{ row.name }}</el-link>
                    </template>
                </el-table-column>
                <el-table-column
                    :label="$t('container.image')"
                    show-overflow-tooltip
                    min-width="100"
                    prop="imageName"
                />
                <el-table-column :label="$t('commons.table.status')" min-width="50" prop="state" fix />
                <el-table-column :label="$t('container.upTime')" min-width="100" prop="runTime" fix />
                <el-table-column
                    prop="createTime"
                    :label="$t('commons.table.date')"
                    :formatter="dateFromat"
                    show-overflow-tooltip
                />
                <fu-table-operations :buttons="buttons" :label="$t('commons.table.operate')" fix />
            </ComplexTable>
        </el-card>

        <el-dialog v-model="detailVisiable" :destroy-on-close="true" :close-on-click-modal="false" width="70%">
            <template #header>
                <div class="card-header">
                    <span>{{ $t('commons.button.view') }}</span>
                </div>
            </template>
            <codemirror
                :autofocus="true"
                placeholder="None data"
                :indent-with-tab="true"
                :tabSize="4"
                style="max-height: 500px"
                :lineWrapping="true"
                :matchBrackets="true"
                theme="cobalt"
                :styleActiveLine="true"
                :extensions="extensions"
                v-model="detailInfo"
                :readOnly="true"
            />
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="detailVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                </span>
            </template>
        </el-dialog>

        <el-dialog
            @close="onCloseLog"
            v-model="logVisiable"
            :destroy-on-close="true"
            :close-on-click-modal="false"
            width="70%"
        >
            <template #header>
                <div class="card-header">
                    <span>{{ $t('commons.button.log') }}</span>
                </div>
            </template>
            <div>
                <el-select @change="searchLogs" style="width: 10%; float: left" v-model="logSearch.mode">
                    <el-option v-for="item in timeOptions" :key="item.label" :value="item.value" :label="item.label" />
                </el-select>
                <div style="margin-left: 20px; float: left">
                    <el-checkbox border v-model="logSearch.isWatch">{{ $t('commons.button.watch') }}</el-checkbox>
                </div>
                <el-button style="margin-left: 20px" @click="onDownload" icon="Download">
                    {{ $t('file.download') }}
                </el-button>
            </div>

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
                v-model="logInfo"
                :readOnly="true"
            />
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="logVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                </span>
            </template>
        </el-dialog>

        <el-dialog
            @close="onCloseLog"
            v-model="newNameVisiable"
            :destroy-on-close="true"
            :close-on-click-modal="false"
            width="30%"
        >
            <template #header>
                <div class="card-header">
                    <span>{{ $t('container.reName') }}</span>
                </div>
            </template>
            <el-form ref="newNameRef" :model="reNameForm">
                <el-form-item label="新名称" :rules="Rules.requiredInput" prop="newName">
                    <el-input v-model="reNameForm.newName"></el-input>
                </el-form-item>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="newNameVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button @click="onSubmitName(newNameRef)">{{ $t('commons.button.confirm') }}</el-button>
                </span>
            </template>
        </el-dialog>
        <CreateDialog ref="dialogCreateRef" />
    </div>
</template>

<script lang="ts" setup>
import ComplexTable from '@/components/complex-table/index.vue';
import CreateDialog from '@/views/container/container/create/index.vue';
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import { reactive, onMounted, ref } from 'vue';
import { dateFromat, dateFromatForName } from '@/utils/util';
import { Rules } from '@/global/form-rules';
import { ContainerOperator, inspect, getContainerLog, getContainerPage } from '@/api/modules/container';
import { Container } from '@/api/interface/container';
import { ElForm, ElMessage, ElMessageBox, FormInstance } from 'element-plus';
import i18n from '@/lang';

const data = ref();
const selects = ref<any>([]);
const paginationConfig = reactive({
    page: 1,
    pageSize: 10,
    total: 0,
});
const containerSearch = reactive({
    page: 1,
    pageSize: 5,
    status: 'all',
});

const detailVisiable = ref<boolean>(false);
const detailInfo = ref();
const extensions = [javascript(), oneDark];
const logVisiable = ref<boolean>(false);
const logInfo = ref();
const logSearch = reactive({
    isWatch: false,
    container: '',
    containerID: '',
    mode: 'all',
});
let timer: NodeJS.Timer | null = null;

const newNameVisiable = ref<boolean>(false);
type FormInstance = InstanceType<typeof ElForm>;
const newNameRef = ref<FormInstance>();
const reNameForm = reactive({
    containerID: '',
    operation: 'reName',
    newName: '',
});

const timeOptions = ref([
    { label: i18n.global.t('container.all'), value: 'all' },
    {
        label: i18n.global.t('container.lastDay'),
        value: new Date(new Date().getTime() - 3600 * 1000 * 24 * 1).getTime() / 1000 + '',
    },
    {
        label: i18n.global.t('container.last4Hour'),
        value: new Date(new Date().getTime() - 3600 * 1000 * 4).getTime() / 1000 + '',
    },
    {
        label: i18n.global.t('container.lastHour'),
        value: new Date(new Date().getTime() - 3600 * 1000).getTime() / 1000 + '',
    },
    {
        label: i18n.global.t('container.last10Min'),
        value: new Date(new Date().getTime() - 600 * 1000).getTime() / 1000 + '',
    },
]);

const search = async () => {
    containerSearch.page = paginationConfig.page;
    containerSearch.pageSize = paginationConfig.pageSize;
    await getContainerPage(containerSearch).then((res) => {
        if (res.data) {
            data.value = res.data.items;
        }
    });
};

const dialogCreateRef = ref<DialogExpose>();

interface DialogExpose {
    acceptParams: () => void;
}
const onCreate = async () => {
    dialogCreateRef.value!.acceptParams();
};

const onInspect = async (id: string) => {
    const res = await inspect({ id: id, type: 'container' });
    detailInfo.value = JSON.stringify(JSON.parse(res.data), null, 2);
    detailVisiable.value = true;
};

const onLog = async (row: Container.ContainerInfo) => {
    logSearch.container = row.name;
    logSearch.containerID = row.containerID;
    searchLogs();
    logVisiable.value = true;
    timer = setInterval(() => {
        if (logVisiable.value && logSearch.isWatch) {
            searchLogs();
        }
    }, 1000 * 5);
};
const onCloseLog = async () => {
    clearInterval(Number(timer));
};
const searchLogs = async () => {
    const res = await getContainerLog(logSearch);
    logInfo.value = res.data;
};
const onDownload = async () => {
    const downloadUrl = window.URL.createObjectURL(new Blob([logInfo.value]));
    const a = document.createElement('a');
    a.style.display = 'none';
    a.href = downloadUrl;
    a.download = logSearch.container + '-' + dateFromatForName(new Date()) + '.log';
    const event = new MouseEvent('click');
    a.dispatchEvent(event);
};

const onRename = async (row: Container.ContainerInfo) => {
    reNameForm.containerID = row.containerID;
    reNameForm.newName = '';
    newNameVisiable.value = true;
};
const onSubmitName = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        ContainerOperator(reNameForm);
        search();
        newNameVisiable.value = false;
        ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
    });
};

const checkStatus = (operation: string) => {
    if (selects.value.length < 1) {
        return true;
    }
    switch (operation) {
        case 'start':
            for (const item of selects.value) {
                if (item.state === 'running') {
                    return true;
                }
            }
            return false;
        case 'stop':
            for (const item of selects.value) {
                if (item.state === 'stopped') {
                    return true;
                }
            }
            return false;
        case 'pause':
            for (const item of selects.value) {
                if (item.state === 'paused') {
                    return true;
                }
            }
            return false;
        case 'unPause':
            for (const item of selects.value) {
                if (item.state !== 'paused') {
                    return true;
                }
            }
            return false;
    }
};
const onOperate = async (operation: string) => {
    ElMessageBox.confirm(
        i18n.global.t('container.operatorHelper', [operation]),
        i18n.global.t('container.' + operation),
        {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        },
    ).then(() => {
        let ps = [];
        for (const item of selects.value) {
            const param = {
                containerID: item.containerID,
                operation: operation,
                newName: '',
            };
            ps.push(ContainerOperator(param));
        }
        Promise.all(ps)
            .then(() => {
                search();
                ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
            })
            .catch(() => {
                search();
            });
    });
};

const buttons = [
    {
        label: i18n.global.t('container.reName'),
        click: (row: Container.ContainerInfo) => {
            onRename(row);
        },
    },
    {
        label: i18n.global.t('commons.button.log'),
        click: (row: Container.ContainerInfo) => {
            onLog(row);
        },
    },
];

onMounted(() => {
    search();
});
</script>
