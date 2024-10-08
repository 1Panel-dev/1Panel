<template>
    <el-row :gutter="20">
        <el-col :span="7">
            <el-form-item :label="$t('runtime.appPort')" prop="params.APP_PORT" :rules="rules.port">
                <el-input v-model.number="params.APP_PORT" />
                <span class="input-help">{{ $t('runtime.appPortHelper') }}</span>
            </el-form-item>
        </el-col>
        <el-col :span="7">
            <el-form-item :label="$t('runtime.externalPort')" prop="params.port" :rules="rules.port">
                <el-input v-model.number="params.port" />
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
                <el-switch v-model="params.HOST_IP" :active-value="'0.0.0.0'" :inactive-value="'127.0.0.1'" />
            </el-form-item>
        </el-col>
    </el-row>
    <el-row :gutter="20" v-for="(port, index) in exposedPorts" :key="index">
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
import { Runtime } from '@/api/interface/runtime';

const props = defineProps({
    params: {
        type: Object,
        required: true,
    },
    exposedPorts: {
        type: Array<Runtime.ExposedPort>,
        required: true,
    },
});

const rules = reactive<FormRules>({
    port: [Rules.requiredInput, Rules.paramPort, checkNumberRange(1, 65535)],
});

const addPort = () => {
    props.exposedPorts.push({
        hostPort: undefined,
        containerPort: undefined,
    });
};

const removePort = (index: number) => {
    props.exposedPorts.splice(index, 1);
};
</script>
