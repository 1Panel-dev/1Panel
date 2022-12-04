<template>
    <div>
        <Submenu activeName="mysql" />
        <AppStatus :app-key="'mysql'" style="margin-top: 20px" @setting="onSetting" @is-exist="checkExist" />
        <div v-if="mysqlIsExist">
            <Setting ref="settingRef" style="margin-top: 20px" />

            <el-card v-if="!isOnSetting" style="margin-top: 20px">
                <ComplexTable
                    :pagination-config="paginationConfig"
                    v-model:selects="selects"
                    @search="search"
                    :data="data"
                >
                    <template #toolbar>
                        <el-button type="primary" icon="Plus" @click="onOpenDialog()">
                            {{ $t('commons.button.create') }}
                        </el-button>
                        <el-button>phpMyAdmin</el-button>
                    </template>
                    <el-table-column type="selection" fix />
                    <el-table-column :label="$t('commons.table.name')" prop="name" />
                    <el-table-column :label="$t('commons.login.username')" prop="username" />
                    <el-table-column :label="$t('commons.login.password')" prop="password">
                        <template #default="{ row }">
                            <div v-if="!row.showPassword">
                                <span style="float: left">***********</span>
                                <div style="margin-top: 2px; cursor: pointer">
                                    <el-icon style="margin-left: 5px" @click="row.showPassword = true" :size="16">
                                        <View />
                                    </el-icon>
                                </div>
                            </div>
                            <div v-else>
                                <span style="float: left">{{ row.password }}</span>
                                <div style="margin-top: 4px; cursor: pointer">
                                    <el-icon style="margin-left: 5px" @click="row.showPassword = false" :size="16">
                                        <Hide />
                                    </el-icon>
                                </div>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.description')" prop="description" />
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
                <el-form ref="changeFormRef" :model="changeForm" label-width="80px">
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
                    <el-button @click="changeVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button @click="submitChangeInfo(changeFormRef)">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-dialog>

        <UploadDialog ref="uploadRef" />
        <OperatrDialog @search="search" ref="dialogRef" />
        <BackupRecords ref="dialogBackupRef" />

        <AppResources ref="checkRef"></AppResources>
    </div>
</template>

<script lang="ts" setup>
import ComplexTable from '@/components/complex-table/index.vue';
import OperatrDialog from '@/views/database/mysql/create/index.vue';
import BackupRecords from '@/views/database/mysql/backup/index.vue';
import UploadDialog from '@/views/database/mysql/upload/index.vue';
import AppResources from '@/views/database/mysql/check/index.vue';
import Setting from '@/views/database/mysql/setting/index.vue';
import AppStatus from '@/components/app-status/index.vue';
import Submenu from '@/views/database/index.vue';
import { dateFromat } from '@/utils/util';
import { reactive, ref } from 'vue';
import { deleteCheckMysqlDB, deleteMysqlDB, searchMysqlDBs, updateMysqlDBInfo } from '@/api/modules/database';
import i18n from '@/lang';
import { useDeleteData } from '@/hooks/use-delete-data';
import { ElForm, ElMessage } from 'element-plus';
import { Database } from '@/api/interface/database';
import { Rules } from '@/global/form-rules';
import { App } from '@/api/interface/app';

const selects = ref<any>([]);
const mysqlName = ref();
const isOnSetting = ref<boolean>();

const checkRef = ref();

const data = ref();
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 10,
    total: 0,
});

const mysqlIsExist = ref(false);

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

const settingRef = ref();
const onSetting = async () => {
    isOnSetting.value = true;
    let params = {
        mysqlName: mysqlName.value,
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
        changeForm.value = changeForm.operation === 'password' ? changeForm.password : changeForm.privilege;
        changeForm.mysqlName = mysqlName.value;
        await updateMysqlDBInfo(changeForm);
        ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
        search();
        changeVisiable.value = false;
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

const checkExist = (data: App.CheckInstalled) => {
    mysqlIsExist.value = data.isExist;
    mysqlName.value = data.name;
    if (mysqlIsExist.value) {
        search();
    }
};

const onDelete = async (row: Database.MysqlDBInfo) => {
    const res = await deleteCheckMysqlDB(row.id);
    if (res.data && res.data.length > 0) {
        checkRef.value.acceptParams({ items: res.data });
    } else {
        await useDeleteData(deleteMysqlDB, row.id, 'app.deleteWarn');
        search();
    }
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
