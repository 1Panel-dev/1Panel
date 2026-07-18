<template>
    <div>
        <ComplexTable :data="data" @search="search()" :heightDiff="350" v-loading="loading">
            <template #toolbar>
                <el-button v-permission type="primary" @click="openOperate">
                    {{ $t('commons.button.create') }}
                </el-button>
                <el-button v-permission type="primary" plain @click="buildNginx">
                    {{ $t('nginx.build') }}
                </el-button>
                <el-text type="warning" class="!ml-2">{{ $t('nginx.buildHelper') }}</el-text>
            </template>
            <el-table-column prop="name" :label="$t('commons.table.name')" />
            <el-table-column :label="$t('nginx.buildMode')" width="150">
                <template #default="{ row }">
                    <el-tag effect="plain" :type="row.buildMode === 'static' ? 'warning' : 'primary'">
                        {{ $t('nginx.buildMode' + capitalize(row.buildMode)) }}
                    </el-tag>
                </template>
            </el-table-column>
            <el-table-column :label="$t('nginx.buildStatus')" width="120">
                <template #default="{ row }">
                    <el-tooltip v-if="row.lastError" :content="row.lastError" placement="top">
                        <el-tag :type="statusType(row.buildStatus)">{{ $t('nginx.' + row.buildStatus) }}</el-tag>
                    </el-tooltip>
                    <el-tag v-else :type="statusType(row.buildStatus)">{{ $t('nginx.' + row.buildStatus) }}</el-tag>
                </template>
            </el-table-column>
            <el-table-column :label="$t('nginx.compatibility')" width="130">
                <template #default="{ row }">
                    <el-tag effect="plain" :type="compatibilityType(row.compatibility)">
                        {{ $t('nginx.' + row.compatibility) }}
                    </el-tag>
                </template>
            </el-table-column>
            <el-table-column :label="$t('commons.table.status')" fix>
                <template #default="{ row }">
                    <el-switch v-permission v-model="row.enable" @change="updateModule(row)" />
                </template>
            </el-table-column>
            <fu-table-operations
                :ellipsis="2"
                width="200px"
                :buttons="buttons"
                :label="$t('commons.table.operate')"
                fixed="right"
                fix
            />
        </ComplexTable>
        <Operate ref="operateRef" @close="search" />
        <OpDialog ref="deleteRef" @search="search" @cancel="search" />
        <Build ref="buildRef" />
    </div>
</template>
<script lang="ts" setup>
import { getNginxModules, updateNginxModule } from '@/api/modules/nginx';
import i18n from '@/lang';
import { Nginx } from '@/api/interface/nginx';
import { MsgSuccess } from '@/utils/message';
import Operate from './operate/index.vue';
import Build from './build/index.vue';
import { onMounted, ref } from 'vue';

const data = ref<Nginx.NginxModule[]>([]);
const loading = ref(false);
const buttons = [
    {
        label: i18n.global.t('commons.button.edit'),
        permission: true,
        click: function (row: Nginx.NginxModule) {
            openEdit(row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        permission: true,
        click: function (row: Nginx.NginxModule) {
            deleteModule(row);
        },
    },
];
const operateRef = ref();
const deleteRef = ref();
const buildRef = ref();

const buildNginx = async () => {
    buildRef.value.acceptParams();
};

const search = () => {
    loading.value = true;
    getNginxModules()
        .then((res) => {
            data.value = res.data.modules;
        })
        .finally(() => {
            loading.value = false;
        });
};

const openOperate = () => {
    operateRef.value.acceptParams('create');
};

const openEdit = (row: Nginx.NginxModule) => {
    operateRef.value.acceptParams('update', row);
};

const updateModule = (row: Nginx.NginxModule) => {
    loading.value = true;
    const data = {
        ...row,
        operate: 'update',
    };
    updateNginxModule(data)
        .then(() => {
            MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
        })
        .catch(() => {
            row.enable = !row.enable;
        })
        .finally(() => {
            loading.value = false;
            search();
        });
};

const capitalize = (value: string) => value.charAt(0).toUpperCase() + value.slice(1);

const statusType = (status: string) => {
    if (status === 'ready') return 'success';
    if (status === 'failed') return 'danger';
    return 'info';
};

const compatibilityType = (status: string) => {
    if (status === 'compatible') return 'success';
    if (status === 'stale') return 'warning';
    if (status === 'static') return 'info';
    return 'info';
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
        api: updateNginxModule,
        params: data,
    });
};

onMounted(() => {
    search();
});
</script>
