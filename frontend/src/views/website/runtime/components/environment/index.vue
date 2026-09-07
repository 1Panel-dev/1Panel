<template>
    <el-text v-if="showHelper" type="info">{{ $t('php.dateTimezoneHelper') }}</el-text>
    <div class="mt-1.5">
        <el-row :gutter="20" v-for="(env, index) in environments" :key="index">
            <el-col :span="7">
                <el-form-item :prop="`environments.${index}.key`" :rules="rules.value">
                    <el-input v-model="env.key" :placeholder="$t('runtime.envKey')" />
                </el-form-item>
            </el-col>
            <el-col :span="1">
                <div class="mt-1">=</div>
            </el-col>
            <el-col :span="7">
                <el-form-item :prop="`environments.${index}.value`">
                    <el-input v-model="env.value" :placeholder="$t('runtime.envValue')" />
                </el-form-item>
            </el-col>
            <el-col :span="3">
                <el-form-item>
                    <el-button v-permission type="primary" @click="removeEnv(index)" link class="mt-1">
                        {{ $t('commons.button.delete') }}
                    </el-button>
                </el-form-item>
            </el-col>
        </el-row>
        <div class="flex flex-wrap gap-2">
            <el-button v-permission @click="addEnv">{{ $t('commons.button.add') }}</el-button>
            <el-button v-permission icon="Upload" @click="openImport">{{ $t('runtime.importEnv') }}</el-button>
        </div>
    </div>
    <el-dialog
        v-model="importVisible"
        :title="$t('runtime.importEnv')"
        width="min(680px, 90vw)"
        append-to-body
        destroy-on-close
        :close-on-click-modal="false"
        @closed="importText = ''"
    >
        <el-input
            v-model="importText"
            type="textarea"
            :rows="8"
            :aria-label="$t('runtime.environment')"
            placeholder="KEY=value"
            spellcheck="false"
        />
        <el-alert
            v-if="parsed.error"
            class="mt-3"
            type="error"
            :closable="false"
            :title="$t('runtime.envImportError', [parsed.error.line, $t('runtime.' + parsed.error.reason)])"
        />
        <el-table v-else-if="parsed.entries.length" :data="parsed.entries" max-height="260" class="mt-3">
            <el-table-column prop="key" :label="$t('runtime.envKey')" min-width="120" show-overflow-tooltip />
            <el-table-column prop="value" :label="$t('runtime.envValue')" min-width="180" show-overflow-tooltip />
        </el-table>
        <template #footer>
            <el-button @click="importVisible = false">{{ $t('commons.button.cancel') }}</el-button>
            <el-button
                v-permission
                type="primary"
                :disabled="!!parsed.error || !parsed.entries.length"
                @click="importEnv"
            >
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </el-dialog>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue';
import { FormRules } from 'element-plus';
import { Rules } from '@/global/form-rules';
import { Runtime } from '@/api/interface/runtime';
import { mergeEnvironments, parseEnvironment } from '@/utils/runtime-environment';

const props = defineProps({
    environments: {
        type: Array<Runtime.Environment>,
        required: true,
    },
    showHelper: {
        type: Boolean,
        default: true,
    },
});

const { showHelper } = props;

const importVisible = ref(false);
const importText = ref('');
const parsed = computed(() => parseEnvironment(importText.value));

const openImport = () => {
    importText.value = '';
    importVisible.value = true;
};

const importEnv = () => {
    if (parsed.value.error || !parsed.value.entries.length) return;
    mergeEnvironments(props.environments, parsed.value.entries);
    importVisible.value = false;
    importText.value = '';
};

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
