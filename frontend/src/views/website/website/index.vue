<template>
    <div v-loading="loading">
        <AppStatus :app-key="'nginx'" @setting="setting" v-model:loading="loading" @is-exist="checkExist"></AppStatus>
        <div v-if="nginxIsExist" :class="{ mask: nginxStatus != 'Running' }">
            <LayoutContent>
                <br />
                <el-card v-if="!openNginxConfig">
                    <LayoutContent :header="$t('website.website')">
                        <ComplexTable :pagination-config="paginationConfig" :data="data" @search="search()">
                            <template #toolbar>
                                <el-button type="primary" icon="Plus" @click="openCreate">
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
                                    <el-link type="primary" @click="openConfig(row.id)">
                                        {{ row.primaryDomain }}
                                    </el-link>
                                </template>
                            </el-table-column>
                            <el-table-column :label="$t('commons.table.type')" fix prop="type">
                                <template #default="{ row }">
                                    {{ $t('website.' + row.type) }}
                                </template>
                            </el-table-column>
                            <el-table-column :label="$t('commons.table.status')" prop="status">
                                <template #default="{ row }">
                                    <el-button
                                        v-if="row.status === 'Running'"
                                        link
                                        type="success"
                                        @click="opWebsite('stop', row.id)"
                                    >
                                        {{ $t('commons.status.running') }}
                                    </el-button>
                                    <el-button v-else link type="danger" @click="opWebsite('start', row.id)">
                                        {{ $t('commons.status.stopped') }}
                                    </el-button>
                                </template>
                            </el-table-column>
                            <el-table-column :label="$t('website.remark')" prop="remark"></el-table-column>
                            <el-table-column :label="$t('website.protocol')" prop="protocol"></el-table-column>
                            <el-table-column :label="$t('website.expireDate')">
                                <template #default="{ row, $index }">
                                    <div v-show="row.showdate">
                                        <el-date-picker
                                            v-model="row.expireDate"
                                            type="date"
                                            :disabled-date="checkDate"
                                            :shortcuts="shortcuts"
                                            :clearable="false"
                                            :default-value="setDate(row.expireDate)"
                                            :ref="(el) => setdateRefs(el, $index)"
                                            @change="submitDate(row)"
                                            @visible-change="(visibility:boolean) => pickerVisibility(visibility, row)"
                                            size="small"
                                        ></el-date-picker>
                                    </div>
                                    <div v-show="!row.showdate">
                                        <span v-if="isEver(row.expireDate)" @click="openDatePicker(row, $index)">
                                            {{ $t('website.neverExpire') }}
                                        </span>
                                        <span v-else @click="openDatePicker(row, $index)">
                                            {{ dateFromatSimple(row.expireDate) }}
                                        </span>
                                    </div>
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

                <CreateWebSite ref="createRef" @close="search"></CreateWebSite>
                <DeleteWebsite ref="deleteRef" @close="search"></DeleteWebsite>
                <WebSiteGroup ref="groupRef"></WebSiteGroup>
                <UploadDialog ref="uploadRef" />
                <BackupRecords ref="dialogBackupRef" />
            </LayoutContent>
        </div>
        <div v-if="nginxIsExist">
            <el-card width="30%" v-if="nginxStatus != 'Running' && !loading" class="mask-prompt">
                <span style="font-size: 14px">{{ $t('commons.service.serviceNotStarted', ['OpenResty']) }}</span>
            </el-card>
        </div>

        <el-card v-if="openNginxConfig">
            <NginxConfig :containerName="containerName" :status="nginxStatus"></NginxConfig>
        </el-card>
    </div>
</template>

<script lang="ts" setup>
import LayoutContent from '@/layout/layout-content.vue';
import BackupRecords from '@/views/website/website/backup/index.vue';
import UploadDialog from '@/views/website/website/upload/index.vue';
import ComplexTable from '@/components/complex-table/index.vue';
import { onMounted, reactive, ref } from '@vue/runtime-core';
import CreateWebSite from './create/index.vue';
import DeleteWebsite from './delete/index.vue';
import WebSiteGroup from './group/index.vue';
import { OpWebsite, SearchWebsites, UpdateWebsite } from '@/api/modules/website';
import { Website } from '@/api/interface/website';
import AppStatus from '@/components/app-status/index.vue';
import NginxConfig from './nginx/index.vue';
import i18n from '@/lang';
import router from '@/routers';
import { App } from '@/api/interface/app';
import { ElMessage, ElMessageBox } from 'element-plus';
import { dateFromatSimple } from '@/utils/util';

const shortcuts = [
    {
        text: i18n.global.t('website.ever'),
        value: () => {
            return new Date('1970-01-01');
        },
    },
    {
        text: i18n.global.t('website.nextYear'),
        value: () => {
            const now = new Date();
            now.setFullYear(now.getFullYear() + 1);
            return now;
        },
    },
];

const loading = ref(false);
const createRef = ref();
const deleteRef = ref();
const groupRef = ref();
let openNginxConfig = ref(false);
let nginxIsExist = ref(false);
let containerName = ref('');
let nginxStatus = ref('');
let installPath = ref('');
const uploadRef = ref();
const dialogBackupRef = ref();
const data = ref();
let dateRefs: Map<number, any> = new Map();

const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 20,
    total: 0,
});

const search = async () => {
    const req = {
        name: '',
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
    };
    SearchWebsites(req).then((res) => {
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

const isEver = (time: string) => {
    const expireDate = new Date(time);
    return expireDate < new Date('1970-01-02');
};

const isBeforeNow = (time: string) => {
    return new Date() > new Date(time);
};

const setDate = (time: string) => {
    if (isEver(time)) {
        return new Date().toLocaleDateString();
    } else {
        return new Date(time);
    }
};

const openDatePicker = (row: any, index: number) => {
    row.showdate = true;
    const ref = dateRefs.get(index);
    if (ref != undefined) {
        if (isBeforeNow(row.expireDate)) {
            row.oldExpireDate = row.expireDate;
            const date = new Date().toLocaleDateString();
            row.expireDate = date;
        }
        ref.handleOpen();
    }
};

const setdateRefs = (ref: any, index: number) => {
    dateRefs.set(index, ref);
};

const pickerVisibility = (visibility: boolean, row: any) => {
    if (!visibility) {
        row.showdate = false;
        if (!row.change) {
            if (row.oldExpireDate) {
                row.expireDate = row.oldExpireDate;
            }
            row.change = false;
        }
    }
};

const submitDate = (row: any) => {
    const reqDate = dateFromatSimple(row.expireDate);
    const req = {
        id: row.id,
        primaryDomain: row.primaryDomain,
        remark: row.remark,
        webSiteGroupId: row.webSiteGroupId,
        expireDate: reqDate,
    };

    UpdateWebsite(req).then(() => {
        row.change = true;
        ElMessage.success(i18n.global.t('commons.msg.updateSuccess'));
    });
};

const buttons = [
    {
        label: i18n.global.t('website.config'),
        click: function (row: Website.Website) {
            openConfig(row.id);
        },
    },
    {
        label: i18n.global.t('database.backupList'),
        click: (row: Website.Website) => {
            let params = {
                id: row.id,
                type: row.type,
                name: row.primaryDomain,
            };
            dialogBackupRef.value!.acceptParams(params);
        },
    },
    {
        label: i18n.global.t('database.loadBackup'),
        click: (row: Website.Website) => {
            let params = {
                websiteName: row.primaryDomain,
                websiteType: row.type,
            };
            uploadRef.value!.acceptParams(params);
        },
    },
    {
        label: i18n.global.t('app.delete'),
        click: function (row: Website.Website) {
            openDelete(row);
        },
    },
];

const openDelete = (website: Website.Website) => {
    deleteRef.value.acceptParams(website);
};

const openCreate = () => {
    createRef.value.acceptParams(installPath.value);
};

const openGroup = () => {
    groupRef.value.acceptParams();
};

const checkExist = (data: App.CheckInstalled) => {
    nginxIsExist.value = data.isExist;
    containerName.value = data.containerName;
    nginxStatus.value = data.status;
    installPath.value = data.installPath;
};

const checkDate = (date: Date) => {
    const now = new Date();
    return date < now;
};

const opWebsite = (op: string, id: number) => {
    ElMessageBox.confirm(i18n.global.t('website.' + op + 'Helper'), i18n.global.t('cronjob.changeStatus'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    }).then(async () => {
        await OpWebsite({ id: id, operate: op });
        ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
        search();
    });
};

onMounted(() => {
    search();
});
</script>
