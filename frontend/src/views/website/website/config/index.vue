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
                    v-if="website.id > 0"
                    :primary-domain="website.primaryDomain"
                    :status="website.status"
                    :expire-date="website.expireDate"
                />
            </template>
            <template #leftToolBar>
                <el-button type="primary" :plain="index !== 'basic'" @click="changeTab('basic')">
                    {{ $t('website.basic') }}
                </el-button>
                <el-button type="primary" :plain="index !== 'log'" @click="changeTab('log')">
                    {{ $t('commons.button.log') }}
                </el-button>
                <el-button type="primary" :plain="index !== 'resource'" @click="changeTab('resource')">
                    {{ $t('website.source', 2) }}
                </el-button>
            </template>
            <template #main>
                <MainDiv :heightDiff="260">
                    <Basic :website="website" v-if="index === 'basic'" :heightDiff="320"></Basic>
                    <Log :id="id" v-if="index === 'log'"></Log>
                    <Resource :id="id" v-if="index === 'resource'"></Resource>
                </MainDiv>
            </template>
        </LayoutContent>
    </div>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from 'vue';
import Basic from './basic/index.vue';
import Resource from './resource/index.vue';
import Log from './log/index.vue';
import WebsiteStatus from '@/views/website/website/status/index.vue';
import { getWebsite } from '@/api/modules/website';
import { GetRuntime } from '@/api/modules/runtime';
import { routerToNameWithParams } from '@/utils/router';

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

const id = ref(0);
const index = ref('basic');
const website = ref<any>({});
const loading = ref(false);
const configPHP = ref(false);

watch(index, (curr, old) => {
    if (curr != old) {
        changeTab(curr);
    }
});

const changeTab = (index: string) => {
    routerToNameWithParams('WebsiteConfig', { id: id.value, tab: index });
};

onMounted(() => {
    index.value = props.tab;
    id.value = Number(props.id);
    loading.value = true;
    getWebsite(id.value)
        .then(async (res) => {
            website.value = res.data;
            if (res.data.type === 'runtime') {
                const runRes = await GetRuntime(res.data.runtimeID);
                if (runRes.data.type == 'php' && runRes.data.resource === 'appstore') {
                    configPHP.value = true;
                }
            }
        })
        .finally(() => {
            loading.value = false;
        });
});
</script>
