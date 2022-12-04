<template>
    <div>
        <Submenu activeName="redis" />

        <AppStatus :app-key="'redis'" style="margin-top: 20px" @setting="onSetting" @is-exist="checkExist"></AppStatus>
        <div v-show="redisIsExist">
            <Setting ref="settingRef" style="margin-top: 20px" />

            <el-button style="margin-top: 20px" type="p" @click="goDashboard" icon="Position">Redis-Command</el-button>
            <Terminal v-show="!isOnSetting" ref="terminalRef" />
        </div>
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

const terminalRef = ref();
const settingRef = ref();
const isOnSetting = ref(false);
const redisIsExist = ref(false);

const redisCommandPort = ref();

const onSetting = async () => {
    isOnSetting.value = true;
    terminalRef.value.onClose();
    settingRef.value!.acceptParams();
};

const goDashboard = async () => {
    window.open('http://localhost:' + redisCommandPort.value, '_blank');
};

const loadDashboardPort = async () => {
    const res = await GetAppPort('phpmyadmin');
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
