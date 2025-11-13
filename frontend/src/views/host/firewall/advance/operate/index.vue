<template>
    <DrawerPro v-model="drawerVisible" :header="title" @close="handleClose" size="large">
        <div class="mb-2">
            <el-alert :closable="false" :title="$t('firewall.ipv4Limit')" />
        </div>
        <el-form ref="formRef" label-position="top" :model="dialogData.rowData" :rules="rules" v-loading="loading">
            <el-form-item :label="$t('commons.table.protocol')" prop="protocol">
                <el-select class="w-full" v-model="dialogData.rowData!.protocol" @change="changeProtocol">
                    <el-option value="all" label="all" />
                    <el-option value="tcp" label="tcp" />
                    <el-option value="udp" label="udp" />
                    <el-option value="icmp" label="icmp" />
                </el-select>
            </el-form-item>

            <el-form-item
                v-if="dialogData.rowData?.chain === '1PANEL_INPUT'"
                :label="$t('firewall.sourceIP')"
                prop="srcIP"
            >
                <el-input clearable v-model.trim="dialogData.rowData!.srcIP" placeholder="0.0.0.0/0" />
                <span class="input-help">{{ $t('firewall.sourceIPHelper') }}</span>
            </el-form-item>

            <el-form-item
                v-if="dialogData.rowData?.chain === '1PANEL_OUTPUT'"
                :label="$t('firewall.destIP')"
                prop="dstIP"
            >
                <el-input clearable v-model.trim="dialogData.rowData!.dstIP" placeholder="0.0.0.0/0" />
                <span class="input-help">{{ $t('firewall.destIPHelper') }}</span>
            </el-form-item>

            <el-form-item :label="$t('firewall.destPort')" prop="dstPort">
                <el-input-number
                    class="w-full"
                    v-model="dialogData.rowData!.dstPort"
                    :min="0"
                    :max="65535"
                    :disabled="dialogData.rowData?.protocol === 'all'"
                />
                <span class="input-help">{{ $t('firewall.portHelper') }}</span>
            </el-form-item>

            <el-form-item :label="$t('firewall.action')" prop="strategy">
                <el-radio-group v-model="dialogData.rowData!.strategy">
                    <el-radio value="accept">{{ $t('firewall.accept') }}</el-radio>
                    <el-radio value="drop">{{ $t('firewall.drop') }}</el-radio>
                </el-radio-group>
            </el-form-item>

            <el-form-item :label="$t('commons.table.description')" prop="description">
                <el-input clearable v-model.trim="dialogData.rowData!.description" />
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="onSubmit(formRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
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
    rowData?: Host.IptablesFilterRuleOp;
}
const title = ref<string>('');
const drawerVisible = ref(false);
const dialogData = ref<DialogProps>({
    title: '',
});

const acceptParams = (params: DialogProps): void => {
    dialogData.value = params;
    title.value = i18n.global.t('firewall.' + dialogData.value.title);
    if (dialogData.value.rowData.chain === '1PANEL_INPUT') {
        dialogData.value.rowData.dstIP = '';
    } else if (dialogData.value.rowData.chain === '1PANEL_OUTPUT') {
        dialogData.value.rowData.srcIP = '';
    }
    drawerVisible.value = true;
};
const emit = defineEmits<{ (e: 'search'): void }>();

const handleClose = () => {
    drawerVisible.value = false;
};

const rules = reactive({
    chain: [Rules.requiredSelect],
    protocol: [Rules.requiredSelect],
    strategy: [Rules.requiredSelect],
    srcIP: [{ validator: checkIPAddress, trigger: 'blur' }],
    dstIP: [{ validator: checkIPAddress, trigger: 'blur' }],
});

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

const changeProtocol = () => {
    if (dialogData.value.rowData.protocol === 'all') {
        dialogData.value.rowData.srcPort = 0;
        dialogData.value.rowData.dstPort = 0;
    }
};

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
</script>
