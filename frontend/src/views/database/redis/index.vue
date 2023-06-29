<template>
    <div v-loading="loading">
        <LayoutContent :title="'Redis ' + $t('menu.database')">
            <template #app>
                <AppStatus
                    :app-key="'redis'"
                    v-model:loading="loading"
                    v-model:mask-show="maskShow"
                    @before="onBefore"
                    @setting="onSetting"
                    @is-exist="checkExist"
                ></AppStatus>
            </template>
            <template #toolbar v-if="!isOnSetting && redisIsExist">
                <div :class="{ mask: redisStatus != 'Running' }">
                    <el-button type="primary" plain @click="onChangePassword">
                        {{ $t('database.databaseConnInfo') }}
                    </el-button>
                    <el-button type="primary" plain @click="goDashboard" icon="Position">Redis-Commander</el-button>
                </div>
            </template>
            <template #main v-if="redisIsExist && !isOnSetting">
                <Terminal
                    style="height: calc(100vh - 370px)"
                    :key="isRefresh"
                    ref="terminalRef"
                    v-show="terminalShow"
                />
            </template>
        </LayoutContent>

        <el-card v-if="redisStatus != 'Running' && !isOnSetting && redisIsExist && maskShow" class="mask-prompt">
            <span>{{ $t('commons.service.serviceNotStarted', ['Redis']) }}</span>
        </el-card>

        <Setting ref="settingRef" style="margin-top: 30px" />
        <Password ref="passwordRef" @check-exist="initTerminal" @close-terminal="closeTerminal(true)" />
        <el-dialog
            v-model="commandVisiable"
            :title="$t('app.checkTitle')"
            width="30%"
            :close-on-click-modal="false"
            :destroy-on-close="true"
        >
            <el-alert :closable="false" :title="$t('app.checkInstalledWarn', ['Redis-Commander'])" type="info">
                <el-link icon="Position" @click="getAppDetail('redis-commander')" type="primary">
                    {{ $t('database.goInstall') }}
                </el-link>
            </el-alert>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="commandVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                </span>
            </template>
        </el-dialog>
    </div>
</template>

<script lang="ts" setup>
import Setting from '@/views/database/redis/setting/index.vue';
import Password from '@/views/database/redis/password/index.vue';
import Terminal from '@/components/terminal/index.vue';
import AppStatus from '@/components/app-status/index.vue';
import { nextTick, onBeforeUnmount, ref } from 'vue';
import { App } from '@/api/interface/app';
import { GetAppPort } from '@/api/modules/app';
import router from '@/routers';
import { MsgError } from '@/utils/message';
import { getSettingInfo } from '@/api/modules/setting';
import i18n from '@/lang';

const loading = ref(false);
const maskShow = ref(true);

const terminalRef = ref<InstanceType<typeof Terminal> | null>(null);
const settingRef = ref();
const isOnSetting = ref(false);
const redisIsExist = ref(false);
const redisStatus = ref();
const redisName = ref();
const terminalShow = ref(false);

const redisCommandPort = ref();
const commandVisiable = ref(false);

const isRefresh = ref();

const onSetting = async () => {
    isOnSetting.value = true;
    terminalRef.value?.onClose(false);
    terminalShow.value = false;
    settingRef.value!.acceptParams({ status: redisStatus.value, redisName: redisName.value });
};

const goDashboard = async () => {
    if (redisCommandPort.value === 0) {
        commandVisiable.value = true;
        return;
    }
    const res = await getSettingInfo();
    if (!res.data.systemIP) {
        MsgError(i18n.global.t('setting.systemIPWarning'));
        return;
    }
    window.open(`http://${res.data.systemIP}:${redisCommandPort.value}`, '_blank');
};
const getAppDetail = (key: string) => {
    router.push({ name: 'AppDetail', params: { appKey: key } });
};

const loadDashboardPort = async () => {
    const res = await GetAppPort('redis-commander');
    redisCommandPort.value = res.data;
};

const passwordRef = ref();
const onChangePassword = async () => {
    passwordRef.value!.acceptParams();
};

const checkExist = (data: App.CheckInstalled) => {
    redisIsExist.value = data.isExist;
    redisName.value = data.name;
    redisStatus.value = data.status;
    loading.value = false;
    if (redisStatus.value === 'Running') {
        loadDashboardPort();
        nextTick(() => {
            terminalShow.value = true;
            terminalRef.value.acceptParams({
                endpoint: '/api/v1/databases/redis/exec',
                args: '',
                error: '',
            });
        });
    }
};

const initTerminal = async () => {
    if (redisStatus.value === 'Running') {
        nextTick(() => {
            terminalShow.value = true;
            terminalRef.value.acceptParams({
                endpoint: '/api/v1/databases/redis/exec',
                args: '',
                error: '',
            });
        });
    }
};
const closeTerminal = async (isKeepShow: boolean) => {
    isRefresh.value = !isRefresh.value;
    terminalRef.value?.onClose(isKeepShow);
    terminalShow.value = isKeepShow;
};

const onBefore = () => {
    closeTerminal(true);
};
onBeforeUnmount(() => {
    closeTerminal(false);
});
</script>
