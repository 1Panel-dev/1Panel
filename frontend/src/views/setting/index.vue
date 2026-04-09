<template>
    <div>
        <RouterButton :buttons="buttons" />
        <LayoutContent>
            <router-view></router-view>
        </LayoutContent>
    </div>
</template>

<script lang="ts" setup>
import i18n from '@/lang';
import { useGlobalStore } from '@/composables/useGlobalStore';
const { isOffLine, isFxplay, isAdmin, isXpackEE } = useGlobalStore();

const buttons = [
    {
        label: i18n.global.t('setting.panel'),
        path: '/settings/panel',
    },
    {
        label: i18n.global.t('setting.safe'),
        path: '/settings/safe',
    },
    {
        label: i18n.global.t('xpack.alert.alertNotice'),
        path: '/settings/alert',
    },
    {
        label: i18n.global.t('setting.backupAccount', 2),
        path: '/settings/backupaccount',
    },
    {
        label: i18n.global.t('setting.snapshot', 2),
        path: '/settings/snapshot',
    },
    {
        label: i18n.global.t('setting.license'),
        path: isXpackEE.value ? '/xpack-ee/license' : '/settings/license',
    },
    {
        label: i18n.global.t('setting.about'),
        path: '/settings/about',
    },
];

onMounted(() => {
    if (isOffLine.value) {
        buttons.splice(5, 1);
    }
    if (isFxplay.value) {
        buttons.splice(6, 1);
    }
    if (isXpackEE.value && !isAdmin.value) {
        buttons.splice(0, 1);
        buttons.splice(0, 1);
        buttons.splice(3, 1);
    }
});
</script>
