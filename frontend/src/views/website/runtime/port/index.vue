<template>
    <el-row :gutter="20">
        <el-col :span="7">
            <el-form-item :label="$t('runtime.appPort')" prop="params.APP_PORT" :rules="rules.port">
                <el-input v-model.number="runtime.params.APP_PORT" />
                <span class="input-help">{{ $t('runtime.appPortHelper') }}</span>
            </el-form-item>
        </el-col>
        <el-col :span="7">
            <el-form-item :label="$t('runtime.externalPort')" prop="params.PANEL_APP_PORT_HTTP" :rules="rules.port">
                <el-input v-model.number="runtime.params.PANEL_APP_PORT_HTTP" />
                <span class="input-help">{{ $t('runtime.externalPortHelper') }}</span>
            </el-form-item>
        </el-col>
        <el-col :span="4">
            <el-form-item :label="$t('commons.button.add') + $t('commons.table.port')">
                <el-button @click="addPort">
                    <el-icon><Plus /></el-icon>
                </el-button>
            </el-form-item>
        </el-col>
        <el-col :span="6">
            <el-form-item :label="$t('app.allowPort')">
                <el-switch v-model="runtime.params.HOST_IP" :active-value="'0.0.0.0'" :inactive-value="'127.0.0.1'" />
            </el-form-item>
        </el-col>
    </el-row>
    <el-row :gutter="20" v-for="(port, index) in runtime.exposedPorts" :key="index">
        <el-col :span="7">
            <el-form-item :prop="`exposedPorts.${index}.containerPort`" :rules="rules.port">
                <el-input v-model.number="port.containerPort" :placeholder="$t('runtime.appPort')" />
            </el-form-item>
        </el-col>
        <el-col :span="7">
            <el-form-item :prop="`exposedPorts.${index}.hostPort`" :rules="rules.port">
                <el-input v-model.number="port.hostPort" :placeholder="$t('runtime.externalPort')" />
            </el-form-item>
        </el-col>
        <el-col :span="4">
            <el-form-item>
                <el-button type="primary" @click="removePort(index)" link>
                    {{ $t('commons.button.delete') }}
                </el-button>
            </el-form-item>
        </el-col>
    </el-row>
</template>

<script setup lang="ts">
import { Rules, checkNumberRange } from '@/global/form-rules';
import { FormRules } from 'element-plus';
import { defineProps } from 'vue';
import { useVModel } from '@vueuse/core';

const props = defineProps({
    mode: {
        type: String,
        required: true,
    },
    modelValue: {
        type: Object,
        required: true,
    },
});
const emit = defineEmits(['update:modelValue']);
const runtime = useVModel(props, 'modelValue', emit);

watch(
    () => runtime.value.params['APP_PORT'],
    (newVal) => {
        if (newVal) {
            runtime.value.params['PANEL_APP_PORT_HTTP'] = newVal;
        }
    },
    { deep: true },
);

const rules = reactive<FormRules>({
    port: [Rules.requiredInput, Rules.paramPort, checkNumberRange(1, 65535)],
});

const addPort = () => {
    runtime.value.exposedPorts.push({
        hostPort: undefined,
        containerPort: undefined,
    });
};

const removePort = (index: number) => {
    runtime.value.exposedPorts.splice(index, 1);
};
</script>
