<template>
    <el-form ref="lbFormRef" label-position="top" :model="formData" :rules="rules">
        <el-form-item :label="$t('commons.table.name')" prop="name">
            <el-input v-model.trim="formData.name" :disabled="disabled"></el-input>
        </el-form-item>
        <el-form-item :label="$t('website.algorithm')" prop="algorithm">
            <el-select v-model="formData.algorithm" @change="handleAlgorithmChange">
                <el-option
                    v-for="(algorithm, index) in Algorithms"
                    :label="algorithm.label"
                    :key="index"
                    :value="algorithm.value"
                ></el-option>
            </el-select>
            <span class="input-help">{{ getHelper(formData.algorithm) }}</span>
        </el-form-item>

        <div>
            <el-card v-for="(server, index) of formData.servers" :key="index" class="server-card" shadow="hover">
                <template #header>
                    <div class="card-header">
                        <span class="server-title">
                            <el-icon><Monitor /></el-icon>
                            <span class="ml-2">{{ $t('website.server') }} - {{ server.server || index + 1 }}</span>
                        </span>
                        <el-button
                            v-if="formData.servers.length > 1"
                            text
                            type="danger"
                            icon="Delete"
                            size="small"
                            @click="removeServer(index)"
                        ></el-button>
                    </div>
                </template>

                <div>
                    <el-row :gutter="16">
                        <el-col :span="24">
                            <el-form-item
                                :label="$t('setting.address')"
                                :prop="`servers.${index}.server`"
                                :rules="rules.server"
                            >
                                <el-input
                                    v-model="formData.servers[index].server"
                                    :placeholder="type === 'stream' ? '127.0.0.1:3306' : '127.0.0.1:8080'"
                                    @input="handleChange"
                                ></el-input>
                            </el-form-item>
                        </el-col>
                    </el-row>
                    <el-row :gutter="16">
                        <el-col :xs="24" :sm="24" :md="12" :lg="12" :xl="12">
                            <el-form-item
                                :label="$t('website.weight')"
                                :prop="`servers.${index}.weight`"
                                :rules="rules.weight"
                            >
                                <el-input
                                    type="number"
                                    v-model.number="formData.servers[index].weight"
                                    @input="handleChange"
                                ></el-input>
                            </el-form-item>
                        </el-col>
                        <el-col :xs="24" :sm="24" :md="12" :lg="12" :xl="12">
                            <el-form-item
                                :label="$t('website.strategy')"
                                :prop="`servers.${index}.flag`"
                                :rules="rules.flag"
                            >
                                <el-select v-model="formData.servers[index].flag" clearable @change="handleChange">
                                    <el-option
                                        v-for="flag in getStatusStrategy()"
                                        :label="flag.label"
                                        :key="flag.value"
                                        :value="flag.value"
                                    ></el-option>
                                </el-select>
                            </el-form-item>
                        </el-col>
                    </el-row>
                    <el-row :gutter="16">
                        <el-col :xs="24" :sm="24" :md="12" :lg="12" :xl="12">
                            <el-form-item
                                :label="$t('website.maxFails')"
                                :prop="`servers.${index}.maxFails`"
                                :rules="rules.maxFails"
                            >
                                <el-input
                                    type="number"
                                    v-model.number="formData.servers[index].maxFails"
                                    @input="handleChange"
                                >
                                    <template #append>{{ $t('commons.units.time') }}</template>
                                </el-input>
                            </el-form-item>
                        </el-col>
                        <el-col :xs="24" :sm="24" :md="12" :lg="12" :xl="12">
                            <el-form-item :prop="`servers.${index}.failTimeout`" :rules="rules.failTimeout">
                                <template #label>
                                    <span class="inline-flex items-center">
                                        {{ $t('website.failTimeout') }}
                                        <el-tooltip :content="$t('website.failTimeoutHelper')" placement="top">
                                            <el-icon><QuestionFilled /></el-icon>
                                        </el-tooltip>
                                    </span>
                                </template>
                                <el-input
                                    type="number"
                                    v-model.number="formData.servers[index].failTimeout"
                                    @input="handleChange"
                                >
                                    <template #append>
                                        <el-select
                                            v-model.number="formData.servers[index].failTimeoutUnit"
                                            class="!w-24"
                                            @change="handleChange"
                                        >
                                            <el-option
                                                v-for="(unit, indexKey) in Units"
                                                :key="indexKey"
                                                :label="unit.label"
                                                :value="unit.value"
                                            />
                                        </el-select>
                                    </template>
                                </el-input>
                            </el-form-item>
                        </el-col>
                    </el-row>
                    <el-row :gutter="16">
                        <el-col :xs="24" :sm="24" :md="12" :lg="12" :xl="12">
                            <el-form-item
                                :label="$t('website.maxConns')"
                                :prop="`servers.${index}.maxConns`"
                                :rules="rules.maxConns"
                            >
                                <el-input
                                    type="number"
                                    v-model.number="formData.servers[index].maxConns"
                                    @input="handleChange"
                                ></el-input>
                            </el-form-item>
                        </el-col>
                    </el-row>
                </div>
            </el-card>

            <el-button class="add-server-btn" type="primary" plain @click="addServer" icon="Plus">
                {{ $t('commons.button.add') + $t('website.server') }}
            </el-button>
        </div>
    </el-form>
</template>

<script lang="ts" setup>
import { ref, computed } from 'vue';
import { FormInstance } from 'element-plus';
import { Rules, checkNumberRange } from '@/global/form-rules';
import { getAlgorithms, getStatusStrategy, Units } from '@/global/mimetype';
import { Monitor, QuestionFilled } from '@element-plus/icons-vue';

interface Server {
    server: string;
    weight?: number;
    maxFails?: number;
    maxConns?: number;
    failTimeout?: number;
    failTimeoutUnit: string;
    flag: string;
}

interface LoadBalanceFormData {
    name: string;
    algorithm: string;
    servers: Server[];
}

interface Props {
    modelValue: LoadBalanceFormData;
    disabled?: boolean;
    type?: string;
}

const props = withDefaults(defineProps<Props>(), {
    disabled: false,
    type: 'loadbanlance',
});

const emit = defineEmits<{
    (e: 'update:modelValue', value: LoadBalanceFormData): void;
}>();

const lbFormRef = ref<FormInstance>();

const formData = computed({
    get: () => props.modelValue,
    set: (value) => emit('update:modelValue', value),
});

const rules = ref<any>({
    name: [Rules.appName],
    algorithm: [Rules.requiredSelect],
    server: [Rules.requiredInput],
    weight: [checkNumberRange(0, 100)],
    servers: {
        type: Array,
    },
    maxFails: [checkNumberRange(0, 1000)],
    maxConns: [checkNumberRange(0, 10000)],
    failTimeout: [checkNumberRange(0, 300)],
});

const Algorithms = getAlgorithms(props.type);

const helper = ref();
const getHelper = (key: string) => {
    Algorithms.forEach((algorithm) => {
        if (algorithm.value === key) {
            helper.value = algorithm.placeHolder;
        }
    });
    return helper.value;
};

const initServer = (): Server => ({
    server: '',
    weight: undefined,
    maxFails: undefined,
    maxConns: undefined,
    failTimeout: undefined,
    failTimeoutUnit: 's',
    flag: '',
});

const addServer = () => {
    const newServers = [...formData.value.servers, initServer()];
    emit('update:modelValue', { ...formData.value, servers: newServers });
};

const removeServer = (index: number) => {
    const newServers = [...formData.value.servers];
    newServers.splice(index, 1);
    emit('update:modelValue', { ...formData.value, servers: newServers });
};

const handleChange = () => {
    emit('update:modelValue', { ...formData.value });
};

const handleAlgorithmChange = () => {
    handleChange();
};

const validate = async () => {
    if (!lbFormRef.value) return false;
    return await lbFormRef.value.validate();
};

const resetFields = () => {
    lbFormRef.value?.resetFields();
};

const clearValidate = () => {
    lbFormRef.value?.clearValidate();
};

defineExpose({
    validate,
    resetFields,
    clearValidate,
    formRef: lbFormRef,
});
</script>

<style scoped lang="scss">
.server-card {
    margin-bottom: 16px;
    border-radius: 8px;
    transition: all 0.3s ease;

    &:hover {
        border-color: var(--el-color-primary);
    }

    :deep(.el-card__header) {
        padding: 12px 20px;
        background-color: var(--el-fill-color-light);
    }
}

.card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    height: 20px;

    .server-title {
        display: flex;
        align-items: center;
        font-weight: 500;
        font-size: 14px;
        color: var(--el-text-color-primary);
    }
}

.add-server-btn {
    width: 100%;
    height: 48px;
    border-style: dashed;
    font-size: 14px;
    margin-top: 8px;

    &:hover {
        border-color: var(--el-color-primary);
        background-color: var(--el-color-primary-light-9);
    }
}
</style>
