<template>
    <DrawerPro
        v-model="open"
        :header="isEdit ? $t('template.edit') : $t('template.create')"
        size="large"
        @close="handleClose"
    >
        <el-form ref="formRef" :model="form" :rules="rules" label-position="top" v-loading="loading">
            <el-form-item :label="$t('template.name')" prop="name">
                <el-input v-model.trim="form.name" />
            </el-form-item>
            <el-form-item :label="$t('template.type')" prop="type">
                <el-radio-group v-model="form.type" :disabled="isEdit">
                    <el-radio value="single">{{ $t('template.single') }}</el-radio>
                    <el-radio value="multi">{{ $t('template.multi') }}</el-radio>
                </el-radio-group>
            </el-form-item>
            <el-form-item v-if="form.type === 'single'" :label="$t('template.content')" prop="content">
                <el-input v-model="form.content" type="textarea" :rows="12" :placeholder="contentPlaceholder" />
                <span class="input-help">{{ $t('template.autoDetectHelper') }}</span>
            </el-form-item>
            <el-form-item v-if="form.type === 'multi'" :label="$t('template.upload')">
                <el-upload :show-file-list="false" :before-upload="handleUpload" accept=".zip" action="#">
                    <el-button type="primary" plain>{{ $t('template.upload') }}</el-button>
                </el-upload>
                <span v-if="form.filePath" class="ml-2 text-xs text-gray-500">{{ form.filePath }}</span>
                <span class="input-help">{{ $t('template.autoDetectHelper') }}</span>
            </el-form-item>
            <el-form-item :label="$t('template.remark')">
                <el-input v-model="form.remark" />
            </el-form-item>
            <el-form-item :label="$t('template.variables')">
                <el-button size="small" type="primary" plain @click="addVariable">
                    {{ $t('commons.button.add') }}
                </el-button>
                <el-table :data="variables" border class="mt-2">
                    <el-table-column :label="$t('template.variableKey')" min-width="110px">
                        <template #default="{ row }">
                            <el-input v-model.trim="row.key" />
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('template.variableLabel')" min-width="110px">
                        <template #default="{ row }">
                            <el-input v-model="row.label" />
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('template.variableType')" width="160px">
                        <template #default="{ row }">
                            <el-select v-model="row.type">
                                <el-option
                                    v-for="item in variableTypes"
                                    :key="item.value"
                                    :value="item.value"
                                    :label="$t('template.varType' + item.name)"
                                >
                                    <div>
                                        <span>{{ $t('template.varType' + item.name) }}</span>
                                        <span class="ml-2 text-xs text-gray-400">
                                            {{ $t('template.varType' + item.name + 'Helper') }}
                                        </span>
                                    </div>
                                </el-option>
                            </el-select>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('template.variableDefault')" min-width="110px">
                        <template #default="{ row }">
                            <el-input v-model="row.default" />
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('template.variableOptions')" min-width="130px">
                        <template #default="{ row }">
                            <el-input v-model="row.options" :disabled="row.type !== 'select'" />
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('template.variableRequired')" width="75px">
                        <template #default="{ row }">
                            <el-switch v-model="row.required" />
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.operate')" width="85px">
                        <template #default="{ $index }">
                            <el-button link type="danger" @click="variables.splice($index, 1)">
                                {{ $t('commons.button.delete') }}
                            </el-button>
                        </template>
                    </el-table-column>
                </el-table>
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" :loading="loading" @click="onSubmit">
                    {{ $t('commons.button.save') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { Website } from '@/api/interface/website';
import { createTemplate, getTemplate, updateTemplate, uploadTemplateZip } from '@/api/modules/website';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { reactive, ref, watch } from 'vue';

const open = ref(false);
const loading = ref(false);
const formRef = ref();
const isEdit = ref(false);
const contentPlaceholder = '<html><body><h1>{{title}}</h1></body></html>';

const variableTypes = [
    { value: 'text', name: 'Text' },
    { value: 'textarea', name: 'Textarea' },
    { value: 'number', name: 'Number' },
    { value: 'select', name: 'Select' },
    { value: 'color', name: 'Color' },
];

const initForm = () => ({
    id: 0,
    name: '',
    type: 'single',
    content: '',
    filePath: '',
    variables: '',
    remark: '',
});
const form = reactive(initForm());

const rules = {
    name: [{ required: true, message: i18n.global.t('commons.rule.requiredInput'), trigger: 'blur' }],
    type: [{ required: true, message: i18n.global.t('commons.rule.requiredSelect'), trigger: 'change' }],
};

const variables = ref<Website.TemplateVariable[]>([]);

const addVariable = (key?: string) => {
    variables.value.push({
        key: typeof key === 'string' ? key : '',
        label: '',
        type: 'text',
        default: '',
        options: '',
        required: false,
    });
};

const mergeDetectedKeys = (keys: string[]) => {
    for (const key of keys) {
        if (!variables.value.some((v) => v.key === key)) {
            addVariable(key);
        }
    }
};

watch(
    () => form.content,
    (content) => {
        if (form.type !== 'single' || !content) return;
        const keys: string[] = [];
        for (const match of content.matchAll(/\{\{(\w+)\}\}/g)) {
            if (!keys.includes(match[1])) {
                keys.push(match[1]);
            }
        }
        mergeDetectedKeys(keys);
    },
);

const handleUpload = async (file: globalThis.File) => {
    try {
        const res = await uploadTemplateZip(file);
        form.filePath = res.data.filePath;
        mergeDetectedKeys(res.data.variables || []);
        MsgSuccess(i18n.global.t('commons.msg.uploadSuccess'));
    } catch {}
    return false;
};

const onSubmit = async () => {
    await formRef.value.validate();
    form.variables = JSON.stringify(variables.value.filter((v) => v.key));
    loading.value = true;
    try {
        if (isEdit.value) {
            await updateTemplate(form);
        } else {
            await createTemplate(form);
        }
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        handleClose();
        em('search');
    } finally {
        loading.value = false;
    }
};

const em = defineEmits(['search']);

const handleClose = () => {
    formRef.value?.resetFields();
    Object.assign(form, initForm());
    variables.value = [];
    open.value = false;
};

const acceptParams = async (id?: number) => {
    Object.assign(form, initForm());
    variables.value = [];
    isEdit.value = !!id;
    if (id) {
        const res = await getTemplate(id);
        const tpl = res.data;
        form.id = tpl.id;
        form.name = tpl.name;
        form.type = tpl.type;
        form.content = tpl.content;
        form.filePath = tpl.filePath;
        form.variables = tpl.variables;
        form.remark = tpl.remark;
        try {
            variables.value = JSON.parse(tpl.variables || '[]');
        } catch {
            variables.value = [];
        }
    }
    open.value = true;
};

defineExpose({
    acceptParams,
});
</script>
