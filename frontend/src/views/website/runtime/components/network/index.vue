<template>
    <el-form-item :label="$t('container.network')">
        <el-select v-model="network" filterable>
            <el-option v-for="item in networkOptions" :key="item" :label="item" :value="item" />
        </el-select>
        <div v-if="network === 'host'" class="input-help">
            {{ $t('app.install.hostModeHelper') }}
        </div>
    </el-form-item>
</template>

<script lang="ts" setup>
import { listNetwork } from '@/api/modules/container';
import { Container } from '@/api/interface/container';
import { useVModel } from '@vueuse/core';
import { onMounted, ref } from 'vue';

const DEFAULT_NETWORK = '1panel-network';

const props = defineProps<{
    modelValue?: string;
}>();

const emit = defineEmits<{
    'update:modelValue': [value: string];
}>();

const network = useVModel(props, 'modelValue', emit);
const networkOptions = ref<string[]>([DEFAULT_NETWORK]);

const loadNetworks = async () => {
    try {
        const res = await listNetwork();
        networkOptions.value = (res.data || []).map((item: Container.Options) => item.option);
    } catch (error) {
        networkOptions.value = [];
    }
    const current = network.value?.trim() || '';
    if (current && !networkOptions.value.includes(current)) {
        networkOptions.value.unshift(current);
    }
    if (!current) {
        network.value = DEFAULT_NETWORK;
    }
    if (!networkOptions.value.includes(DEFAULT_NETWORK)) {
        networkOptions.value.unshift(DEFAULT_NETWORK);
    }
};

onMounted(loadNetworks);
</script>
