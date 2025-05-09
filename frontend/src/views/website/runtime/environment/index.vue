<template>
    <el-text type="info">{{ $t('php.date_timezone') }}: TZ=Asia/Shaghai</el-text>
    <div class="mt-1.5">
        <el-row :gutter="20" v-for="(env, index) in environments" :key="index">
            <el-col :span="7">
                <el-form-item :prop="`environments.${index}.key`" :rules="rules.value">
                    <el-input v-model="env.key" :placeholder="$t('runtime.envKey')" />
                </el-form-item>
            </el-col>
            <el-col :span="7">
                <el-form-item :prop="`environments.${index}.value`" :rules="rules.value">
                    <el-input v-model="env.value" :placeholder="$t('runtime.envValue')" />
                </el-form-item>
            </el-col>
            <el-col :span="4">
                <el-form-item>
                    <el-button type="primary" @click="removeEnv(index)" link class="mt-1">
                        {{ $t('commons.button.delete') }}
                    </el-button>
                </el-form-item>
            </el-col>
        </el-row>
        <el-row :gutter="20">
            <el-col :span="4">
                <el-button @click="addEnv">{{ $t('commons.button.add') }}</el-button>
            </el-col>
        </el-row>
    </div>
</template>

<script setup lang="ts">
import { reactive } from 'vue';
import { FormRules } from 'element-plus';
import { Rules } from '@/global/form-rules';
import { Runtime } from '@/api/interface/runtime';

const props = defineProps({
    environments: {
        type: Array<Runtime.Environment>,
        required: true,
    },
});

const rules = reactive<FormRules>({
    value: [Rules.requiredInput],
});

const addEnv = () => {
    props.environments.push({
        key: '',
        value: '',
    });
};

const removeEnv = (index: number) => {
    props.environments.splice(index, 1);
};
</script>
