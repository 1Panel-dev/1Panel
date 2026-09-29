<template>
    <DrawerPro v-model="drawerVisible" :header="$t('firewall.portWhiteList')" size="large">
        <template #content>
            <p class="input-help mb-3">{{ $t('firewall.whitelistConfigHelper') }}</p>
            <div class="mb-3">
                <el-button type="primary" :disabled="disabled" @click="openEditor()">
                    {{ $t('commons.button.create') }}
                </el-button>
            </div>
            <ComplexTable :data="data" v-loading="busy">
                <el-table-column :label="$t('commons.table.protocol')" width="100">
                    <template #default="{ row }">{{ (row.protocol || 'tcp').toUpperCase() }}</template>
                </el-table-column>
                <el-table-column :label="$t('firewall.portOrRange')" width="150">
                    <template #default="{ row }">
                        <div v-if="row.type" class="flex items-center gap-1 whitespace-nowrap">
                            <span class="truncate" :title="row.port">
                                {{ row.port || $t('commons.status.unknown') }}
                            </span>
                            <span class="shrink-0 text-xs text-[var(--el-text-color-secondary)]">
                                {{ serviceLabel(row.type) }}
                            </span>
                        </div>
                        <span v-else>{{ row.port }}</span>
                    </template>
                </el-table-column>
                <el-table-column :label="$t('firewall.allowedSources')" min-width="260">
                    <template #default="{ row }">
                        <span class="whitelist-sources">{{ formatHostAddressList(row.sources || []) }}</span>
                    </template>
                </el-table-column>
                <el-table-column :label="$t('commons.table.operate')" width="140" fixed="right">
                    <template #default="{ row }">
                        <el-button link type="primary" :disabled="disabled" @click="openEditor(row)">
                            {{ $t('commons.button.edit') }}
                        </el-button>
                        <el-button link type="primary" :disabled="disabled" @click="removeRule(row)">
                            {{ $t('commons.button.delete') }}
                        </el-button>
                    </template>
                </el-table-column>
            </ComplexTable>
        </template>
    </DrawerPro>
    <DialogPro
        v-model="dialogVisible"
        :title="$t(editingRule ? 'commons.button.edit' : 'commons.button.create')"
        :show-close="!saving"
    >
        <el-form
            ref="formRef"
            :model="form"
            :rules="rules"
            label-position="top"
            :disabled="saving"
            @submit.prevent="saveRule"
        >
            <el-form-item :label="$t('commons.table.type')">
                <el-select v-model="form.type" :disabled="!!editingRule" @change="changeType">
                    <el-option value="custom" :label="$t('website.other')" />
                    <el-option
                        v-for="type in serviceTypes"
                        :key="type"
                        :value="type"
                        :label="`${serviceLabel(type)} (${servicePortLabel(type)})`"
                    />
                </el-select>
            </el-form-item>
            <el-form-item :label="$t('commons.table.protocol')" required>
                <el-select v-model="form.protocol">
                    <el-option value="tcp" label="TCP" />
                    <el-option value="udp" label="UDP" />
                </el-select>
            </el-form-item>
            <template v-if="form.type === 'custom'">
                <el-form-item :label="$t('firewall.portOrRange')" prop="port">
                    <el-input v-model.trim="form.port" placeholder="80 / 8000-8100" clearable />
                    <span class="input-help">{{ $t('firewall.portWhiteListHelper') }}</span>
                </el-form-item>
            </template>
            <el-form-item v-else :label="$t('commons.table.port')" prop="port">
                <el-input v-model.trim="form.port" :placeholder="form.type === 'ssh' ? '2222' : '18443'" clearable />
                <span class="input-help">{{ $t('firewall.whitelistServicePortsHelper') }}</span>
            </el-form-item>
            <el-form-item :label="$t('firewall.allowedSources')" prop="sourceInput">
                <el-input
                    v-model.trim="form.sourceInput"
                    type="textarea"
                    :autosize="{ minRows: 3, maxRows: 8 }"
                    :placeholder="$t('firewall.sourceAddressPlaceholder')"
                />
                <span class="input-help">{{ $t('firewall.whitelistSourcesHelper') }}</span>
            </el-form-item>
        </el-form>
        <template #footer>
            <el-button :disabled="saving" @click="dialogVisible = false">{{ $t('commons.button.cancel') }}</el-button>
            <el-button :loading="saving" :disabled="disabled" type="primary" @click="saveRule">
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DialogPro>
</template>

<script lang="ts" setup>
import { computed, ref } from 'vue';
import {
    createFirewallPortWhitelist,
    deleteFirewallPortWhitelist,
    updateFirewallPortWhitelist,
} from '@/api/modules/firewall';
import i18n from '@/lang';
import { MsgError, MsgSuccess } from '@/utils/message';
import {
    formatHostAddressList,
    isValidIPOrCIDR,
    isValidPortRange,
    splitTagValues,
} from '@/views/host/firewall/utils/validation';
import { ElMessageBox, type FormInstance, type FormRules } from 'element-plus';
import { normalizeWhiteListRule, WhiteListProtocol, WhiteListRule, WhiteListType, whiteListRuleKey } from './model';

const props = defineProps<{
    rules?: WhiteListRule[];
    panelPort?: string;
    sshPort?: string;
    loading: boolean;
}>();
const emit = defineEmits<{ (e: 'saved'): void }>();
const drawerVisible = ref(false);
const dialogVisible = ref(false);
const saving = ref(false);
const busy = computed(() => props.loading || saving.value);
const disabled = computed(() => busy.value || !props.rules);
const data = computed(() => props.rules || []);
const editingRule = ref<WhiteListRule>();
const formRef = ref<FormInstance>();
const form = ref({
    type: 'custom' as WhiteListType | 'custom',
    protocol: 'tcp' as WhiteListProtocol,
    port: '',
    sourceInput: '',
});
const rules: FormRules = {
    port: [
        {
            required: true,
            validator: (_rule, value: string, callback) => {
                if (!value) {
                    callback(new Error(i18n.global.t('commons.rule.requiredInput')));
                    return;
                }
                if (!isValidPortRange(value) || (form.value.type !== 'custom' && !/^\d+$/.test(value))) {
                    const message = form.value.type === 'custom' ? 'firewall.portFormatError' : 'commons.rule.port';
                    callback(new Error(i18n.global.t(message)));
                    return;
                }
                callback();
            },
            trigger: ['blur', 'change'],
        },
    ],
    sourceInput: [
        {
            validator: (_rule, value: string, callback) => {
                if (splitTagValues([value]).some((source) => !isValidIPOrCIDR(source))) {
                    callback(new Error(i18n.global.t('commons.rule.ip')));
                    return;
                }
                callback();
            },
            trigger: ['blur', 'change'],
        },
    ],
};
const serviceTypes: WhiteListType[] = ['ssh', 'panel'];
const serviceLabel = (type: WhiteListType) => (type === 'panel' ? '1Panel' : 'SSH');
const servicePortLabel = (type: WhiteListType) =>
    (type === 'panel' ? props.panelPort : props.sshPort) || i18n.global.t('commons.status.unknown');

const changeType = () => {
    if (form.value.type === 'custom') return;
    form.value.port = (form.value.type === 'panel' ? props.panelPort : props.sshPort) || '';
    formRef.value?.clearValidate('port');
};

const acceptParams = () => {
    drawerVisible.value = true;
    dialogVisible.value = false;
};

const openEditor = (rule?: WhiteListRule) => {
    if (disabled.value) return;
    editingRule.value = rule;
    form.value = {
        type: rule?.type || 'custom',
        protocol: rule?.protocol || 'tcp',
        port: rule?.port || '',
        sourceInput: formatHostAddressList(rule ? rule.sources || [] : ['0.0.0.0/0', '::/0']),
    };
    formRef.value?.clearValidate();
    dialogVisible.value = true;
};

const saveRule = async () => {
    if (disabled.value || !formRef.value) return;
    const valid = await formRef.value.validate().catch(() => false);
    if (!valid) return;
    const sources = splitTagValues([form.value.sourceInput]);
    if (!sources.length) {
        sources.push('0.0.0.0/0', '::/0');
    }
    let rule: WhiteListRule;
    try {
        rule = normalizeWhiteListRule({
            ...form.value,
            type: form.value.type === 'custom' ? undefined : form.value.type,
            port: form.value.port,
            sources,
        });
    } catch {
        MsgError(i18n.global.t('firewall.portFormatError'));
        return;
    }
    const key = whiteListRuleKey(rule);
    if (data.value.some((item) => item !== editingRule.value && whiteListRuleKey(item) === key)) {
        MsgError(i18n.global.t('commons.rule.duplicate'));
        return;
    }
    const oldRule = editingRule.value;
    await submit(() => (oldRule ? updateFirewallPortWhitelist({ oldRule, rule }) : createFirewallPortWhitelist(rule)));
};

const removeRule = async (rule: WhiteListRule) => {
    if (disabled.value) return;
    const confirmed = await ElMessageBox.confirm(
        i18n.global.t('firewall.whitelistDeleteConfirm'),
        i18n.global.t('commons.button.delete'),
        {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'warning',
        },
    )
        .then(() => true)
        .catch(() => false);
    if (confirmed) await submit(() => deleteFirewallPortWhitelist(rule));
};

const submit = async (request: () => ReturnType<typeof createFirewallPortWhitelist>) => {
    if (disabled.value) return;
    saving.value = true;
    try {
        await request();
        dialogVisible.value = false;
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        emit('saved');
    } catch {
    } finally {
        saving.value = false;
    }
};

defineExpose({ acceptParams });
</script>

<style scoped>
.whitelist-sources {
    overflow-wrap: anywhere;
}
</style>
