<template>
    <div :style="{ '--main-height': mainHeight + 'px' }">
        <el-tabs tab-position="left" v-model="tabIndex" v-if="id > 0" class="custom-tabs" ref="tabsRef">
            <el-tab-pane :label="$t('website.domainConfig')" name="0">
                <Domain :key="id" :id="id" v-if="tabIndex == '0'"></Domain>
            </el-tab-pane>
            <el-tab-pane :label="$t('website.sitePath')" name="1">
                <SitePath :id="id" v-if="tabIndex == '1'"></SitePath>
            </el-tab-pane>
            <el-tab-pane :label="$t('website.defaultDoc')" name="2">
                <Default :id="id" v-if="tabIndex == '2'"></Default>
            </el-tab-pane>
            <el-tab-pane :label="$t('website.rate')" name="3">
                <LimitConn :id="id" v-if="tabIndex == '3'"></LimitConn>
            </el-tab-pane>
            <el-tab-pane :label="$t('website.proxy')" name="4">
                <Proxy :id="id" v-if="tabIndex == '4'"></Proxy>
            </el-tab-pane>
            <el-tab-pane :label="$t('website.loadBalance')" name="5">
                <LoadBalance :id="id" v-if="tabIndex == '5'"></LoadBalance>
            </el-tab-pane>
            <el-tab-pane :label="$t('website.basicAuth')" name="6">
                <AuthBasic :id="id" v-if="tabIndex == '6'"></AuthBasic>
            </el-tab-pane>
            <el-tab-pane :label="$t('website.cors')" name="16">
                <Cors :id="id" v-if="tabIndex == '16'"></Cors>
            </el-tab-pane>
            <el-tab-pane :label="'HTTPS'" name="7">
                <HTTPS :id="id" v-if="tabIndex == '7'"></HTTPS>
            </el-tab-pane>
            <el-tab-pane :label="$t('website.realIP')" name="8">
                <RealIP :id="id" v-if="tabIndex == '8'"></RealIP>
            </el-tab-pane>
            <el-tab-pane :label="$t('website.rewrite')" name="9">
                <Rewrite :id="id" v-if="tabIndex == '9'"></Rewrite>
            </el-tab-pane>
            <el-tab-pane :label="$t('website.antiLeech')" name="10">
                <AntiLeech :id="id" v-if="tabIndex == '10'"></AntiLeech>
            </el-tab-pane>
            <el-tab-pane :label="$t('website.redirect')" name="11">
                <Redirect :id="id" v-if="tabIndex == '11'"></Redirect>
            </el-tab-pane>

            <el-tab-pane
                :label="'PHP'"
                name="13"
                v-if="(website.type === 'runtime' && website.runtimeType === 'php') || website.type === 'static'"
            >
                <PHP :id="id" v-if="tabIndex == '13'"></PHP>
            </el-tab-pane>
            <el-tab-pane
                :label="$t('logs.resource')"
                name="14"
                v-if="website.type === 'runtime' || website.type === 'static'"
            >
                <Resource :id="id" v-if="tabIndex == '14'"></Resource>
            </el-tab-pane>
            <el-tab-pane :label="$t('website.other')" name="12">
                <Other :id="id" v-if="tabIndex == '12'"></Other>
            </el-tab-pane>
        </el-tabs>
    </div>
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
import Cors from './cors/index.vue';

const props = defineProps({
    website: {
        type: Object,
    },
    heightDiff: {
        type: Number,
        default: 0,
    },
});
const windowHeight = ref(window.innerHeight);
const mainHeight = computed(() => windowHeight.value - props.heightDiff);

const id = computed(() => {
    return props.website.id;
});
const tabIndex = ref('0');

watch(tabIndex, (newVal) => {
    localStorage.setItem('site-tabIndex', newVal);
});

const handleResize = () => {
    windowHeight.value = window.innerHeight;
};

const tabsRef = ref();
const handleScroll = (event: WheelEvent) => {
    if (!tabsRef.value) return;
    const tabContainer = tabsRef.value.$el.querySelector('.el-tabs__nav-scroll');
    if (!tabContainer) return;

    const target = event.target as HTMLElement;
    if (!target.classList.contains('el-tabs__item')) {
        return;
    }

    event.preventDefault();
    tabContainer.scrollTop += event.deltaY;
};

onMounted(() => {
    const storedTabIndex = localStorage.getItem('site-tabIndex');
    if (storedTabIndex !== null) {
        tabIndex.value = storedTabIndex;
    }
    window.addEventListener('resize', handleResize);
    document.addEventListener('wheel', handleScroll, { passive: false });
});

onUnmounted(() => {
    window.removeEventListener('resize', handleResize);
    document.removeEventListener('wheel', handleScroll);
});
</script>
<style scoped>
.custom-tabs {
    height: var(--main-height);
    display: flex;
    overflow: hidden;
}

.custom-tabs :deep(.el-tabs__header.is-left) {
    height: 100%;
    overflow-y: auto;
    overflow-x: hidden;
    flex-shrink: 0;
}

.custom-tabs :deep(.el-tabs__content) {
    height: 100%;
    overflow-y: auto;
    overflow-x: hidden;
    flex-grow: 1;
}
</style>
