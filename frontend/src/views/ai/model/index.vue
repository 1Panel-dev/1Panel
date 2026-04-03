<template>
    <div>
        <RouterButton :buttons="buttons" />
        <component :is="currentComponent" />
    </div>
</template>

<script lang="ts" setup>
import { computed } from 'vue';
import { useRoute } from 'vue-router';
import i18n from '@/lang';
import AccountView from '@/views/ai/agents/model/index.vue';
import LocalView from '@/views/ai/model/local/index.vue';

const route = useRoute();

const buttons = [
    {
        label: i18n.global.t('aiTools.agents.account'),
        path: '/ai/model/account',
    },
    {
        label: i18n.global.t('aiTools.model.localModel'),
        path: '/ai/model/local',
    },
];

const currentComponent = computed(() => {
    return route.path === '/ai/model/local' ? LocalView : AccountView;
});
</script>
