<template>
    <div>
        <Submenu activeName="redis" />
        <div v-show="!redisInfo.isExist" style="margin-top: 20px">
            <el-alert :closable="false" :title="$t('database.noMysql', ['Redis'])" type="info">
                <el-link icon="Position" @click="goRouter('/apps')" type="primary">
                    {{ $t('database.goInstall') }}
                </el-link>
            </el-alert>
        </div>
        <div v-show="redisInfo.isExist">
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
    </div>
</template>

<script lang="ts" setup>
import Submenu from '@/views/database/index.vue';
import Status from '@/views/database/redis/status/index.vue';
import Setting from '@/views/database/redis/setting/index.vue';
import Persistence from '@/views/database/redis/persistence/index.vue';
import Terminal from '@/views/database/redis/terminal/index.vue';
import { onMounted, reactive, ref } from 'vue';
import { CheckAppInstalled } from '@/api/modules/app';
import { useRouter } from 'vue-router';
const router = useRouter();

const statusRef = ref();
const settingRef = ref();
const persistenceRef = ref();
const terminalRef = ref();

const redisInfo = reactive({
    name: '',
    version: '',
    isExist: false,
});

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
const checkRedisInstalled = async () => {
    const res = await CheckAppInstalled('redis');
    redisInfo.isExist = res.data.isExist;
    redisInfo.name = res.data.name;
    redisInfo.version = res.data.version;
    if (redisInfo.isExist) {
        changeView('status');
    }
};

const goRouter = async (path: string) => {
    router.push({ path: path });
};

onMounted(() => {
    checkRedisInstalled();
});
</script>
