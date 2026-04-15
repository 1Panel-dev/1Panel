<template>
    <div class="mt-1.5">
        <el-row :gutter="20" v-for="(volume, index) in extraHosts" :key="index">
            <el-col :span="7">
                <el-form-item :prop="`volumes.${index}.hostname`" :rules="rules.hostname">
                    <el-input v-model="volume.hostname" :placeholder="$t('toolbox.device.hostname')" />
                </el-form-item>
            </el-col>
            <el-col :span="7">
                <el-form-item :prop="`volumes.${index}.ip`" :rules="rules.ip">
                    <el-input v-model="volume.ip" :placeholder="'IP'" />
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
    extraHosts: {
        type: Array<Runtime.ExtraHost>,
        required: true,
    },
});

const rules = reactive<FormRules>({
    value: [Rules.requiredInput],
});

const addEnv = () => {
    props.extraHosts.push({
        hostname: '',
        ip: '',
    });
};

const removeEnv = (index: number) => {
    props.extraHosts.splice(index, 1);
};
</script>
