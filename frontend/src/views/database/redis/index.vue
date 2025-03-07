<template>
    <div v-loading="loading">
        <div class="app-status" style="margin-top: 20px" v-if="currentDB && currentDB.from === 'remote'">
            <el-card>
                <div class="flex w-full flex-col gap-4 md:flex-row">
                    <div class="flex flex-wrap gap-4">
                        <el-tag style="float: left" effect="dark" type="success">Redis</el-tag>
                        <el-tag>{{ $t('app.version') }}: {{ currentDB?.version }}</el-tag>
                    </div>
                </div>
            </el-card>
        </div>
        <LayoutContent>
            <template #app v-if="currentDB?.from === 'local'">
                <AppStatus
                    :app-key="'redis'"
                    :app-name="appName"
                    v-model:loading="loading"
                    @before="onBefore"
                    @after="onAfter"
                    @setting="onSetting"
                    ref="appStatusRef"
                ></AppStatus>
            </template>
            <template #search v-if="!isOnSetting && currentDB">
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
            <template #toolbar v-if="!isOnSetting">
                <div class="flex justify-between gap-2 flex-wrap sm:flex-row">
                    <div class="flex flex-wrap gap-3">
                        <el-button v-if="currentDB" type="primary" plain @click="onLoadConn">
                            {{ $t('database.databaseConnInfo') }}
                        </el-button>
                        <el-button @click="goRemoteDB" type="primary" plain>
                            {{ $t('database.manageRemoteDB') }}
                        </el-button>
                    </div>
                </div>
            </template>
        </LayoutContent>

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
                    <el-option v-for="cmd in quickCmdList" :key="cmd.id" :label="cmd.name" :value="cmd.command" />
                </el-select>
                <el-button @click="onSetQuickCmd" icon="Setting" style="width: 10%">
                    {{ $t('commons.button.set') }}
                </el-button>
            </div>
        </div>

        <div v-if="dbOptionsLocal.length === 0 && dbOptionsRemote.length === 0">
            <LayoutContent :title="'Redis ' + $t('menu.database').toLowerCase()" :divider="true">
                <template #main>
                    <div class="app-warn">
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
        </div>

        <Setting ref="settingRef" style="margin-top: 30px" />
        <Conn ref="connRef" @check-exist="reOpenTerminal" @close-terminal="closeTerminal(true)" />
        <el-dialog
            v-model="commandVisible"
            :title="$t('app.checkTitle')"
            width="30%"
            :close-on-click-modal="false"
            :destroy-on-close="true"
        >
            <div class="flex justify-center items-center gap-2 flex-wrap">
                {{ $t('app.checkInstalledWarn', ['Redis-Commander']) }}
                <el-link icon="Position" @click="getAppDetail('redis-commander')" type="primary">
                    {{ $t('database.goInstall') }}
                </el-link>
            </div>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="commandVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                </span>
            </template>
        </el-dialog>

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
import { CheckAppInstalled } from '@/api/modules/app';
import router from '@/routers';
import { GlobalStore } from '@/store';
import { listDatabases, checkRedisCli, installRedisCli } from '@/api/modules/database';
import { Database } from '@/api/interface/database';
import { MsgSuccess } from '@/utils/message';
import i18n from '@/lang';
import { getRedisCommandList } from '@/api/modules/host';
const globalStore = GlobalStore();

const loading = ref(false);

const terminalRef = ref<InstanceType<typeof Terminal> | null>(null);
const settingRef = ref();
const isOnSetting = ref(false);
const redisIsExist = ref(false);
const redisStatus = ref();
const terminalShow = ref(false);

const appStatusRef = ref();
const commandVisible = ref(false);

const redisCliExist = ref();

const appKey = ref('redis');
const appName = ref();
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
    settingRef.value!.acceptParams({ status: redisStatus.value, database: currentDBName.value });
};

const loadHeight = () => {
    return globalStore.openMenuTabs ? '470px' : '380px';
};

const getAppDetail = (key: string) => {
    router.push({ name: 'AppAll', query: { install: key } });
};
const goRemoteDB = async () => {
    if (currentDB.value) {
        globalStore.setCurrentRedisDB(currentDBName.value);
    }
    router.push({ name: 'Redis-Remote' });
};

const connRef = ref();
const onLoadConn = async () => {
    connRef.value!.acceptParams({
        from: currentDB.value.from,
        type: currentDB.value.type,
        database: currentDBName.value,
    });
};
//
// const mobile = computed(() => {
//     return globalStore.isMobile();
// });

const goRouter = async (target: string) => {
    if (target === 'app') {
        router.push({ name: 'AppAll', query: { install: 'redis' } });
        return;
    }
    router.push({ name: 'Redis-Remote' });
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
    const res = await listDatabases('redis');
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
                endpoint: '/api/v1/containers/exec',
                args: `source=redis&name=${currentDBName.value}&from=${currentDB.value.from}`,
                error: '',
                initCmd: '',
            });
        });
        isRefresh.value = !isRefresh.value;
        return;
    }
    await CheckAppInstalled('redis', currentDBName.value)
        .then((res) => {
            redisIsExist.value = res.data.isExist;
            redisStatus.value = res.data.status;
            loading.value = false;
            nextTick(() => {
                if (res.data.status === 'Running') {
                    terminalShow.value = true;
                    terminalRef.value.acceptParams({
                        endpoint: '/api/v1/containers/exec',
                        args: `source=redis&name=${currentDBName.value}&from=${currentDB.value.from}`,
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
    const res = await getRedisCommandList();
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
