<template>
    <div class="custom-webhook-form">
        <el-alert
            v-if="modelValue.state"
            class="mb-3"
            :closable="false"
            type="warning"
            show-icon
            :title="$t('xpack.alert.customWebhookRecoveryRequired')"
        />
        <el-form-item
            :label="$t('xpack.alert.displayName')"
            prop="customWebhook.displayName"
            :error="errorFor('displayName')"
        >
            <el-input
                :model-value="modelValue.displayName"
                maxlength="64"
                show-word-limit
                @update:model-value="updateDisplayName"
            />
        </el-form-item>

        <el-form-item :label="$t('xpack.alert.webhookPreset')">
            <el-select :model-value="modelValue.preset" class="w-full" @change="changePreset">
                <el-option value="genericJson" :label="$t('xpack.alert.genericJsonPreset')" />
                <el-option value="slack" label="Slack" />
                <el-option value="discord" label="Discord" />
                <el-option value="teamsWorkflows" label="Teams Workflows" />
                <el-option value="custom" :label="$t('xpack.alert.customPreset')" />
            </el-select>
        </el-form-item>

        <el-form-item :label="$t('xpack.alert.webhookUrl')" :error="errorFor('url')">
            <template v-if="modelValue.url.action === 'keep'">
                <div class="secret-editor">
                    <el-input
                        :model-value="secretEditorValue(modelValue.url)"
                        type="password"
                        show-password
                        autocomplete="new-password"
                        :maxlength="CUSTOM_WEBHOOK_LIMITS.url"
                        :placeholder="secretEditorPlaceholder(modelValue.url, 'https://example.com/webhook')"
                        @update:model-value="updateUrlValue"
                    />
                    <div class="secret-editor__actions">
                        <el-button v-if="allowClearUrl" plain type="danger" @click="clearUrl">
                            {{ $t('xpack.alert.clearSecret') }}
                        </el-button>
                    </div>
                </div>
            </template>
            <template v-else-if="modelValue.url.action === 'clear'">
                <div class="secret-editor">
                    <el-alert :closable="false" type="warning" :title="$t('xpack.alert.secretCleared')" />
                    <el-input
                        model-value=""
                        type="password"
                        show-password
                        autocomplete="new-password"
                        :maxlength="CUSTOM_WEBHOOK_LIMITS.url"
                        placeholder="https://example.com/webhook"
                        @update:model-value="updateUrlValue"
                    />
                    <div class="secret-editor__actions">
                        <el-button v-if="modelValue.url.configured" plain @click="keepUrl">
                            {{ $t('xpack.alert.keepSecret') }}
                        </el-button>
                    </div>
                </div>
            </template>
            <template v-else>
                <div class="secret-editor">
                    <el-input
                        :model-value="modelValue.url.value"
                        type="password"
                        show-password
                        autocomplete="new-password"
                        :maxlength="CUSTOM_WEBHOOK_LIMITS.url"
                        placeholder="https://example.com/webhook"
                        @update:model-value="updateUrlValue"
                    />
                    <div v-if="modelValue.url.configured" class="secret-editor__actions">
                        <el-button plain @click="keepUrl">{{ $t('xpack.alert.keepSecret') }}</el-button>
                        <el-button v-if="allowClearUrl" plain type="danger" @click="clearUrl">
                            {{ $t('xpack.alert.clearSecret') }}
                        </el-button>
                    </div>
                </div>
            </template>
            <span class="input-help">{{ $t('xpack.alert.webhookUrlSecretHelper') }}</span>
            <span class="input-help">{{ $t('xpack.alert.webhookPublicAddressHelper') }}</span>
        </el-form-item>

        <el-form-item :label="$t('xpack.alert.bodyType')">
            <el-radio-group :model-value="modelValue.body.type" class="body-type-group" @change="changeBodyType">
                <el-radio-button value="json">JSON</el-radio-button>
                <el-radio-button value="form">Form</el-radio-button>
                <el-radio-button value="text">Text</el-radio-button>
            </el-radio-group>
            <span class="input-help">POST · {{ derivedContentType }}</span>
        </el-form-item>

        <el-form-item :label="$t('xpack.alert.bodyTemplate')" :error="errorFor('body')">
            <template v-if="modelValue.body.type === 'form'">
                <div class="key-value-list">
                    <div v-for="(field, index) in modelValue.body.fields" :key="field.uid" class="key-value-row">
                        <el-input
                            :model-value="field.key"
                            :maxlength="CUSTOM_WEBHOOK_LIMITS.headerName"
                            :placeholder="$t('xpack.alert.formFieldName')"
                            @update:model-value="updateFormField(index, 'key', $event)"
                        />
                        <el-input
                            :model-value="field.value"
                            :maxlength="CUSTOM_WEBHOOK_LIMITS.headerValue"
                            :placeholder="$t('xpack.alert.formFieldValue')"
                            @focus="activeFormFieldIndex = index"
                            @update:model-value="updateFormField(index, 'value', $event)"
                        />
                        <el-button plain type="danger" @click="removeFormField(index)">
                            {{ $t('commons.button.delete') }}
                        </el-button>
                    </div>
                    <el-button
                        plain
                        type="primary"
                        :disabled="modelValue.body.fields.length >= CUSTOM_WEBHOOK_LIMITS.formFields"
                        @click="addFormField"
                    >
                        {{ $t('xpack.alert.addFormField') }}
                    </el-button>
                </div>
            </template>
            <el-input
                v-else
                ref="bodyTemplateInputRef"
                :model-value="modelValue.body.template"
                type="textarea"
                :rows="modelValue.body.type === 'json' ? 10 : 6"
                :maxlength="CUSTOM_WEBHOOK_LIMITS.body"
                resize="vertical"
                @focus="bodyTemplateFocused = true"
                @update:model-value="updateBodyTemplate"
            />
        </el-form-item>

        <el-collapse v-model="activeSections" class="advanced-collapse">
            <el-collapse-item name="advanced">
                <template #title>
                    <div class="advanced-title">
                        <span>{{ $t('xpack.alert.webhookAdvanced') }}</span>
                        <span class="advanced-title__summary">
                            POST · {{ derivedContentType }} · {{ modelValue.headers.length }}
                            {{ $t('xpack.alert.headers') }}
                        </span>
                    </div>
                </template>

                <el-form-item :label="$t('xpack.alert.headers')" :error="errorFor('headers')">
                    <div class="header-list">
                        <div v-for="(header, index) in modelValue.headers" :key="header.uid" class="header-card">
                            <el-input
                                :model-value="header.key"
                                :maxlength="CUSTOM_WEBHOOK_LIMITS.headerName"
                                :placeholder="$t('xpack.alert.headerName')"
                                @update:model-value="updateHeaderKey(index, $event)"
                            />

                            <template v-if="!header.secret">
                                <el-input
                                    :model-value="header.value"
                                    :maxlength="CUSTOM_WEBHOOK_LIMITS.headerValue"
                                    :placeholder="$t('xpack.alert.headerValue')"
                                    @update:model-value="updateHeader(index, { value: $event })"
                                />
                            </template>
                            <template v-else-if="header.action === 'keep'">
                                <el-input
                                    :model-value="secretEditorValue(header)"
                                    :maxlength="CUSTOM_WEBHOOK_LIMITS.headerValue"
                                    type="password"
                                    show-password
                                    autocomplete="new-password"
                                    :placeholder="secretEditorPlaceholder(header, $t('xpack.alert.headerValue'))"
                                    @update:model-value="updateSecretHeaderValue(index, $event)"
                                />
                            </template>
                            <template v-else-if="header.action === 'clear'">
                                <div class="secret-editor">
                                    <el-alert
                                        :closable="false"
                                        type="warning"
                                        :title="$t('xpack.alert.secretCleared')"
                                    />
                                    <el-input
                                        model-value=""
                                        :maxlength="CUSTOM_WEBHOOK_LIMITS.headerValue"
                                        type="password"
                                        show-password
                                        autocomplete="new-password"
                                        :placeholder="$t('xpack.alert.headerValue')"
                                        @update:model-value="updateSecretHeaderValue(index, $event)"
                                    />
                                </div>
                            </template>
                            <template v-else>
                                <el-input
                                    :model-value="header.value"
                                    :maxlength="CUSTOM_WEBHOOK_LIMITS.headerValue"
                                    type="password"
                                    show-password
                                    autocomplete="new-password"
                                    :placeholder="$t('xpack.alert.headerValue')"
                                    @update:model-value="updateHeader(index, { value: $event })"
                                />
                            </template>

                            <div class="header-card__footer">
                                <el-checkbox
                                    :model-value="header.secret"
                                    :disabled="isCustomWebhookSecretHeader(header.key)"
                                    @update:model-value="toggleHeaderSecret(index, Boolean($event))"
                                >
                                    {{ $t('xpack.alert.secretValue') }}
                                </el-checkbox>
                                <div v-if="header.secret" class="header-card__secret-actions">
                                    <el-button
                                        v-if="header.configured && header.action !== 'keep'"
                                        link
                                        @click="setHeaderSecretAction(index, 'keep')"
                                    >
                                        {{ $t('xpack.alert.keepSecret') }}
                                    </el-button>
                                    <el-button
                                        v-if="header.configured && header.action !== 'clear'"
                                        link
                                        type="danger"
                                        @click="setHeaderSecretAction(index, 'clear')"
                                    >
                                        {{ $t('xpack.alert.clearSecret') }}
                                    </el-button>
                                </div>
                                <el-button link type="danger" @click="removeHeader(index)">
                                    {{ $t('commons.button.delete') }}
                                </el-button>
                            </div>
                        </div>
                        <el-button
                            plain
                            type="primary"
                            :disabled="modelValue.headers.length >= CUSTOM_WEBHOOK_LIMITS.headers"
                            @click="addHeader"
                        >
                            {{ $t('xpack.alert.addHeader') }}
                        </el-button>
                    </div>
                </el-form-item>

                <el-form-item :label="$t('xpack.alert.templateVariables')">
                    <div class="variable-list">
                        <el-tag
                            v-for="variable in templateVariables"
                            :key="variable.token"
                            class="variable-tag"
                            effect="plain"
                            :title="variable.token"
                            @click="insertVariable(variable.token)"
                        >
                            {{ $t(variable.labelKey) }}
                        </el-tag>
                    </div>
                    <span class="input-help">{{ $t('xpack.alert.templateVariablesHelper') }}</span>
                </el-form-item>
            </el-collapse-item>
        </el-collapse>
    </div>
</template>

<script setup lang="ts">
import { computed, nextTick, reactive, ref, watch } from 'vue';
import { ElInput, ElMessageBox } from 'element-plus';
import i18n from '@/lang';
import {
    CUSTOM_WEBHOOK_VARIABLES,
    CUSTOM_WEBHOOK_LIMITS,
    CustomWebhookBody,
    CustomWebhookBodyType,
    CustomWebhookDraft,
    CustomWebhookPreset,
    CustomWebhookSecretAction,
    CustomWebhookValidationIssue,
    applyCustomWebhookPreset,
    contentTypeForBodyType,
    createCustomWebhookFormField,
    createCustomWebhookHeader,
    insertCustomWebhookVariable,
    isCustomWebhookSecretHeader,
    isCustomWebhookPresetPristine,
    updateCustomWebhookBodyTemplate,
} from './custom-webhook';
import {
    clearSecretDraft,
    keepSecretDraft,
    replaceSecretDraft,
    secretEditorPlaceholder,
    secretEditorValue,
} from './secret-field';

const props = withDefaults(
    defineProps<{
        modelValue: CustomWebhookDraft;
        validationIssues?: CustomWebhookValidationIssue[];
        allowClearUrl?: boolean;
    }>(),
    {
        validationIssues: () => [],
        allowClearUrl: false,
    },
);

const emit = defineEmits<{
    (e: 'update:modelValue', value: CustomWebhookDraft): void;
}>();

const activeSections = ref<string[]>([]);
const derivedContentType = computed(() => contentTypeForBodyType(props.modelValue.body.type));
const bodyTemplateInputRef = ref<InstanceType<typeof ElInput>>();
const bodyTemplateFocused = ref(false);
const activeFormFieldIndex = ref<number | null>(null);
const templateVariables = [
    { token: CUSTOM_WEBHOOK_VARIABLES[0], labelKey: 'xpack.alert.templateVariableTitle' },
    { token: CUSTOM_WEBHOOK_VARIABLES[1], labelKey: 'xpack.alert.templateVariableMessage' },
    { token: CUSTOM_WEBHOOK_VARIABLES[2], labelKey: 'xpack.alert.templateVariableType' },
    { token: CUSTOM_WEBHOOK_VARIABLES[3], labelKey: 'xpack.alert.templateVariableNodeName' },
    { token: CUSTOM_WEBHOOK_VARIABLES[4], labelKey: 'xpack.alert.templateVariableTimestamp' },
];
const bodyDrafts = reactive<Record<CustomWebhookBodyType, CustomWebhookBody>>({
    json: { type: 'json', template: '', fields: [] },
    form: { type: 'form', template: '', fields: [] },
    text: { type: 'text', template: '{{message}}', fields: [] },
});

watch(
    () => props.modelValue.body,
    (body) => {
        bodyDrafts[body.type] = {
            type: body.type,
            template: body.template,
            fields: body.fields.map((field) => ({ ...field })),
        };
    },
    { deep: true, immediate: true },
);

const emitValue = (value: CustomWebhookDraft) => emit('update:modelValue', value);

const updateDisplayName = (displayName: string) => emitValue({ ...props.modelValue, displayName });

const changePreset = async (value: string | number | boolean | undefined) => {
    const preset = value as CustomWebhookPreset;
    if (preset === props.modelValue.preset) return;
    const hasBodyContent = Boolean(props.modelValue.body.template || props.modelValue.body.fields.length);
    const overwritesBody = preset !== 'custom' && !isCustomWebhookPresetPristine(props.modelValue) && hasBodyContent;
    if (overwritesBody) {
        try {
            await ElMessageBox.confirm(
                i18n.global.t('xpack.alert.presetOverwriteHelper'),
                i18n.global.t('xpack.alert.webhookPreset'),
                {
                    confirmButtonText: i18n.global.t('commons.button.confirm'),
                    cancelButtonText: i18n.global.t('commons.button.cancel'),
                },
            );
        } catch {
            return;
        }
    }
    const next = applyCustomWebhookPreset(props.modelValue, preset);
    bodyDrafts[next.body.type] = { ...next.body, fields: next.body.fields.map((field) => ({ ...field })) };
    bodyTemplateFocused.value = false;
    activeFormFieldIndex.value = null;
    emitValue(next);
};

const updateUrlValue = (value: string) => {
    emitValue({
        ...props.modelValue,
        url: replaceSecretDraft(props.modelValue.url, value),
    });
};

const keepUrl = () => {
    emitValue({ ...props.modelValue, url: keepSecretDraft(props.modelValue.url) });
};

const clearUrl = () => {
    if (!props.allowClearUrl) return;
    emitValue({ ...props.modelValue, url: clearSecretDraft(props.modelValue.url) });
};

const defaultBodyForType = (type: CustomWebhookBodyType): CustomWebhookBody => {
    if (type === 'form') return { type, template: '', fields: [createCustomWebhookFormField()] };
    if (type === 'text') return { type, template: '{{message}}', fields: [] };
    return { type, template: '{}', fields: [] };
};

const changeBodyType = (value: string | number | boolean | undefined) => {
    const type = value as CustomWebhookBodyType;
    bodyDrafts[props.modelValue.body.type] = {
        ...props.modelValue.body,
        fields: props.modelValue.body.fields.map((field) => ({ ...field })),
    };
    const cached = bodyDrafts[type];
    const body = cached.template || cached.fields.length ? cached : defaultBodyForType(type);
    bodyTemplateFocused.value = false;
    activeFormFieldIndex.value = null;
    emitValue({
        ...props.modelValue,
        preset: 'custom',
        body: { ...body, fields: body.fields.map((field) => ({ ...field })) },
    });
};

const updateBodyTemplate = (template: string) => {
    emitValue(updateCustomWebhookBodyTemplate(props.modelValue, template));
};

const addFormField = () => {
    emitValue({
        ...props.modelValue,
        preset: 'custom',
        body: { ...props.modelValue.body, fields: [...props.modelValue.body.fields, createCustomWebhookFormField()] },
    });
};

const updateFormField = (index: number, field: 'key' | 'value', value: string) => {
    const fields = props.modelValue.body.fields.map((item, itemIndex) =>
        itemIndex === index ? { ...item, [field]: value } : item,
    );
    emitValue({ ...props.modelValue, preset: 'custom', body: { ...props.modelValue.body, fields } });
};

const removeFormField = (index: number) => {
    emitValue({
        ...props.modelValue,
        preset: 'custom',
        body: {
            ...props.modelValue.body,
            fields: props.modelValue.body.fields.filter((_, itemIndex) => itemIndex !== index),
        },
    });
    if (activeFormFieldIndex.value === index) {
        activeFormFieldIndex.value = null;
    } else if (activeFormFieldIndex.value !== null && activeFormFieldIndex.value > index) {
        activeFormFieldIndex.value -= 1;
    }
};

const addHeader = () =>
    emitValue({ ...props.modelValue, headers: [...props.modelValue.headers, createCustomWebhookHeader()] });

const updateHeader = (index: number, patch: Partial<CustomWebhookDraft['headers'][number]>) => {
    const headers = props.modelValue.headers.map((header, itemIndex) =>
        itemIndex === index ? { ...header, ...patch } : header,
    );
    emitValue({ ...props.modelValue, headers });
};

const updateSecretHeaderValue = (index: number, value: string) => {
    updateHeader(index, replaceSecretDraft(props.modelValue.headers[index], value));
};

const updateHeaderKey = (index: number, key: string) => {
    const current = props.modelValue.headers[index];
    if (!isCustomWebhookSecretHeader(key) || current.secret) {
        updateHeader(index, { key });
        return;
    }
    updateHeader(index, {
        key,
        secret: true,
        configured: false,
        masked: '',
        action: 'replace',
    });
};

const removeHeader = (index: number) => {
    emitValue({ ...props.modelValue, headers: props.modelValue.headers.filter((_, itemIndex) => itemIndex !== index) });
};

const toggleHeaderSecret = (index: number, secret: boolean) => {
    const current = props.modelValue.headers[index];
    if (!secret && isCustomWebhookSecretHeader(current.key)) return;
    updateHeader(index, {
        secret,
        configured: false,
        masked: '',
        action: 'replace',
        value: current.action === 'replace' ? current.value : '',
    });
};

const setHeaderSecretAction = (index: number, action: CustomWebhookSecretAction) => {
    const header = props.modelValue.headers[index];
    if (action === 'keep') {
        updateHeader(index, keepSecretDraft(header));
        return;
    }
    if (action === 'clear') {
        updateHeader(index, clearSecretDraft(header));
        return;
    }
    updateHeader(index, replaceSecretDraft(header, ''));
};

const insertVariable = (variable: string) => {
    if (props.modelValue.body.type === 'form') {
        if (props.modelValue.body.fields.length === 0) {
            const field = createCustomWebhookFormField();
            field.value = variable;
            emitValue({
                ...props.modelValue,
                preset: 'custom',
                body: { ...props.modelValue.body, fields: [field] },
            });
            return;
        }
        const fallbackIndex = props.modelValue.body.fields.length - 1;
        const targetIndex = activeFormFieldIndex.value ?? fallbackIndex;
        updateFormField(targetIndex, 'value', props.modelValue.body.fields[targetIndex].value + variable);
        return;
    }
    const textarea = bodyTemplateInputRef.value?.textarea;
    const selection =
        bodyTemplateFocused.value && textarea
            ? { start: textarea.selectionStart, end: textarea.selectionEnd }
            : undefined;
    const insertion = insertCustomWebhookVariable(
        props.modelValue.body.template,
        variable,
        props.modelValue.body.type,
        selection,
    );
    updateBodyTemplate(insertion.template);
    void nextTick(() => {
        bodyTemplateInputRef.value?.focus();
        bodyTemplateInputRef.value?.textarea?.setSelectionRange(insertion.cursor, insertion.cursor);
        bodyTemplateFocused.value = true;
    });
};

const errorFor = (field: string): string => {
    const issue = props.validationIssues.find((item) => item.field === field || item.field.startsWith(`${field}.`));
    return issue ? i18n.global.t(`xpack.alert.customWebhookValidation.${issue.code}`) : '';
};
</script>

<style scoped lang="scss">
.custom-webhook-form,
.secret-editor,
.key-value-list,
.header-list {
    width: 100%;
}

.secret-editor {
    display: flex;
    flex-direction: column;
    gap: 8px;
}

.secret-editor__actions,
.header-card__footer,
.header-card__secret-actions {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 8px;
}

.key-value-list,
.header-list {
    display: flex;
    flex-direction: column;
    gap: 10px;
}

.key-value-row {
    display: grid;
    grid-template-columns: minmax(120px, 0.8fr) minmax(180px, 1.4fr) auto;
    gap: 8px;
    align-items: start;
}

.header-card {
    display: grid;
    grid-template-columns: minmax(150px, 0.8fr) minmax(220px, 1.4fr);
    gap: 8px;
    padding: 12px;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 8px;
}

.header-card__footer {
    grid-column: 1 / -1;
    justify-content: space-between;
}

.advanced-collapse {
    margin-top: 4px;
    border-top: 0;
}

.advanced-title {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    width: 100%;
    padding-right: 12px;
}

.advanced-title__summary {
    color: var(--el-text-color-secondary);
    font-size: 12px;
    font-weight: 400;
}

.variable-list {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
}

.variable-tag {
    cursor: pointer;
}

@media (max-width: 640px) {
    .key-value-row,
    .header-card {
        grid-template-columns: 1fr;
    }

    .header-card__footer {
        grid-column: 1;
        align-items: flex-start;
        flex-direction: column;
    }

    .body-type-group {
        display: flex;
        width: 100%;
    }

    .body-type-group :deep(.el-radio-button) {
        flex: 1;
    }

    .body-type-group :deep(.el-radio-button__inner) {
        width: 100%;
        min-height: 44px;
    }

    .secret-editor__actions > .el-button,
    .key-value-row > .el-button {
        min-height: 44px;
    }

    .advanced-title {
        align-items: flex-start;
        flex-direction: column;
        gap: 2px;
    }
}
</style>
