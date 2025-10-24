<template>
    <el-form-item :label="$t('commons.table.group')" :prop="prop">
        <el-select v-model="selectedValue">
            <el-option
                v-for="(group, index) in groups"
                :key="index"
                :label="group.name === 'Default' ? $t('commons.table.default') : group.name"
                :value="group.id"
            />
        </el-select>
    </el-form-item>
</template>

<script lang="ts" setup>
import { ref, computed, onMounted } from 'vue';
import { Group } from '@/api/interface/group';
import { getAgentGroupList } from '@/api/modules/group';

const props = defineProps({
    modelValue: {
        type: Number,
        required: true,
    },
    groupType: {
        type: String,
        default: 'website',
    },
    prop: {
        type: String,
        default: '',
    },
});

const emit = defineEmits(['update:modelValue']);

const groups = ref<Group.GroupInfo[]>([]);

const selectedValue = computed({
    get: () => props.modelValue,
    set: (value) => emit('update:modelValue', value),
});

onMounted(async () => {
    const res = await getAgentGroupList(props.groupType);
    groups.value = res.data;
    if (groups.value.length > 0 && !props.modelValue && props.modelValue == 0) {
        selectedValue.value = groups.value[0].id;
    }
});
</script>
