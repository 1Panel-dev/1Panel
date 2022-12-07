<template>
    <LayoutContent :header="$t('nginx.nginxConfig')" :reload="true">
        <el-collapse v-model="activeName" accordion>
            <el-collapse-item :title="$t('nginx.configResource')" name="1">
                <Source v-if="activeName === '1'"></Source>
            </el-collapse-item>
            <el-collapse-item :title="$t('nginx.status')" name="2">
                <Status v-if="activeName === '2'" :status="status" />
            </el-collapse-item>
            <el-collapse-item :title="$t('website.nginxPer')" name="3">
                <NginxPer v-if="activeName === '3'" />
            </el-collapse-item>
            <el-collapse-item :title="$t('website.log')" name="4">
                <ContainerLog ref="dialogContainerLogRef" />
            </el-collapse-item>
        </el-collapse>
    </LayoutContent>
</template>

<script lang="ts" setup>
import LayoutContent from '@/layout/layout-content.vue';
import Source from './source/index.vue';
import { ref, watch } from 'vue';
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

watch(
    activeName,
    (newvalue) => {
        if (newvalue === '4') {
            dialogContainerLogRef.value!.acceptParams({
                containerID: props.containerName,
                container: props.containerName,
            });
        }
    },
    { immediate: true },
);
</script>
