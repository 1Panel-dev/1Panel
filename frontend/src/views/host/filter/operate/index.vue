<template>
    <DrawerPro v-model="drawerVisible" :header="title" @close="handleClose" size="large">
        <el-form ref="formRef" label-position="top" :model="dialogData.rowData" :rules="rules" v-loading="loading">
            <el-form-item :label="$t('firewall.targetChain')" prop="chain">
                <el-select class="w-full" v-model="dialogData.rowData!.chain">
                    <el-option value="1PANEL_INPUT" label="1PANEL_INPUT" />
                    <el-option value="1PANEL_OUTPUT" label="1PANEL_OUTPUT" />
                </el-select>
            </el-form-item>

            <el-form-item :label="$t('commons.table.protocol')" prop="protocol">
                <el-select class="w-full" v-model="dialogData.rowData!.protocol">
                    <el-option value="all" label="all" />
                    <el-option value="tcp" label="tcp" />
                    <el-option value="udp" label="udp" />
                    <el-option value="icmp" label="icmp" />
                </el-select>
            </el-form-item>

            <el-form-item :label="$t('firewall.sourceIP')" prop="sourceIP">
                <el-input clearable v-model.trim="dialogData.rowData!.sourceIP" />
                <span class="input-help">{{ $t('firewall.sourceIPHelper') }}</span>
            </el-form-item>

            <el-form-item :label="$t('firewall.sourcePort')" prop="sourcePort">
                <el-input-number
                    class="w-full"
                    v-model="dialogData.rowData!.sourcePort"
                    :min="0"
                    :max="65535"
                    controls-position="right"
                />
                <span class="input-help">{{ $t('firewall.portHelper') }}</span>
            </el-form-item>

            <el-form-item :label="$t('firewall.destIP')" prop="destIP">
                <el-input clearable v-model.trim="dialogData.rowData!.destIP" />
                <span class="input-help">{{ $t('firewall.destIPHelper') }}</span>
            </el-form-item>

            <el-form-item :label="$t('firewall.destPort')" prop="destPort">
                <el-input-number
                    class="w-full"
                    v-model="dialogData.rowData!.destPort"
                    :min="0"
                    :max="65535"
                    controls-position="right"
                />
                <span class="input-help">{{ $t('firewall.portHelper') }}</span>
            </el-form-item>

            <el-form-item :label="$t('firewall.action')" prop="action">
                <el-radio-group v-model="dialogData.rowData!.action">
                    <el-radio value="ACCEPT">{{ $t('firewall.accept') }}</el-radio>
                    <el-radio value="DROP">阻止</el-radio>
                    <el-radio value="REJECT">{{ $t('firewall.reject') }}</el-radio>
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
    rowData?: Host.IptablesFilterRuleOperate;
    getTableList?: () => Promise<any>;
}
const title = ref<string>('');
const drawerVisible = ref(false);
const dialogData = ref<DialogProps>({
    title: '',
});
const acceptParams = (params: DialogProps): void => {
    dialogData.value = params;
    title.value = i18n.global.t('firewall.' + dialogData.value.title);
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
</script>
