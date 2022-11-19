<template>
    <LayoutContent :header="'网站设置'" :back-name="'Website'">
        <el-tabs v-model="index" @click="changeTab(index)">
            <el-tab-pane label="基本" name="basic">
                <Basic :key="id" :id="id" v-if="index === 'basic'"></Basic>
            </el-tab-pane>
            <el-tab-pane label="安全" name="safety">
                <Safety :key="id" :id="id" v-if="index === 'safety'"></Safety>
            </el-tab-pane>
            <el-tab-pane label="备份">反代</el-tab-pane>
            <el-tab-pane label="源文" name="resource">
                <Resource :key="id" :id="id" v-if="index === 'resource'"></Resource>
            </el-tab-pane>
        </el-tabs>
    </LayoutContent>
</template>

<script setup lang="ts">
import LayoutContent from '@/layout/layout-content.vue';
import { onMounted, ref } from 'vue';
import Basic from './basic/index.vue';
import Safety from './safety/index.vue';
import Resource from './resource/index.vue';
import router from '@/routers';

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

const changeTab = (index: string) => {
    router.replace({ name: 'WebsiteConfig', params: { id: id.value, tab: index } });
};

onMounted(() => {
    index.value = props.tab;
    id.value = Number(props.id);
});
</script>
