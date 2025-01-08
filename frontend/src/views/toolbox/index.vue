<template>
    <div>
        <RouterButton :buttons="buttons">
            <template #route-button>
                <div class="router-button mr-0 sm:mr-7">
                    <el-button link type="primary" @click="onRestart('1panel')">
                        {{ $t('home.restart_1panel') }}
                    </el-button>
                    <el-divider direction="vertical" />
                    <el-button link type="primary" @click="onRestart('system')">
                        {{ $t('home.restart_system') }}
                    </el-button>
                </div>
            </template>
        </RouterButton>
        <LayoutContent>
            <router-view></router-view>
        </LayoutContent>

        <ConfirmDialog ref="confirmDialogRef" @confirm="onSave"></ConfirmDialog>
    </div>
</template>

<script lang="ts" setup>
import i18n from '@/lang';
import { ref } from 'vue';
import ConfirmDialog from '@/components/confirm-dialog/index.vue';
import { GlobalStore } from '@/store';
import { MsgSuccess } from '@/utils/message';
import { systemRestart } from '@/api/modules/dashboard';

const restartType = ref();
const confirmDialogRef = ref();
const globalStore = GlobalStore();

const onRestart = (type: string) => {
    restartType.value = type;
    let params = {
        header: i18n.global.t('home.restart_' + type),
        operationInfo: '',
        submitInputInfo: i18n.global.t('database.restartNow'),
    };
    confirmDialogRef.value!.acceptParams(params);
};
const onSave = async () => {
    globalStore.isOnRestart = true;
    MsgSuccess(i18n.global.t('home.operationSuccess'));
    await systemRestart(restartType.value);
};

const buttons = [
    {
        label: i18n.global.t('toolbox.device.toolbox'),
        path: '/toolbox/device',
    },
    {
        label: i18n.global.t('setting.diskClean'),
        path: '/toolbox/clean',
    },
    {
        label: i18n.global.t('menu.supervisor'),
        path: '/toolbox/supervisor',
    },
    {
        label: i18n.global.t('toolbox.clam.clam'),
        path: '/toolbox/clam',
    },
    {
        label: 'FTP',
        path: '/toolbox/ftp',
    },
    {
        label: 'Fail2ban',
        path: '/toolbox/fail2ban',
    },
];
</script>
