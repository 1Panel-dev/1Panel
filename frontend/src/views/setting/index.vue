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
import { hasPermission } from '@/utils/rbac';
const { isOffLine, isFxplay, isAdmin, isXpackEE } = useGlobalStore();

const buttons = computed(() => {
    const items = [
        ...(isAdmin.value || hasPermission('setting_view')
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
        ...(isOffLine.value || !(isAdmin.value || hasPermission('setting_view'))
            ? []
            : [
                  {
                      label: i18n.global.t('setting.license'),
                      path: isXpackEE.value ? '/xpack-ee/license' : '/settings/license',
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
