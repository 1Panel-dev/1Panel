<template>
    <div v-loading="loading">
        <div class="app-status mt-5" v-if="currentDB && currentDB.from === 'remote'">
            <el-card>
                <div class="flex w-full flex-col gap-4 md:flex-row">
                    <div class="flex flex-wrap gap-4 ml-3">
                        <el-tag class="float-left" effect="dark" type="success">Redis</el-tag>
                        <el-tag>{{ $t('app.version') }}: {{ currentDB?.version }}</el-tag>
                    </div>
                </div>
            </el-card>
        </div>
        <LayoutContent title="Redis">
            <template #app v-if="currentDB?.from === 'local'">
                <AppStatus
                    :app-key="currentDB.type"
                    :app-name="appName"
                    v-model:loading="loading"
                    @before="onBefore"
                    @after="onAfter"
                    @setting="onSetting"
                    ref="appStatusRef"
                ></AppStatus>
            </template>
            <template #leftToolBar v-if="!isOnSetting">
                <el-button v-if="currentDB" type="primary" plain @click="onLoadConn">
                    {{ $t('database.databaseConnInfo') }}
                </el-button>
                <el-button @click="goRemoteDB()" type="primary" plain>
                    {{ $t('database.remoteDB') }}
                </el-button>
            </template>
            <template #rightToolBar v-if="!isOnSetting">
                <el-select
                    v-model="currentDBName"
                    @change="changeDatabase()"
                    class="p-w-200 ml-5"
                    v-if="currentDB"
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
            </template>
            <template #main v-if="!isOnSetting">
                <div v-if="currentDB && !isOnSetting" class="mt-5">
                    <Terminal
                        :style="{ height: `calc(100vh - ${loadHeight()})` }"
                        :key="isRefresh"
                        ref="terminalRef"
                        v-show="redisStatus === 'Running' && terminalShow"
                    />
                    <el-empty
                        v-if="redisStatus !== 'Running' || (currentDB.from === 'remote' && !redisCliExist)"
                        :image-size="80"
                        :style="{ height: `calc(100vh - ${loadHeight()})`, 'background-color': '#000' }"
                        :description="loadErrMsg()"
                    >
                        <el-button v-if="currentDB.from === 'remote'" type="primary" @click="installCli">
                            {{ $t('commons.button.enable') }}
                        </el-button>
                    </el-empty>
                    <div>
                        <el-select v-model="quickCmd" clearable filterable @change="quickInput" style="width: 90%">
                            <template #prefix>{{ $t('terminal.quickCommand') }}</template>
                            <el-option
                                v-for="cmd in quickCmdList"
                                :key="cmd.id"
                                :label="cmd.name"
                                :value="cmd.command"
                            />
                        </el-select>
                        <el-button @click="onSetQuickCmd" icon="Setting" style="width: 10%">
                            {{ $t('commons.button.set') }}
                        </el-button>
                    </div>
                </div>

                <div class="app-warn" v-if="isLoaded && dbOptionsLocal.length === 0 && dbOptionsRemote.length === 0">
                    <div class="flex flex-col gap-2 items-center justify-center w-full sm:flex-row">
                        <span>{{ $t('app.checkInstalledWarn', ['Redis']) }}</span>
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

        <Setting ref="settingRef" style="margin-top: 30px" />
        <Conn ref="connRef" @check-exist="reOpenTerminal" @close-terminal="closeTerminal(true)" />

        <DialogPro v-model="open" :title="$t('app.checkTitle')" size="small">
            <div class="flex justify-center items-center gap-2 flex-wrap">
                {{ $t('app.checkInstalledWarn', ['Redis-Commander']) }}
                <el-link icon="Position" @click="getAppDetail('redis-commander')" type="primary">
                    {{ $t('database.goInstall') }}
                </el-link>
            </div>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="open = false">{{ $t('commons.button.cancel') }}</el-button>
                </span>
            </template>
        </DialogPro>

        <QuickCmd ref="dialogQuickCmdRef" @reload="loadQuickCmd" />
    </div>
</template>

<script lang="ts" setup>
import Setting from '@/views/database/redis/setting/index.vue';
import Conn from '@/views/database/redis/conn/index.vue';
import Terminal from '@/components/terminal/index.vue';
import AppStatus from '@/components/app-status/index.vue';
import QuickCmd from '@/views/database/redis/command/index.vue';
import { nextTick, onBeforeUnmount, onMounted, ref } from 'vue';
import { checkAppInstalled } from '@/api/modules/app';
import { GlobalStore } from '@/store';
import { listDatabases, checkRedisCli, installRedisCli } from '@/api/modules/database';
import { Database } from '@/api/interface/database';
import { MsgSuccess } from '@/utils/message';
import i18n from '@/lang';
import { getCommandList } from '@/api/modules/command';
import { routerToName, routerToNameWithQuery } from '@/utils/router';
const globalStore = GlobalStore();

const loading = ref(false);

const terminalRef = ref<InstanceType<typeof Terminal> | null>(null);
const settingRef = ref();
const isOnSetting = ref(false);
const redisIsExist = ref(false);
const redisStatus = ref();
const terminalShow = ref(false);

const appStatusRef = ref();

const open = ref(false);

const redisCliExist = ref();

const appKey = ref('redis');
const appName = ref();

const isLoaded = ref(false);
const dbOptionsLocal = ref<Array<Database.DatabaseOption>>([]);
const dbOptionsRemote = ref<Array<Database.DatabaseOption>>([]);
const currentDB = ref<Database.DatabaseOption>();
const currentDBName = ref();

const quickCmd = ref();
const quickCmdList = ref([]);
const dialogQuickCmdRef = ref();

const isRefresh = ref();

const onSetting = async () => {
    isOnSetting.value = true;
    terminalRef.value?.onClose(false);
    terminalShow.value = false;
    settingRef.value!.acceptParams({ status: redisStatus.value, database: currentDBName.value, type: appKey.value });
};

const loadHeight = () => {
    return globalStore.openMenuTabs ? '470px' : '380px';
};

const getAppDetail = (key: string) => {
    routerToNameWithQuery('AppAll', { install: key });
};
const goRemoteDB = async () => {
    if (currentDB.value) {
        globalStore.setCurrentRedisDB(currentDBName.value);
    }
    routerToName('Redis-Remote');
};

const connRef = ref();
const onLoadConn = async () => {
    connRef.value!.acceptParams({
        from: currentDB.value.from,
        type: currentDB.value.type,
        database: currentDBName.value,
    });
};

const goRouter = async (target: string) => {
    if (target === 'app') {
        routerToNameWithQuery('AppAll', { install: 'redis' });
        return;
    }
    routerToName('Redis-Remote');
};

const changeDatabase = async () => {
    for (const item of dbOptionsLocal.value) {
        if (item.database == currentDBName.value) {
            currentDB.value = item;
            appKey.value = item.type;
            appName.value = item.database;
            appStatusRef.value?.onCheck(appKey.value, appName.value);
            reOpenTerminal();
            return;
        }
    }
    for (const item of dbOptionsRemote.value) {
        if (item.database == currentDBName.value) {
            currentDB.value = item;
            break;
        }
    }
    reOpenTerminal();
};

const loadDBOptions = async () => {
    try {
        const res = await listDatabases('redis,redis-cluster');
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
            reOpenTerminal();
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
            reOpenTerminal();
        }
    } finally {
        isLoaded.value = true;
    }
};

const loadErrMsg = () => {
    return currentDB.value.from === 'local'
        ? i18n.global.t('commons.service.serviceNotStarted', ['Redis'])
        : i18n.global.t('database.redisCliHelper');
};
const reOpenTerminal = async () => {
    closeTerminal(false);
    initTerminal();
};

const initTerminal = async () => {
    loading.value = true;
    if (currentDB.value.from === 'remote') {
        if (!redisCliExist.value) {
            loading.value = false;
            return;
        }
        isRefresh.value = !isRefresh.value;
        loading.value = false;
        redisIsExist.value = true;
        nextTick(() => {
            terminalShow.value = true;
            redisStatus.value = 'Running';
            terminalRef.value.acceptParams({
                endpoint: '/api/v2/containers/exec',
                args: `source=redis&name=${currentDBName.value}&from=${currentDB.value.from}`,
                error: '',
                initCmd: '',
            });
        });
        isRefresh.value = !isRefresh.value;
        return;
    }
    await checkAppInstalled(currentDB.value.type, currentDBName.value)
        .then((res) => {
            redisIsExist.value = res.data.isExist;
            redisStatus.value = res.data.status;
            loading.value = false;
            nextTick(() => {
                if (res.data.status === 'Running') {
                    terminalShow.value = true;
                    terminalRef.value.acceptParams({
                        endpoint: '/api/v2/containers/exec',
                        args: `source=${currentDB.value.type}&name=${currentDBName.value}&from=${currentDB.value.from}`,
                        error: '',
                        initCmd: '',
                    });
                }
            });
            isRefresh.value = !isRefresh.value;
        })
        .catch(() => {
            closeTerminal(false);
            loading.value = false;
        });
};
const closeTerminal = async (isKeepShow: boolean) => {
    isRefresh.value = !isRefresh.value;
    terminalRef.value?.onClose(isKeepShow);
    terminalShow.value = isKeepShow;
};

const checkCliValid = async () => {
    await checkRedisCli()
        .then((res) => {
            redisCliExist.value = res.data;
            loadDBOptions();
        })
        .catch(() => {
            loadDBOptions();
        });
};
const installCli = async () => {
    loading.value = true;
    await installRedisCli()
        .then(() => {
            loading.value = false;
            redisCliExist.value = true;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            reOpenTerminal();
        })
        .catch(() => {
            loading.value = false;
        });
};

const loadQuickCmd = async () => {
    const res = await getCommandList('redis');
    quickCmdList.value = res.data || [];
};

const quickInput = (val: any) => {
    if (val) {
        terminalRef.value?.sendMsg(val);
        quickCmd.value = '';
    }
};

const onSetQuickCmd = () => {
    dialogQuickCmdRef.value.acceptParams();
};

onMounted(() => {
    loadQuickCmd();
    checkCliValid();
});
const onBefore = () => {
    closeTerminal(false);
};
const onAfter = () => {
    initTerminal();
};
onBeforeUnmount(() => {
    closeTerminal(false);
});
</script>

<style lang="scss" scoped>
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
