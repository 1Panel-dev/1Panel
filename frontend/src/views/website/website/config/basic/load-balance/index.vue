<template>
    <div>
        <ComplexTable :data="data" @search="search" v-loading="loading" :heightDiff="420">
            <template #toolbar>
                <el-button type="primary" plain @click="create()">
                    {{ $t('commons.button.create') }}
                </el-button>
                <el-alert :closable="false" class="!mt-2">
                    <template #default>
                        <span style="white-space: pre-line">{{ $t('website.loadBalanceHelper') }}</span>
                    </template>
                </el-alert>
            </template>
            <el-table-column :label="$t('commons.table.name')" prop="name"></el-table-column>
            <el-table-column :label="$t('website.algorithm')" prop="algorithm">
                <template #default="{ row }">
                    {{ getAlgorithm(row.algorithm) }}
                </template>
            </el-table-column>
            <el-table-column :label="$t('website.server')" prop="servers" minWidth="400px">
                <template #default="{ row }">
                    <table>
                        <tr v-for="(item, index) in row.servers" :key="index">
                            <td>
                                <el-tag>
                                    {{ item.server }}
                                </el-tag>
                            </td>
                            <td v-if="item.weight > 0">
                                <el-tag type="success">{{ $t('website.weight') }}: {{ item.weight }}</el-tag>
                            </td>
                            <td v-if="item.failTimeout != ''">
                                <el-tag type="warning">{{ $t('website.failTimeout') }}: {{ item.failTimeout }}</el-tag>
                            </td>
                            <td v-if="item.maxFails > 0">
                                <el-tag type="danger">{{ $t('website.maxFails') }}: {{ item.maxFails }}</el-tag>
                            </td>
                            <td v-if="item.maxConns > 0">
                                <el-tag>{{ $t('website.maxConns') }}: {{ item.maxConns }}</el-tag>
                            </td>
                            <td v-if="item.flag != ''">
                                <el-tag type="info">{{ $t('website.strategy') }}: {{ item.flag }}</el-tag>
                            </td>
                        </tr>
                    </table>
                </template>
            </el-table-column>
            <fu-table-operations
                :ellipsis="10"
                width="260px"
                :buttons="buttons"
                :label="$t('commons.table.operate')"
                fix
            />
        </ComplexTable>
    </div>
    <Operate ref="operateRef" @close="search()" />
    <OpDialog ref="delRef" @search="search()" />
    <File ref="fileRef" @search="search()" />
</template>

<script setup lang="ts">
import { deleteLoadBalance, getLoadBalances } from '@/api/modules/website';
import { onMounted, ref } from 'vue';
import Operate from './operate/index.vue';
import i18n from '@/lang';
import { Website } from '@/api/interface/website';
import { getAlgorithms } from '@/global/mimetype';
import File from './file/index.vue';

const props = defineProps({
    id: {
        type: Number,
        default: 0,
    },
});

const data = ref([]);
const loading = ref(false);
const operateRef = ref();
const delRef = ref();
const fileRef = ref();
const Algorithms = getAlgorithms();

const buttons = [
    {
        label: i18n.global.t('website.sourceFile'),
        click: function (row: any) {
            openEditFile(row);
        },
    },
    {
        label: i18n.global.t('commons.button.edit'),
        click: (row: any) => {
            update(row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: any) => {
            deleteLb(row);
        },
    },
];

const search = () => {
    getLoadBalances(props.id).then((res) => {
        data.value = res.data;
    });
};

const getAlgorithm = (key: string) => {
    let label = '';
    Algorithms.forEach((algorithm) => {
        if (algorithm.value === key) {
            label = algorithm.label;
        }
    });
    if (label === '') {
        return i18n.global.t('commons.table.default');
    }
    return label;
};

const deleteLb = async (row: Website.NginxUpstream) => {
    delRef.value.acceptParams({
        title: i18n.global.t('commons.button.delete'),
        names: [row.name],
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('website.loadBalance'),
            i18n.global.t('commons.button.delete'),
        ]),
        api: deleteLoadBalance,
        params: { websiteID: props.id, name: row.name },
    });
};

const create = () => {
    operateRef.value.acceptParams({
        websiteID: props.id,
        operate: 'create',
    });
};

const update = (row: Website.NginxUpstream) => {
    operateRef.value.acceptParams({
        websiteID: props.id,
        operate: 'edit',
        upstream: row,
    });
};

const openEditFile = (row: Website.NginxUpstream) => {
    fileRef.value.acceptParams({
        websiteID: props.id,
        name: row.name,
        content: row.content,
    });
};

onMounted(() => {
    search();
});
</script>
