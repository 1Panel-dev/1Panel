<template>
    <LayoutContent :header="$t('website.nginxConfig')" :reload="true">
        <el-collapse v-model="activeName" accordion>
            <el-collapse-item :title="$t('website.source')" name="1">
                <Source></Source>
            </el-collapse-item>
            <el-collapse-item :title="$t('website.log')" name="2">
                <ContainerLog ref="dialogContainerLogRef" />
            </el-collapse-item>
            <el-collapse-item :title="$t('website.nginxPer')" name="3">
                <NginxPer />
            </el-collapse-item>
            <el-collapse-item :title="$t('nginx.status')" name="4">
                <Status />
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
});

watch(
    activeName,
    (newvalue) => {
        if (newvalue === '2') {
            dialogContainerLogRef.value!.acceptParams({
                containerID: props.containerName,
                container: props.containerName,
            });
        }
    },
    { immediate: true },
);
</script>
