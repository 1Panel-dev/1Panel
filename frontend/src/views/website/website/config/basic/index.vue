<template>
    <el-tabs tab-position="left" v-model="tabIndex" v-if="id > 0">
        <el-tab-pane :label="$t('website.domainConfig')">
            <Domain :key="id" :id="id" v-if="tabIndex == '0'"></Domain>
        </el-tab-pane>
        <el-tab-pane :label="$t('website.sitePath')">
            <SitePath :id="id" v-if="tabIndex == '1'"></SitePath>
        </el-tab-pane>
        <el-tab-pane :label="$t('website.defaultDoc')">
            <Default :id="id" v-if="tabIndex == '2'"></Default>
        </el-tab-pane>
        <el-tab-pane :label="$t('website.rate')">
            <LimitConn :id="id" v-if="tabIndex == '3'"></LimitConn>
        </el-tab-pane>
        <el-tab-pane :label="$t('website.proxy')">
            <Proxy :id="id" v-if="tabIndex == '4'"></Proxy>
        </el-tab-pane>
        <el-tab-pane :label="$t('website.loadBalance')">
            <LoadBalance :id="id" v-if="tabIndex == '5'"></LoadBalance>
        </el-tab-pane>
        <el-tab-pane :label="$t('website.basicAuth')">
            <AuthBasic :id="id" v-if="tabIndex == '6'"></AuthBasic>
        </el-tab-pane>
        <el-tab-pane :label="'HTTPS'">
            <HTTPS :id="id" v-if="tabIndex == '7'"></HTTPS>
        </el-tab-pane>
        <el-tab-pane :label="$t('website.realIP')">
            <RealIP :id="id" v-if="tabIndex == '8'"></RealIP>
        </el-tab-pane>
        <el-tab-pane :label="$t('website.rewrite')">
            <Rewrite :id="id" v-if="tabIndex == '9'"></Rewrite>
        </el-tab-pane>
        <el-tab-pane :label="$t('website.antiLeech')">
            <AntiLeech :id="id" v-if="tabIndex == '10'"></AntiLeech>
        </el-tab-pane>
        <el-tab-pane :label="$t('website.redirect')">
            <Redirect :id="id" v-if="tabIndex == '11'"></Redirect>
        </el-tab-pane>

        <el-tab-pane
            :label="'PHP'"
            name="13"
            v-if="(website.type === 'runtime' && website.runtimeType === 'php') || website.type === 'static'"
        >
            <PHP :id="id" v-if="tabIndex == '13'"></PHP>
        </el-tab-pane>
        <el-tab-pane :label="$t('logs.resource')" name="14">
            <Resource :id="id" v-if="tabIndex == '14'"></Resource>
        </el-tab-pane>
        <el-tab-pane :label="$t('website.other')" name="12">
            <Other :id="id" v-if="tabIndex == '12'"></Other>
        </el-tab-pane>
    </el-tabs>
</template>

<script lang="ts" setup name="Basic">
import { computed, onMounted, ref, watch } from 'vue';

import Domain from './domain/index.vue';
import Default from './default-doc/index.vue';
import LimitConn from './limit-conn/index.vue';
import Other from './other/index.vue';
import HTTPS from './https/index.vue';
import SitePath from './site-folder/index.vue';
import Rewrite from './rewrite/index.vue';
import Proxy from './proxy/index.vue';
import AuthBasic from './auth-basic/index.vue';
import AntiLeech from './anti-Leech/index.vue';
import Redirect from './redirect/index.vue';
import LoadBalance from './load-balance/index.vue';
import PHP from './php/index.vue';
import RealIP from './real-ip/index.vue';
import Resource from './resource/index.vue';

const props = defineProps({
    website: {
        type: Object,
    },
});
const id = computed(() => {
    return props.website.id;
});
const tabIndex = ref('0');

watch(tabIndex, (newVal) => {
    localStorage.setItem('site-tabIndex', newVal);
});

onMounted(() => {
    const storedTabIndex = localStorage.getItem('site-tabIndex');
    if (storedTabIndex !== null) {
        tabIndex.value = storedTabIndex;
    }
});
</script>
