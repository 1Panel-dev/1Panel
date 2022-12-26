<template>
    <div v-loading="loading">
        <Submenu activeName="mysql" />
        <AppStatus
            :app-key="'mysql'"
            style="margin-top: 20px"
            v-model:loading="loading"
            @setting="onSetting"
            @is-exist="checkExist"
        />
        <Setting ref="settingRef" style="margin-top: 20px" />

        <el-card
            width="30%"
            v-if="mysqlStatus != 'Running' && !isOnSetting && mysqlIsExist && !loading"
            class="mask-prompt"
        >
            <span style="font-size: 14px">{{ $t('commons.service.serviceNotStarted', ['Mysql']) }}</span>
        </el-card>
        <div v-if="mysqlIsExist" :class="{ mask: mysqlStatus != 'Running' }">
            <el-card v-if="!isOnSetting" style="margin-top: 20px">
                <ComplexTable :pagination-config="paginationConfig" @search="search" :data="data">
                    <template #toolbar>
                        <el-button type="primary" icon="Plus" @click="onOpenDialog()">
                            {{ $t('commons.button.create') }}
                        </el-button>
                        <el-button @click="onChangeRootPassword" type="primary" plain>
                            {{ $t('database.rootPassword') }}
                        </el-button>
                        <el-button @click="onChangeAccess" type="primary" plain>
                            {{ $t('database.remoteAccess') }}
                        </el-button>
                        <el-button @click="goDashboard" type="primary" plain icon="Position">phpMyAdmin</el-button>
                    </template>
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
                        :formatter="dateFromat"
                        show-overflow-tooltip
                    />
                    <fu-table-operations
                        width="300px"
                        :buttons="buttons"
                        :ellipsis="10"
                        :label="$t('commons.table.operate')"
                        fix
                    />
                </ComplexTable>
            </el-card>
        </div>
        <el-dialog v-model="changeVisiable" :destroy-on-close="true" :close-on-click-modal="false" width="30%">
            <template #header>
                <div class="card-header">
                    <span>{{ $t('database.changePassword') }}</span>
                </div>
            </template>
            <el-form>
                <el-form v-loading="loading" ref="changeFormRef" :model="changeForm" label-width="80px">
                    <div v-if="changeForm.operation === 'password'">
                        <el-form-item :label="$t('commons.login.username')" prop="userName">
                            <el-input disabled v-model="changeForm.userName"></el-input>
                        </el-form-item>
                        <el-form-item
                            :label="$t('commons.login.password')"
                            prop="password"
                            :rules="Rules.requiredInput"
                        >
                            <el-input type="password" clearable show-password v-model="changeForm.password"></el-input>
                        </el-form-item>
                    </div>
                    <div v-if="changeForm.operation === 'privilege'">
                        <el-form-item :label="$t('database.permission')" prop="privilege">
                            <el-select style="width: 100%" v-model="changeForm.privilege">
                                <el-option value="localhost" :label="$t('database.permissionLocal')" />
                                <el-option value="%" :label="$t('database.permissionAll')" />
                                <el-option value="ip" :label="$t('database.permissionForIP')" />
                            </el-select>
                        </el-form-item>
                        <el-form-item
                            v-if="changeForm.privilege === 'ip'"
                            prop="privilegeIPs"
                            :rules="Rules.requiredInput"
                        >
                            <el-input clearable v-model="changeForm.privilegeIPs" />
                        </el-form-item>
                    </div>
                </el-form>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button :disabled="loading" @click="changeVisiable = false">
                        {{ $t('commons.button.cancel') }}
                    </el-button>
                    <el-button :disabled="loading" @click="submitChangeInfo(changeFormRef)">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-dialog>

        <el-dialog
            v-model="phpVisiable"
            :title="$t('app.checkTitle')"
            width="30%"
            :close-on-click-modal="false"
            :destroy-on-close="true"
        >
            <el-alert :closable="false" :title="$t('app.checkInstalledWarn', ['phpMyAdmin'])" type="info">
                <el-link icon="Position" @click="goRouter('/apps/installed')" type="primary">
                    {{ $t('database.goInstall') }}
                </el-link>
            </el-alert>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="phpVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                </span>
            </template>
        </el-dialog>

        <RootPasswordDialog ref="rootPasswordRef" />
        <RemoteAccessDialog ref="remoteAccessRef" />
        <UploadDialog ref="uploadRef" />
        <OperateDialog @search="search" ref="dialogRef" />
        <BackupRecords ref="dialogBackupRef" />

        <AppResources ref="checkRef"></AppResources>
        <DeleteDialog ref="deleteRef" @search="search" />

        <ConfirmDialog ref="confirmDialogRef" @confirm="onSubmit"></ConfirmDialog>
    </div>
</template>

<script lang="ts" setup>
import ComplexTable from '@/components/complex-table/index.vue';
import OperateDialog from '@/views/database/mysql/create/index.vue';
import DeleteDialog from '@/views/database/mysql/delete/index.vue';
import RootPasswordDialog from '@/views/database/mysql/password/index.vue';
import RemoteAccessDialog from '@/views/database/mysql/remote/index.vue';
import BackupRecords from '@/views/database/mysql/backup/index.vue';
import ConfirmDialog from '@/components/confirm-dialog/index.vue';
import UploadDialog from '@/views/database/mysql/upload/index.vue';
import AppResources from '@/views/database/mysql/check/index.vue';
import Setting from '@/views/database/mysql/setting/index.vue';
import AppStatus from '@/components/app-status/index.vue';
import Submenu from '@/views/database/index.vue';
import { dateFromat } from '@/utils/util';
import { reactive, ref } from 'vue';
import {
    deleteCheckMysqlDB,
    loadRemoteAccess,
    searchMysqlDBs,
    updateMysqlAccess,
    updateMysqlDescription,
    updateMysqlPassword,
} from '@/api/modules/database';
import i18n from '@/lang';
import { ElForm, ElMessage } from 'element-plus';
import { Database } from '@/api/interface/database';
import { Rules } from '@/global/form-rules';
import { App } from '@/api/interface/app';
import { GetAppPort } from '@/api/modules/app';
import router from '@/routers';

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
        mysqlName: mysqlName.value,
        dbName: dbName,
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

const changeVisiable = ref(false);
type FormInstance = InstanceType<typeof ElForm>;
const changeFormRef = ref<FormInstance>();
const changeForm = reactive({
    id: 0,
    mysqlName: '',
    userName: '',
    password: '',
    operation: '',
    privilege: '',
    privilegeIPs: '',
    value: '',
});
const submitChangeInfo = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        let param = {
            id: changeForm.id,
            value: '',
        };
        if (changeForm.operation === 'password') {
            const res = await deleteCheckMysqlDB(changeForm.id);
            if (res.data && res.data.length > 0) {
                let params = {
                    header: i18n.global.t('database.changePassword'),
                    operationInfo: i18n.global.t('database.changePasswordHelper'),
                    submitInputInfo: i18n.global.t('database.restartNow'),
                };
                confirmDialogRef.value!.acceptParams(params);
            } else {
                param.value = changeForm.password;
                loading.value = true;
                await updateMysqlPassword(param)
                    .then(() => {
                        loading.value = false;
                        search();
                        changeVisiable.value = false;
                        ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
                    })
                    .catch(() => {
                        loading.value = false;
                    });
            }
            return;
        }
        param.value = changeForm.privilege;
        changeForm.mysqlName = mysqlName.value;
        loading.value = true;
        await updateMysqlAccess(param)
            .then(() => {
                loading.value = false;
                search();
                changeVisiable.value = false;
                ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

const search = async () => {
    let params = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
    };
    const res = await searchMysqlDBs(params);
    data.value = res.data.items || [];
    paginationConfig.total = res.data.total;
};

const onChange = async (info: any) => {
    if (!info.edit) {
        await updateMysqlDescription({ id: info.id, description: info.description });
        ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
    }
};

const goRouter = async (path: string) => {
    router.push({ path: path });
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
    ElMessage.success(i18n.global.t('commons.msg.copySuccess'));
};

const onDelete = async (row: Database.MysqlDBInfo) => {
    const res = await deleteCheckMysqlDB(row.id);
    if (res.data && res.data.length > 0) {
        checkRef.value.acceptParams({ items: res.data });
    } else {
        deleteRef.value.acceptParams({ id: row.id, name: row.name });
    }
};

const confirmDialogRef = ref();
const onSubmit = async () => {
    let param = {
        id: changeForm.id,
        value: changeForm.password,
    };
    loading.value = true;
    await updateMysqlPassword(param)
        .then(() => {
            loading.value = false;
            search();
            changeVisiable.value = false;
            ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
        })
        .catch(() => {
            loading.value = false;
        });
};

const buttons = [
    {
        label: i18n.global.t('database.changePassword'),
        click: (row: Database.MysqlDBInfo) => {
            changeForm.id = row.id;
            changeForm.operation = 'password';
            changeForm.userName = row.username;
            changeForm.password = row.password;
            changeVisiable.value = true;
        },
    },
    {
        label: i18n.global.t('database.permission'),
        click: (row: Database.MysqlDBInfo) => {
            changeForm.id = row.id;
            changeForm.operation = 'privilege';
            if (row.permission === '%' || row.permission === 'localhost') {
                changeForm.privilege = row.permission;
            } else {
                changeForm.privilegeIPs = row.permission;
                changeForm.privilege = 'ip';
            }
            changeVisiable.value = true;
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
