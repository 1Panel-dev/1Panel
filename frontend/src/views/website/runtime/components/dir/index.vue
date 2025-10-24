<template>
    <el-form-item :label="$t('runtime.codeDir')" prop="codeDir">
        <el-input v-model.trim="runtime.codeDir" :disabled="mode === 'edit'" @blur="changeDir">
            <template #prepend>
                <el-button
                    icon="Folder"
                    @click="fileRef.acceptParams({ path: runtime.codeDir, dir: true, disabled: mode === 'edit' })"
                />
            </template>
        </el-input>
        <span class="input-help">
            {{ dirHelper }}
        </span>
    </el-form-item>
    <div v-if="appKey == 'node'">
        <el-row :gutter="20">
            <el-col :span="18">
                <el-form-item :label="$t('runtime.runScript')" prop="params.EXEC_SCRIPT">
                    <el-select v-model="runtime.params['EXEC_SCRIPT']" v-if="runtime.params['CUSTOM_SCRIPT'] == '0'">
                        <el-option
                            v-for="(script, index) in scripts"
                            :key="index"
                            :label="script.name + ' 【 ' + script.script + ' 】'"
                            :value="script.name"
                        >
                            <el-row :gutter="10">
                                <el-col :span="4">{{ script.name }}</el-col>
                                <el-col :span="10">{{ ' 【 ' + script.script + ' 】' }}</el-col>
                            </el-row>
                        </el-option>
                    </el-select>
                    <el-input v-else v-model="runtime.params['EXEC_SCRIPT']"></el-input>
                    <span class="input-help" v-if="runtime.params['CUSTOM_SCRIPT'] == '0'">
                        {{ $t('runtime.runScriptHelper') }}
                    </span>
                    <span class="input-help" v-else>
                        {{ scriptHelper }}
                    </span>
                </el-form-item>
            </el-col>
            <el-col :span="6">
                <el-form-item :label="$t('runtime.customScript')" prop="params.CUSTOM_SCRIPT">
                    <el-switch
                        v-model="runtime.params['CUSTOM_SCRIPT']"
                        :active-value="'1'"
                        :inactive-value="'0'"
                        @change="changeScriptType"
                    />
                </el-form-item>
            </el-col>
        </el-row>
    </div>
    <div v-else>
        <el-form-item :label="$t('runtime.runScript')" prop="params.EXEC_SCRIPT">
            <el-input v-model="runtime.params['EXEC_SCRIPT']"></el-input>
            <span class="input-help">
                {{ scriptHelper }}
            </span>
        </el-form-item>
    </div>
    <FileList ref="fileRef" @choose="getPath" />
</template>

<script setup lang="ts">
import { Runtime } from '@/api/interface/runtime';
import FileList from '@/components/file-list/index.vue';
import { GetNodeScripts } from '@/api/modules/runtime';
import { useVModel } from '@vueuse/core';

const fileRef = ref();

const props = defineProps({
    mode: {
        type: String,
        required: true,
    },
    modelValue: {
        type: Object,
        required: true,
    },
    dirHelper: {
        type: String,
        required: false,
    },
    scriptHelper: {
        type: String,
        required: true,
    },
    appKey: {
        type: String,
        required: false,
    },
});
const emit = defineEmits(['update:modelValue']);
const runtime = useVModel(props, 'modelValue', emit);
const scripts = ref<Runtime.NodeScripts[]>([]);

watch(
    () => runtime.value.name,
    (newVal) => {
        if (newVal && props.mode == 'create') {
            runtime.value.params['CONTAINER_NAME'] = newVal;
        }
    },
    { deep: true },
);

const changeDir = () => {
    if (props.appKey == 'node') {
        getScripts();
    }
};

const getPath = (codeDir: string) => {
    runtime.value.codeDir = codeDir;
    if (props.appKey == 'node') {
        getScripts();
    }
};

const changeScriptType = () => {
    runtime.value.params['EXEC_SCRIPT'] = '';
    if (runtime.value.params['CUSTOM_SCRIPT'] == '0') {
        getScripts();
    }
};

const getScripts = () => {
    GetNodeScripts({ codeDir: runtime.value.codeDir }).then((res) => {
        scripts.value = res.data;
        if (props.mode == 'create' && scripts.value.length > 0) {
            runtime.value.params['EXEC_SCRIPT'] = scripts.value[0].name;
        }
    });
};

onMounted(() => {
    if (props.mode == 'edit' && props.appKey == 'node') {
        getScripts();
    }
});
</script>
