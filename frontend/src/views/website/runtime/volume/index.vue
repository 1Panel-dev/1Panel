<template>
    <div class="mt-2">
        <el-text>{{ $t('container.mount') }}</el-text>
        <div class="mt-2">
            <el-row :gutter="20" v-for="(volume, index) in volumes" :key="index">
                <el-col :span="7">
                    <el-form-item :prop="`volumes.${index}.source`" :rules="rules.value">
                        <el-input v-model="volume.source" :placeholder="$t('container.hostOption')" />
                    </el-form-item>
                </el-col>
                <el-col :span="7">
                    <el-form-item :prop="`volumes.${index}.target`" :rules="rules.value">
                        <el-input v-model="volume.target" :placeholder="$t('container.containerDir')" />
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
    </div>
</template>

<script setup lang="ts">
import { defineProps, reactive } from 'vue';
import { FormRules } from 'element-plus';
import { Rules } from '@/global/form-rules';
import { Runtime } from '@/api/interface/runtime';

const props = defineProps({
    volumes: {
        type: Array<Runtime.Volume>,
        required: true,
    },
});

const rules = reactive<FormRules>({
    value: [Rules.requiredInput],
});

const addEnv = () => {
    props.volumes.push({
        source: '',
        target: '',
    });
};

const removeEnv = (index: number) => {
    props.volumes.splice(index, 1);
};
</script>
