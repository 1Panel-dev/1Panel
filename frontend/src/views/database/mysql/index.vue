<template>
    <div v-loading="loading">
        <LayoutContent :title="'MySQL ' + $t('menu.database')">
            <template #app v-if="mysqlIsExist">
                <AppStatus
                    :app-key="'mysql'"
                    v-model:loading="loading"
                    v-model:mask-show="maskShow"
                    @setting="onSetting"
                    @is-exist="checkExist"
                ></AppStatus>
            </template>

            <template v-if="!isOnSetting" #search>
                <el-select v-model="paginationConfig.from" @change="search()">
                    <template #prefix>{{ $t('commons.table.type') }}</template>
                    <el-option-group>
                        <el-option :label="$t('database.localDB')" value="local" />
                    </el-option-group>
                    <el-option-group :label="$t('database.remote')" v-if="dbOptions.length !== 0">
                        <el-option
                            v-for="(item, index) in dbOptions"
                            :key="index"
                            :value="item.name"
                            :label="item.name"
                        ></el-option>
                    </el-option-group>
                </el-select>
            </template>

            <template #toolbar v-if="!isOnSetting">
                <el-row>
                    <el-col :xs="24" :sm="20" :md="20" :lg="20" :xl="20">
                        <el-button
                            v-if="(mysqlIsExist && mysqlStatus === 'Running') || !isLocal()"
                            type="primary"
                            @click="onOpenDialog()"
                        >
                            {{ $t('database.create') }}
                        </el-button>
                        <el-button @click="onChangeConn" type="primary" plain>
                            {{ $t('database.databaseConnInfo') }}
                        </el-button>
                        <el-button
                            v-if="(mysqlIsExist && mysqlStatus === 'Running') || !isLocal()"
                            @click="loadDB"
                            type="primary"
                            plain
                        >
                            {{ $t('database.loadFromRemote') }}
                        </el-button>
                        <el-button @click="goRemoteDB" type="primary" plain>
                            {{ $t('database.remoteDB') }}
                        </el-button>
                        <el-button
                            v-if="mysqlIsExist && mysqlStatus === 'Running' && isLocal()"
                            @click="goDashboard"
                            icon="Position"
                            type="primary"
                            plain
                        >
                            phpMyAdmin
                        </el-button>
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
            <template #main v-if="(mysqlIsExist && !isOnSetting) || !isLocal()">
                <ComplexTable
                    :pagination-config="paginationConfig"
                    @sort-change="search"
                    @search="search"
                    :data="data"
                    :class="{ mask: mysqlStatus != 'Running' && isLocal() }"
                >
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

        <div v-if="!mysqlIsExist && isLocal()">
            <LayoutContent :title="'MySQL ' + $t('menu.database')" :divider="true">
                <template #main>
                    <div class="app-warn">
                        <div>
                            <span>{{ $t('app.checkInstalledWarn', ['Mysql']) }}</span>
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

        <el-card
            v-if="mysqlStatus != 'Running' && !isOnSetting && mysqlIsExist && !loading && maskShow && isLocal"
            class="mask-prompt"
        >
            <span>{{ $t('commons.service.serviceNotStarted', ['MySQL']) }}</span>
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
import Setting from '@/views/database/mysql/setting/index.vue';
import AppStatus from '@/components/app-status/index.vue';
import Backups from '@/components/backup/index.vue';
import UploadDialog from '@/components/upload/index.vue';
import PortJumpDialog from '@/components/port-jump/index.vue';
import { dateFormat } from '@/utils/util';
import { onMounted, reactive, ref } from 'vue';
import {
    deleteCheckMysqlDB,
    listRemoteDBs,
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

const dbOptions = ref<Array<Database.RemoteDBOption>>([]);

const mysqlName = ref();
const isOnSetting = ref<boolean>();

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
    from: 'local',
});
const searchName = ref();

const mysqlIsExist = ref(true);
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

const uploadRef = ref();

const connRef = ref();
const onChangeConn = async () => {
    connRef.value!.acceptParams({ from: paginationConfig.from });
};

const goRemoteDB = async () => {
    router.push({ name: 'MySQL-Remote' });
};

function isLocal() {
    return paginationConfig.from === 'local';
}

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

const search = async (column?: any) => {
    let params = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
        info: searchName.value,
        from: paginationConfig.from,
        orderBy: column?.order ? column.prop : 'created_at',
        order: column?.order ? column.order : 'null',
    };
    const res = await searchMysqlDBs(params);
    data.value = res.data.items || [];
    paginationConfig.total = res.data.total;
};

const loadDB = async () => {
    loading.value = true;
    await loadDBFromRemote(paginationConfig.from)
        .then(() => {
            loading.value = false;
            search();
        })
        .catch(() => {
            loading.value = false;
        });
};

const goRouter = async () => {
    router.push({ name: 'AppDetail', params: { appKey: 'mysql' } });
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
    dialogPortJumpRef.value.acceptParams({ port: phpadminPort.value });
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

const loadDBOptions = async () => {
    const res = await listRemoteDBs('mysql');
    dbOptions.value = res.data || [];
    for (let i = 0; i < dbOptions.value.length; i++) {
        if (dbOptions.value[i].name === 'local') {
            dbOptions.value.splice(i, 1);
        }
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
    const res = await deleteCheckMysqlDB(row.id);
    if (res.data && res.data.length > 0) {
        checkRef.value.acceptParams({ items: res.data });
    } else {
        deleteRef.value.acceptParams({ id: row.id, name: row.name });
    }
};

const onChangePassword = async (row: Database.MysqlDBInfo) => {
    let param = {
        id: row.id,
        from: row.from,
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
                type: 'mysql',
                name: row.mysqlName,
                detailName: row.name,
            };
            dialogBackupRef.value!.acceptParams(params);
        },
    },
    {
        label: i18n.global.t('database.loadBackup'),
        click: (row: Database.MysqlDBInfo) => {
            let params = {
                type: 'mysql',
                name: mysqlName.value || row.name,
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
});
</script>

<style lang="scss" scoped>
.iconInTable {
    margin-left: 5px;
    margin-top: 3px;
}
</style>
