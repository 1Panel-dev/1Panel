<template>
    <DrawerPro v-model="drawerVisible" :header="title" @close="handleClose" size="large">
        <el-form ref="formRef" label-position="top" :model="dialogData.rowData" :rules="rules" v-loading="loading">
            <el-form-item :label="$t('firewall.targetChain')" prop="chain">
                <el-select class="w-full" v-model="dialogData.rowData!.chain">
                    <el-option value="1PANEL_INPUT" :label="$t('firewall.inboundDirection')" />
                    <el-option value="1PANEL_OUTPUT" :label="$t('firewall.outboundDirection')" />
                </el-select>
            </el-form-item>

            <el-form-item :label="$t('firewall.ruleTemplate')">
                <el-select class="w-full" v-model="templateValue" clearable>
                    <el-option
                        v-for="item in ruleTemplates"
                        :key="item.value"
                        :label="$t(item.label)"
                        :value="item.value"
                    />
                </el-select>
            </el-form-item>

            <el-form-item :label="$t('commons.table.protocol')" prop="protocol">
                <el-select class="w-full" v-model="dialogData.rowData!.protocol">
                    <el-option value="all" label="ALL" />
                    <el-option value="tcp" label="TCP" />
                    <el-option value="udp" label="UDP" />
                    <el-option value="icmp" label="ICMP" />
                </el-select>
            </el-form-item>

            <el-form-item v-if="!isOutbound" :label="$t('firewall.sourceIP')" prop="sourceIP">
                <el-input clearable v-model.trim="dialogData.rowData!.sourceIP" placeholder="0.0.0.0/0" />
                <span class="input-help">{{ $t('firewall.sourceIPHelper') }}</span>
            </el-form-item>

            <el-form-item v-if="!isInbound" :label="$t('firewall.destIP')" prop="destIP">
                <el-input clearable v-model.trim="dialogData.rowData!.destIP" placeholder="0.0.0.0/0" />
                <span class="input-help">{{ $t('firewall.destIPHelper') }}</span>
            </el-form-item>

            <el-form-item :label="$t('firewall.destPort')" prop="destPort">
                <el-input-number
                    class="w-full"
                    v-model="dialogData.rowData!.destPort"
                    :min="0"
                    :max="65535"
                    controls-position="right"
                    :disabled="isProtocolAll"
                />
                <span class="input-help">{{ $t('firewall.portHelper') }}</span>
            </el-form-item>

            <el-form-item :label="$t('firewall.action')" prop="action">
                <el-radio-group v-model="dialogData.rowData!.action">
                    <el-radio value="ACCEPT">{{ $t('firewall.accept') }}</el-radio>
                    <el-radio value="DROP">{{ $t('firewall.drop') }}</el-radio>
                </el-radio-group>
            </el-form-item>

            <el-form-item :label="$t('commons.table.description')" prop="description">
                <el-input clearable v-model.trim="dialogData.rowData!.description" />
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="onSubmit(formRef)" :disabled="isDangerousRule">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { computed, reactive, ref, watch } from 'vue';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import { MsgSuccess } from '@/utils/message';
import { Host } from '@/api/interface/host';
import { operateFilterRule } from '@/api/modules/host';
import { checkCidr, checkCidrV6, checkIpV4V6 } from '@/utils/util';

const loading = ref();

interface DialogProps {
    title: string;
    rowData?: Host.IptablesFilterRuleOperate;
    getTableList?: () => Promise<any>;
}
const title = ref<string>('');
const drawerVisible = ref(false);
const dialogData = ref<DialogProps>({
    title: '',
});
const templateValue = ref<string>('');
const ruleTemplates = [
    { value: 'http', label: 'HTTP(80)', protocol: 'tcp', port: 80 },
    { value: 'https', label: 'HTTPS(443)', protocol: 'tcp', port: 443 },
    { value: 'mysql', label: 'MYSQL(3306)', protocol: 'tcp', port: 3306 },
] as const;

const isInbound = computed(() => dialogData.value.rowData?.chain === '1PANEL_INPUT');
const isOutbound = computed(() => dialogData.value.rowData?.chain === '1PANEL_OUTPUT');

const isDangerousRule = computed(() => {
    const rowData = dialogData.value.rowData;
    if (!rowData) return false;

    // 当规则为拒绝、端口为0、且IP留空时认为是危险规则
    if (rowData.action === 'DROP' && rowData.destPort === 0) {
        const isSourceIPEmpty = !rowData.sourceIP || rowData.sourceIP.trim() === '';
        const isDestIPEmpty = !rowData.destIP || rowData.destIP.trim() === '';

        if ((isInbound.value && isSourceIPEmpty) || (isOutbound.value && isDestIPEmpty)) {
            return true;
        }
    }
    return false;
});

const acceptParams = (params: DialogProps): void => {
    dialogData.value = params;
    title.value = i18n.global.t('firewall.' + dialogData.value.title);
    templateValue.value = detectTemplate(dialogData.value.rowData);
    drawerVisible.value = true;
};
const emit = defineEmits<{ (e: 'search'): void }>();

const handleClose = () => {
    drawerVisible.value = false;
};

const rules = reactive({
    chain: [Rules.requiredSelect],
    protocol: [Rules.requiredSelect],
    action: [Rules.requiredSelect],
    sourceIP: [{ validator: checkIPAddress, trigger: 'blur' }],
    destIP: [{ validator: checkIPAddress, trigger: 'blur' }],
});

const isProtocolAll = computed(() => dialogData.value.rowData?.protocol === 'all');

watch(
    () => dialogData.value.rowData?.chain,
    (chain) => {
        const rowData = dialogData.value.rowData;
        if (!rowData) return;
        if (chain === '1PANEL_INPUT') {
            rowData.destIP = '';
        } else if (chain === '1PANEL_OUTPUT') {
            rowData.sourceIP = '';
        }
    },
    { immediate: true },
);

watch(
    () => dialogData.value.rowData?.protocol,
    (protocol) => {
        if (protocol === 'all' && dialogData.value.rowData) {
            dialogData.value.rowData.sourcePort = 0;
            dialogData.value.rowData.destPort = 0;
        }
    },
    { immediate: true },
);

watch(templateValue, (value) => {
    const rowData = dialogData.value.rowData;
    if (!rowData || !value) {
        return;
    }
    const template = ruleTemplates.find((item) => item.value === value);
    if (!template) return;
    rowData.protocol = template.protocol;
    rowData.destPort = template.port;
});

watch(
    () => [dialogData.value.rowData?.protocol, dialogData.value.rowData?.destPort],
    ([protocol, port]) => {
        const matched = ruleTemplates.find((item) => item.protocol === protocol && item.port === Number(port));
        const nextValue = matched?.value || '';
        if (templateValue.value !== nextValue) {
            templateValue.value = nextValue;
        }
    },
    { immediate: true },
);

function checkIPAddress(_rule: any, value: any, callback: any) {
    if (!value) {
        return callback();
    }
    if (value.indexOf('/') !== -1) {
        if (value.indexOf(':') !== -1) {
            if (checkCidrV6(value)) {
                return callback(new Error(i18n.global.t('firewall.addressFormatError')));
            }
        } else {
            if (checkCidr(value)) {
                return callback(new Error(i18n.global.t('firewall.addressFormatError')));
            }
        }
    } else {
        if (checkIpV4V6(value)) {
            return callback(new Error(i18n.global.t('firewall.addressFormatError')));
        }
    }
    callback();
}

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        dialogData.value.rowData.operation = 'add';
        if (!dialogData.value.rowData) return;

        loading.value = true;
        await operateFilterRule(dialogData.value.rowData)
            .then(() => {
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                emit('search');
                drawerVisible.value = false;
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

defineExpose({
    acceptParams,
});

function detectTemplate(rowData?: Host.IptablesFilterRuleOperate) {
    if (!rowData) return '';
    const match = ruleTemplates.find(
        (item) => item.protocol === rowData.protocol && item.port === Number(rowData.destPort),
    );
    return match?.value || '';
}
</script>
