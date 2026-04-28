<template>
    <div>
        <RouterButton :buttons="buttons" />
        <component :is="currentComponent" />
    </div>
</template>

<script lang="ts" setup>
import { computed, defineAsyncComponent } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import i18n from '@/lang';
import AccountView from '@/views/ai/agents/model/index.vue';
import LocalView from '@/views/ai/model/local/index.vue';
import { loadOptionalComponent } from '@/extensions/optional';

const AIProxyView = defineAsyncComponent(() => loadOptionalComponent('/src/enterprise/views/ai-proxy/index.vue'));

const route = useRoute();
const router = useRouter();

const buttons = computed<RouterButton[]>(() => {
    const items: RouterButton[] = [
        {
            label: i18n.global.t('aiTools.agents.account'),
            path: '/ai/model/account',
        },
        {
            label: i18n.global.t('aiTools.model.localModel'),
            path: '/ai/model/local',
        },
    ];
    if (router.hasRoute('AIProxyManagement')) {
        items.push({
            label: i18n.global.t('aiTools.aiProxy.title'),
            path: '/ai/model/ai-proxy',
        });
    }
    return items;
});

const currentComponent = computed(() => {
    if (route.path === '/ai/model/local') {
        return LocalView;
    }
    if (route.path === '/ai/model/ai-proxy') {
        return AIProxyView;
    }
    return AccountView;
});
</script>
