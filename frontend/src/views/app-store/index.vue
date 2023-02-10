<template>
    <div>
        <div v-if="showButton">
            <RouterButton :buttons="buttons" />
        </div>
        <LayoutContent>
            <router-view></router-view>
        </LayoutContent>
    </div>
</template>

<script lang="ts" setup>
import LayoutContent from '@/layout/layout-content.vue';
import RouterButton from '@/components/router-button/index.vue';
import i18n from '@/lang';
import { onMounted, ref } from 'vue';
import { SearchAppInstalled } from '@/api/modules/app';
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
        label: i18n.global.t('app.canUpdate'),
        path: '/apps/update',
        count: 0,
    },
];
onMounted(() => {
    SearchAppInstalled({ update: true, page: 1, pageSize: 100 }).then((res) => {
        if (res.data.items) {
            buttons[2].count = res.data.items.length;
        }
        showButton.value = true;
    });
});
</script>
