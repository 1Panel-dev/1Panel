<template>
    <div v-loading="loading">
        <div class="app-status mt-5" v-if="currentDB?.from === 'remote'">
            <el-card>
                <div class="flex w-full flex-col gap-4 md:flex-row">
                    <div class="flex flex-wrap gap-4 ml-3">
                        <el-tag class="float-left" effect="dark" type="success">MongoDB</el-tag>
                        <el-tag>{{ $t('app.version') }}: {{ currentDB?.version }}</el-tag>
                    </div>
                </div>
            </el-card>
        </div>
        <LayoutContent title="MongoDB">
            <template #app v-if="currentDB?.from === 'local'">
                <AppStatus
                    ref="appStatusRef"
                    app-key="mongodb"
                    :app-name="appName"
                    :hide-setting="true"
                    v-model:loading="loading"
                    v-model:mask-show="maskShow"
                    @is-exist="checkExist"
                />
            </template>

            <template #leftToolBar>
                <el-button
                    v-if="currentDB && (currentDB.from !== 'local' || mongodbStatus === 'Running')"
                    type="primary"
                    @click="openCreateDrawer"
                >
                    {{ $t('commons.button.create') }}
                </el-button>
                <el-button v-if="currentDB" type="primary" plain @click="onLoadConn">
                    {{ $t('database.databaseConnInfo') }}
                </el-button>
                <el-button
                    v-if="currentDB && (currentDB.from !== 'local' || mongodbStatus === 'Running')"
                    type="primary"
                    plain
                    @click="onLoadFromRemote"
                >
                    {{ $t('database.loadFromRemote') }}
                </el-button>
                <el-button type="primary" plain @click="goRemoteDB">
                    {{ $t('database.remoteDB') }}
                </el-button>
                <el-button
                    type="primary"
                    plain
                    :disabled="!currentDB || currentDB.from !== 'local' || mongodbStatus !== 'Running'"
                    @click="goTerminal"
                >
                    {{ $t('menu.terminal') }}
                </el-button>
                <el-button type="primary" plain :disabled="currentDB?.from !== 'local'" @click="goDashboard">
                    {{ $t('database.manage') }}
                </el-button>
            </template>

            <template #rightToolBar>
                <el-select
                    v-if="currentDB"
                    v-model="currentDBName"
                    @change="changeDatabase"
                    class="p-w-200"
                    placement="bottom-end"
                >
                    <template #prefix>{{ $t('commons.table.type') }}</template>
                    <el-option-group :label="$t('commons.table.local')">
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
                <TableSearch @search="search()" v-model:searchName="searchName" />
                <TableRefresh @search="search()" />
            </template>

            <template #main>
                <ComplexTable
                    v-if="currentDB"
                    :class="{ mask: maskShow }"
                    :pagination-config="paginationConfig"
                    @sort-change="search"
                    @search="search"
                    :data="tableRows"
                    :heightDiff="370"
                >
                    <el-table-column :label="$t('commons.table.name')" prop="name" min-width="180" sortable>
                        <template #default="{ row }">
                            <Tooltip v-if="!row.isDelete" :islink="false" :text="row.name" />
                            <div v-else>
                                <span>{{ row.name }}</span>
                                <el-tag round type="info" class="ml-1" size="small">
                                    {{ $t('database.isDelete') }}
                                </el-tag>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('commons.login.username')"
                        prop="username"
                        min-width="180"
                        show-overflow-tooltip
                    >
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
                    <el-table-column :label="$t('commons.login.password')" prop="password" min-width="180">
                        <template #default="{ row }">
                            <span v-if="row.username === ''">-</span>
                            <div v-else-if="row.password" class="flex items-center flex-wrap">
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
                                    <CopyButton :content="row.password" />
                                </div>
                            </div>
                            <div v-else>
                                <el-button
                                    :disabled="row.isDelete"
                                    style="margin-left: -3px"
                                    link
                                    type="primary"
                                    @click="onChangePassword(row)"
                                >
                                    {{ $t('database.passwordHelper') }}
                                </el-button>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('commons.table.description')"
                        prop="description"
                        min-width="220"
                        show-overflow-tooltip
                    >
                        <template #default="{ row }">
                            <fu-input-rw-switch
                                v-model="row.description"
                                @enter="onChange(row)"
                                @blur="onChange(row)"
                            />
                        </template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('commons.table.date')"
                        prop="createdAt"
                        min-width="200"
                        sortable
                        :formatter="dateFormat"
                    />
                    <fu-table-operations
                        :ellipsis="isMobile ? 0 : 10"
                        :min-width="isMobile ? 'auto' : 300"
                        :buttons="buttons"
                        :label="$t('commons.table.operate')"
                        fixed="right"
                        fix
                    />
                </ComplexTable>
                <div v-if="isLoaded && dbOptionsLocal.length === 0 && dbOptionsRemote.length === 0" class="app-warn">
                    <div class="flex flex-col gap-2 items-center justify-center w-full sm:flex-row">
                        <span>{{ $t('app.checkInstalledWarn', ['MongoDB']) }}</span>
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

        <el-card
            v-if="mongodbStatus !== 'Running' && currentDB && !loading && maskShow && currentDB.from === 'local'"
            class="mask-prompt"
        >
            <span>{{ $t('commons.service.serviceNotStarted', ['MongoDB']) }}</span>
        </el-card>

        <DrawerPro
            v-model="createVisible"
            :header="$t('commons.button.create')"
            @close="handleCreateClose"
            size="normal"
        >
            <el-form ref="formRef" v-loading="submitLoading" :model="createForm" :rules="rules" label-position="top">
                <el-form-item :label="$t('commons.table.name')" prop="name">
                    <el-input v-model.trim="createForm.name" clearable @input="handleNameInput" />
                </el-form-item>
                <el-form-item :label="$t('commons.login.username')" prop="username">
                    <el-input v-model.trim="createForm.username" clearable />
                </el-form-item>
                <el-form-item :label="$t('commons.login.password')" prop="password">
                    <el-input v-model.trim="createForm.password" type="password" clearable show-password>
                        <template #append>
                            <el-button @click="random">{{ $t('commons.button.random') }}</el-button>
                        </template>
                    </el-input>
                </el-form-item>
                <el-form-item :label="$t('database.permission')" prop="permission">
                    <el-select v-model="createForm.permission" class="w-full">
                        <el-option
                            v-for="item in permissionOptions"
                            :key="item.value"
                            :label="item.label"
                            :value="item.value"
                        />
                    </el-select>
                </el-form-item>
                <el-form-item :label="$t('commons.table.description')" prop="description">
                    <el-input
                        v-model.trim="createForm.description"
                        type="textarea"
                        clearable
                        :autosize="{ minRows: 3, maxRows: 5 }"
                    />
                </el-form-item>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button :disabled="submitLoading" @click="createVisible = false">
                        {{ $t('commons.button.cancel') }}
                    </el-button>
                    <el-button :disabled="submitLoading" type="primary" @click="submitCreate(formRef)">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </DrawerPro>

        <DialogPro v-model="open" :title="$t('app.checkTitle')" size="small">
            <div class="flex justify-center items-center gap-2 flex-wrap">
                {{ $t('app.checkInstalledWarn', [dashboardName]) }}
                <el-link icon="Position" type="primary" @click="getAppDetail">
                    {{ $t('database.goInstall') }}
                </el-link>
            </div>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="open = false">{{ $t('commons.button.cancel') }}</el-button>
                </span>
            </template>
        </DialogPro>

        <Conn ref="connRef" />
        <BindDialog ref="bindRef" @search="search" />
        <UploadDialog ref="uploadRef" />
        <Backups ref="dialogBackupRef" />
        <PasswordDialog ref="passwordRef" @search="search" />
        <PrivilegesDialog ref="privilegesRef" @search="search" />
        <PortJumpDialog ref="dialogPortJumpRef" />
        <TerminalDialog ref="dialogTerminalRef" />
        <AppResources ref="checkRef" />
        <DeleteDialog ref="deleteRef" @search="search" />
    </div>
</template>

<script lang="ts" setup>
import { nextTick, onMounted, reactive, ref } from 'vue';
import { dateFormat } from '@/utils/date';
import { getRandomStr } from '@/utils/id';
import { Position } from '@element-plus/icons-vue';
import { ElForm, ElMessageBox } from 'element-plus';
import i18n from '@/lang';
import { App } from '@/api/interface/app';
import { Database } from '@/api/interface/database';
import AppStatus from '@/components/app-status/index.vue';
import Backups from '@/components/backup/index.vue';
import PortJumpDialog from '@/components/port-jump/index.vue';
import TerminalDialog from '@/components/terminal/database.vue';
import UploadDialog from '@/components/upload/index.vue';
import BindDialog from '@/views/database/mongodb/bind/index.vue';
import Conn from '@/views/database/mongodb/conn/index.vue';
import DeleteDialog from '@/views/database/mongodb/delete/index.vue';
import PasswordDialog from '@/views/database/mongodb/password/index.vue';
import PrivilegesDialog from '@/views/database/mongodb/permission/index.vue';
import AppResources from '@/views/database/postgresql/check/index.vue';
import { getAppConnInfo, getAppPort } from '@/api/modules/app';
import { Rules } from '@/global/form-rules';
import {
    addMongodbDB,
    deleteCheckMongodbDB,
    listDatabases,
    loadMongodbFromRemote,
    searchMongodbDBs,
    updateMongodbDescription,
} from '@/api/modules/database';
import { MsgSuccess } from '@/utils/message';
import { routerToName, routerToNameWithQuery } from '@/utils/router';
import { GlobalStore } from '@/store';
import Tooltip from '@/components/tooltip/index.vue';
import { useGlobalStore } from '@/composables/useGlobalStore';

const { isMobile } = useGlobalStore();

const globalStore = GlobalStore();

const loading = ref(false);
const maskShow = ref(true);
const submitLoading = ref(false);
const searchName = ref('');
const createVisible = ref(false);
const appStatusRef = ref();
const bindRef = ref();
const connRef = ref();
const passwordRef = ref();
const uploadRef = ref();
const checkRef = ref();
const deleteRef = ref();
const dialogBackupRef = ref();
const dialogTerminalRef = ref();
const dialogPortJumpRef = ref();
const privilegesRef = ref();
const mongoExpressPort = ref(0);
const dashboardName = ref('mongo-express');
const dashboardKey = ref('mongo-express');
const mongodbStatus = ref('');
const open = ref(false);
const appName = ref('');
const isLoaded = ref(false);
const currentDB = ref<Database.DatabaseOption>();
const currentDBName = ref('');
const dbOptionsLocal = ref<Array<Database.DatabaseOption>>([]);
const dbOptionsRemote = ref<Array<Database.DatabaseOption>>([]);

const tableRows = ref<Array<Database.MongodbDBInfo>>([]);
const paginationConfig = reactive({
    cacheSizeKey: 'mongodb-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('mongodb-page-size')) || 20,
    total: 0,
    orderBy: 'createdAt',
    order: 'null',
});

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

const createForm = reactive({
    name: '',
    username: '',
    password: '',
    permission: 'readWrite',
    description: '',
});

const rules = reactive({
    name: [Rules.requiredInput, Rules.dbName],
    username: [Rules.requiredInput, Rules.name],
    password: [Rules.requiredInput, Rules.noSpace, Rules.illegal],
    permission: [Rules.requiredSelect],
});

const permissionOptions = [
    { label: i18n.global.t('database.mongodbPermissionDbOwner'), value: 'dbOwner' },
    { label: i18n.global.t('database.mongodbPermissionRead'), value: 'read' },
    { label: i18n.global.t('database.mongodbPermissionReadWrite'), value: 'readWrite' },
    { label: i18n.global.t('database.mongodbPermissionUserAdmin'), value: 'userAdmin' },
];

const onLoadConn = () => {
    if (!currentDB.value) {
        return;
    }
    connRef.value?.acceptParams({
        from: currentDB.value.from,
        type: currentDB.value.type,
        database: currentDBName.value,
    });
};

const onLoadFromRemote = async () => {
    if (!currentDB.value) {
        return;
    }
    ElMessageBox.confirm(i18n.global.t('database.loadFromRemoteHelper'), i18n.global.t('commons.msg.infoTitle'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        loading.value = true;
        await loadMongodbFromRemote({
            from: currentDB.value.from,
            type: currentDB.value.type,
            database: currentDBName.value,
        })
            .then(async () => {
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                await loadData(true);
            })
            .finally(() => {
                loading.value = false;
            });
    });
};

const goDashboard = () => {
    if (currentDB.value?.from !== 'local') {
        return;
    }
    if (mongoExpressPort.value === 0) {
        dashboardName.value = 'mongo-express';
        dashboardKey.value = 'mongo-express';
        open.value = true;
        return;
    }
    dialogPortJumpRef.value?.acceptParams({ port: mongoExpressPort.value });
};

const getAppDetail = () => {
    routerToNameWithQuery('AppAll', { install: dashboardKey.value });
};

const loadMongoExpressPort = async () => {
    const res = await getAppPort('mongo-express', '');
    mongoExpressPort.value = res.data;
};

const goRemoteDB = async () => {
    if (currentDB.value) {
        globalStore.currentMongodbDB = currentDBName.value;
    }
    routerToName('MongoDB-Remote');
};

const goRouter = async (target: string) => {
    if (target === 'app') {
        routerToNameWithQuery('AppAll', { install: 'mongodb' });
        return;
    }
    routerToName('MongoDB-Remote');
};

const loadData = async (resetPage = false) => {
    if (!currentDBName.value) {
        tableRows.value = [];
        paginationConfig.total = 0;
        return;
    }
    if (resetPage) {
        paginationConfig.currentPage = 1;
    }
    loading.value = true;
    try {
        const res = await searchMongodbDBs({
            page: paginationConfig.currentPage,
            pageSize: paginationConfig.pageSize,
            info: searchName.value,
            database: currentDBName.value,
            orderBy: paginationConfig.orderBy,
            order: paginationConfig.order,
        });
        tableRows.value = (res.data.items || []).map((item) => ({
            ...item,
            showPassword: false,
        }));
        paginationConfig.total = res.data.total || 0;
    } finally {
        loading.value = false;
    }
};

const search = (column?: { prop?: string; order?: string }) => {
    if (column) {
        paginationConfig.orderBy = column.order ? column.prop || 'createdAt' : 'createdAt';
        paginationConfig.order = column.order || 'null';
        loadData();
        return;
    }
    loadData(true);
};

const buttons = [
    {
        label: i18n.global.t('database.changePassword'),
        disabled: (row: Database.MongodbDBInfo) => {
            return !row.username || row.isDelete;
        },
        click: (row: Database.MongodbDBInfo) => {
            onChangePassword(row);
        },
    },
    {
        label: i18n.global.t('database.permission'),
        disabled: (row: Database.MongodbDBInfo) => {
            return !row.username || row.isDelete;
        },
        click: (row: Database.MongodbDBInfo) => {
            let param = {
                database: currentDBName.value,
                name: row.name,
                username: row.username,
            };
            privilegesRef.value.acceptParams(param);
        },
    },
    {
        label: i18n.global.t('database.backupList'),
        disabled: (row: Database.MongodbDBInfo) => {
            return row.isDelete;
        },
        click: (row: Database.MongodbDBInfo) => {
            openBackupList(row);
        },
    },
    {
        label: i18n.global.t('database.loadBackup'),
        disabled: (row: Database.MongodbDBInfo) => {
            return row.isDelete;
        },
        click: (row: Database.MongodbDBInfo) => {
            openUploadDialog(row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: Database.MongodbDBInfo) => {
            onDelete(row);
        },
    },
];

const checkExist = (data: App.CheckInstalled | boolean) => {
    if (!data || typeof data === 'boolean') {
        mongodbStatus.value = '';
        maskShow.value = false;
        tableRows.value = [];
        paginationConfig.total = 0;
        return;
    }
    mongodbStatus.value = data.status || '';
    if (data.isExist) {
        loadData(true);
    }
};

const onChange = async (row: Database.MongodbDBInfo) => {
    await updateMongodbDescription({ id: row.id, description: row.description });
    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
};

const changeDatabase = async () => {
    for (const item of dbOptionsLocal.value) {
        if (item.database === currentDBName.value) {
            currentDB.value = item;
            appName.value = item.database;
            globalStore.currentMongodbDB = item.database;
            await nextTick();
            appStatusRef.value?.onCheck('mongodb', item.database);
            return;
        }
    }
    for (const item of dbOptionsRemote.value) {
        if (item.database === currentDBName.value) {
            currentDB.value = item;
            appName.value = '';
            mongodbStatus.value = '';
            maskShow.value = false;
            globalStore.currentMongodbDB = item.database;
            await loadData(true);
            return;
        }
    }
};

const loadDBOptions = async () => {
    try {
        const res = await listDatabases('mongodb');
        const datas = res.data || [];
        dbOptionsLocal.value = [];
        dbOptionsRemote.value = [];
        currentDB.value = undefined;
        currentDBName.value = globalStore.currentMongodbDB;
        for (const item of datas) {
            if (currentDBName.value && item.database === currentDBName.value) {
                currentDB.value = item;
                if (item.from === 'local') {
                    appName.value = item.database;
                }
            }
            if (item.from === 'local') {
                dbOptionsLocal.value.push(item);
            } else {
                dbOptionsRemote.value.push(item);
            }
        }
        if (!currentDB.value && dbOptionsLocal.value.length !== 0) {
            currentDB.value = dbOptionsLocal.value[0];
            currentDBName.value = dbOptionsLocal.value[0].database;
            appName.value = dbOptionsLocal.value[0].database;
        }
        if (!currentDB.value && dbOptionsRemote.value.length !== 0) {
            currentDB.value = dbOptionsRemote.value[0];
            currentDBName.value = dbOptionsRemote.value[0].database;
            appName.value = '';
        }
        if (currentDB.value) {
            globalStore.currentMongodbDB = currentDBName.value;
            if (currentDB.value.from === 'remote') {
                maskShow.value = false;
                await loadData(true);
            } else {
                await nextTick();
                appStatusRef.value?.onCheck('mongodb', currentDBName.value);
            }
        } else {
            maskShow.value = false;
            tableRows.value = [];
            paginationConfig.total = 0;
        }
    } finally {
        isLoaded.value = true;
    }
};

const openBackupList = (row: { name: string }) => {
    if (!currentDB.value) {
        return;
    }
    const params: {
        type: string;
        name: string;
        detailName: string;
        status?: string;
    } = {
        type: 'mongodb',
        name: currentDBName.value,
        detailName: row.name,
    };
    if (currentDB.value.from === 'local') {
        params.status = mongodbStatus.value;
    }
    dialogBackupRef.value?.acceptParams(params);
};

const openUploadDialog = (row: { name: string }) => {
    if (!currentDB.value) {
        return;
    }
    uploadRef.value?.acceptParams({
        type: 'mongodb',
        name: currentDBName.value,
        detailName: row.name,
        remark: '.gz',
    });
};

const onBind = async (row: Database.MongodbDBInfo) => {
    bindRef.value.acceptParams({
        database: currentDBName.value,
        name: row.name,
    });
};

const onChangePassword = async (row: Database.MongodbDBInfo) => {
    passwordRef.value.acceptParams({
        database: currentDBName.value,
        name: row.name,
        username: row.username,
    });
};

const shellQuote = (value: string) => {
    return `'${String(value || '').replace(/'/g, `'\\''`)}'`;
};

const buildMongoTerminalCommand = (username: string, password: string) => {
    return `mongosh "mongodb://127.0.0.1:27017/admin?authSource=admin" --username ${shellQuote(
        username,
    )} --password ${shellQuote(password)}\r`;
};

const goTerminal = async () => {
    if (!currentDB.value || currentDB.value.from !== 'local' || mongodbStatus.value !== 'Running') {
        return;
    }
    const res = await getAppConnInfo('mongodb', currentDBName.value);
    const connInfo = res.data;
    if (!connInfo?.containerName) {
        return;
    }
    dialogTerminalRef.value?.acceptParams({
        containerID: connInfo.containerName,
        command: '/bin/sh',
        initCmd: buildMongoTerminalCommand(connInfo.username || 'root', connInfo.password || ''),
        waitForPrompt: 'admin> ',
        resourceName: currentDBName.value,
    });
};

const openCreateDrawer = () => {
    createForm.name = '';
    createForm.username = '';
    random();
    createForm.permission = 'readWrite';
    createForm.description = '';
    createVisible.value = true;
};

const handleCreateClose = () => {
    createVisible.value = false;
};

const handleNameInput = () => {
    createForm.username = createForm.name;
};

const random = () => {
    createForm.password = getRandomStr(16);
};

const submitCreate = async (formEl: FormInstance | undefined) => {
    if (!formEl || !currentDB.value) {
        return;
    }
    const valid = await formEl.validate().catch(() => false);
    if (!valid) {
        return;
    }
    submitLoading.value = true;
    await addMongodbDB({
        name: createForm.name,
        from: currentDB.value.from,
        database: currentDBName.value,
        username: createForm.username,
        password: createForm.password,
        permission: createForm.permission,
        description: createForm.description,
    })
        .then(async () => {
            createVisible.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            search();
        })
        .finally(() => {
            submitLoading.value = false;
        });
};

const onDelete = async (row: Database.MongodbDBInfo) => {
    if (!currentDB.value) {
        return;
    }
    const res = await deleteCheckMongodbDB({
        id: row.id,
        type: currentDB.value.type,
        database: currentDBName.value,
    });
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

onMounted(() => {
    loadMongoExpressPort();
    loadDBOptions();
});
</script>

<style lang="scss" scoped>
.jumpAdd {
    margin-top: 10px;
    margin-left: 15px;
    margin-bottom: 5px;
    font-size: 12px;
}

.optionClass {
    min-width: 350px;
}
</style>
