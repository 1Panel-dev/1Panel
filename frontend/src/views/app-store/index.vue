<template>
    <div>
        <div v-if="showButton">
            <RouterButton :buttons="buttons">
                <template #route-button>
                    <div class="pr-5">
                        <el-badge is-dot :hidden="!canUpdate">
                            <el-button @click="sync" type="primary" plain :disabled="syncing">
                                {{ $t('app.syncAppList') }}
                            </el-button>
                        </el-badge>
                    </div>
                </template>
            </RouterButton>
        </div>
        <LayoutContent>
            <router-view></router-view>
        </LayoutContent>
    </div>
</template>

<script lang="ts" setup>
import i18n from '@/lang';
import { onMounted, ref } from 'vue';
import { SearchAppInstalled, SyncApp } from '@/api/modules/app';
import bus from './bus';
import { MsgSuccess } from '@/utils/message';
let showButton = ref(false);
const syncing = ref(false);
const canUpdate = ref(false);

const buttons = [
    {
        label: i18n.global.t('app.all'),
        path: '/apps/all',
    },
    {
        label: i18n.global.t('app.installed'),
        path: '/apps/installed',
    },
    {
        label: i18n.global.t('app.canUpgrade'),
        path: '/apps/upgrade',
        count: 0,
    },
];

const sync = () => {
    syncing.value = true;
    SyncApp()
        .then((res) => {
            if (res.message != '') {
                MsgSuccess(res.message);
            } else {
                MsgSuccess(i18n.global.t('app.syncStart'));
            }
            canUpdate.value = false;
        })
        .finally(() => {
            syncing.value = false;
        });
};

const search = () => {
    SearchAppInstalled({ update: true, page: 1, pageSize: 100 })
        .then((res) => {
            if (res.data.items) {
                buttons[2].count = res.data.items.length;
            } else {
                buttons[2].count = 0;
            }
        })
        .finally(() => {
            showButton.value = true;
        });
};

onMounted(() => {
    search();
    bus.on('upgrade', () => {
        showButton.value = false;
        search();
    });
});
</script>
