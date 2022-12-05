<template>
    <div>
        <Submenu activeName="redis" />

        <AppStatus :app-key="'redis'" style="margin-top: 20px" @setting="onSetting" @is-exist="checkExist"></AppStatus>
        <div v-show="redisIsExist">
            <el-button style="margin-top: 20px" type="p" @click="goDashboard" icon="Position">
                Redis-Commander
            </el-button>

            <Setting ref="settingRef" style="margin-top: 10px" />

            <Terminal style="margin-top: 10px" v-show="!isOnSetting" ref="terminalRef" />
        </div>

        <el-dialog
            v-model="commandVisiable"
            :title="$t('app.checkTitle')"
            width="30%"
            :close-on-click-modal="false"
            :destroy-on-close="true"
        >
            <el-alert :closable="false" :title="$t('app.checkInstalledWarn', ['Redis-Commander'])" type="info">
                <el-link icon="Position" @click="goRouter('/apps')" type="primary">
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
import Submenu from '@/views/database/index.vue';
import Setting from '@/views/database/redis/setting/index.vue';
import Terminal from '@/views/database/redis/terminal/index.vue';
import AppStatus from '@/components/app-status/index.vue';
import { ref } from 'vue';
import { App } from '@/api/interface/app';
import { GetAppPort } from '@/api/modules/app';
import router from '@/routers';

const terminalRef = ref();
const settingRef = ref();
const isOnSetting = ref(false);
const redisIsExist = ref(false);

const redisCommandPort = ref();
const commandVisiable = ref(false);

const onSetting = async () => {
    isOnSetting.value = true;
    terminalRef.value.onClose();
    settingRef.value!.acceptParams();
};

const goRouter = async (path: string) => {
    router.push({ path: path });
};

const goDashboard = async () => {
    if (redisCommandPort.value === 0) {
        commandVisiable.value = true;
        return;
    }
    window.open('http://localhost:' + redisCommandPort.value, '_blank');
};

const loadDashboardPort = async () => {
    const res = await GetAppPort('redis-commander');
    redisCommandPort.value = res.data;
};

const checkExist = (data: App.CheckInstalled) => {
    redisIsExist.value = data.isExist;
    if (redisIsExist.value) {
        loadDashboardPort();
        terminalRef.value.acceptParams();
    }
};
</script>
