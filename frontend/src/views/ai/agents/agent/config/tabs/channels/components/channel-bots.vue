<template>
    <div class="channel-bots">
        <div class="channel-bots__header">
            <span class="channel-bots__title">{{ title || t('aiTools.agents.bots') }}</span>
            <el-button type="primary" link :disabled="addDisabled" @click="openCreate">
                {{ t('aiTools.agents.addBot') }}
            </el-button>
        </div>
        <el-table :data="bots" border>
            <el-table-column :label="t('commons.table.name')" min-width="140">
                <template #default="{ row }">
                    <div class="channel-bots__name">
                        <span>{{ row.name || row.accountId }}</span>
                        <el-tag v-if="row.isDefault" type="success" size="small">
                            {{ t('commons.table.default') }}
                        </el-tag>
                    </div>
                </template>
            </el-table-column>
            <el-table-column prop="accountId" :label="t('aiTools.agents.accountId')" min-width="140" />
            <el-table-column v-if="summaryFormatter" :label="summaryLabel" min-width="200">
                <template #default="{ row }">
                    <span>{{ summaryFormatter(row) }}</span>
                </template>
            </el-table-column>
            <el-table-column :label="t('commons.table.status')" width="120">
                <template #default="{ row, $index }">
                    <el-switch
                        :model-value="row.enabled"
                        :disabled="disabled || isBotActionDisabled(row)"
                        @change="updateEnabled($index, $event)"
                    />
                </template>
            </el-table-column>
            <el-table-column :label="t('commons.table.operate')" min-width="320" fixed="right">
                <template #default="{ row, $index }">
                    <div class="channel-bots__actions">
                        <el-button link type="primary" :disabled="disabled" @click="openEdit(row, $index)">
                            {{ t('commons.button.edit') }}
                        </el-button>
                        <el-button
                            v-if="defaultable && !row.isDefault"
                            link
                            type="primary"
                            :disabled="disabled || isBotActionDisabled(row)"
                            @click="setDefault(row.accountId)"
                        >
                            {{ t('aiTools.agents.setDefaultBot') }}
                        </el-button>
                        <el-button
                            v-if="approvable"
                            link
                            type="primary"
                            :disabled="
                                disabled ||
                                !row.enabled ||
                                isBotActionDisabled(row) ||
                                (approveDisabled ? approveDisabled(row) : false)
                            "
                            @click="emit('approve', row)"
                        >
                            {{ t('aiTools.agents.approvePairing') }}
                        </el-button>
                        <el-button
                            link
                            type="primary"
                            :disabled="disabled || undeletableAccountIds.includes(row.accountId)"
                            @click="removeBot($index)"
                        >
                            {{ t('commons.button.delete') }}
                        </el-button>
                    </div>
                </template>
            </el-table-column>
        </el-table>

        <DialogPro v-model="dialogVisible" :title="dialogTitle">
            <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
                <el-form-item v-if="showNameField" :label="t('commons.table.name')" prop="name">
                    <el-input v-model="form.name" :disabled="disabled" />
                </el-form-item>
                <el-form-item :label="t('aiTools.agents.accountId')" prop="accountId">
                    <el-input v-model="form.accountId" :disabled="disabled || accountIdLocked" />
                </el-form-item>
                <el-form-item :label="t('commons.table.status')">
                    <el-switch v-model="form.enabled" :disabled="disabled" />
                </el-form-item>
                <el-form-item v-for="field in fields" :key="field.prop" :label="field.label" :prop="field.prop">
                    <el-select
                        v-if="field.type === 'select'"
                        v-model="form[field.prop]"
                        :disabled="disabled"
                        :placeholder="field.placeholder"
                    >
                        <el-option
                            v-for="option in field.options || []"
                            :key="option.value"
                            :label="option.label"
                            :value="option.value"
                        />
                    </el-select>
                    <el-input
                        v-else-if="field.type === 'password'"
                        v-model="form[field.prop]"
                        type="password"
                        show-password
                        :disabled="disabled"
                        :placeholder="field.placeholder"
                    />
                    <el-input
                        v-else-if="field.type === 'textarea'"
                        v-model="form[field.prop]"
                        type="textarea"
                        :rows="3"
                        :disabled="disabled"
                        :placeholder="field.placeholder"
                    />
                    <el-input v-else v-model="form[field.prop]" :disabled="disabled" :placeholder="field.placeholder" />
                </el-form-item>
            </el-form>
            <template #footer>
                <el-button @click="dialogVisible = false">{{ t('commons.button.cancel') }}</el-button>
                <el-button type="primary" :disabled="disabled" @click="saveBot">
                    {{ t('commons.button.save') }}
                </el-button>
            </template>
        </DialogPro>
    </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue';
import type { FormInstance, FormRules } from 'element-plus';
import type { PropType } from 'vue';
import { useI18n } from 'vue-i18n';
import { Rules } from '@/global/form-rules';
import { MsgWarning } from '@/utils/message';

interface ChannelBotField {
    prop: string;
    label: string;
    type?: 'text' | 'password' | 'select' | 'textarea';
    placeholder?: string;
    required?: boolean;
    options?: Array<{
        label: string;
        value: string;
    }>;
}

interface ChannelBotItem {
    accountId: string;
    name: string;
    enabled: boolean;
    isDefault: boolean;
    [key: string]: any;
}

const props = defineProps({
    title: {
        type: String,
        default: '',
    },
    bots: {
        type: Array as PropType<ChannelBotItem[]>,
        required: true,
    },
    fields: {
        type: Array as PropType<ChannelBotField[]>,
        default: () => [],
    },
    createBot: {
        type: Function as PropType<() => ChannelBotItem>,
        required: true,
    },
    summaryLabel: {
        type: String,
        default: '',
    },
    summaryFormatter: {
        type: Function as PropType<(bot: ChannelBotItem) => string>,
        default: undefined,
    },
    defaultable: {
        type: Boolean,
        default: false,
    },
    approvable: {
        type: Boolean,
        default: false,
    },
    approveDisabled: {
        type: Function as PropType<(bot: ChannelBotItem) => boolean>,
        default: undefined,
    },
    addDisabled: {
        type: Boolean,
        default: false,
    },
    disabled: {
        type: Boolean,
        default: false,
    },
    showNameField: {
        type: Boolean,
        default: true,
    },
    undeletableAccountIds: {
        type: Array as PropType<string[]>,
        default: () => [],
    },
    fixedAccountIds: {
        type: Array as PropType<string[]>,
        default: () => [],
    },
});

const emit = defineEmits<{
    (e: 'update:bots', bots: ChannelBotItem[]): void;
    (e: 'approve', bot: ChannelBotItem): void;
    (e: 'save', action?: 'delete' | 'save'): void;
}>();

const { t } = useI18n();
const dialogVisible = ref(false);
const editIndex = ref(-1);
const editingAccountId = ref('');
const formRef = ref<FormInstance>();

const form = reactive<ChannelBotItem>(props.createBot());

const accountIdLocked = computed(() => props.fixedAccountIds.includes(editingAccountId.value));
const dialogTitle = computed(() => (editIndex.value >= 0 ? t('commons.button.edit') : t('commons.button.create')));

const rules = computed<FormRules>(() => {
    const config: FormRules = {
        accountId: [
            Rules.appName,
            {
                validator: (_rule, value, callback) => {
                    const exists = props.bots.some(
                        (bot, index) => index !== editIndex.value && bot.accountId === value,
                    );
                    if (exists) {
                        callback(new Error(t('aiTools.agents.botDuplicateAccountId')));
                        return;
                    }
                    callback();
                },
                trigger: 'blur',
            },
        ],
    };
    if (props.showNameField) {
        config.name = [Rules.requiredInput];
    }
    for (const field of props.fields) {
        if (!field.required) {
            continue;
        }
        config[field.prop] = [Rules.requiredInput];
    }
    return config;
});

const resetForm = (bot?: ChannelBotItem) => {
    Object.keys(form).forEach((key) => {
        delete form[key];
    });
    Object.assign(form, props.createBot(), bot ? { ...bot } : {});
};

const emitBots = (bots: ChannelBotItem[]) => {
    const nextBots = bots.map((bot) => ({ ...bot }));
    if (props.defaultable && nextBots.length > 0 && nextBots.every((bot) => !bot.isDefault)) {
        nextBots[0].isDefault = true;
    }
    emit('update:bots', nextBots);
};

const isBotActionDisabled = (bot: ChannelBotItem) => {
    return props.fields.some((field) => field.required && !bot[field.prop]);
};

const openCreate = () => {
    editIndex.value = -1;
    editingAccountId.value = '';
    resetForm();
    dialogVisible.value = true;
};

const openEdit = (bot: ChannelBotItem, index: number) => {
    editIndex.value = index;
    editingAccountId.value = bot.accountId;
    resetForm(bot);
    dialogVisible.value = true;
};

const saveBot = async () => {
    if (!formRef.value) {
        return;
    }
    await formRef.value.validate();
    const nextBots = props.bots.map((bot) => ({ ...bot }));
    const nextBot = { ...form };
    if (!props.showNameField) {
        nextBot.name = nextBot.accountId;
    }
    if (editIndex.value >= 0) {
        nextBots.splice(editIndex.value, 1, nextBot);
    } else {
        if (props.defaultable && nextBots.every((bot) => !bot.isDefault)) {
            nextBot.isDefault = true;
        }
        nextBots.push(nextBot);
    }
    emitBots(nextBots);
    dialogVisible.value = false;
    emit('save', 'save');
};

const removeBot = (index: number) => {
    const nextBots = props.bots.filter((_, botIndex) => botIndex !== index).map((bot) => ({ ...bot }));
    if (nextBots.length === 0) {
        MsgWarning(t('aiTools.agents.botRequired'));
        return;
    }
    emitBots(nextBots);
    emit('save', 'delete');
};

const setDefault = (accountId: string) => {
    const nextBots = props.bots.map((bot) => ({
        ...bot,
        isDefault: bot.accountId === accountId,
    }));
    emitBots(nextBots);
    emit('save', 'save');
};

const updateEnabled = (index: number, enabled: boolean | string | number) => {
    const nextBots = props.bots.map((bot) => ({ ...bot }));
    nextBots[index].enabled = Boolean(enabled);
    emitBots(nextBots);
    emit('save', 'save');
};
</script>

<style lang="scss" scoped>
.channel-bots {
    margin-top: 20px;
}

.channel-bots__header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 12px;
}

.channel-bots__title {
    font-size: 14px;
    font-weight: 500;
    color: var(--el-text-color-primary);
}

.channel-bots__name {
    display: flex;
    align-items: center;
    gap: 8px;
}

.channel-bots__actions {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 6px 12px;

    :deep(.el-button) {
        margin-left: 0;
        white-space: nowrap;
    }
}
</style>
