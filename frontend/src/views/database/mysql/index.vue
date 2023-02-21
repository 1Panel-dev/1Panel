<template>
    <div v-loading="loading">
        <LayoutContent :title="'MySQL ' + $t('menu.database')">
            <template #app>
                <AppStatus
                    :app-key="'mysql'"
                    style="margin-top: 20px"
                    @setting="onSetting"
                    @is-exist="checkExist"
                ></AppStatus>
            </template>

            <template #toolbar v-if="mysqlIsExist && !isOnSetting">
                <el-row :class="{ mask: mysqlStatus != 'Running' }">
                    <el-col :span="20">
                        <el-button type="primary" @click="onOpenDialog()">
                            {{ $t('database.create') }}
                        </el-button>
                        <el-button @click="onChangeRootPassword" type="primary" plain>
                            {{ $t('database.rootPassword') }}
                        </el-button>
                        <el-button @click="onChangeAccess" type="primary" plain>
                            {{ $t('database.remoteAccess') }}
                        </el-button>
                        <el-button @click="goDashboard" type="primary" plain>phpMyAdmin</el-button>
                    </el-col>
                    <el-col :span="4">
                        <div class="search-button">
                            <el-input
                                v-model="searchName"
                                clearable
                                @clear="search()"
                                suffix-icon="Search"
                                @keyup.enter="search()"
                                @blur="search()"
                                :placeholder="$t('commons.button.search')"
                            ></el-input>
                        </div>
                    </el-col>
                </el-row>
            </template>
            <template #main v-if="mysqlIsExist && !isOnSetting">
                <ComplexTable
                    :pagination-config="paginationConfig"
                    @search="search"
                    :data="data"
                    :class="{ mask: mysqlStatus != 'Running' }"
                >
                    <el-table-column :label="$t('commons.table.name')" prop="name" />
                    <el-table-column :label="$t('commons.login.username')" prop="username" />
                    <el-table-column :label="$t('commons.login.password')" prop="password">
                        <template #default="{ row }">
                            <div>
                                <span style="float: left; line-height: 25px" v-if="!row.showPassword">***********</span>
                                <div style="cursor: pointer; float: left" v-if="!row.showPassword">
                                    <el-icon
                                        style="margin-left: 5px; margin-top: 3px"
                                        @click="row.showPassword = true"
                                        :size="16"
                                    >
                                        <View />
                                    </el-icon>
                                </div>
                                <span style="float: left" v-if="row.showPassword">{{ row.password }}</span>
                                <div style="cursor: pointer; float: left" v-if="row.showPassword">
                                    <el-icon
                                        style="margin-left: 5px; margin-top: 3px"
                                        @click="row.showPassword = false"
                                        :size="16"
                                    >
                                        <Hide />
                                    </el-icon>
                                </div>
                                <div style="cursor: pointer; float: left">
                                    <el-icon
                                        style="margin-left: 5px; margin-top: 3px"
                                        :size="16"
                                        @click="onCopyPassword(row)"
                                    >
                                        <DocumentCopy />
                                    </el-icon>
                                </div>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.description')" prop="description">
                        <template #default="{ row }">
                            <fu-read-write-switch :data="row.description" v-model="row.edit" @change="onChange(row)">
                                <el-input v-model="row.description" @blur="row.edit = false" />
                            </fu-read-write-switch>
                        </template>
                    </el-table-column>
                    <el-table-column
                        prop="createdAt"
                        :label="$t('commons.table.date')"
                        :formatter="dateFormat"
                        show-overflow-tooltip
                    />
                    <fu-table-operations
                        width="370px"
                        :buttons="buttons"
                        :ellipsis="10"
                        :label="$t('commons.table.operate')"
                        fix
                    />
                </ComplexTable>
            </template>
        </LayoutContent>

        <el-card
            width="30%"
            v-if="mysqlStatus != 'Running' && !isOnSetting && mysqlIsExist && !loading"
            class="mask-prompt"
        >
            <span style="font-size: 14px">{{ $t('commons.service.serviceNotStarted', ['MySQL']) }}</span>
        </el-card>

        <Setting ref="settingRef" style="margin-top: 20px" />
        <el-dialog
            v-model="phpVisiable"
            :title="$t('app.checkTitle')"
            width="30%"
            :close-on-click-modal="false"
            :destroy-on-close="true"
        >
            <el-alert :closable="false" :title="$t('app.checkInstalledWarn', ['phpMyAdmin'])" type="info">
                <el-link icon="Position" @click="getAppDetail('phpmyadmin')" type="primary">
                    {{ $t('database.goInstall') }}
                </el-link>
            </el-alert>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="phpVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                </span>
            </template>
        </el-dialog>

        <PasswordDialog ref="passwordRef" @search="search" />
        <RootPasswordDialog ref="rootPasswordRef" />
        <RemoteAccessDialog ref="remoteAccessRef" />
        <UploadDialog ref="uploadRef" />
        <OperateDialog @search="search" ref="dialogRef" />
        <Backups ref="dialogBackupRef" />

        <AppResources ref="checkRef"></AppResources>
        <DeleteDialog ref="deleteRef" @search="search" />
    </div>
</template>

<script lang="ts" setup>
import LayoutContent from '@/layout/layout-content.vue';
import ComplexTable from '@/components/complex-table/index.vue';
import OperateDialog from '@/views/database/mysql/create/index.vue';
import DeleteDialog from '@/views/database/mysql/delete/index.vue';
import PasswordDialog from '@/views/database/mysql/password/index.vue';
import RootPasswordDialog from '@/views/database/mysql/root-password/index.vue';
import RemoteAccessDialog from '@/views/database/mysql/remote/index.vue';
import UploadDialog from '@/views/database/mysql/upload/index.vue';
import AppResources from '@/views/database/mysql/check/index.vue';
import Setting from '@/views/database/mysql/setting/index.vue';
import AppStatus from '@/components/app-status/index.vue';
import Backups from '@/components/backup/index.vue';
import { dateFormat } from '@/utils/util';
import { reactive, ref } from 'vue';
import { deleteCheckMysqlDB, loadRemoteAccess, searchMysqlDBs, updateMysqlDescription } from '@/api/modules/database';
import i18n from '@/lang';
import { Database } from '@/api/interface/database';
import { App } from '@/api/interface/app';
import { GetAppPort } from '@/api/modules/app';
import router from '@/routers';
import { MsgSuccess } from '@/utils/message';

const loading = ref(false);

const mysqlName = ref();
const isOnSetting = ref<boolean>();

const checkRef = ref();
const deleteRef = ref();

const phpadminPort = ref();
const phpVisiable = ref(false);

const data = ref();
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 10,
    total: 0,
});
const searchName = ref();

const mysqlIsExist = ref(false);
const mysqlContainer = ref();
const mysqlStatus = ref();
const mysqlVersion = ref();

const dialogRef = ref();
const onOpenDialog = async () => {
    let params = {
        mysqlName: mysqlName.value,
    };
    dialogRef.value!.acceptParams(params);
};

const dialogBackupRef = ref();
const onOpenBackupDialog = async (dbName: string) => {
    let params = {
        type: 'mysql',
        name: mysqlName.value,
        detailName: dbName,
    };
    dialogBackupRef.value!.acceptParams(params);
};

const uploadRef = ref();

const rootPasswordRef = ref();
const onChangeRootPassword = async () => {
    rootPasswordRef.value!.acceptParams();
};

const remoteAccessRef = ref();
const onChangeAccess = async () => {
    const res = await loadRemoteAccess();
    let param = {
        privilege: res.data,
    };
    remoteAccessRef.value!.acceptParams(param);
};

const passwordRef = ref();

const settingRef = ref();
const onSetting = async () => {
    isOnSetting.value = true;
    let params = {
        status: mysqlStatus.value,
        mysqlName: mysqlName.value,
        mysqlVersion: mysqlVersion.value,
    };
    settingRef.value!.acceptParams(params);
};

const search = async () => {
    let params = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
        info: searchName.value,
    };
    const res = await searchMysqlDBs(params);
    data.value = res.data.items || [];
    paginationConfig.total = res.data.total;
};

const onChange = async (info: any) => {
    if (!info.edit) {
        await updateMysqlDescription({ id: info.id, description: info.description });
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    }
};

const goDashboard = async () => {
    if (phpadminPort.value === 0) {
        phpVisiable.value = true;
        return;
    }
    let href = window.location.href;
    let ipLocal = href.split('//')[1].split(':')[0];
    window.open(`http://${ipLocal}:${phpadminPort.value}`, '_blank');
};
const getAppDetail = (key: string) => {
    router.push({ name: 'AppDetail', params: { appKey: key } });
};

const loadDashboardPort = async () => {
    const res = await GetAppPort('phpmyadmin');
    phpadminPort.value = res.data;
};

const checkExist = (data: App.CheckInstalled) => {
    mysqlIsExist.value = data.isExist;
    mysqlName.value = data.name;
    mysqlStatus.value = data.status;
    mysqlVersion.value = data.version;
    mysqlContainer.value = data.containerName;
    if (mysqlIsExist.value) {
        search();
        loadDashboardPort();
    }
};

const onCopyPassword = (row: Database.MysqlDBInfo) => {
    let input = document.createElement('input');
    input.value = row.password;
    document.body.appendChild(input);
    input.select();
    document.execCommand('Copy');
    document.body.removeChild(input);
    MsgSuccess(i18n.global.t('commons.msg.copySuccess'));
};

const onDelete = async (row: Database.MysqlDBInfo) => {
    const res = await deleteCheckMysqlDB(row.id);
    if (res.data && res.data.length > 0) {
        checkRef.value.acceptParams({ items: res.data });
    } else {
        deleteRef.value.acceptParams({ id: row.id, name: row.name });
    }
};

const buttons = [
    {
        label: i18n.global.t('database.changePassword'),
        click: (row: Database.MysqlDBInfo) => {
            let param = {
                id: row.id,
                operation: 'password',
                username: row.username,
                password: row.password,
            };
            passwordRef.value.acceptParams(param);
        },
    },
    {
        label: i18n.global.t('database.permission'),
        click: (row: Database.MysqlDBInfo) => {
            let param = {
                id: row.id,
                operation: 'privilege',
                privilege: '',
                privilegeIPs: '',
                password: '',
            };
            if (row.permission === '%' || row.permission === 'localhost') {
                param.privilege = row.permission;
            } else {
                param.privilegeIPs = row.permission;
                param.privilege = 'ip';
            }
            passwordRef.value.acceptParams(param);
        },
    },
    {
        label: i18n.global.t('database.backupList'),
        click: (row: Database.MysqlDBInfo) => {
            onOpenBackupDialog(row.name);
        },
    },
    {
        label: i18n.global.t('database.loadBackup'),
        click: (row: Database.MysqlDBInfo) => {
            let params = {
                mysqlName: mysqlName.value,
                dbName: row.name,
            };
            uploadRef.value!.acceptParams(params);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: Database.MysqlDBInfo) => {
            onDelete(row);
        },
    },
];
</script>
