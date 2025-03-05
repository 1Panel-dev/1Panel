<template>
    <div>
        <RouterButton
            :buttons="[
                {
                    label: i18n.global.t('menu.website'),
                    path: '/websites',
                },
            ]"
        />
        <LayoutContent :title="$t('menu.website')" v-loading="loading">
            <template #app>
                <AppStatus
                    :app-key="'openresty'"
                    @setting="setting"
                    v-model:mask-show="maskShow"
                    v-model:loading="loading"
                    @is-exist="checkExist"
                ></AppStatus>
            </template>
            <template v-if="!openNginxConfig && nginxIsExist" #leftToolBar>
                <el-button type="primary" @click="openCreate" :disabled="nginxStatus != 'Running'">
                    {{ $t('website.create') }}
                </el-button>
                <el-button type="primary" plain @click="openGroup" :disabled="nginxStatus != 'Running'">
                    {{ $t('commons.table.group') }}
                </el-button>
                <el-button type="primary" plain @click="openDefault" :disabled="nginxStatus != 'Running'">
                    {{ $t('website.defaultServer') }}
                </el-button>
                <el-button type="primary" plain @click="openDefaultHtml" :disabled="nginxStatus != 'Running'">
                    {{ $t('website.defaultHtml') }}
                </el-button>
            </template>
            <template v-if="!openNginxConfig && nginxIsExist" #rightToolBar>
                <el-select
                    v-model="req.websiteGroupId"
                    @change="search()"
                    class="p-w-200 mr-2.5"
                    :disabled="nginxStatus != 'Running'"
                >
                    <template #prefix>{{ $t('commons.table.group') }}</template>
                    <el-option :label="$t('commons.table.all')" :value="0"></el-option>
                    <div v-for="item in groups" :key="item.id">
                        <el-option
                            v-if="item.name === 'default'"
                            :label="$t('commons.table.default')"
                            :value="item.id"
                        />
                        <el-option v-else :label="item.name" :value="item.id" />
                    </div>
                </el-select>
                <TableSearch @search="search()" v-model:searchName="req.name" />
                <div class="!ml-2.5">
                    <fu-table-column-select
                        :columns="columns"
                        trigger="hover"
                        :title="$t('commons.table.selectColumn')"
                        popper-class="popper-class"
                        :only-icon="true"
                    />
                </div>
            </template>
            <template v-if="!openNginxConfig" #main>
                <ComplexTable
                    :pagination-config="paginationConfig"
                    :data="data"
                    @sort-change="changeSort"
                    @search="search()"
                    :class="{ mask: nginxStatus != 'Running' }"
                    :heightDiff="370"
                    :columns="columns"
                    localKey="websiteColumn"
                >
                    <el-table-column
                        :label="$t('commons.table.name')"
                        fix
                        prop="primaryDomain"
                        min-width="120px"
                        :width="mobile ? 220 : 'auto'"
                        sortable
                        show-overflow-tooltip
                    >
                        <template #default="{ row }">
                            <el-text type="primary" class="cursor-pointer" @click="openConfig(row.id)">
                                {{ row.primaryDomain }}
                            </el-text>
                            <el-popover
                                placement="right"
                                trigger="hover"
                                :width="300"
                                @before-enter="searchDomains(row.id)"
                            >
                                <template #reference>
                                    <el-button link icon="Promotion" class="ml-2.5"></el-button>
                                </template>
                                <table>
                                    <tbody>
                                        <tr v-for="(domain, index) in domains" :key="index">
                                            <td>
                                                <el-button type="primary" link @click="openUrl(getUrl(domain, row))">
                                                    {{ getUrl(domain, row) }}
                                                </el-button>
                                            </td>
                                            <td>
                                                <CopyButton :content="getUrl(domain, row)" type="icon" />
                                            </td>
                                        </tr>
                                    </tbody>
                                </table>
                            </el-popover>
                        </template>
                    </el-table-column>
                    <el-table-column
                        min-width="120px"
                        :label="$t('commons.table.type')"
                        fix
                        show-overflow-tooltip
                        prop="type"
                        sortable
                    >
                        <template #default="{ row }">
                            <div v-if="row.type">
                                {{ $t('website.' + row.type) }}
                                <span v-if="row.type === 'deployment'">[{{ row.appName }}]</span>
                                <span v-if="row.type === 'runtime'">[{{ row.runtimeName }}]</span>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('website.sitePath')" prop="sitePath" width="90px">
                        <template #default="{ row }">
                            <el-button type="primary" link @click="toFolder(row.sitePath + '/index')">
                                <el-icon>
                                    <FolderOpened />
                                </el-icon>
                            </el-button>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.status')" prop="status" width="120px" sortable>
                        <template #default="{ row }">
                            <Status
                                v-if="row.status === 'Running'"
                                :operate="true"
                                :status="row.status"
                                @click="operateWebsite('stop', row.id)"
                            />
                            <Status
                                v-else
                                :status="row.status"
                                :operate="true"
                                @click="operateWebsite('start', row.id)"
                            />
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('website.remark')" prop="remark" min-width="120px">
                        <template #default="{ row }">
                            <fu-read-write-switch>
                                <template #read>
                                    <MsgInfo :info="row.remark" width="180px" />
                                </template>
                                <template #default="{ read }">
                                    <el-input v-model="row.remark" @blur="updateRemark(row, read)" />
                                </template>
                            </fu-read-write-switch>
                        </template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('commons.table.protocol')"
                        prop="protocol"
                        width="90px"
                    ></el-table-column>
                    <el-table-column
                        :label="$t('website.expireDate')"
                        prop="expireDate"
                        :sortable="'custom'"
                        min-width="150px"
                    >
                        <template #default="{ row }">
                            <div v-if="row.showdate">
                                <el-date-picker
                                    :key="row.id"
                                    class="p-w-100"
                                    v-model="row.expireDate"
                                    type="date"
                                    :disabled-date="checkDate"
                                    :shortcuts="shortcuts"
                                    :clearable="false"
                                    @change="updateWebsitConfig(row)"
                                    :ref="(el) => setdateRefs(el)"
                                    @visible-change="(visibility:boolean) => pickerVisibility(visibility, row)"
                                    size="small"
                                    :mounted="initDatePicker(row)"
                                ></el-date-picker>
                            </div>
                            <div v-else>
                                <el-link type="primary" :underline="false" @click.stop="openDatePicker(row)">
                                    <span v-if="isEver(row.expireDate)">
                                        {{ $t('website.neverExpire') }}
                                    </span>
                                    <span v-else>
                                        {{ dateFormatSimple(row.expireDate) }}
                                    </span>
                                </el-link>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('website.sslExpireDate')" prop="sslExpireDate" width="150px">
                        <template #default="{ row }">
                            <el-tag v-if="row.protocol == 'HTTPS'" :type="row.sslStatus">
                                {{ dateFormatSimple(row.sslExpireDate) }}
                            </el-tag>
                            <span v-else></span>
                        </template>
                    </el-table-column>
                    <fu-table-operations
                        :ellipsis="10"
                        width="300px"
                        :buttons="buttons"
                        :label="$t('commons.table.operate')"
                        :fixed="mobile ? false : 'right'"
                        fix
                    />
                </ComplexTable>
                <el-card width="30%" v-if="nginxStatus != 'Running' && maskShow" class="mask-prompt">
                    <span v-if="nginxIsExist">{{ $t('commons.service.serviceNotStarted', ['OpenResty']) }}</span>
                    <span v-else>
                        {{ $t('app.checkInstalledWarn', ['OpenResty']) }}
                        <el-button @click="goRouter('openresty')" link icon="Position" type="primary">
                            {{ $t('database.goInstall') }}
                        </el-button>
                    </span>
                </el-card>
            </template>
        </LayoutContent>
        <CreateWebSite ref="createRef" @close="search" />
        <DeleteWebsite ref="deleteRef" @close="search" />
        <UploadDialog ref="uploadRef" />
        <Backups ref="dialogBackupRef" />
        <DefaultServer ref="defaultRef" />
        <GroupDialog @search="listGroup" ref="groupRef" />
        <NginxConfig v-if="openNginxConfig" v-loading="loading" :containerName="containerName" :status="nginxStatus" />
        <DefaultHtml ref="defaultHtmlRef" />
    </div>
</template>

<script lang="ts" setup>
import Backups from '@/components/backup/index.vue';
import UploadDialog from '@/components/upload/index.vue';
import DefaultServer from '@/views/website/website/default/index.vue';
import DefaultHtml from '@/views/website/website/html/index.vue';
import CreateWebSite from '@/views/website/website/create/index.vue';
import DeleteWebsite from '@/views/website/website/delete/index.vue';
import NginxConfig from '@/views/website/website/nginx/index.vue';
import GroupDialog from '@/components/agent-group/index.vue';
import AppStatus from '@/components/app-status/index.vue';
import i18n from '@/lang';
import router from '@/routers';
import { onMounted, reactive, ref, computed } from 'vue';
import { listDomains, opWebsite, searchWebsites, updateWebsite } from '@/api/modules/website';
import { Website } from '@/api/interface/website';
import { App } from '@/api/interface/app';
import { ElMessageBox } from 'element-plus';
import { dateFormatSimple } from '@/utils/util';
import { MsgError, MsgSuccess } from '@/utils/message';
import { useI18n } from 'vue-i18n';
import { getAgentGroupList } from '@/api/modules/group';
import { Group } from '@/api/interface/group';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();

const shortcuts = [
    {
        text: useI18n().t('website.ever'),
        value: () => {
            return new Date('9999-12-31');
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
const maskShow = ref(true);
const createRef = ref();
const deleteRef = ref();
const groupRef = ref();
const defaultHtmlRef = ref();
const openNginxConfig = ref(false);
const nginxIsExist = ref(false);
const containerName = ref('');
const nginxStatus = ref('');
const installPath = ref('');
const uploadRef = ref();
const dialogBackupRef = ref();
const defaultRef = ref();
const data = ref();
let groups = ref<Group.GroupInfo[]>([]);
const dataRef = ref();
const domains = ref<Website.Domain[]>([]);
const columns = ref([]);

const paginationConfig = reactive({
    cacheSizeKey: 'website-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('website-page-size')) || 10,
    total: 0,
});
let req = reactive({
    name: '',
    page: 1,
    pageSize: 10,
    orderBy: 'createdAt',
    order: 'null',
    websiteGroupId: 0,
});
const mobile = computed(() => {
    return globalStore.isMobile();
});

const goRouter = async (key: string) => {
    router.push({ name: 'AppAll', query: { install: key } });
};

const changeSort = ({ prop, order }) => {
    if (order) {
        switch (prop) {
            case 'primaryDomain':
                prop = 'primary_domain';
                break;
            case 'expireDate':
                prop = 'expire_date';
                break;
            default:
                break;
        }
        req.orderBy = prop;
        req.order = order;
    } else {
        req.orderBy = 'created_at';
        req.order = 'null';
    }
    search();
};

const search = async () => {
    req.page = paginationConfig.currentPage;
    req.pageSize = paginationConfig.pageSize;

    loading.value = true;
    data.value = [];
    await searchWebsites(req)
        .then((res) => {
            data.value = res.data.items;
            paginationConfig.total = res.data.total;
        })
        .finally(() => {
            loading.value = false;
        });
};

const listGroup = async () => {
    const res = await getAgentGroupList('website');
    groups.value = res.data;
};

const setting = () => {
    openNginxConfig.value = true;
};

const openConfig = (id: number) => {
    router.push({ name: 'WebsiteConfig', params: { id: id, tab: 'basic' } });
};

const isEver = (time: string) => {
    const expireDate = new Date(time);
    return expireDate > new Date('9999-12-30');
};

const isBeforeNow = (time: string) => {
    return new Date() > new Date(time);
};

const refreshData = () => {
    for (let i = 0; i < data.value.length; i++) {
        if (data.value[i].showdate) {
            data.value[i].showdate = false;
            if (data.value[i].oldExpireDate) {
                data.value[i].expireDate = data.value[i].oldExpireDate;
                data.value[i].oldExpireDate = undefined;
            }
        }
    }
    if (dataRef.value) {
        dataRef.value.handleClose();
        dataRef.value = undefined;
    }
};

const openDatePicker = (row: any) => {
    refreshData();
    row.showdate = true;
};

const setdateRefs = (ref: any) => {
    dataRef.value = ref;
    if (dataRef.value != undefined) {
        dataRef.value.handleOpen();
    }
};

const initDatePicker = (row: any) => {
    if (
        (dataRef.value == undefined && row.oldExpireDate == undefined && isBeforeNow(row.expireDate)) ||
        isEver(row.expireDate)
    ) {
        row.oldExpireDate = row.expireDate;
        const date = new Date().toLocaleDateString();
        row.expireDate = date;
        return;
    }
};

const pickerVisibility = (visibility: boolean, row: any) => {
    if (!visibility) {
        dataRef.value = undefined;
        if (row.oldExpireDate) {
            row.expireDate = row.oldExpireDate;
            row.oldExpireDate = undefined;
        }
        row.showdate = false;
    }
};

const updateWebsitConfig = (row: any) => {
    const reqDate = dateFormatSimple(row.expireDate);
    const req = {
        id: row.id,
        primaryDomain: row.primaryDomain,
        remark: row.remark,
        webSiteGroupId: row.webSiteGroupId,
        expireDate: reqDate,
        IPV6: row.IPV6,
    };

    updateWebsite(req).then(() => {
        MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
        search();
    });
};

const buttons = [
    {
        label: i18n.global.t('menu.config'),
        click: function (row: Website.Website) {
            openConfig(row.id);
        },
    },
    {
        label: i18n.global.t('database.backupList'),
        click: (row: Website.Website) => {
            let params = {
                type: 'website',
                name: row.alias,
                detailName: row.alias,
            };
            dialogBackupRef.value!.acceptParams(params);
        },
    },
    {
        label: i18n.global.t('database.loadBackup'),
        click: (row: Website.Website) => {
            let params = {
                type: 'website',
                name: row.primaryDomain,
                detailName: row.alias,
            };
            uploadRef.value!.acceptParams(params);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
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
    groupRef.value.acceptParams({ type: 'website' });
};

const openDefault = () => {
    defaultRef.value.acceptParams();
};

const openDefaultHtml = () => {
    defaultHtmlRef.value.acceptParams();
};

const checkExist = (data: App.CheckInstalled) => {
    nginxIsExist.value = data.isExist;
    containerName.value = data.containerName;
    nginxStatus.value = data.status;
    installPath.value = data.installPath;
};

const checkDate = (date: Date) => {
    const now = new Date();
    return date.getTime() < now.getTime();
};

const operateWebsite = (op: string, id: number) => {
    ElMessageBox.confirm(i18n.global.t('website.' + op + 'Helper'), i18n.global.t('cronjob.changeStatus'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    }).then(async () => {
        await opWebsite({ id: id, operate: op });
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        search();
    });
};

const toFolder = (folder: string) => {
    router.push({ path: '/hosts/files', query: { path: folder } });
};

const searchDomains = (id: number) => {
    listDomains(id).then((res) => {
        domains.value = res.data;
    });
};

const openUrl = (url: string) => {
    window.open(url);
};

const getUrl = (domain: Website.Domain, website: Website.Website): string => {
    const protocol = website.protocol.toLowerCase();
    let url = protocol + '://' + domain.domain;
    if (protocol == 'http' && domain.port != 80) {
        url = url + ':' + domain.port;
    }
    if (protocol == 'https' && domain.ssl) {
        url = url + ':' + domain.port;
    }
    return url;
};

const updateRemark = (row: Website.Website, bulr: Function) => {
    bulr();
    if (row.remark && row.remark.length > 128) {
        MsgError(i18n.global.t('commons.rule.length128Err'));
        return;
    }
    updateWebsitConfig(row);
};

onMounted(() => {
    search();
    listGroup();
});
</script>
