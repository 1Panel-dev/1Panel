<template>
    <DrawerPro v-model="drawerVisible" :header="$t('firewall.portWhiteList')" size="large">
        <template #content>
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
        <el-form label-position="top" :disabled="saving" @submit.prevent="saveRule">
            <el-form-item :label="$t('commons.table.type')">
                <el-select v-model="form.type" :disabled="!!editingRule">
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
                <el-form-item :label="$t('firewall.portOrRange')" required>
                    <el-input v-model.trim="form.port" placeholder="80 / 8000-8100" clearable />
                    <span class="input-help">{{ $t('firewall.portWhiteListHelper') }}</span>
                </el-form-item>
            </template>
            <el-form-item v-else-if="editingRule" :label="$t('commons.table.port')" required>
                <el-input v-model.trim="form.port" :placeholder="form.type === 'ssh' ? '2222' : '18443'" clearable />
                <span class="input-help">{{ $t('firewall.whitelistServicePortsHelper') }}</span>
            </el-form-item>
            <el-form-item :label="$t('firewall.allowedSources')">
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
import { MsgError } from '@/utils/message';
import { formatHostAddressList, isValidIPOrCIDR, splitTagValues } from '@/views/host/firewall/utils/validation';
import { ElMessageBox } from 'element-plus';
import { normalizeWhiteListRule, WhiteListProtocol, WhiteListRule, WhiteListType, whiteListRuleKey } from './model';

const props = defineProps<{
    rules?: WhiteListRule[];
    panelPort?: string;
    sshPort?: string;
    loading: boolean;
}>();
const emit = defineEmits<{ (e: 'created', taskID: string): void }>();
const drawerVisible = ref(false);
const dialogVisible = ref(false);
const saving = ref(false);
const busy = computed(() => props.loading || saving.value);
const disabled = computed(() => busy.value || !props.rules);
const data = computed(() => props.rules || []);
const editingRule = ref<WhiteListRule>();
const form = ref({
    type: 'custom' as WhiteListType | 'custom',
    protocol: 'tcp' as WhiteListProtocol,
    port: '',
    sourceInput: '',
});
const serviceTypes: WhiteListType[] = ['ssh', 'panel'];
const serviceLabel = (type: WhiteListType) => (type === 'panel' ? '1Panel' : 'SSH');
const servicePortLabel = (type: WhiteListType) =>
    (type === 'panel' ? props.panelPort : props.sshPort) || i18n.global.t('commons.status.unknown');

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
    dialogVisible.value = true;
};

const saveRule = async () => {
    if (disabled.value) return;
    const sources = splitTagValues([form.value.sourceInput]);
    if (!sources.length) {
        sources.push('0.0.0.0/0', '::/0');
    }
    if (sources.some((source) => !isValidIPOrCIDR(source))) {
        MsgError(i18n.global.t('firewall.systemAccessSourceError', ['IPv4 / IPv6']));
        return;
    }
    let rule: WhiteListRule;
    try {
        rule = normalizeWhiteListRule({
            ...form.value,
            type: form.value.type === 'custom' ? undefined : form.value.type,
            port: form.value.type === 'custom' || editingRule.value ? form.value.port : undefined,
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
    await submit(
        () => (oldRule ? updateFirewallPortWhitelist({ oldRule, rule }) : createFirewallPortWhitelist(rule)),
        oldRule ? { operation: 'edit', rule: oldRule } : undefined,
    );
};

const removeRule = async (rule: WhiteListRule) => {
    await submit(() => deleteFirewallPortWhitelist(rule), { operation: 'delete', rule });
};

const submit = async (
    request: () => ReturnType<typeof createFirewallPortWhitelist>,
    change?: { operation: 'edit' | 'delete'; rule: WhiteListRule },
) => {
    if (disabled.value) return;
    saving.value = true;
    try {
        if (change && (change.operation === 'delete' || change.rule.type)) {
            const messages: string[] = [];
            if (change.operation !== 'delete' || !change.rule.type) {
                messages.push(
                    change.operation === 'delete'
                        ? i18n.global.t('commons.msg.delete')
                        : i18n.global.t('firewall.editRuleConfirm'),
                );
            }
            if (change.rule.type) {
                const service = `${serviceLabel(change.rule.type)} (${change.rule.port || servicePortLabel(change.rule.type)} / ${(change.rule.protocol || 'tcp').toUpperCase()})`;
                messages.push(i18n.global.t('firewall.systemAccessChangeConfirm', [service]));
            }
            const confirmed = await ElMessageBox.confirm(
                messages.join('\n'),
                i18n.global.t(`commons.button.${change.operation}`),
                {
                    confirmButtonText: i18n.global.t('commons.button.confirm'),
                    cancelButtonText: i18n.global.t('commons.button.cancel'),
                    type: 'warning',
                },
            )
                .then(() => true)
                .catch(() => false);
            if (!confirmed) return;
        }
        const { data: result } = await request();
        if (!result.taskID || !result.queued) {
            MsgError(i18n.global.t('commons.msg.operationFailed'));
            return;
        }
        dialogVisible.value = false;
        emit('created', result.taskID);
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
