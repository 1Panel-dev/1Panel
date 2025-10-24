<template>
    <el-tabs type="border-card">
        <el-tab-pane :label="$t('commons.table.port')">
            <PortConfig :exposedPorts="runtime.exposedPorts" />
        </el-tab-pane>
        <el-tab-pane :label="$t('runtime.environment')">
            <Environment :environments="runtime.environments" />
        </el-tab-pane>
        <el-tab-pane :label="$t('container.mount')"><Volumes :volumes="runtime.volumes" /></el-tab-pane>
        <el-tab-pane :label="$t('runtime.extraHosts')">
            <ExtraHosts :extraHosts="runtime.extraHosts" />
        </el-tab-pane>
    </el-tabs>
</template>

<script lang="ts" setup>
import PortConfig from '@/views/website/runtime/components/port/index.vue';
import Environment from '@/views/website/runtime/components/environment/index.vue';
import Volumes from '@/views/website/runtime/components/volume/index.vue';
import ExtraHosts from '@/views/website/runtime/components/extra_hosts/index.vue';

import { Runtime } from '@/api/interface/runtime';
import { useVModel } from '@vueuse/core';

const props = defineProps({
    modelValue: {
        type: Object as PropType<Runtime.Runtime>,
        required: true,
    },
});

const emit = defineEmits(['update:modelValue']);
const runtime = useVModel(props, 'modelValue', emit);
</script>
