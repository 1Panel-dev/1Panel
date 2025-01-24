<template>
    <div>
        <ComplexTable :data="data" @search="search()" :heightDiff="350" v-loading="loading">
            <template #toolbar>
                <el-button type="primary" @click="openOperate">{{ $t('commons.button.create') }}</el-button>
                <el-button type="primary" plain @click="buildNginx">{{ $t('nginx.build') }}</el-button>
                <el-text type="warning" class="!ml-2">{{ $t('nginx.buildHelper') }}</el-text>
            </template>
            <el-table-column prop="name" :label="$t('commons.table.name')" />
            <el-table-column prop="params" :label="$t('nginx.params')" />
            <el-table-column :label="$t('commons.table.status')" fix>
                <template #default="{ row }">
                    <el-switch v-model="row.enable" @click="updateModule(row)" />
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

const data = ref([]);
const loading = ref(false);
const buttons = [
    {
        label: i18n.global.t('commons.button.edit'),
        click: function (row: Nginx.NginxModule) {
            openEdit(row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
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
        .finally(() => {
            loading.value = false;
        });
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
