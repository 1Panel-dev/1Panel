<template>
    <LayoutContent :header="'网站设置'" :back-name="'Website'">
        <el-tabs type="card">
            <el-tab-pane label="基本">
                <Basic :id="id"></Basic>
            </el-tab-pane>
            <el-tab-pane label="优化">优化</el-tab-pane>
            <el-tab-pane label="反代">反代</el-tab-pane>
            <el-tab-pane label="安全">反代</el-tab-pane>
            <el-tab-pane label="备份">反代</el-tab-pane>
            <el-tab-pane label="日志">反代</el-tab-pane>
            <el-tab-pane label="源文">反代</el-tab-pane>
        </el-tabs>
    </LayoutContent>
</template>

<script setup lang="ts">
import LayoutContent from '@/layout/layout-content.vue';
import { computed } from 'vue';
import Basic from './basic/index.vue';

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
