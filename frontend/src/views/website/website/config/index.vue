<template>
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
            <el-button type="primary" :plain="index !== 'safety'" @click="changeTab('safety')">
                {{ $t('website.security') }}
            </el-button>
            <el-button type="primary" :plain="index !== 'log'" @click="changeTab('log')">
                {{ $t('website.log') }}
            </el-button>
            <el-button type="primary" :plain="index !== 'resource'" @click="changeTab('resource')">
                {{ $t('website.source') }}
            </el-button>
        </template>
        <template #main>
            <Basic :id="id" v-if="index === 'basic'"></Basic>
            <Safety :id="id" v-if="index === 'safety'"></Safety>
            <Log :id="id" v-if="index === 'log'"></Log>
            <Resource :id="id" v-if="index === 'resource'"></Resource>
        </template>
    </LayoutContent>
</template>

<script setup lang="ts">
import LayoutContent from '@/layout/layout-content.vue';
import { onMounted, ref, watch } from 'vue';
import Basic from './basic/index.vue';
import Safety from './safety/index.vue';
import Resource from './resource/index.vue';
import Log from './log/index.vue';
import router from '@/routers';
import WebsiteStatus from '@/views/website/website/status/index.vue';
import { GetWebsite } from '@/api/modules/website';

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
        .then((res) => {
            website.value = res.data;
        })
        .finally(() => {
            loading.value = false;
        });
});
</script>
