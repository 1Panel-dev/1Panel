<template>
    <div>
        <ComplexTable :data="data" @search="search()" :heightDiff="350" v-loading="loading">
            <template #toolbar>
                <el-button type="primary" @click="openOperate">{{ $t('commons.button.create') }}</el-button>
                <el-button type="primary" plain @click="buildNginx">{{ $t('nginx.build') }}</el-button>
            </template>
            <el-table-column prop="name" :label="$t('commons.table.name')" />
            <el-table-column prop="params" :label="$t('nginx.params')" />
            <el-table-column :label="$t('commons.table.status')" fix>
                <template #default="{ row }">
                    <el-switch v-model="row.enable" />
                </template>
            </el-table-column>
            <fu-table-operations
                :ellipsis="2"
                width="100px"
                :buttons="buttons"
                :label="$t('commons.table.operate')"
                fixed="right"
                fix
            />
        </ComplexTable>
        <TaskLog ref="taskLogRef" />
        <Operate ref="operateRef" @close="search" />
        <OpDialog ref="deleteRef" @search="search" @cancel="search" />
    </div>
</template>
<script lang="ts" setup>
import { BuildNginx, GetNginxModules, UpdateNginxModule } from '@/api/modules/nginx';
import { newUUID } from '@/utils/util';
import TaskLog from '@/components/task-log/index.vue';
import Operate from './operate/index.vue';
import i18n from '@/lang';
import { Nginx } from '@/api/interface/nginx';

const taskLogRef = ref();
const data = ref([]);
const loading = ref(false);
const buttons = [
    {
        label: i18n.global.t('commons.button.delete'),
        click: function (row: Nginx.NginxModule) {
            deleteModule(row);
        },
    },
];
const operateRef = ref();
const deleteRef = ref();

const buildNginx = async () => {
    ElMessageBox.confirm(i18n.global.t('nginx.buildWarn'), i18n.global.t('nginx.build'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    }).then(async () => {
        const taskID = newUUID();
        try {
            await BuildNginx({
                taskID: taskID,
            });
            openTaskLog(taskID);
        } catch (error) {}
    });
};

const search = () => {
    loading.value = true;
    GetNginxModules()
        .then((res) => {
            data.value = res.data;
        })
        .finally(() => {
            loading.value = false;
        });
};

const openTaskLog = (taskID: string) => {
    taskLogRef.value.openWithTaskID(taskID);
};

const openOperate = () => {
    operateRef.value.acceptParams();
};

const deleteModule = async (row: Nginx.NginxModule) => {
    const data = {
        name: row.name,
        operate: 'delete',
    };
    deleteRef.value.acceptParams({
        title: i18n.global.t('commons.button.delete'),
        names: [row.name],
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('nginx.module'),
            i18n.global.t('commons.button.delete'),
        ]),
        api: UpdateNginxModule,
        params: data,
    });
};

onMounted(() => {
    search();
});
</script>
