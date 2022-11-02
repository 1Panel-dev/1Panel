<template>
    <div>
        <Submenu activeName="redis" />
        <el-button style="margin-top: 20px" size="default" icon="Tickets" @click="changeView('status')">
            {{ $t('database.status') }}
        </el-button>
        <el-button style="margin-top: 20px" size="default" icon="Setting" @click="changeView('setting')">
            {{ $t('database.setting') }}
        </el-button>
        <el-button style="margin-top: 20px" size="default" icon="Files" @click="changeView('persistence')">
            {{ $t('database.persistence') }}
        </el-button>
        <el-button style="margin-top: 20px" size="default" icon="Setting" @click="changeView('terminal')">
            {{ $t('database.terminal') }}
        </el-button>

        <el-card style="height: calc(100vh - 178px); margin-top: 5px">
            <Status ref="statusRef"></Status>
            <Setting ref="settingRef"></Setting>
            <Terminal ref="terminalRef"></Terminal>
        </el-card>
    </div>
</template>

<script lang="ts" setup>
import Submenu from '@/views/database/index.vue';
import Status from '@/views/database/redis/status/index.vue';
import Setting from '@/views/database/redis/setting/index.vue';
import Terminal from '@/views/database/redis/terminal/index.vue';
import { onMounted, ref } from 'vue';

const statusRef = ref();
const settingRef = ref();
const terminalRef = ref();

const changeView = async (params: string) => {
    switch (params) {
        case 'status':
            settingRef.value!.onClose();
            terminalRef.value!.onClose();
            statusRef.value!.acceptParams(params);
            break;
        case 'setting':
            statusRef.value!.onClose();
            terminalRef.value!.onClose();
            settingRef.value!.acceptParams(params);
            break;
        case 'terminal':
            statusRef.value!.onClose();
            settingRef.value!.onClose();
            terminalRef.value!.acceptParams(params);
            break;
    }
};

onMounted(() => {
    changeView('status');
});
</script>
