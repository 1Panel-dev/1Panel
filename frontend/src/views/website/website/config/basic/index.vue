<template>
    <el-tabs tab-position="left" type="border-card" v-model="tabIndex">
        <el-tab-pane :label="$t('website.domainConfig')">
            <Doamin :id="id" v-if="tabIndex == '0'"></Doamin>
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
        <el-tab-pane :label="'HTTPS'">
            <HTTPS :id="id" v-if="tabIndex == '4'"></HTTPS>
        </el-tab-pane>
        <el-tab-pane :label="$t('website.other')">
            <Other :id="id" v-if="tabIndex == '5'"></Other>
        </el-tab-pane>
    </el-tabs>
</template>

<script lang="ts" setup name="Basic">
import { computed, ref } from 'vue';

import Doamin from './domain/index.vue';
import Default from './default-doc/index.vue';
import LimitConn from './limit-conn/index.vue';
import Other from './other/index.vue';
import HTTPS from './https/index.vue';
import SitePath from './site-folder/index.vue';

const props = defineProps({
    id: {
        type: Number,
        default: 0,
    },
});

const id = computed(() => {
    return props.id;
});

let tabIndex = ref('0');
</script>
