<template>
    <el-drawer
        v-model="drawerVisible"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        size="50%"
    >
        <template #header>
            <DrawerHeader :header="title" :back="handleClose" />
        </template>
        <div v-loading="loading">
            <el-form ref="formRef" label-position="top" :model="dialogData.rowData" :rules="rules">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item :label="$t('commons.table.protocol')" prop="protocol">
                            <el-select style="width: 100%" v-model="dialogData.rowData!.protocol">
                                <el-option value="tcp" label="tcp" />
                                <el-option value="udp" label="udp" />
                                <el-option value="tcp/udp" label="tcp/udp" />
                            </el-select>
                        </el-form-item>

                        <el-form-item :label="$t('commons.table.port')" prop="port">
                            <el-input
                                :disabled="dialogData.title === 'edit'"
                                clearable
                                v-model.trim="dialogData.rowData!.port"
                            />
                            <span class="input-help">{{ $t('firewall.portHelper1') }}</span>
                            <span class="input-help">{{ $t('firewall.portHelper2') }}</span>
                        </el-form-item>

                        <el-form-item :label="$t('firewall.source')" prop="source">
                            <el-radio-group v-model="dialogData.rowData!.source">
                                <el-radio value="anyWhere">{{ $t('firewall.anyWhere') }}</el-radio>
                                <el-radio value="address">{{ $t('firewall.address') }}</el-radio>
                            </el-radio-group>
                        </el-form-item>

                        <el-form-item
                            :label="$t('firewall.address')"
                            v-if="dialogData.rowData!.source === 'address'"
                            prop="address"
                        >
                            <el-input v-model.trim="dialogData.rowData!.address" />
                            <span class="input-help">{{ $t('firewall.addressHelper1') }}</span>
                            <span class="input-help">{{ $t('firewall.addressHelper2') }}</span>
                        </el-form-item>

                        <el-form-item :label="$t('firewall.strategy')" prop="strategy">
                            <el-radio-group v-model="dialogData.rowData!.strategy">
                                <el-radio value="accept">{{ $t('firewall.accept') }}</el-radio>
                                <el-radio value="drop">{{ $t('firewall.drop') }}</el-radio>
                            </el-radio-group>
                        </el-form-item>
                        <el-form-item :label="$t('commons.table.description')" prop="description">
                            <el-input clearable v-model.trim="dialogData.rowData!.description" />
                        </el-form-item>
                    </el-col>
                </el-row>
            </el-form>
        </div>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="onSubmit(formRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-drawer>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { MsgError, MsgSuccess } from '@/utils/message';
import { Host } from '@/api/interface/host';
import { operatePortRule, updatePortRule } from '@/api/modules/host';
import { checkCidr, checkCidrV6, checkIpV4V6, checkPort, deepCopy } from '@/utils/util';

const loading = ref();
const oldRule = ref<Host.RulePort>();

interface DialogProps {
    title: string;
    rowData?: Host.RulePort;
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
        if (params.rowData.address && params.rowData.address !== 'Anywhere') {
            dialogData.value.rowData.source = 'address';
        } else {
            dialogData.value.rowData.source = 'anyWhere';
        }
        oldRule.value = deepCopy(params.rowData);
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
    port: [Rules.requiredInput],
    address: [{ validator: checkAddress, trigger: 'blur' }],
});

function checkAddress(rule: any, value: any, callback: any) {
    if (!dialogData.value.rowData.address) {
        return callback(new Error(i18n.global.t('firewall.addressFormatError')));
    }
    let addrs = dialogData.value.rowData.address.split(',');
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
            if (checkIpV4V6(item)) {
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
        dialogData.value.rowData.operation = 'add';
        if (!dialogData.value.rowData) return;
        if (dialogData.value.rowData.source === 'anyWhere') {
            dialogData.value.rowData.address = '';
        }
        let ports = [];
        if (dialogData.value.rowData.port.indexOf('-') !== -1 && !dialogData.value.rowData.port.startsWith('-')) {
            ports = dialogData.value.rowData.port.split('-');
        } else if (
            dialogData.value.rowData.port.indexOf(',') !== -1 &&
            !dialogData.value.rowData.port.startsWith(',')
        ) {
            ports = dialogData.value.rowData.port.split(',');
        } else {
            ports.push(dialogData.value.rowData.port);
        }
        for (const port of ports) {
            if (checkPort(port)) {
                MsgError(i18n.global.t('firewall.portFormatError'));
                return;
            }
        }
        loading.value = true;
        if (dialogData.value.title === 'create') {
            await operatePortRule(dialogData.value.rowData)
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
        oldRule.value.operation = 'remove';
        dialogData.value.rowData.operation = 'add';
        await updatePortRule({ oldRule: oldRule.value, newRule: dialogData.value.rowData })
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
