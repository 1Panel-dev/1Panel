<template>
    <div v-loading="loading">
        <div class="app-status" style="margin-top: 20px" v-if="currentDB?.from === 'remote'">
            <el-card>
                <div>
                    <el-tag style="float: left" effect="dark" type="success">
                        {{ currentDB?.type === 'mysql' ? 'Mysql' : 'MariaDB' }}
                    </el-tag>
                    <el-tag class="status-content">{{ $t('app.version') }}: {{ currentDB?.version }}</el-tag>
                </div>
            </el-card>
        </div>
        <LayoutContent :title="(currentDB?.type === 'mysql' ? 'MySQL ' : 'MariaDB ') + $t('menu.database')">
            <template #app v-if="currentDB?.from === 'local'">
                <AppStatus
                    :app-key="appKey"
                    :app-name="appName"
                    v-model:loading="loading"
                    v-model:mask-show="maskShow"
                    @setting="onSetting"
                    @is-exist="checkExist"
                ></AppStatus>
            </template>

            <template #search v-if="currentDB">
                <el-select v-model="currentDBName" @change="changeDatabase()" class="p-w-200">
                    <template #prefix>{{ $t('commons.table.type') }}</template>
                    <el-option-group :label="$t('database.local')">
                        <div v-for="(item, index) in dbOptionsLocal" :key="index">
                            <el-option v-if="item.from === 'local'" :value="item.database" class="optionClass">
                                <span v-if="item.database.length < 25">{{ item.database }}</span>
                                <el-tooltip v-else :content="item.database" placement="top">
                                    <span>{{ item.database.substring(0, 25) }}...</span>
                                </el-tooltip>
                                <el-tag class="tagClass">
                                    {{ item.type === 'mysql' ? 'MySQL' : 'MariaDB' }}
                                </el-tag>
                            </el-option>
                        </div>
                        <el-button link type="primary" class="jumpAdd" @click="goRouter('app')" icon="Position">
                            {{ $t('database.goInstall') }}
                        </el-button>
                    </el-option-group>
                    <el-option-group :label="$t('database.remote')">
                        <div v-for="(item, index) in dbOptionsRemote" :key="index">
                            <el-option v-if="item.from === 'remote'" :value="item.database" class="optionClass">
                                <span v-if="item.database.length < 25">{{ item.database }}</span>
                                <el-tooltip v-else :content="item.database" placement="top">
                                    <span>{{ item.database.substring(0, 25) }}...</span>
                                </el-tooltip>
                                <el-tag class="tagClass">
                                    {{ item.type === 'mysql' ? 'MySQL' : 'MariaDB' }}
                                </el-tag>
                            </el-option>
                        </div>
                        <el-button link type="primary" class="jumpAdd" @click="goRouter('remote')" icon="Position">
                            {{ $t('database.createRemoteDB') }}
                        </el-button>
                    </el-option-group>
                </el-select>
            </template>

            <template #toolbar>
                <el-row>
                    <el-col :xs="24" :sm="20" :md="20" :lg="20" :xl="20">
                        <el-button
                            v-if="currentDB && (currentDB.from !== 'local' || mysqlStatus === 'Running')"
                            type="primary"
                            @click="onOpenDialog()"
                        >
                            {{ $t('database.create') }}
                        </el-button>
                        <el-button
                            v-if="currentDB && (currentDB.from !== 'local' || mysqlStatus === 'Running')"
                            @click="onChangeConn"
                            type="primary"
                            plain
                        >
                            {{ $t('database.databaseConnInfo') }}
                        </el-button>
                        <el-button
                            v-if="currentDB && (currentDB.from !== 'local' || mysqlStatus === 'Running')"
                            @click="loadDB"
                            type="primary"
                            plain
                        >
                            {{ $t('database.loadFromRemote') }}
                        </el-button>
                        <el-button @click="goRemoteDB" type="primary" plain>
                            {{ $t('database.remoteDB') }}
                        </el-button>
                        <el-dropdown class="ml-3">
                            <el-button type="primary" plain>
                                {{ $t('database.manage') }}
                                <el-icon class="el-icon--right"><arrow-down /></el-icon>
                            </el-button>
                            <template #dropdown>
                                <el-dropdown-menu>
                                    <el-dropdown-item icon="Position" @click="goDashboard('phpMyAdmin')">
                                        phpMyAdmin
                                    </el-dropdown-item>
                                    <el-dropdown-item icon="Position" @click="goDashboard('Adminer')" divided>
                                        Adminer
                                    </el-dropdown-item>
                                </el-dropdown-menu>
                            </template>
                        </el-dropdown>
                    </el-col>
                    <el-col :xs="24" :sm="4" :md="4" :lg="4" :xl="4">
                        <TableSearch @search="search()" v-model:searchName="searchName" />
                    </el-col>
                </el-row>
            </template>
            <template #main v-if="currentDB">
                <ComplexTable :pagination-config="paginationConfig" @sort-change="search" @search="search" :data="data">
                    <el-table-column :label="$t('commons.table.name')" prop="name" sortable />
                    <el-table-column :label="$t('commons.login.username')" prop="username">
                        <template #default="{ row }">
                            <div class="flex items-center" v-if="row.username">
                                <span>
                                    {{ row.username }}
                                </span>
                            </div>
                            <div v-else>
                                <el-button style="margin-left: -3px" type="primary" link @click="onBind(row)">
                                    {{ $t('database.userBind') }}
                                </el-button>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.login.password')" prop="password">
                        <template #default="{ row }">
                            <span v-if="row.username === ''">-</span>
                            <div class="flex items-center" v-if="row.password && row.username">
                                <div class="star-center" v-if="!row.showPassword">
                                    <span>**********</span>
                                </div>
                                <div>
                                    <span v-if="row.showPassword">
                                        {{ row.password }}
                                    </span>
                                </div>
                                <el-button
                                    v-if="!row.showPassword"
                                    link
                                    @click="row.showPassword = true"
                                    icon="View"
                                    class="ml-1.5"
                                ></el-button>
                                <el-button
                                    v-if="row.showPassword"
                                    link
                                    @click="row.showPassword = false"
                                    icon="Hide"
                                    class="ml-1.5"
                                ></el-button>
                                <div>
                                    <CopyButton :content="row.password" type="icon" />
                                </div>
                            </div>
                            <div v-if="row.password === '' && row.username">
                                <el-button style="margin-left: -3px" link type="primary" @click="onChangePassword(row)">
                                    {{ $t('database.passwordHelper') }}
                                </el-button>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.description')" prop="description">
                        <template #default="{ row }">
                            <fu-input-rw-switch v-model="row.description" @blur="onChange(row)" />
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

        <div v-if="dbOptionsLocal.length === 0 && dbOptionsRemote.length === 0">
            <LayoutContent :title="'MySQL ' + $t('menu.database')" :divider="true">
                <template #main>
                    <div class="app-warn">
                        <div>
                            <span>{{ $t('app.checkInstalledWarn', [$t('database.noMysql')]) }}</span>
                            <span @click="goRouter('app')">
                                <el-icon class="ml-2"><Position /></el-icon>
                                {{ $t('database.goInstall') }}
                            </span>
                            <div>
                                <img src="@/assets/images/no_app.svg" />
                            </div>
                        </div>
                    </div>
                </template>
            </LayoutContent>
        </div>

        <el-dialog
            v-model="dashboardVisible"
            :title="$t('app.checkTitle')"
            width="30%"
            :close-on-click-modal="false"
            :destroy-on-close="true"
        >
            <el-alert :closable="false" :title="$t('app.checkInstalledWarn', [dashboardName])" type="info">
                <el-link icon="Position" @click="getAppDetail" type="primary">
                    {{ $t('database.goInstall') }}
                </el-link>
            </el-alert>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="dashboardVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                </span>
            </template>
        </el-dialog>

        <BindDialog ref="bindRef" @search="search" />
        <PasswordDialog ref="passwordRef" @search="search" />
        <RootPasswordDialog ref="connRef" />
        <UploadDialog ref="uploadRef" />
        <OperateDialog @search="search" ref="dialogRef" />
        <Backups ref="dialogBackupRef" />

        <AppResources ref="checkRef"></AppResources>
        <DeleteDialog ref="deleteRef" @search="search" />

        <PortJumpDialog ref="dialogPortJumpRef" />
    </div>
</template>

<script lang="ts" setup>
import BindDialog from '@/views/database/mysql/bind/index.vue';
import OperateDialog from '@/views/database/mysql/create/index.vue';
import DeleteDialog from '@/views/database/mysql/delete/index.vue';
import PasswordDialog from '@/views/database/mysql/password/index.vue';
import RootPasswordDialog from '@/views/database/mysql/conn/index.vue';
import AppResources from '@/views/database/mysql/check/index.vue';
import AppStatus from '@/components/app-status/index.vue';
import Backups from '@/components/backup/index.vue';
import UploadDialog from '@/components/upload/index.vue';
import PortJumpDialog from '@/components/port-jump/index.vue';
import { dateFormat } from '@/utils/util';
import { ElMessageBox } from 'element-plus';
import { onMounted, reactive, ref } from 'vue';
import {
    deleteCheckMysqlDB,
    listDatabases,
    loadDBFromRemote,
    searchMysqlDBs,
    updateMysqlDescription,
} from '@/api/modules/database';
import i18n from '@/lang';
import { Database } from '@/api/interface/database';
import { App } from '@/api/interface/app';
import { GetAppPort } from '@/api/modules/app';
import router from '@/routers';
import { MsgSuccess } from '@/utils/message';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();

const loading = ref(false);
const maskShow = ref(true);

const appKey = ref('mysql');
const appName = ref();

const dbOptionsLocal = ref<Array<Database.DatabaseOption>>([]);
const dbOptionsRemote = ref<Array<Database.DatabaseOption>>([]);
const currentDB = ref<Database.DatabaseOption>();
const currentDBName = ref();

const bindRef = ref();
const checkRef = ref();
const deleteRef = ref();

const phpadminPort = ref();
const adminerPort = ref();
const dashboardName = ref();
const dashboardKey = ref();
const dashboardVisible = ref(false);

const dialogPortJumpRef = ref();

const data = ref();
const paginationConfig = reactive({
    cacheSizeKey: 'mysql-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('mysql-page-size')) || 10,
    total: 0,
    orderBy: 'created_at',
    order: 'null',
});
const searchName = ref();

const mysqlContainer = ref();
const mysqlStatus = ref();
const mysqlVersion = ref();

const dialogRef = ref();
const onOpenDialog = async () => {
    let params = {
        from: currentDB.value.from,
        type: currentDB.value.type,
        database: currentDBName.value,
    };
    dialogRef.value!.acceptParams(params);
};

const dialogBackupRef = ref();

const uploadRef = ref();

const connRef = ref();
const onChangeConn = async () => {
    connRef.value!.acceptParams({
        from: currentDB.value.from,
        type: currentDB.value.type,
        database: currentDBName.value,
    });
};

const goRemoteDB = async () => {
    if (currentDB.value) {
        globalStore.setCurrentDB(currentDB.value.database);
    }
    router.push({ name: 'MySQL-Remote' });
};

const passwordRef = ref();

const onSetting = async () => {
    if (currentDB.value) {
        globalStore.setCurrentDB(currentDB.value.database);
    }
    router.push({ name: 'MySQL-Setting', params: { type: currentDB.value.type, database: currentDB.value.database } });
};

const changeDatabase = async () => {
    for (const item of dbOptionsLocal.value) {
        if (item.database == currentDBName.value) {
            currentDB.value = item;
            appKey.value = item.type;
            appName.value = item.database;
            search();
            return;
        }
    }
    for (const item of dbOptionsRemote.value) {
        if (item.database == currentDBName.value) {
            currentDB.value = item;
            break;
        }
    }
    search();
};

const search = async (column?: any) => {
    paginationConfig.orderBy = column?.order ? column.prop : paginationConfig.orderBy;
    paginationConfig.order = column?.order ? column.order : paginationConfig.order;
    let params = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
        info: searchName.value,
        database: currentDB.value.database,
        orderBy: paginationConfig.orderBy,
        order: paginationConfig.order,
    };
    const res = await searchMysqlDBs(params);
    data.value = res.data.items || [];
    paginationConfig.total = res.data.total;
};

const loadDB = async () => {
    ElMessageBox.confirm(i18n.global.t('database.loadFromRemoteHelper'), i18n.global.t('commons.msg.infoTitle'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        loading.value = true;
        let params = {
            from: currentDB.value.from,
            type: currentDB.value.type,
            database: currentDBName.value,
        };
        await loadDBFromRemote(params)
            .then(() => {
                loading.value = false;
                search();
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

const goRouter = async (target: string) => {
    if (target === 'app') {
        router.push({ name: 'AppAll', query: { install: 'mysql' } });
        return;
    }
    router.push({ name: 'MySQL-Remote' });
};

const onChange = async (info: any) => {
    await updateMysqlDescription({ id: info.id, description: info.description });
    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
};

const goDashboard = async (name: string) => {
    if (name === 'phpMyAdmin') {
        if (phpadminPort.value === 0) {
            dashboardName.value = 'phpMyAdmin';
            dashboardKey.value = 'phpmyadmin';
            dashboardVisible.value = true;
            return;
        }
        dialogPortJumpRef.value.acceptParams({ port: phpadminPort.value });
        return;
    }
    if (adminerPort.value === 0) {
        dashboardName.value = 'Adminer';
        dashboardKey.value = 'adminer';
        dashboardVisible.value = true;
        return;
    }
    dialogPortJumpRef.value.acceptParams({ port: adminerPort.value });
};

const getAppDetail = () => {
    router.push({ name: 'AppAll', query: { install: dashboardKey.value } });
};

const loadPhpMyAdminPort = async () => {
    const res = await GetAppPort('phpmyadmin', '');
    phpadminPort.value = res.data;
};

const loadAdminerPort = async () => {
    const res = await GetAppPort('adminer', '');
    adminerPort.value = res.data;
};

const checkExist = (data: App.CheckInstalled) => {
    mysqlStatus.value = data.status;
    mysqlVersion.value = data.version;
    mysqlContainer.value = data.containerName;
};

const loadDBOptions = async () => {
    const res = await listDatabases('mysql,mariadb');
    let datas = res.data || [];
    dbOptionsLocal.value = [];
    dbOptionsRemote.value = [];
    currentDBName.value = globalStore.currentDB;
    for (const item of datas) {
        if (currentDBName.value && item.database === currentDBName.value) {
            currentDB.value = item;
            if (item.from === 'local') {
                appKey.value = item.type;
                appName.value = item.database;
            }
        }
        if (item.from === 'local') {
            dbOptionsLocal.value.push(item);
        } else {
            dbOptionsRemote.value.push(item);
        }
    }
    if (currentDB.value) {
        globalStore.setCurrentDB('');
        search();
        return;
    }
    if (dbOptionsLocal.value.length !== 0) {
        currentDB.value = dbOptionsLocal.value[0];
        currentDBName.value = dbOptionsLocal.value[0].database;
        appKey.value = dbOptionsLocal.value[0].type;
        appName.value = dbOptionsLocal.value[0].database;
    }
    if (!currentDB.value && dbOptionsRemote.value.length !== 0) {
        currentDB.value = dbOptionsRemote.value[0];
        currentDBName.value = dbOptionsRemote.value[0].database;
    }
    if (currentDB.value) {
        search();
    }
};
const onDelete = async (row: Database.MysqlDBInfo) => {
    let param = {
        id: row.id,
        type: currentDB.value.type,
        database: currentDBName.value,
    };
    const res = await deleteCheckMysqlDB(param);
    if (res.data && res.data.length > 0) {
        checkRef.value.acceptParams({ items: res.data });
    } else {
        deleteRef.value.acceptParams({
            id: row.id,
            type: currentDB.value.type,
            database: currentDBName.value,
            name: row.name,
        });
    }
};

const onBind = async (row: Database.MysqlDBInfo) => {
    let param = {
        database: currentDBName.value,
        mysqlName: row.name,
        from: row.from,
    };
    bindRef.value.acceptParams(param);
};

const onChangePassword = async (row: Database.MysqlDBInfo) => {
    let param = {
        id: row.id,
        from: row.from,
        type: currentDB.value.type,
        database: currentDBName.value,
        mysqlName: row.name,
        operation: 'password',
        username: row.username,
        password: row.password,
    };
    passwordRef.value.acceptParams(param);
};

const buttons = [
    {
        label: i18n.global.t('database.changePassword'),
        disabled: (row: Database.MysqlDBInfo) => {
            return !row.username;
        },
        click: (row: Database.MysqlDBInfo) => {
            onChangePassword(row);
        },
    },
    {
        label: i18n.global.t('database.permission'),
        disabled: (row: Database.MysqlDBInfo) => {
            return !row.password;
        },
        click: (row: Database.MysqlDBInfo) => {
            let param = {
                id: row.id,
                from: row.from,
                type: currentDB.value.type,
                database: currentDBName.value,
                mysqlName: row.name,
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
            let params = {
                type: currentDB.value.type,
                name: currentDBName.value,
                detailName: row.name,
            };
            dialogBackupRef.value!.acceptParams(params);
        },
    },
    {
        label: i18n.global.t('database.loadBackup'),
        click: (row: Database.MysqlDBInfo) => {
            let params = {
                type: currentDB.value.type,
                name: currentDBName.value,
                detailName: row.name,
                remark: row.format,
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

onMounted(() => {
    loadDBOptions();
    loadPhpMyAdminPort();
    loadAdminerPort();
});
</script>

<style lang="scss" scoped>
.iconInTable {
    margin-left: 5px;
    margin-top: 3px;
}
.jumpAdd {
    margin-top: 10px;
    margin-left: 15px;
    margin-bottom: 5px;
    font-size: 12px;
}
.tagClass {
    float: right;
    font-size: 12px;
    margin-top: 5px;
}
.optionClass {
    min-width: 350px;
}
</style>
