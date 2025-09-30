<template>
    <DrawerPro v-model="drawerVisible" :header="title" @close="handleClose" size="large">
        <el-form ref="formRef" label-position="top" :model="dialogData.rowData" :rules="rules" v-loading="loading">
            <el-form-item :label="$t('commons.table.protocol')" prop="protocol">
                <el-select class="w-full" v-model="dialogData.rowData!.protocol">
                    <el-option value="tcp" label="tcp" />
                    <el-option value="udp" label="udp" />
                    <el-option value="tcp/udp" label="tcp/udp" />
                </el-select>
            </el-form-item>

            <el-form-item :label="$t('firewall.sourcePort')" prop="port">
                <el-input clearable v-model.trim="dialogData.rowData!.port" />
            </el-form-item>

            <el-form-item :label="$t('firewall.targetIP')" prop="targetIP">
                <el-input v-model.trim="dialogData.rowData!.targetIP" />
                <span class="input-help">{{ $t('firewall.forwardHelper1') }}</span>
                <span class="input-help">{{ $t('firewall.forwardHelper2') }}</span>
                <span class="input-help">{{ $t('firewall.forwardHelper3') }}</span>
            </el-form-item>

            <el-form-item :label="$t('firewall.targetPort')" prop="targetPort">
                <el-input clearable v-model.trim="dialogData.rowData!.targetPort" />
            </el-form-item>

            <template v-if="dialogData.fireName === 'ufw'">
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
            </template>
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
import { operateForwardRule, getNetworkOptions } from '@/api/modules/host';
import { checkCidr, checkCidrV6, checkIp, checkPort, deepCopy } from '@/utils/util';

const loading = ref();
const oldRule = ref<Host.RuleForward>();

const interfaceOptions = ref<Array<{ label: string; value: string }>>([]);

interface DialogProps {
    title: string;
    rowData?: Host.RuleForward;
    fireName?: string;
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
    if (dialogData.value.fireName === 'ufw') {
        getNetworkOptions().then((res) => {
            interfaceOptions.value = res.data.map((item) => ({ label: item, value: item }));
            dialogData.value.rowData!.interface = dialogData.value.rowData!.interface || 'all';
        });
    }
    title.value = i18n.global.t('firewall.' + dialogData.value.title);
    drawerVisible.value = true;
};
const emit = defineEmits<{ (e: 'search'): void }>();

const handleClose = () => {
    drawerVisible.value = false;
};

const rules = reactive({
    protocol: [Rules.requiredSelect],
    port: [{ validator: checkPortRule, trigger: 'blur', required: true }],
    targetPort: [{ validator: checkPortRule, trigger: 'blur', required: true }],
    targetIP: [{ validator: checkAddress, trigger: 'blur' }],
});

function checkPortRule(rule: any, value: string, callback: any) {
    if (!value) {
        return callback(new Error(i18n.global.t('firewall.portFormatError')));
    }
    if (checkPort(value)) {
        return callback(new Error(i18n.global.t('firewall.portFormatError')));
    }
    callback();
}
function checkAddress(rule: any, value: string, callback: any) {
    if (!value) {
        return callback();
    }
    let addrs = value.split(',');
    for (const item of addrs) {
        if (item.indexOf('/') !== -1) {
            if (item.indexOf(':') !== -1) {
                if (checkCidrV6(item)) {
                    return callback(new Error(i18n.global.t('firewall.addressFormatError')));
                }
            } else {
                if (checkCidr(item)) {
                    return callback(new Error(i18n.global.t('firewall.addressFormatError')));
                }
            }
        } else {
            if (checkIp(item)) {
                return callback(new Error(i18n.global.t('firewall.addressFormatError')));
            }
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
        const { rowData } = dialogData.value;
        let rules = [];
        if (!rowData) return;
        rowData.operation = 'add';
        if (rowData.targetIP === '') {
            rowData.targetIP = '127.0.0.1';
        }
        if (rowData.interface === 'all') {
            rowData.interface = '';
        }
        rules.push(rowData);
        loading.value = true;
        if (dialogData.value.title === 'create') {
            await operateForwardRule({ rules: rules })
                .then(() => {
                    loading.value = false;
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                    emit('search');
                    drawerVisible.value = false;
                })
                .catch(() => {
                    loading.value = false;
                });
            return;
        }
        rules = [];
        oldRule.value.operation = 'remove';
        dialogData.value.rowData.operation = 'add';
        rules.push(oldRule.value);
        rules.push(dialogData.value.rowData);
        await operateForwardRule({ rules: rules })
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
