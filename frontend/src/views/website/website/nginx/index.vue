<template>
    <LayoutContent :title="$t('commons.button.set')" :reload="true">
        <template #leftToolBar>
            <el-button
                type="primary"
                :plain="activeName !== '1'"
                @click="changeTab('1')"
                :disabled="status != 'Running'"
            >
                {{ $t('nginx.status') }}
            </el-button>
            <el-button type="primary" :plain="activeName !== '2'" @click="changeTab('2')">
                {{ $t('nginx.configResource') }}
            </el-button>
            <el-button type="primary" :plain="activeName !== '3'" @click="changeTab('3')">
                {{ $t('website.nginxPer') }}
            </el-button>
            <el-button
                type="primary"
                :plain="activeName !== '4'"
                @click="changeTab('4')"
                :disabled="status != 'Running'"
            >
                {{ $t('commons.button.log') }}
            </el-button>
            <el-button type="primary" :plain="activeName !== '5'" @click="changeTab('5')">
                {{ $t('runtime.module') }}
            </el-button>
            <el-button type="primary" :plain="activeName !== '6'" @click="changeTab('6')">
                {{ $t('website.other') }}
            </el-button>
        </template>
        <template #main>
            <Status v-if="activeName === '1'" :status="status" />
            <Source v-if="activeName === '2'" />
            <NginxPer v-if="activeName === '3'" />
            <ContainerLog v-if="activeName === '4'" :container="containerName" :highlightDiff="350" />
            <Module v-if="activeName === '5'" />
            <Other v-if="activeName === '6'" />
        </template>
    </LayoutContent>
</template>

<script lang="ts" setup>
import Source from './source/index.vue';
import { ref } from 'vue';
import ContainerLog from '@/components/log/container/index.vue';
import NginxPer from './performance/index.vue';
import Status from './status/index.vue';
import Module from './module/index.vue';
import Other from './other/index.vue';

const activeName = ref('1');

const props = defineProps({
    containerName: {
        type: String,
        default: '',
    },
    status: {
        type: String,
        default: 'Running',
    },
});
const changeTab = (index: string) => {
    activeName.value = index;
};

onMounted(() => {
    if (props.status != 'Running') {
        activeName.value = '2';
    }
});
</script>
