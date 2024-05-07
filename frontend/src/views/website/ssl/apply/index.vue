<template>
    <el-dialog
        v-model="open"
        :title="$t('ssl.apply')"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        width="50%"
        :before-close="handleClose"
    >
        <div v-if="loading">
            <el-alert type="info" :closable="false" center>{{ $t('ssl.getDnsResolve') }}</el-alert>
        </div>

        <div v-if="dnsResolve.length > 0">
            <span>{{ $t('ssl.dnsResolveHelper') }}</span>
            <el-table :data="dnsResolve" border :table-layout="'auto'">
                <el-table-column prop="domain" :label="$t('website.domain')" />
                <el-table-column prop="resolve" :label="$t('ssl.resolveDomain')">
                    <template #default="{ row }">
                        <span>{{ row.resolve }}</span>
                        <CopyButton :content="row.resolve" type="icon" />
                    </template>
                </el-table-column>
                <el-table-column prop="value" :label="$t('ssl.value')">
                    <template #default="{ row }">
                        <span>{{ row.value }}</span>
                        <CopyButton :content="row.value" type="icon" />
                    </template>
                </el-table-column>
                <el-table-column :label="$t('commons.table.type')">TXT</el-table-column>
            </el-table>
        </div>

        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit()" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script lang="ts" setup>
import { Website } from '@/api/interface/website';
import { GetDnsResolve, ObtainSSL } from '@/api/modules/website';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { ref } from 'vue';

interface RenewProps {
    ssl: Website.SSL;
}

const open = ref(false);
const loading = ref(false);
const dnsResolve = ref<Website.DNSResolve[]>([]);
const sslID = ref(0);
const em = defineEmits(['close', 'submit']);
const handleClose = () => {
    open.value = false;
    em('close', false);
};

const acceptParams = async (props: RenewProps) => {
    open.value = true;
    dnsResolve.value = [];
    sslID.value = props.ssl.id;
    getDnsResolve(props.ssl);
};

const getDnsResolve = async (row: Website.SSL) => {
    loading.value = true;

    let domains = [row.primaryDomain];
    if (row.domains != '') {
        let otherDomains = row.domains.split(',');
        domains = domains.concat(otherDomains);
    }
    try {
        const res = await GetDnsResolve({ acmeAccountId: row.acmeAccountId, domains: domains });
        if (res.data) {
            dnsResolve.value = res.data;
        }
    } finally {
        loading.value = false;
    }
};

const submit = () => {
    ObtainSSL({ ID: sslID.value })
        .then(() => {
            MsgSuccess(i18n.global.t('ssl.applyStart'));
            handleClose();
            em('submit', sslID.value);
        })
        .finally(() => {});
};

defineExpose({
    acceptParams,
});
</script>
