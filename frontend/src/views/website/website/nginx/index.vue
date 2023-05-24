<template>
    <LayoutContent :title="$t('nginx.nginxConfig')" :reload="true">
        <template #buttons>
            <el-button type="primary" :plain="activeName !== '1'" @click="changeTab('1')">
                {{ $t('nginx.status') }}
            </el-button>
            <el-button type="primary" :plain="activeName !== '2'" @click="changeTab('2')">
                {{ $t('nginx.configResource') }}
            </el-button>
            <el-button type="primary" :plain="activeName !== '3'" @click="changeTab('3')">
                {{ $t('website.nginxPer') }}
            </el-button>
            <el-button type="primary" :plain="activeName !== '4'" @click="changeTab('4')">
                {{ $t('website.log') }}
            </el-button>
        </template>
        <template #main>
            <Status v-if="activeName === '1'" :status="status" />
            <Source v-if="activeName === '2'"></Source>
            <NginxPer v-if="activeName === '3'" />
            <ContainerLog v-if="activeName === '4'" ref="dialogContainerLogRef" />
        </template>
    </LayoutContent>
</template>

<script lang="ts" setup>
import Source from './source/index.vue';
import { nextTick, ref } from 'vue';
import ContainerLog from '@/components/container-log/index.vue';
import NginxPer from './performance/index.vue';
import Status from './status/index.vue';

let activeName = ref('1');
let dialogContainerLogRef = ref();

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

    if (index === '4') {
        nextTick(() => {
            dialogContainerLogRef.value!.acceptParams({
                containerID: props.containerName,
                container: props.containerName,
            });
        });
    }
};
</script>
