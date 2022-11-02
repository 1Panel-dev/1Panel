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

        <Status ref="statusRef"></Status>
        <Setting ref="settingRef"></Setting>
        <Persistence ref="persistenceRef"></Persistence>
        <Terminal ref="terminalRef"></Terminal>
    </div>
</template>

<script lang="ts" setup>
import Submenu from '@/views/database/index.vue';
import Status from '@/views/database/redis/status/index.vue';
import Setting from '@/views/database/redis/setting/index.vue';
import Persistence from '@/views/database/redis/persistence/index.vue';
import Terminal from '@/views/database/redis/terminal/index.vue';
import { onMounted, ref } from 'vue';

const statusRef = ref();
const settingRef = ref();
const persistenceRef = ref();
const terminalRef = ref();

const changeView = async (params: string) => {
    switch (params) {
        case 'status':
            settingRef.value!.onClose();
            terminalRef.value!.onClose();
            persistenceRef.value!.onClose();
            statusRef.value!.acceptParams(params);
            break;
        case 'setting':
            statusRef.value!.onClose();
            terminalRef.value!.onClose();
            persistenceRef.value!.onClose();
            settingRef.value!.acceptParams(params);
            break;
        case 'persistence':
            statusRef.value!.onClose();
            settingRef.value!.onClose();
            terminalRef.value!.onClose();
            persistenceRef.value!.acceptParams(params);
            break;
        case 'terminal':
            statusRef.value!.onClose();
            settingRef.value!.onClose();
            persistenceRef.value!.onClose();
            terminalRef.value!.acceptParams(params);
            break;
    }
};

onMounted(() => {
    changeView('status');
});
</script>
