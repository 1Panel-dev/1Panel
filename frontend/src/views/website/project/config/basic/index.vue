<template>
    <el-tabs tab-position="left" type="border-card">
        <el-tab-pane label="域名设置">
            <Doamin :id="id"></Doamin>
        </el-tab-pane>
        <el-tab-pane label="目录设置">Config</el-tab-pane>
        <el-tab-pane label="HTTPS">Role</el-tab-pane>
        <el-tab-pane label="域名跳转">Task</el-tab-pane>
        <el-tab-pane label="错误页面">Task</el-tab-pane>
        <el-tab-pane label="过期时间">Task</el-tab-pane>
        <el-tab-pane label="流量限制">Task</el-tab-pane>
        <el-tab-pane label="默认文档">Task</el-tab-pane>
    </el-tabs>
</template>

<script lang="ts" setup name="Basic">
import { computed } from 'vue';

import Doamin from './domain/index.vue';

const props = defineProps({
    id: {
        type: Number,
        default: 0,
    },
});

const id = computed(() => {
    return props.id;
});
</script>
