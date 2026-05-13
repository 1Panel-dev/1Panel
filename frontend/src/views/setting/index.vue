<template>
    <div>
        <RouterButton :buttons="buttons" />
        <LayoutContent>
            <router-view></router-view>
        </LayoutContent>
    </div>
</template>

<script lang="ts" setup>
import { computed } from 'vue';
import i18n from '@/lang';
import { useGlobalStore } from '@/composables/useGlobalStore';
const { isOffline, isFxplay, isAdmin, isEnterprise } = useGlobalStore();

const buttons = computed(() => {
    const items = [
        ...(isAdmin.value
            ? [
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
              ]
            : []),
        ...((isOffline.value && !isEnterprise.value) || (isEnterprise.value && !isAdmin.value)
            ? []
            : [
                  {
                      label: i18n.global.t('setting.license'),
                      path: isEnterprise.value ? '/enterprise/license' : '/settings/license',
                  },
              ]),
        ...(isFxplay.value
            ? []
            : [
                  {
                      label: i18n.global.t('setting.about'),
                      path: '/settings/about',
                  },
              ]),
    ];
    return items;
});
</script>
