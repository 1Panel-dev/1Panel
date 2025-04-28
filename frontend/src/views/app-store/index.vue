<template>
    <div>
        <div v-if="showButton">
            <RouterButton :buttons="buttons"></RouterButton>
        </div>
        <LayoutContent>
            <router-view></router-view>
        </LayoutContent>
    </div>
</template>

<script lang="ts" setup>
import i18n from '@/lang';
import { onMounted, ref } from 'vue';
import { searchAppInstalled } from '@/api/modules/app';
import bus from '@/global/bus';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();
let showButton = ref(false);

let buttons = [
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

const settingButtons = [
    {
        label: i18n.global.t('commons.button.set'),
        path: '/apps/setting',
    },
];

const search = () => {
    searchAppInstalled({ update: true, page: 1, pageSize: 100 })
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
    if (globalStore.isProductPro) {
        buttons = buttons.concat(settingButtons);
    }

    search();
    bus.on('upgrade', () => {
        showButton.value = false;
        search();
    });
});
</script>
