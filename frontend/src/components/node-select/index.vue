<template>
    <el-select
        :model-value="modelValue"
        @update:model-value="handleChange"
        class="p-w-200"
        :placeholder="$t('setting.selectNode')"
    >
        <template #prefix>{{ $t('xpack.node.node') }}</template>
        <el-option
            v-for="item in nodes"
            :key="item.id"
            :label="item.name === 'local' ? globalStore.getMasterAlias() : item.name"
            :value="item.name"
        ></el-option>
    </el-select>
</template>

<script setup>
import { defineProps, defineEmits } from 'vue';
import { listAllNodes } from '@/api/modules/setting';
import { useGlobalStore } from '@/composables/useGlobalStore';
const { globalStore } = useGlobalStore();

defineProps({
    modelValue: {
        type: String,
        default: '',
    },
});

const nodes = ref([]);
const emit = defineEmits(['update:modelValue', 'change']);

const handleChange = (value) => {
    emit('update:modelValue', value);
    emit('change', value);
};

const listNodes = async () => {
    try {
        const res = await listAllNodes();
        nodes.value = res.data || [];
    } catch (error) {}
};

onMounted(() => {
    listNodes();
});
</script>
