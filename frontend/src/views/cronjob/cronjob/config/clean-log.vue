<template>
    <el-card>
        <el-form-item :label="$t('cronjob.cleanLogscope')" prop="scopes">
            <el-select :model-value="modelValue" multiple @change="handleChange" clearable>
                <el-option v-for="item in scopes" :key="item.value" :value="item.value" :label="item.label" />
            </el-select>
        </el-form-item>
    </el-card>
</template>

<script lang="ts" setup>
import i18n from '@/lang';

interface Props {
    modelValue?: string[];
}

withDefaults(defineProps<Props>(), {
    modelValue: () => [],
});

const emit = defineEmits<{
    (e: 'update:modelValue', value: string[]): void;
}>();

const handleChange = (value: string[]) => {
    emit('update:modelValue', value);
};

const scopes = [{ value: 'website', label: i18n.global.t('logs.websiteLog') }];
</script>
