<template>
    <el-tabs tab-position="left" type="border-card" v-model="index">
        <el-tab-pane :label="$t('website.domainConfig')">
            <Doamin :id="id" v-if="index == '0'"></Doamin>
        </el-tab-pane>
        <el-tab-pane :label="$t('website.defaultDoc')">
            <Default :id="id" v-if="index == '1'"></Default>
        </el-tab-pane>
        <el-tab-pane :label="$t('website.rate')">
            <LimitConn :id="id" v-if="index == '2'"></LimitConn>
        </el-tab-pane>
        <el-tab-pane :label="$t('website.other')">
            <Other :id="id" v-if="index == '3'"></Other>
        </el-tab-pane>
    </el-tabs>
</template>

<script lang="ts" setup name="Basic">
import { computed, ref } from 'vue';

import Doamin from './domain/index.vue';
import Default from './default-doc/index.vue';
import LimitConn from './limit-conn/index.vue';
import Other from './other/index.vue';

const props = defineProps({
    id: {
        type: Number,
        default: 0,
    },
});

const id = computed(() => {
    return props.id;
});

let index = ref('0');
</script>
