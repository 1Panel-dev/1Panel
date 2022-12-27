<template>
    <el-card>
        <LayoutContent :header="$t('website.websiteConfig')" :back-name="'Website'">
            <el-tabs v-model="index">
                <el-tab-pane :label="$t('website.basic')" name="basic">
                    <Basic :id="id" v-if="index === 'basic'"></Basic>
                </el-tab-pane>
                <el-tab-pane :label="$t('website.security')" name="safety">
                    <Safety :id="id" v-if="index === 'safety'"></Safety>
                </el-tab-pane>
                <el-tab-pane :label="$t('website.log')" name="log">
                    <Log :id="id" v-if="index === 'log'"></Log>
                </el-tab-pane>
                <el-tab-pane :label="$t('website.source')" name="resource">
                    <Resource :id="id" v-if="index === 'resource'"></Resource>
                </el-tab-pane>
            </el-tabs>
        </LayoutContent>
    </el-card>
</template>

<script setup lang="ts">
import LayoutContent from '@/layout/layout-content.vue';
import { onMounted, ref, watch } from 'vue';
import Basic from './basic/index.vue';
import Safety from './safety/index.vue';
import Resource from './resource/index.vue';
import Log from './log/index.vue';
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
});
</script>
