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

        <el-tab-pane label="HTTPS">Role</el-tab-pane>
        <el-tab-pane label="域名跳转">Task</el-tab-pane>
        <el-tab-pane label="错误页面">Task</el-tab-pane>
        <el-tab-pane label="过期时间">Task</el-tab-pane>
        <el-tab-pane label="默认文档">Task</el-tab-pane>
    </el-tabs>
</template>

<script lang="ts" setup name="Basic">
import { computed, ref } from 'vue';

import Doamin from './domain/index.vue';
import Default from './default-doc/index.vue';
import LimitConn from './limit-conn/index.vue';

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
