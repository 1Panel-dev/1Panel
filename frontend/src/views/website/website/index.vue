<template>
    <div>
        <AppStatus :app-key="'nginx'" @setting="setting" @is-exist="checkExist"></AppStatus>
        <LayoutContent v-if="nginxIsExist">
            <br />
            <el-card v-if="!openNginxConfig">
                <LayoutContent :header="$t('website.website')">
                    <ComplexTable :pagination-config="paginationConfig" :data="data" @search="search()">
                        <template #toolbar>
                            <el-button type="primary" plain @click="openCreate">
                                {{ $t('commons.button.create') }}
                            </el-button>
                            <el-button type="primary" plain @click="openGroup">{{ $t('website.group') }}</el-button>
                        </template>
                        <el-table-column
                            :label="$t('commons.table.name')"
                            fix
                            show-overflow-tooltip
                            prop="primaryDomain"
                        >
                            <template #default="{ row }">
                                <el-link @click="openConfig(row.id)">{{ row.primaryDomain }}</el-link>
                            </template>
                        </el-table-column>
                        <el-table-column :label="$t('commons.table.status')" prop="status"></el-table-column>
                        <el-table-column :label="$t('website.remark')" prop="remark"></el-table-column>
                        <el-table-column :label="$t('website.protocol')" prop="protocol"></el-table-column>
                        <el-table-column :label="$t('website.expireDate')">
                            <template #default="{ row }">
                                <span v-if="row.protocol === 'HTTP'">{{ $t('website.neverExpire') }}</span>
                                <span v-else>{{ dateFromat(1, 1, row.webSiteSSL.expireDate) }}</span>
                            </template>
                        </el-table-column>
                        <fu-table-operations
                            :ellipsis="10"
                            width="260px"
                            :buttons="buttons"
                            :label="$t('commons.table.operate')"
                            fixed="right"
                            fix
                        />
                    </ComplexTable>
                </LayoutContent>
            </el-card>
            <el-card v-if="openNginxConfig">
                <NginxConfig :containerName="containerName"></NginxConfig>
            </el-card>

            <CreateWebSite ref="createRef" @close="search"></CreateWebSite>
            <DeleteWebsite ref="deleteRef" @close="search"></DeleteWebsite>
            <WebSiteGroup ref="groupRef"></WebSiteGroup>
            <UploadDialog ref="uploadRef" />
            <BackupRecords ref="dialogBackupRef" />
        </LayoutContent>
    </div>
</template>

<script lang="ts" setup>
import LayoutContent from '@/layout/layout-content.vue';
import BackupRecords from '@/views/database/mysql/backup/index.vue';
import UploadDialog from '@/views/database/mysql/upload/index.vue';
import ComplexTable from '@/components/complex-table/index.vue';
import { onMounted, reactive, ref } from '@vue/runtime-core';
import CreateWebSite from './create/index.vue';
import DeleteWebsite from './delete/index.vue';
import WebSiteGroup from './group/index.vue';
import { SearchWebSites } from '@/api/modules/website';
import { WebSite } from '@/api/interface/website';
import AppStatus from '@/components/app-status/index.vue';
import NginxConfig from './nginx/index.vue';
import { dateFromat } from '@/utils/util';

import i18n from '@/lang';
import router from '@/routers';
import { App } from '@/api/interface/app';

const createRef = ref();
const deleteRef = ref();
const groupRef = ref();
let openNginxConfig = ref(false);
let nginxIsExist = ref(false);
let containerName = ref('');

const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 20,
    total: 0,
});

const data = ref();
const search = async () => {
    const req = {
        name: '',
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
    };

    SearchWebSites(req).then((res) => {
        data.value = res.data.items;
        paginationConfig.total = res.data.total;
    });
};
const setting = () => {
    openNginxConfig.value = true;
};

const openConfig = (id: number) => {
    router.push({ name: 'WebsiteConfig', params: { id: id, tab: 'basic' } });
};

const uploadRef = ref();
const dialogBackupRef = ref();
const onOpenBackupDialog = async (dbName: string) => {
    let params = {
        mysqlName: 'test',
        dbName: dbName,
    };
    dialogBackupRef.value!.acceptParams(params);
};
const buttons = [
    {
        label: i18n.global.t('website.config'),
        click: function (row: WebSite.WebSite) {
            openConfig(row.id);
        },
    },
    {
        label: i18n.global.t('database.backupList'),
        click: (row: WebSite.WebSite) => {
            onOpenBackupDialog(row.primaryDomain);
        },
    },
    {
        label: i18n.global.t('database.loadBackup'),
        click: (row: WebSite.WebSite) => {
            let params = {
                mysqlName: 'test',
                dbName: row.primaryDomain,
            };
            uploadRef.value!.acceptParams(params);
        },
    },
    {
        label: i18n.global.t('app.delete'),
        click: function (row: WebSite.WebSite) {
            openDelete(row.id);
        },
    },
];

const openDelete = (id: number) => {
    deleteRef.value.acceptParams(id);
};

const openCreate = () => {
    createRef.value.acceptParams();
};

const openGroup = () => {
    groupRef.value.acceptParams();
};

const checkExist = (data: App.CheckInstalled) => {
    nginxIsExist.value = data.isExist;
    containerName.value = data.containerName;
};

onMounted(() => {
    search();
});
</script>
