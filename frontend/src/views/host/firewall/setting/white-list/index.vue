<template>
    <DrawerPro v-model="drawerVisible" :header="$t('firewall.portWhiteList')" @close="handleClose" size="large">
        <template #content>
            <el-button type="primary" @click="openCreate">
                {{ $t('commons.button.add') }}
            </el-button>
            <ComplexTable :data="data" v-loading="loading">
                <el-table-column :label="$t('firewall.addressFamily')" width="120">
                    <template #default="{ row }">
                        <span v-if="!row.edit">{{ row.family.toUpperCase() }}</span>
                        <el-select v-else v-model="row.family">
                            <el-option value="ipv4" label="IPv4" />
                            <el-option value="ipv6" label="IPv6" />
                        </el-select>
                    </template>
                </el-table-column>
                <el-table-column :label="$t('commons.table.protocol')" width="120">
                    <template #default="{ row }">
                        <span v-if="!row.edit">{{ row.protocol.toUpperCase() }}</span>
                        <el-select v-else v-model="row.protocol">
                            <el-option value="tcp" label="TCP" />
                            <el-option value="udp" label="UDP" />
                        </el-select>
                    </template>
                </el-table-column>
                <el-table-column :label="$t('firewall.portOrRange')" prop="port">
                    <template #default="{ row }">
                        <span v-if="!row.edit">{{ row.port }}</span>
                        <el-input v-else v-model.trim="row.port" placeholder="80 / 8000-8100" clearable />
                    </template>
                </el-table-column>
                <el-table-column :label="$t('commons.table.operate')" width="160">
                    <template #default="{ row, $index }">
                        <el-button v-if="row.edit" link type="primary" @click="saveRow(row)">
                            {{ $t('commons.button.save') }}
                        </el-button>
                        <el-button v-if="!row.edit" link type="primary" @click="editRow(row)">
                            {{ $t('commons.button.edit') }}
                        </el-button>
                        <el-button v-if="!row.edit" link type="primary" @click="removeRow($index)">
                            {{ $t('commons.button.delete') }}
                        </el-button>
                        <el-button v-if="row.edit" link type="primary" @click="cancelEdit(row, $index)">
                            {{ $t('commons.button.cancel') }}
                        </el-button>
                    </template>
                </el-table-column>
            </ComplexTable>
            <span class="input-help">{{ $t('firewall.portWhiteListHelper') }}</span>
        </template>
        <template #footer>
            <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
            <el-button :loading="loading" type="primary" @click="onSubmit">
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { getAgentSettingInfo } from '@/api/modules/setting';
import { updateFirewallPortWhitelist } from '@/api/modules/firewall';
import i18n from '@/lang';
import { MsgError } from '@/utils/message';
import {
    normalizeWhiteListRule,
    parseWhiteList,
    serializeWhiteList,
    WhiteListFamily,
    WhiteListProtocol,
    WhiteListRule,
    whiteListRuleKey,
} from './model';

interface WhiteListItem extends WhiteListRule {
    oldRule: WhiteListRule;
    edit: boolean;
    isNew: boolean;
}
const emit = defineEmits<{ (e: 'created', taskID: string): void }>();

const drawerVisible = ref(false);
const loading = ref(false);
const data = ref<WhiteListItem[]>([]);
const defaultWhiteList = '80/tcp,443/tcp,443/udp';

const acceptParams = async (): Promise<void> => {
    drawerVisible.value = true;
    loading.value = true;
    await getAgentSettingInfo()
        .then((res) => {
            data.value = parseWhiteList(res.data.firewallPortWhiteList ?? defaultWhiteList).map((rule) => ({
                ...rule,
                oldRule: { ...rule },
                edit: false,
                isNew: false,
            }));
        })
        .finally(() => {
            loading.value = false;
        });
};

const openCreate = () => {
    for (const item of data.value) {
        if (item.isNew && item.port === '') {
            return;
        }
    }
    data.value.unshift({
        family: 'ipv4',
        protocol: 'tcp',
        port: '',
        oldRule: { family: 'ipv4', protocol: 'tcp', port: '' },
        edit: true,
        isNew: true,
    });
};

const editRow = (row: WhiteListItem) => {
    row.oldRule = { family: row.family, protocol: row.protocol, port: row.port };
    row.edit = true;
};

const saveRow = (row: WhiteListItem) => {
    const rule = validateRule(row);
    if (!rule) return;
    if (hasDuplicate(rule, row)) {
        MsgError(i18n.global.t('commons.rule.duplicate'));
        return;
    }
    Object.assign(row, rule);
    row.oldRule = { ...rule };
    row.edit = false;
    row.isNew = false;
};

const cancelEdit = (row: WhiteListItem, index: number) => {
    if (row.isNew) {
        data.value.splice(index, 1);
        return;
    }
    Object.assign(row, row.oldRule);
    row.edit = false;
};

const removeRow = (index: number) => {
    data.value.splice(index, 1);
};

const validateRule = (row: Pick<WhiteListItem, 'family' | 'protocol' | 'port'>): WhiteListRule | undefined => {
    try {
        return normalizeWhiteListRule({
            family: row.family as WhiteListFamily,
            protocol: row.protocol as WhiteListProtocol,
            port: row.port,
        });
    } catch {
        MsgError(i18n.global.t('firewall.portFormatError'));
        return undefined;
    }
};

const hasDuplicate = (rule: WhiteListRule, row?: WhiteListItem): boolean => {
    return data.value.some(
        (item) => item !== row && item.port !== '' && whiteListRuleKey(rule) === whiteListRuleKey(item),
    );
};

const validateRules = (): WhiteListRule[] | undefined => {
    const rules: WhiteListRule[] = [];
    for (const item of data.value) {
        if (!item.port) continue;
        const rule = validateRule(item);
        if (!rule) return undefined;
        if (rules.some((existing) => whiteListRuleKey(existing) === whiteListRuleKey(rule))) {
            MsgError(i18n.global.t('commons.rule.duplicate'));
            return undefined;
        }
        rules.push(rule);
    }
    return rules;
};

const onSubmit = async () => {
    const rules = validateRules();
    if (!rules || loading.value) return;
    loading.value = true;
    try {
        const { data: result } = await updateFirewallPortWhitelist(serializeWhiteList(rules));
        if (!result.taskID || !result.queued) return;
        drawerVisible.value = false;
        emit('created', result.taskID);
    } finally {
        loading.value = false;
    }
};

const handleClose = () => {
    drawerVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>
