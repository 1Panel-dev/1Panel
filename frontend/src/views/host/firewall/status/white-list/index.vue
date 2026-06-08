<template>
    <DrawerPro v-model="drawerVisible" :header="$t('firewall.portWhiteList')" @close="handleClose" size="small">
        <template #content>
            <el-alert type="info" :closable="false" :title="$t('firewall.portWhiteListAlter')" />

            <el-button class="mt-5" type="primary" @click="openCreate">
                {{ $t('commons.button.add') }}
            </el-button>
            <ComplexTable :data="data" v-loading="loading">
                <el-table-column :label="$t('commons.table.port')" prop="port">
                    <template #default="{ row }">
                        <span v-if="!row.edit">{{ row.port }}</span>
                        <el-input v-else v-model.trim="row.port" placeholder="80/tcp" clearable />
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
import { getAgentSettingInfo, updateAgentSetting } from '@/api/modules/setting';
import i18n from '@/lang';
import { MsgError, MsgSuccess } from '@/utils/message';
import { checkPort } from '@/utils/validate';

interface WhiteListItem {
    port: string;
    oldPort: string;
    edit: boolean;
    isNew: boolean;
}
const emit = defineEmits<{ (e: 'search'): void }>();

const drawerVisible = ref(false);
const loading = ref(false);
const data = ref<WhiteListItem[]>([]);
const defaultWhiteList = '80/tcp,443/tcp,443/udp';

const acceptParams = async (): Promise<void> => {
    drawerVisible.value = true;
    loading.value = true;
    await getAgentSettingInfo()
        .then((res) => {
            data.value = parseWhiteList(res.data.firewallPortWhiteList ?? defaultWhiteList);
        })
        .finally(() => {
            loading.value = false;
        });
};

const parseWhiteList = (value: string): WhiteListItem[] => {
    return value
        .split(/[\s,;]+/)
        .filter((item) => item !== '')
        .map((item) => ({
            port: item,
            oldPort: item,
            edit: false,
            isNew: false,
        }));
};

const openCreate = () => {
    for (const item of data.value) {
        if (item.isNew && item.port === '') {
            return;
        }
    }
    data.value.unshift({
        port: '',
        oldPort: '',
        edit: true,
        isNew: true,
    });
};

const editRow = (row: WhiteListItem) => {
    row.oldPort = row.port;
    row.edit = true;
};

const saveRow = (row: WhiteListItem) => {
    if (!validatePort(row.port)) {
        return;
    }
    row.port = normalizePort(row.port);
    row.oldPort = row.port;
    row.edit = false;
    row.isNew = false;
};

const cancelEdit = (row: WhiteListItem, index: number) => {
    if (row.isNew) {
        data.value.splice(index, 1);
        return;
    }
    row.port = row.oldPort;
    row.edit = false;
};

const removeRow = (index: number) => {
    data.value.splice(index, 1);
};

const normalizePort = (value: string): string => {
    const segments = value.split('/');
    const port = segments[0];
    const protocol = segments[1] ? segments[1].toLowerCase() : '';
    return protocol ? `${port}/${protocol}` : port;
};

const validatePort = (value: string): boolean => {
    if (value === '') {
        MsgError(i18n.global.t('firewall.portFormatError'));
        return false;
    }
    const segments = value.split('/');
    const [port, protocol = 'tcp'] = segments;
    if (checkPort(port) || !['tcp', 'udp'].includes(protocol.toLowerCase()) || segments.length > 2) {
        MsgError(i18n.global.t('firewall.portFormatError'));
        return false;
    }
    return true;
};

const onSubmit = async () => {
    const ports = data.value.map((item) => item.port).filter((item) => item !== '');
    for (const port of ports) {
        if (!validatePort(port)) {
            return;
        }
    }
    loading.value = true;
    await updateAgentSetting({
        key: 'FirewallPortWhiteList',
        value: ports.map((item) => normalizePort(item)).join('\n'),
    })
        .then(() => {
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            emit('search');
            drawerVisible.value = false;
        })
        .finally(() => {
            loading.value = false;
        });
};

const handleClose = () => {
    drawerVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>
