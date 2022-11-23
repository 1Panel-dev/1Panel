<template>
    <LayoutContent :header="$t('website.nginxConfig')" :reload="true">
        <el-collapse v-model="activeName" accordion>
            <el-collapse-item :title="$t('website.source')" name="1">
                <Source></Source>
            </el-collapse-item>
            <el-collapse-item :title="$t('website.log')" name="2">
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
            dialogContainerLogRef.value!.acceptParams({ containerID: props.containerName });
        }
    },
    { immediate: true },
);
</script>
