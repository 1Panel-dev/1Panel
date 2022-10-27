<template>
    <div>
        <Submenu activeName="mysql" />
        <el-dropdown size="default" split-button style="margin-top: 20px; margin-bottom: 5px">
            {{ version }}
            <template #dropdown>
                <el-dropdown-menu v-model="version">
                    <el-dropdown-item v-for="item in mysqlVersions" :key="item" @click="onChangeVersion(item)">
                        {{ item }}
                    </el-dropdown-item>
                </el-dropdown-menu>
            </template>
        </el-dropdown>
        <el-button
            v-if="!isOnSetting"
            style="margin-top: 20px; margin-left: 10px"
            size="default"
            icon="Setting"
            @click="onSetting"
        >
            {{ $t('database.setting') }}
        </el-button>
        <el-button
            v-if="isOnSetting"
            style="margin-top: 20px; margin-left: 10px"
            size="default"
            icon="Back"
            @click="onBacklist"
        >
            {{ $t('commons.button.back') }}列表
        </el-button>

        <Setting ref="settingRef"></Setting>

        <el-card v-if="!isOnSetting">
            <ComplexTable :pagination-config="paginationConfig" v-model:selects="selects" @search="search" :data="data">
                <template #toolbar>
                    <el-button type="primary" @click="onOpenDialog()">{{ $t('commons.button.create') }}</el-button>
                    <el-button @click="onOpenDialog()">phpMyAdmin</el-button>
                    <el-button type="danger" plain :disabled="selects.length === 0" @click="onBatchDelete(null)">
                        {{ $t('commons.button.delete') }}
                    </el-button>
                </template>
                <el-table-column type="selection" fix />
                <el-table-column :label="$t('commons.table.name')" prop="name" />
                <el-table-column :label="$t('auth.username')" prop="username" />
                <el-table-column :label="$t('auth.password')" prop="password" />
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

        <el-dialog v-model="changeVisiable" :destroy-on-close="true" width="30%">
            <template #header>
                <div class="card-header">
                    <span>{{ $t('database.changePassword') }}</span>
                </div>
            </template>
            <el-form>
                <el-form ref="changeFormRef" :model="changeForm" label-width="80px">
                    <div v-if="changeForm.operation === 'password'">
                        <el-form-item :label="$t('auth.username')" prop="userName">
                            <el-input disabled v-model="changeForm.userName"></el-input>
                        </el-form-item>
                        <el-form-item :label="$t('auth.password')" prop="password" :rules="Rules.requiredInput">
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

        <OperatrDialog @search="search" ref="dialogRef" />
        <BackupRecords ref="dialogBackupRef" />
    </div>
</template>

<script lang="ts" setup>
import ComplexTable from '@/components/complex-table/index.vue';
import OperatrDialog from '@/views/database/mysql/create/index.vue';
import BackupRecords from '@/views/database/mysql/backup/index.vue';
import Setting from '@/views/database/mysql/setting/index.vue';
import Submenu from '@/views/database/index.vue';
import { dateFromat } from '@/utils/util';
import { onMounted, reactive, ref } from 'vue';
import { deleteMysqlDB, loadVersions, searchMysqlDBs, updateMysqlDBInfo } from '@/api/modules/database';
import i18n from '@/lang';
import { useDeleteData } from '@/hooks/use-delete-data';
import { ElForm, ElMessage } from 'element-plus';
import { Database } from '@/api/interface/database';
import { Rules } from '@/global/form-rules';

const selects = ref<any>([]);
const mysqlVersions = ref();
const version = ref<string>('5.7');
const isOnSetting = ref<boolean>();

const data = ref();
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 10,
    total: 0,
});

const dialogRef = ref();
const onOpenDialog = async () => {
    let params = {
        version: version.value,
    };
    dialogRef.value!.acceptParams(params);
};

const dialogBackupRef = ref();
const onOpenBackupDialog = async (dbName: string) => {
    let params = {
        version: version.value,
        dbName: dbName,
    };
    dialogBackupRef.value!.acceptParams(params);
};

const settingRef = ref();
const onSetting = async () => {
    isOnSetting.value = true;
    let params = {
        version: version.value,
    };
    settingRef.value!.acceptParams(params);
};
const onBacklist = async () => {
    isOnSetting.value = false;
    search();
    settingRef.value!.onClose();
};

const changeVisiable = ref(false);
type FormInstance = InstanceType<typeof ElForm>;
const changeFormRef = ref<FormInstance>();
const changeForm = reactive({
    id: 0,
    version: '',
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
        changeForm.version = version.value;
        await updateMysqlDBInfo(changeForm);
        ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
        search();
        changeVisiable.value = false;
    });
};

const loadRunningOptions = async () => {
    const res = await loadVersions();
    mysqlVersions.value = res.data;
    if (mysqlVersions.value.length != 0) {
        version.value = mysqlVersions.value[0];
        search();
    }
};

const onChangeVersion = async (val: string) => {
    version.value = val;
    search();
    if (isOnSetting.value) {
        let params = {
            version: version.value,
        };
        settingRef.value!.acceptParams(params);
    }
};

const search = async () => {
    let params = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
        version: version.value,
    };
    const res = await searchMysqlDBs(params);
    data.value = res.data.items || [];
    paginationConfig.total = res.data.total;
};

const onBatchDelete = async (row: Database.MysqlDBInfo | null) => {
    let ids: Array<number> = [];
    if (row) {
        ids.push(row.id);
    } else {
        selects.value.forEach((item: Database.MysqlDBInfo) => {
            ids.push(item.id);
        });
    }
    await useDeleteData(deleteMysqlDB, { ids: ids }, 'commons.msg.delete', true);
    search();
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
            onBatchDelete(row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: Database.MysqlDBInfo) => {
            onBatchDelete(row);
        },
    },
];

onMounted(() => {
    loadRunningOptions();
});
</script>
