<template>
    <div>
        <RouterButton
            :buttons="[
                {
                    label: $t('menu.website'),
                    path: '/websites',
                },
            ]"
        />
        <LayoutContent :title="$t('website.websiteConfig')" :back-name="'Website'" v-loading="loading">
            <template #app>
                <WebsiteStatus
                    v-if="website"
                    :id="id"
                    @select="selectWebsite"
                    :primary-domain="website.primaryDomain"
                    :status="website.status"
                    :expire-date="website.expireDate"
                />
            </template>
            <template #leftToolBar>
                <el-button type="primary" :plain="props.tab !== 'basic'" @click="changeTab('basic')">
                    {{ $t('website.basic') }}
                </el-button>
                <el-button type="primary" :plain="props.tab !== 'log'" @click="changeTab('log')">
                    {{ $t('commons.button.log') }}
                </el-button>
                <el-button type="primary" :plain="props.tab !== 'resource'" @click="changeTab('resource')">
                    {{ $t('website.source', 2) }}
                </el-button>
            </template>
            <template #main>
                <MainDiv v-if="website" :heightDiff="260">
                    <Basic :website="website" v-if="props.tab === 'basic'" :heightDiff="320"></Basic>
                    <Log :id="id" v-if="props.tab === 'log'"></Log>
                    <Resource :id="id" v-if="props.tab === 'resource'"></Resource>
                </MainDiv>
            </template>
        </LayoutContent>
    </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import Basic from './basic/index.vue';
import Resource from './resource/index.vue';
import Log from './log/index.vue';
import WebsiteStatus from '@/views/website/website/status/index.vue';
import { getWebsite } from '@/api/modules/website';
import type { Website } from '@/api/interface/website';
import { routerToNameWithParams } from '@/utils/router';
import MainDiv from '@/components/main-div/index.vue';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { useOperateNodeContext } from '@/composables/useOperateNodeContext';

const { currentNode } = useGlobalStore();
useOperateNodeContext(currentNode);

const props = defineProps({
    id: {
        type: String,
        default: '0',
    },
    tab: {
        type: String,
        default: 'basic',
    },
});

const id = Number(props.id);
const website = ref<Website.WebsiteDTO>();
const loading = ref(true);

const changeTab = (tab: string) => routerToNameWithParams('WebsiteConfig', { id, tab });
const selectWebsite = (id: number) => routerToNameWithParams('WebsiteConfig', { id, tab: props.tab });

onMounted(async () => {
    try {
        const { data } = await getWebsite(id);
        website.value = data;
    } finally {
        loading.value = false;
    }
});
</script>
