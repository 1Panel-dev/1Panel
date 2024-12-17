<template>
    <div v-loading="loading">
        <div class="app-status" style="margin-top: 20px" v-if="currentDB?.from === 'remote'">
            <el-card>
                <div class="flex w-full flex-col gap-4 md:flex-row">
                    <div class="flex flex-wrap gap-4">
                        <el-tag style="float: left" effect="dark" type="success">PostgreSQL</el-tag>
                        <el-tag>{{ $t('app.version') }}: {{ currentDB?.version }}</el-tag>
                    </div>
                </div>
            </el-card>
        </div>
        <LayoutContent :title="'PostgreSQL'">
            <template #app v-if="currentDB?.from === 'local'">
                <AppStatus
                    :app-key="appKey"
                    :app-name="appName"
                    v-model:loading="loading"
                    v-model:mask-show="maskShow"
                    @setting="onSetting"
                    @is-exist="checkExist"
                    ref="appStatusRef"
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
                            </el-option>
                        </div>
                        <el-button link type="primary" class="jumpAdd" @click="goRouter('remote')" icon="Position">
                            {{ $t('database.createRemoteDB') }}
                        </el-button>
                    </el-option-group>
                </el-select>
            </template>

            <template #toolbar>
                <div class="flex justify-between gap-2 flex-wrap sm:flex-row">
                    <div class="flex flex-wrap gap-3">
                        <el-button
                            v-if="currentDB && (currentDB.from !== 'local' || postgresqlStatus === 'Running')"
                            type="primary"
                            @click="onOpenDialog()"
                        >
                            {{ $t('database.create') }}
                        </el-button>
                        <el-button v-if="currentDB" @click="onChangeConn" type="primary" plain>
                            {{ $t('database.databaseConnInfo') }}
                        </el-button>
                        <el-button
                            v-if="currentDB && (currentDB.from !== 'local' || postgresqlStatus === 'Running')"
                            @click="loadDB"
                            type="primary"
                            plain
                        >
                            {{ $t('database.loadFromRemote') }}
                        </el-button>
                        <el-button @click="goRemoteDB" type="primary" plain>
                            {{ $t('database.manageRemoteDB') }}
                        </el-button>
                        <el-button @click="goDashboard()" type="primary" plain>PGAdmin4</el-button>
                    </div>
                    <div><TableSearch @search="search()" v-model:searchName="searchName" /></div>
                </div>
            </template>
            <template #main v-if="currentDB">
                <ComplexTable
                    :class="{ mask: maskShow }"
                    :pagination-config="paginationConfig"
                    @sort-change="search"
                    @search="search"
                    :data="data"
                >
                    <el-table-column :label="$t('commons.table.name')" prop="name" sortable>
                        <template #default="{ row }">
                            <Tooltip v-if="!row.isDelete" :islink="false" :text="row.name" />
                            <div v-else>
                                <span v-if="row.name.length < 15">{{ row.name }}</span>
                                <el-tooltip v-else :content="row.name">{{ row.name.substring(0, 10) }}...</el-tooltip>
                                <el-tag round type="info" class="ml-1" size="small">
                                    {{ $t('database.isDelete') }}
                                </el-tag>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.login.username')" prop="username">
                        <template #default="{ row }">
                            <div class="flex items-center" v-if="row.username">
                                <span>
                                    {{ row.username }}
                                </span>
                            </div>
                            <div v-else>
                                <el-button
                                    :disabled="row.isDelete"
                                    style="margin-left: -3px"
                                    type="primary"
                                    link
                                    @click="onBind(row)"
                                >
                                    {{ $t('database.userBind') }}
                                </el-button>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.login.password')" prop="password">
                        <template #default="{ row }">
                            <span v-if="row.username === '' || row.password === ''">-</span>
                            <div class="flex items-center flex-wrap" v-else>
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
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.description')" prop="description" show-overflow-tooltip>
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
                        :ellipsis="mobile ? 0 : 10"
                        :min-width="mobile ? 'auto' : 400"
                        :buttons="buttons"
                        :label="$t('commons.table.operate')"
                        fixed="right"
                        fix
                    />
                </ComplexTable>
            </template>
        </LayoutContent>

        <el-card
            v-if="postgresqlStatus != 'Running' && currentDB && !loading && maskShow && currentDB?.from === 'local'"
            class="mask-prompt"
        >
            <span>{{ $t('commons.service.serviceNotStarted', ['PostgreSQL']) }}</span>
        </el-card>

        <div v-if="dbOptionsLocal.length === 0 && dbOptionsRemote.length === 0">
            <LayoutContent :title="'PostgreSQL ' + $t('menu.database').toLowerCase()" :divider="true">
                <template #main>
                    <div class="app-warn">
                        <div class="flex flex-col gap-2 items-center justify-center w-full sm:flex-row">
                            <span>{{ $t('app.checkInstalledWarn', [$t('database.noPostgresql')]) }}</span>
                            <span @click="goRouter('app')" class="flex items-center justify-center gap-0.5">
                                <el-icon><Position /></el-icon>
                                {{ $t('database.goInstall') }}
                            </span>
                        </div>
                        <div>
                            <img src="@/assets/images/no_app.svg" />
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
            <div class="flex justify-center items-center gap-2 flex-wrap">
                {{ $t('app.checkInstalledWarn', [dashboardName]) }}
                <el-link icon="Position" @click="getAppDetail" type="primary">
                    {{ $t('database.goInstall') }}
                </el-link>
            </div>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="dashboardVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                </span>
            </template>
        </el-dialog>

        <PrivilegesDialog ref="privilegesRef" @search="search" />
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
import BindDialog from '@/views/database/postgresql/bind/index.vue';
import OperateDialog from '@/views/database/postgresql/create/index.vue';
import DeleteDialog from '@/views/database/postgresql/delete/index.vue';
import PasswordDialog from '@/views/database/postgresql/password/index.vue';
import PrivilegesDialog from '@/views/database/postgresql/privileges/index.vue';
import RootPasswordDialog from '@/views/database/postgresql/conn/index.vue';
import AppResources from '@/views/database/postgresql/check/index.vue';
import AppStatus from '@/components/app-status/index.vue';
import Backups from '@/components/backup/index.vue';
import UploadDialog from '@/components/upload/index.vue';
import PortJumpDialog from '@/components/port-jump/index.vue';
import { dateFormat } from '@/utils/util';
import { computed, onMounted, reactive, ref } from 'vue';
import {
    deleteCheckPostgresqlDB,
    listDatabases,
    loadPgFromRemote,
    searchPostgresqlDBs,
    updatePostgresqlDescription,
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

const appKey = ref('postgresql');
const appName = ref();

const dbOptionsLocal = ref<Array<Database.DatabaseOption>>([]);
const dbOptionsRemote = ref<Array<Database.DatabaseOption>>([]);
const currentDB = ref<Database.DatabaseOption>();
const currentDBName = ref();

const checkRef = ref();
const deleteRef = ref();
const bindRef = ref();
const privilegesRef = ref();

const pgadminPort = ref();
const dashboardName = ref();
const dashboardKey = ref();
const dashboardVisible = ref(false);

const dialogPortJumpRef = ref();

const data = ref();
const paginationConfig = reactive({
    cacheSizeKey: 'postgresql-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('postgresql-page-size')) || 10,
    total: 0,
    orderBy: 'created_at',
    order: 'null',
});
const searchName = ref();

const postgresqlContainer = ref();
const postgresqlStatus = ref();
const postgresqlVersion = ref();

const appStatusRef = ref();
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

const mobile = computed(() => {
    return globalStore.isMobile();
});

const goRemoteDB = async () => {
    if (currentDB.value) {
        globalStore.setCurrentDB(currentDB.value.database);
    }
    router.push({ name: 'PostgreSQL-Remote' });
};

const passwordRef = ref();

const onSetting = async () => {
    if (currentDB.value) {
        globalStore.setCurrentDB(currentDB.value.database);
    }
    router.push({
        name: 'PostgreSQL-Setting',
        params: { type: currentDB.value.type, database: currentDB.value.database },
    });
};

const changeDatabase = async () => {
    for (const item of dbOptionsLocal.value) {
        if (item.database == currentDBName.value) {
            currentDB.value = item;
            appKey.value = item.type;
            appName.value = item.database;
            search();
            appStatusRef.value?.onCheck(appKey.value, appName.value);
            return;
        }
    }
    for (const item of dbOptionsRemote.value) {
        if (item.database == currentDBName.value) {
            maskShow.value = false;
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
    const res = await searchPostgresqlDBs(params);
    data.value = res.data.items || [];
    paginationConfig.total = res.data.total;
};

const onBind = async (row: Database.PostgresqlDBInfo) => {
    let param = {
        name: row.name,
        database: currentDBName.value,
    };
    bindRef.value.acceptParams(param);
};

const loadDB = async () => {
    ElMessageBox.confirm(i18n.global.t('database.loadFromRemoteHelper'), i18n.global.t('commons.msg.infoTitle'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        loading.value = true;
        await loadPgFromRemote(currentDBName.value)
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
        router.push({ name: 'AppAll', query: { install: 'postgresql' } });
        return;
    }
    router.push({ name: 'PostgreSQL-Remote' });
};

const onChange = async (info: any) => {
    await updatePostgresqlDescription({ id: info.id, description: info.description });
    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
};

const goDashboard = async () => {
    if (pgadminPort.value === 0) {
        dashboardName.value = 'PGAdmin4';
        dashboardKey.value = 'pgadmin4';
        dashboardVisible.value = true;
        return;
    }
    dialogPortJumpRef.value.acceptParams({ port: pgadminPort.value });
    return;
};

const getAppDetail = () => {
    router.push({ name: 'AppAll', query: { install: dashboardKey.value } });
};

const loadPGAdminPort = async () => {
    const res = await GetAppPort('pgadmin4', '');
    pgadminPort.value = res.data;
};

const checkExist = (data: App.CheckInstalled) => {
    postgresqlStatus.value = data.status;
    postgresqlVersion.value = data.version;
    postgresqlContainer.value = data.containerName;
};

const loadDBOptions = async () => {
    const res = await listDatabases('postgresql');
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
        if (currentDB.value.from === 'remote') {
            maskShow.value = false;
        }
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
    if (currentDB.value.from === 'remote') {
        maskShow.value = false;
    }
};
const onDelete = async (row: Database.PostgresqlDBInfo) => {
    let param = {
        id: row.id,
        type: currentDB.value.type,
        database: currentDBName.value,
    };
    const res = await deleteCheckPostgresqlDB(param);
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

const onChangePassword = async (row: Database.PostgresqlDBInfo) => {
    let param = {
        id: row.id,
        from: row.from,
        type: currentDB.value.type,
        database: currentDBName.value,
        postgresqlName: row.name,
        operation: 'password',
        username: row.username,
        password: row.password,
    };
    passwordRef.value.acceptParams(param);
};

const buttons = [
    {
        label: i18n.global.t('database.changePassword'),
        disabled: (row: Database.PostgresqlDBInfo) => {
            return !row.username || row.isDelete;
        },
        click: (row: Database.PostgresqlDBInfo) => {
            onChangePassword(row);
        },
    },
    {
        label: i18n.global.t('database.permission'),
        disabled: (row: Database.PostgresqlDBInfo) => {
            return !row.username || row.isDelete;
        },
        click: (row: Database.PostgresqlDBInfo) => {
            let param = {
                database: currentDBName.value,
                name: row.name,
                username: row.username,
                superUser: row.superUser,
            };
            privilegesRef.value.acceptParams(param);
        },
    },
    {
        label: i18n.global.t('database.backupList'),
        disabled: (row: Database.PostgresqlDBInfo) => {
            return row.isDelete;
        },
        click: (row: Database.PostgresqlDBInfo) => {
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
        disabled: (row: Database.PostgresqlDBInfo) => {
            return row.isDelete;
        },
        click: (row: Database.PostgresqlDBInfo) => {
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
        click: (row: Database.PostgresqlDBInfo) => {
            onDelete(row);
        },
    },
];

onMounted(() => {
    loadDBOptions();
    loadPGAdminPort();
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
