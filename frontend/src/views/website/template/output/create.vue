<template>
    <DrawerPro v-model="open" :header="$t('template.outputCreate')" size="60%" @close="handleClose">
        <el-form ref="formRef" :model="form" :rules="rules" label-position="top" v-loading="loading">
            <el-form-item :label="$t('template.selectTemplate')" prop="templateID">
                <el-select v-model="form.templateID" class="w-full" filterable @change="onTemplateChange">
                    <el-option v-for="tpl in templates" :key="tpl.id" :value="tpl.id" :label="tpl.name" />
                </el-select>
            </el-form-item>
            <el-form-item :label="$t('template.outputName')" prop="name">
                <el-input v-model.trim="form.name" />
            </el-form-item>
            <div v-if="currentTemplate">
                <el-divider content-position="left">
                    <el-text type="info" size="small">{{ $t('template.fillVariables') }}</el-text>
                </el-divider>
                <span v-if="!templateVariables.length" class="input-help">{{ $t('template.noVariables') }}</span>
                <el-form-item
                    v-for="variable in templateVariables"
                    :key="variable.key"
                    :label="variable.label || variable.key"
                    :required="variable.required"
                >
                    <el-input v-if="variable.type === 'text'" v-model="variableValues[variable.key]" />
                    <el-input
                        v-else-if="variable.type === 'textarea'"
                        v-model="variableValues[variable.key]"
                        type="textarea"
                        :rows="4"
                    />
                    <el-input
                        v-else-if="variable.type === 'number'"
                        v-model="variableValues[variable.key]"
                        type="number"
                    />
                    <el-select
                        v-else-if="variable.type === 'select'"
                        v-model="variableValues[variable.key]"
                        class="w-full"
                    >
                        <el-option v-for="opt in getOptions(variable)" :key="opt" :value="opt" :label="opt" />
                    </el-select>
                    <el-color-picker v-else-if="variable.type === 'color'" v-model="variableValues[variable.key]" />
                    <el-input v-else v-model="variableValues[variable.key]" />
                </el-form-item>
                <el-divider content-position="left">
                    <el-text type="info" size="small">{{ $t('template.preview') }}</el-text>
                </el-divider>
                <div class="preview-box">
                    <iframe
                        v-if="previewHTML"
                        :srcdoc="previewHTML"
                        class="preview-frame"
                        sandbox="allow-same-origin"
                    ></iframe>
                    <div v-else class="preview-empty">{{ $t('template.previewEmpty') }}</div>
                </div>
            </div>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" :loading="loading" @click="onGenerate">
                    {{ $t('template.generate') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { Website } from '@/api/interface/website';
import { createTemplateOutput, previewTemplate, searchTemplates } from '@/api/modules/website';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { computed, reactive, ref, watch } from 'vue';

const open = ref(false);
const loading = ref(false);
const formRef = ref();

const templates = ref<Website.Template[]>([]);
const variableValues = reactive<Record<string, string>>({});
const previewHTML = ref('');
let previewTimer: ReturnType<typeof setTimeout> | null = null;

const form = reactive({
    templateID: undefined as number | undefined,
    name: '',
});

const rules = {
    templateID: [{ required: true, message: i18n.global.t('commons.rule.requiredSelect'), trigger: 'change' }],
    name: [{ required: true, message: i18n.global.t('commons.rule.requiredInput'), trigger: 'blur' }],
};

const currentTemplate = computed(() => templates.value.find((tpl) => tpl.id === form.templateID));

const templateVariables = computed<Website.TemplateVariable[]>(() => {
    if (!currentTemplate.value) return [];
    try {
        const arr = JSON.parse(currentTemplate.value.variables || '[]');
        return Array.isArray(arr) ? arr : [];
    } catch {
        return [];
    }
});

const getOptions = (variable: Website.TemplateVariable) => {
    return (variable.options || '')
        .split(',')
        .map((opt) => opt.trim())
        .filter((opt) => opt);
};

const onTemplateChange = () => {
    for (const key of Object.keys(variableValues)) {
        delete variableValues[key];
    }
    for (const variable of templateVariables.value) {
        variableValues[variable.key] = variable.default || '';
    }
    refreshPreview();
};

const renderLocal = (content: string) => {
    let html = content;
    for (const key of Object.keys(variableValues)) {
        html = html.replace(new RegExp('\\{\\{' + key + '\\}\\}', 'g'), variableValues[key] || '');
    }
    return html.replace(/\{\{\w+\}\}/g, '');
};

const refreshPreview = () => {
    if (!currentTemplate.value) {
        previewHTML.value = '';
        return;
    }
    if (currentTemplate.value.type === 'single') {
        previewHTML.value = renderLocal(currentTemplate.value.content || '');
        return;
    }
    if (previewTimer) {
        clearTimeout(previewTimer);
    }
    previewTimer = setTimeout(async () => {
        try {
            const res = await previewTemplate({
                templateID: form.templateID,
                variableValues: { ...variableValues },
            });
            previewHTML.value = res.data.html;
        } catch {
            previewHTML.value = '';
        }
    }, 500);
};

watch(variableValues, () => {
    refreshPreview();
});

const onGenerate = async () => {
    await formRef.value.validate();
    loading.value = true;
    try {
        await createTemplateOutput({
            templateID: form.templateID,
            name: form.name,
            variableValues: { ...variableValues },
        });
        MsgSuccess(i18n.global.t('template.outputSuccess'));
        handleClose();
        em('search');
    } finally {
        loading.value = false;
    }
};

const em = defineEmits(['search']);

const handleClose = () => {
    formRef.value?.resetFields();
    form.templateID = undefined;
    form.name = '';
    for (const key of Object.keys(variableValues)) {
        delete variableValues[key];
    }
    previewHTML.value = '';
    open.value = false;
};

const acceptParams = async (templateID?: number) => {
    const res = await searchTemplates({ page: 1, pageSize: 1000, name: '', type: '' });
    templates.value = res.data.items || [];
    if (templateID && templates.value.some((tpl) => tpl.id === templateID)) {
        form.templateID = templateID;
        onTemplateChange();
    }
    open.value = true;
};

defineExpose({
    acceptParams,
});
</script>

<style lang="scss" scoped>
.preview-box {
    width: 100%;
    border: 1px solid var(--el-border-color);
    border-radius: 4px;
    overflow: hidden;

    .preview-frame {
        display: block;
        width: 100%;
        height: 420px;
        border: none;
        background-color: #fff;
    }

    .preview-empty {
        display: flex;
        align-items: center;
        justify-content: center;
        height: 200px;
        color: var(--el-text-color-secondary);
        font-size: 13px;
    }
}
</style>
