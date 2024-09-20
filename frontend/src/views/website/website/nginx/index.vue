<template>
    <LayoutContent :title="$t('nginx.nginxConfig')" :reload="true">
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
                {{ $t('website.log') }}
            </el-button>
            <el-button type="primary" :plain="activeName !== '5'" @click="changeTab('5')">
                {{ $t('runtime.module') }}
            </el-button>
        </template>
        <template #main>
            <Status v-if="activeName === '1'" :status="status" />
            <Source v-if="activeName === '2'" />
            <NginxPer v-if="activeName === '3'" />
            <ContainerLog v-if="activeName === '4'" ref="dialogContainerLogRef" />
            <Module v-if="activeName === '5'" />
        </template>
    </LayoutContent>
</template>

<script lang="ts" setup>
import Source from './source/index.vue';
import { nextTick, ref } from 'vue';
import ContainerLog from '@/components/container-log/index.vue';
import NginxPer from './performance/index.vue';
import Status from './status/index.vue';
import Module from './module/index.vue';

const activeName = ref('1');
const dialogContainerLogRef = ref();

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

onMounted(() => {
    if (props.status != 'Running') {
        activeName.value = '2';
    }
});
</script>
