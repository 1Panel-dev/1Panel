<template>
    <div class="mt-1.5">
        <el-text>{{ $t('commons.table.port') }}</el-text>
        <div class="mt-1.5">
            <el-row :gutter="20" v-for="(port, index) in runtime.exposedPorts" :key="index">
                <el-col :span="7">
                    <el-form-item :prop="`exposedPorts.${index}.hostPort`" :rules="rules.port">
                        <el-input v-model.number="port.hostPort" :placeholder="$t('runtime.externalPort')" />
                    </el-form-item>
                </el-col>
                <el-col :span="7">
                    <el-form-item :prop="`exposedPorts.${index}.containerPort`" :rules="rules.port">
                        <el-input v-model.number="port.containerPort" :placeholder="$t('runtime.appPort')" />
                    </el-form-item>
                </el-col>
                <el-col :span="7">
                    <el-text>{{ $t('app.allowPort') }}</el-text>
                    <el-switch
                        class="ml-1"
                        v-model="port.hostIP"
                        :active-value="'0.0.0.0'"
                        :inactive-value="'127.0.0.1'"
                    />
                </el-col>
                <el-col :span="2">
                    <el-form-item>
                        <el-button type="primary" @click="removePort(index)" link class="mt-1">
                            {{ $t('commons.button.delete') }}
                        </el-button>
                    </el-form-item>
                </el-col>
            </el-row>
        </div>
        <el-row :gutter="20">
            <el-col :span="4">
                <el-button @click="addPort">{{ $t('commons.button.add') }}</el-button>
            </el-col>
        </el-row>
    </div>
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

const rules = reactive<FormRules>({
    port: [Rules.requiredInput, Rules.paramPort, checkNumberRange(1, 65535)],
});

const addPort = () => {
    runtime.value.exposedPorts.push({
        hostPort: undefined,
        containerPort: undefined,
        hostIP: '0.0.0.0',
    });
};

const removePort = (index: number) => {
    runtime.value.exposedPorts.splice(index, 1);
};
</script>
