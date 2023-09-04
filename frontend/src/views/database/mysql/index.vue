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
        <LayoutContent :title="(currentDB?.type === 'mysql' ? 'MySQL' : 'MariaDB') + $t('menu.database')">
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
                <el-select v-model="currentDBName" @change="changeDatabase()">
                    <template #prefix>{{ $t('commons.table.type') }}</template>
                    <el-option-group :label="$t('database.local')" v-if="dbOptionsLocal.length !== 0">
                        <div v-for="(item, index) in dbOptionsLocal" :key="index">
                            <el-option
                                v-if="item.from === 'local'"
                                :value="item.database"
                                :label="item.database + ' [' + item.type + ']'"
                            ></el-option>
                        </div>
                    </el-option-group>
                    <el-option-group :label="$t('database.remote')" v-if="dbOptionsRemote.length !== 0">
                        <div v-for="(item, index) in dbOptionsRemote" :key="index">
                            <el-option
                                v-if="item.from === 'remote'"
                                :value="item.database"
                                :label="item.database + ' [' + item.type + ']'"
                            ></el-option>
                        </div>
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
                        <el-button @click="goDashboard" icon="Position" type="primary" plain>phpMyAdmin</el-button>
                    </el-col>
                    <el-col :xs="24" :sm="4" :md="4" :lg="4" :xl="4">
                        <div class="search-button">
                            <el-input
                                v-model="searchName"
                                clearable
                                @clear="search()"
                                suffix-icon="Search"
                                @keyup.enter="search()"
                                @change="search()"
                                :placeholder="$t('commons.button.search')"
                            ></el-input>
                        </div>
                    </el-col>
                </el-row>
            </template>
            <template #main v-if="currentDB">
                <ComplexTable :pagination-config="paginationConfig" @sort-change="search" @search="search" :data="data">
                    <el-table-column :label="$t('commons.table.name')" prop="name" sortable />
                    <el-table-column :label="$t('commons.login.username')" prop="username" />
                    <el-table-column :label="$t('commons.login.password')" prop="password">
                        <template #default="{ row }">
                            <div v-if="row.password">
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
                                    <el-icon class="iconInTable" @click="row.showPassword = false" :size="16">
                                        <Hide />
                                    </el-icon>
                                </div>
                                <div style="cursor: pointer; float: left">
                                    <el-icon class="iconInTable" :size="16" @click="onCopy(row)">
                                        <DocumentCopy />
                                    </el-icon>
                                </div>
                            </div>
                            <div v-else>
                                <el-link @click="onChangePassword(row)">
                                    <span style="font-size: 12px">{{ $t('database.passwordHelper') }}</span>
                                </el-link>
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

        <div v-if="!currentDB">
            <LayoutContent :title="'MySQL ' + $t('menu.database')" :divider="true">
                <template #main>
                    <div class="app-warn">
                        <div>
                            <span>{{ $t('app.checkInstalledWarn', [$t('database.noMysql')]) }}</span>
                            <span @click="goRouter">
                                <el-icon><Position /></el-icon>
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
import { MsgError, MsgSuccess } from '@/utils/message';
import useClipboard from 'vue-clipboard3';
const { toClipboard } = useClipboard();

const loading = ref(false);
const maskShow = ref(true);

const appKey = ref('mysql');
const appName = ref();

const dbOptionsLocal = ref<Array<Database.DatabaseOption>>([]);
const dbOptionsRemote = ref<Array<Database.DatabaseOption>>([]);
const currentDB = ref<Database.DatabaseOption>();
const currentDBName = ref();

const checkRef = ref();
const deleteRef = ref();

const phpadminPort = ref();
const phpVisiable = ref(false);

const dialogPortJumpRef = ref();

const data = ref();
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 10,
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
    router.push({ name: 'MySQL-Remote' });
};

const passwordRef = ref();

const onSetting = async () => {
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

const goRouter = async () => {
    router.push({ name: 'AppDetail', params: { appKey: 'mysql' } });
};

const onChange = async (info: any) => {
    await updateMysqlDescription({ id: info.id, description: info.description });
    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
};

const goDashboard = async () => {
    if (phpadminPort.value === 0) {
        phpVisiable.value = true;
        return;
    }
    dialogPortJumpRef.value.acceptParams({ port: phpadminPort.value });
};

const getAppDetail = (key: string) => {
    router.push({ name: 'AppDetail', params: { appKey: key } });
};

const loadDashboardPort = async () => {
    const res = await GetAppPort('phpmyadmin', '');
    phpadminPort.value = res.data;
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
    for (const item of datas) {
        if (item.from === 'local') {
            dbOptionsLocal.value.push(item);
        } else {
            dbOptionsRemote.value.push(item);
        }
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

const onCopy = async (row: any) => {
    try {
        await toClipboard(row.password);
        MsgSuccess(i18n.global.t('commons.msg.copySuccess'));
    } catch (e) {
        MsgError(i18n.global.t('commons.msg.copyfailed'));
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
    loadDashboardPort();
});
</script>

<style lang="scss" scoped>
.iconInTable {
    margin-left: 5px;
    margin-top: 3px;
}
</style>
