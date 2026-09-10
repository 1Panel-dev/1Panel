<template>
    <DrawerPro v-model="drawerVisible" :header="title" @close="handleClose" size="large">
        <el-form ref="formRef" label-position="top" :model="dialogData.rowData" :rules="rules" v-loading="loading">
            <el-form-item label="IP" prop="family">
                <el-select class="w-full" v-model="dialogData.rowData!.family">
                    <el-option value="ipv4" label="IPv4" />
                    <el-option value="ipv6" label="IPv6" />
                </el-select>
            </el-form-item>

            <el-form-item :label="$t('commons.table.protocol')" prop="protocol">
                <el-select class="w-full" v-model="dialogData.rowData!.protocol">
                    <el-option value="tcp" label="tcp" />
                    <el-option value="udp" label="udp" />
                    <el-option value="tcp/udp" label="tcp/udp" />
                </el-select>
            </el-form-item>

            <el-form-item :label="$t('firewall.sourcePort')" prop="port">
                <el-input clearable v-model.trim="dialogData.rowData!.port" />
                <span class="input-help">{{ $t('firewall.forwardPortHelper') }}</span>
            </el-form-item>

            <el-form-item :label="$t('firewall.targetIP')" prop="targetIP">
                <el-input v-model.trim="dialogData.rowData!.targetIP" />
                <span class="input-help">{{ $t('firewall.forwardHelper1') }}</span>
                <span class="input-help">{{ $t('firewall.forwardHelper2') }}</span>
            </el-form-item>

            <el-form-item :label="$t('firewall.targetPort')" prop="targetPort">
                <el-input clearable v-model.trim="dialogData.rowData!.targetPort" />
                <span class="input-help">{{ $t('firewall.forwardPortHelper') }}</span>
            </el-form-item>

            <el-form-item :label="$t('firewall.forwardInboundInterface')" prop="interface">
                <el-select class="w-full" v-model="dialogData.rowData!.interface">
                    <el-option
                        v-for="item in interfaceOptions"
                        :key="item.value"
                        :label="item.label === 'all' ? $t('commons.table.all') : item.label"
                        :value="item.value"
                    />
                </el-select>
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
import { getErrorMessage } from '@/utils/misc';
import { isAxiosError } from 'axios';
import { ElForm } from 'element-plus';
import { MsgError } from '@/utils/message';
import { Firewall } from '@/api/interface/firewall';
import { getNetworkOptions } from '@/api/modules/host';
import { operateForwardRule } from '@/api/modules/firewall';
import { deepCopy } from '@/utils/misc';
import { isValidAddressForFamily, isValidPortRange, normalizePortRange } from '@/views/host/firewall/utils/validation';
const loading = ref();
const oldRule = ref<Firewall.RuleForward>();

const interfaceOptions = ref<Array<{ label: string; value: string }>>([]);

interface DialogProps {
    title: string;
    rowData?: Firewall.RuleForward;
    getTableList?: () => Promise<any>;
}
const title = ref<string>('');
const drawerVisible = ref(false);
const dialogData = ref<DialogProps>({
    title: '',
});
const acceptParams = (params: DialogProps): void => {
    dialogData.value = params;
    if (dialogData.value.title === 'edit') {
        oldRule.value = deepCopy(params.rowData);
    }
    getNetworkOptions().then((res) => {
        interfaceOptions.value = res.data.map((item) => ({ label: item, value: item }));
        dialogData.value.rowData!.interface = dialogData.value.rowData!.interface || 'all';
    });
    title.value = i18n.global.t('firewall.' + dialogData.value.title);
    drawerVisible.value = true;
};
const emit = defineEmits<{ (e: 'created', taskID: string): void }>();

const handleClose = () => {
    drawerVisible.value = false;
};

const rules = reactive({
    family: [Rules.requiredSelect],
    protocol: [Rules.requiredSelect],
    port: [{ validator: checkPortRule, trigger: 'blur', required: true }],
    targetPort: [{ validator: checkPortRule, trigger: 'blur', required: true }],
    targetIP: [{ validator: checkAddress, trigger: 'blur' }],
});

function checkPortRule(rule: any, value: string, callback: any) {
    if (!isValidPortRange(value)) {
        return callback(new Error(i18n.global.t('firewall.portFormatError')));
    }
    callback();
}

function checkAddress(rule: any, value: string, callback: any) {
    if (!value) {
        return callback();
    }
    const family = dialogData.value.rowData?.family || 'ipv4';
    if (!isValidAddressForFamily(family, value, false)) {
        return callback(new Error(i18n.global.t('firewall.addressFormatError')));
    }
    callback();
}

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl || loading.value) return;
    const { rowData } = dialogData.value;
    if (!rowData) return;
    if (!isValidPortRange(rowData.port) || !isValidPortRange(rowData.targetPort)) {
        await formEl.validateField(['port', 'targetPort']).catch(() => undefined);
        MsgError(i18n.global.t('commons.rule.port'));
        return;
    }
    const valid = await formEl.validate().catch(() => false);
    if (!valid) return;
    rowData.port = normalizePortRange(rowData.port);
    rowData.targetPort = normalizePortRange(rowData.targetPort);
    rowData.operation = 'add';
    if (rowData.targetIP === '') {
        rowData.targetIP = rowData.family === 'ipv6' ? '::1' : '127.0.0.1';
    }
    if (rowData.interface === 'all') {
        rowData.interface = '';
    }
    const operations: Firewall.RuleForward[] = [];
    if (dialogData.value.title !== 'create') {
        operations.push({ ...oldRule.value, operation: 'remove' });
    }
    operations.push(rowData);
    loading.value = true;
    try {
        const result = (await operateForwardRule({ rules: operations })).data;
        if (!result.taskID || !result.queued) {
            MsgError(i18n.global.t('commons.msg.operationFailed'));
            return;
        }
        drawerVisible.value = false;
        emit('created', result.taskID);
    } catch (error) {
        MsgError(
            (isAxiosError(error) && error.response?.data?.message) ||
                (error && getErrorMessage(error)) ||
                i18n.global.t('commons.res.commonError'),
        );
    } finally {
        loading.value = false;
    }
};

defineExpose({
    acceptParams,
});
</script>
