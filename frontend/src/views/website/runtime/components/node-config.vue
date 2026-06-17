<template>
    <el-tabs type="border-card">
        <el-tab-pane :label="$t('commons.table.port')">
            <PortConfig :exposedPorts="exposedPorts" />
        </el-tab-pane>
        <el-tab-pane :label="$t('runtime.environment')">
            <Environment :environments="environments" />
        </el-tab-pane>
        <el-tab-pane :label="$t('container.mount')">
            <Volumes :volumes="volumes" />
        </el-tab-pane>
        <el-tab-pane :label="$t('runtime.extraHosts')">
            <ExtraHosts :extraHosts="extraHosts" />
        </el-tab-pane>
    </el-tabs>
</template>

<script lang="ts" setup>
import PortConfig from '@/views/website/runtime/components/port/index.vue';
import Environment from '@/views/website/runtime/components/environment/index.vue';
import Volumes from '@/views/website/runtime/components/volume/index.vue';
import ExtraHosts from '@/views/website/runtime/components/extra-hosts/index.vue';
import { useVModel } from '@vueuse/core';
import { computed, type PropType } from 'vue';

interface RuntimeConfigFields {
    exposedPorts?: any[];
    environments?: any[];
    volumes?: any[];
    extraHosts?: any[];
}

const props = defineProps({
    modelValue: {
        type: Object as PropType<RuntimeConfigFields>,
        required: true,
    },
});

const emit = defineEmits<{
    'update:modelValue': [value: RuntimeConfigFields];
}>();

const runtime = useVModel(props, 'modelValue', emit);

const ensureConfigArray = (key: keyof RuntimeConfigFields): any[] => {
    if (!runtime.value[key]) {
        runtime.value[key] = [];
    }
    return runtime.value[key] as any[];
};

const exposedPorts = computed(() => ensureConfigArray('exposedPorts'));
const environments = computed(() => ensureConfigArray('environments'));
const volumes = computed(() => ensureConfigArray('volumes'));
const extraHosts = computed(() => ensureConfigArray('extraHosts'));
</script>
