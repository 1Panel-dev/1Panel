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
let showButton = ref(false);

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
    search();
    bus.on('upgrade', () => {
        showButton.value = false;
        search();
    });
});
</script>
