<template>
    <div>
        <RouterButton
            :buttons="[
                {
                    label: $t('website.website'),
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
            <template #buttons>
                <el-button type="primary" :plain="index !== 'basic'" @click="changeTab('basic')">
                    {{ $t('website.basic') }}
                </el-button>
                <el-button type="primary" :plain="index !== 'log'" @click="changeTab('log')">
                    {{ $t('website.log') }}
                </el-button>
                <el-button type="primary" :plain="index !== 'resource'" @click="changeTab('resource')">
                    {{ $t('website.source', 2) }}
                </el-button>
                <el-button type="primary" v-if="configPHP" :plain="index !== 'php'" @click="changeTab('php')">
                    PHP
                </el-button>
            </template>
            <template #main>
                <Basic :id="id" v-if="index === 'basic'"></Basic>
                <Log :id="id" v-if="index === 'log'"></Log>
                <Resource :id="id" v-if="index === 'resource'"></Resource>
                <PHP :id="id" v-if="index === 'php'"></PHP>
            </template>
        </LayoutContent>
    </div>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from 'vue';
import Basic from './basic/index.vue';
import Resource from './resource/index.vue';
import Log from './log/index.vue';
import PHP from './php/index.vue';
import router from '@/routers';
import WebsiteStatus from '@/views/website/website/status/index.vue';
import { GetWebsite } from '@/api/modules/website';
import { GetRuntime } from '@/api/modules/runtime';

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

let id = ref(0);
let index = ref('basic');
let website = ref<any>({});
let loading = ref(false);
const configPHP = ref(false);

watch(index, (curr, old) => {
    if (curr != old) {
        changeTab(curr);
    }
});

const changeTab = (index: string) => {
    router.push({ name: 'WebsiteConfig', params: { id: id.value, tab: index } });
};

onMounted(() => {
    index.value = props.tab;
    id.value = Number(props.id);
    loading.value = true;
    GetWebsite(id.value)
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
