<template>
    <DrawerPro v-model="open" :header="$t('commons.button.' + mode)" size="large" @close="handleClose">
        <el-form v-loading="loading" ref="recordForm" label-position="top" :model="record" :rules="rules">
            <el-form-item :label="$t('commons.table.type')" prop="type">
                <el-select v-model="record.type">
                    <el-option v-for="t in recordTypes" :key="t" :label="t" :value="t"></el-option>
                </el-select>
            </el-form-item>
            <el-form-item :label="$t('website.dnsRecordHost')" prop="name">
                <el-input v-model.trim="record.name" placeholder="@">
                    <template #append>.{{ zoneName }}</template>
                </el-input>
                <span class="input-help">{{ $t('website.dnsRecordHostHelper') }}</span>
            </el-form-item>
            <el-form-item :label="$t('website.dnsRecordValue')" prop="value">
                <div class="flex w-full gap-2">
                    <el-input
                        class="flex-1"
                        v-model.trim="record.value"
                        :type="accountType === 'AWSRoute53' ? 'textarea' : 'text'"
                    />
                    <el-button v-if="['A', 'AAAA'].includes(record.type)" :loading="ipLoading" @click="useServerIP">
                        {{ $t('website.dnsRecordUseIP') }}
                    </el-button>
                </div>
                <span class="input-help" v-if="accountType === 'AWSRoute53'">
                    {{ $t('website.dnsRecordValueHelper') }}
                </span>
            </el-form-item>
            <el-form-item :label="$t('website.dnsRecordTTL')" prop="ttl">
                <el-input-number v-model="record.ttl" :min="0" :max="2147483647" />
                <span class="input-help" v-if="accountType === 'CloudFlare'">
                    {{ $t('website.dnsRecordCfTTLHelper') }}
                </span>
            </el-form-item>
            <el-form-item :label="$t('website.dnsRecordPriority')" prop="priority" v-if="record.type === 'MX'">
                <el-input-number v-model="record.priority" :min="0" :max="65535" />
            </el-form-item>
            <el-form-item
                :label="$t('website.dnsRecordProxied')"
                prop="proxied"
                v-if="accountType === 'CloudFlare' && ['A', 'AAAA', 'CNAME'].includes(record.type)"
            >
                <el-switch v-model="record.proxied" />
            </el-form-item>
            <el-form-item :label="$t('website.dnsRecordRemark')" prop="remark" v-if="accountType === 'CloudFlare'">
                <el-input v-model.trim="record.remark" />
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button v-permission type="primary" @click="submit(recordForm)" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { createDnsRecord, updateDnsRecord } from '@/api/modules/website';
import { Website } from '@/api/interface/website';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { MsgError, MsgSuccess } from '@/utils/message';
import { getAgentSettingInfo } from '@/api/modules/setting';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { FormInstance } from 'element-plus';
import { ref } from 'vue';

interface RecordProps {
    mode: string;
    accountId: number;
    accountType: string;
    zone: Website.DnsZone;
    form?: Website.DnsRecord;
}

const recordTypes = ['A', 'AAAA', 'CNAME', 'TXT', 'MX', 'NS'];
const open = ref(false);
const loading = ref(false);
const mode = ref('create');
const accountId = ref(0);
const accountType = ref('');
const zoneId = ref('');
const zoneName = ref('');
const recordForm = ref<FormInstance>();
const rules = ref<any>({
    type: [Rules.requiredSelect],
    name: [Rules.requiredInput],
    value: [Rules.requiredInput],
});

const emptyRecord = (): Website.DnsRecord => ({
    id: '',
    name: '@',
    type: 'A',
    value: '',
    ttl: 600,
    priority: 10,
    proxied: false,
    remark: '',
});
const record = ref<Website.DnsRecord>(emptyRecord());
const em = defineEmits(['saved']);

const { currentNodeAddr } = useGlobalStore();
const ipLoading = ref(false);

const useServerIP = async () => {
    ipLoading.value = true;
    try {
        const res = await getAgentSettingInfo();
        const ip = res.data.systemIP || currentNodeAddr.value || '';
        if (!ip) {
            MsgError(i18n.global.t('website.dnsRecordUseIPEmpty'));
            return;
        }
        const isV4 = /^(\d{1,3}\.){3}\d{1,3}$/.test(ip);
        const isV6 = ip.includes(':');
        if ((record.value.type === 'A' && !isV4) || (record.value.type === 'AAAA' && !isV6)) {
            MsgError(i18n.global.t('website.dnsRecordUseIPMismatch', [ip]));
            return;
        }
        record.value.value = ip;
        recordForm.value?.clearValidate('value');
    } finally {
        ipLoading.value = false;
    }
};

const handleClose = () => {
    recordForm.value?.resetFields();
    record.value = emptyRecord();
    open.value = false;
};

const acceptParams = (props: RecordProps) => {
    record.value = emptyRecord();
    mode.value = props.mode;
    accountId.value = props.accountId;
    accountType.value = props.accountType;
    zoneId.value = props.zone.id;
    zoneName.value = props.zone.name;
    if (props.mode === 'edit' && props.form) {
        record.value = { ...props.form };
    } else if (props.accountType === 'CloudFlare') {
        record.value.ttl = 1;
    }
    open.value = true;
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl || loading.value) return;
    loading.value = true;
    try {
        if (!(await formEl.validate().catch(() => false))) return;
        const req: Website.DnsRecordOperate = {
            ...record.value,
            accountId: accountId.value,
            zoneId: zoneId.value,
            zoneName: zoneName.value,
        };
        const save = mode.value === 'edit' ? updateDnsRecord : createDnsRecord;
        await save(req);
        MsgSuccess(i18n.global.t(mode.value === 'edit' ? 'commons.msg.updateSuccess' : 'commons.msg.createSuccess'));
        handleClose();
        em('saved');
    } finally {
        loading.value = false;
    }
};

defineExpose({
    acceptParams,
});
</script>
