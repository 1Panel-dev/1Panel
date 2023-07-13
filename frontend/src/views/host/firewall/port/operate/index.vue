<template>
    <el-drawer v-model="drawerVisiable" :destroy-on-close="true" :close-on-click-modal="false" size="50%">
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
                                <el-radio label="anyWhere">{{ $t('firewall.anyWhere') }}</el-radio>
                                <el-radio label="address">{{ $t('firewall.address') }}</el-radio>
                            </el-radio-group>
                        </el-form-item>

                        <el-form-item
                            :label="$t('firewall.address')"
                            v-if="dialogData.rowData!.source === 'address'"
                            prop="address"
                        >
                            <el-input
                                :placeholder="$t('firewall.addressHelper')"
                                v-model="dialogData.rowData!.address"
                            />
                        </el-form-item>

                        <el-form-item :label="$t('firewall.strategy')" prop="strategy">
                            <el-radio-group v-model="dialogData.rowData!.strategy">
                                <el-radio label="accept">{{ $t('firewall.accept') }}</el-radio>
                                <el-radio label="drop">{{ $t('firewall.drop') }}</el-radio>
                            </el-radio-group>
                        </el-form-item>
                    </el-col>
                </el-row>
            </el-form>
        </div>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="drawerVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
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
import { checkIpV4V6, checkPort, deepCopy } from '@/utils/util';

const loading = ref();
const oldRule = ref<Host.RulePort>();

interface DialogProps {
    title: string;
    rowData?: Host.RulePort;
    getTableList?: () => Promise<any>;
}
const title = ref<string>('');
const drawerVisiable = ref(false);
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
    drawerVisiable.value = true;
};
const emit = defineEmits<{ (e: 'search'): void }>();

const handleClose = () => {
    drawerVisiable.value = false;
};

const rules = reactive({
    protocol: [Rules.requiredSelect],
    port: [Rules.requiredInput],
    address: [Rules.requiredInput],
});

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
        } else {
            if (dialogData.value.rowData.address.indexOf('/') !== -1) {
                if (checkIpV4V6(dialogData.value.rowData.address.split('/')[0])) {
                    MsgError(i18n.global.t('firewall.addressFormatError'));
                    return;
                }
            } else {
                if (checkIpV4V6(dialogData.value.rowData.address)) {
                    MsgError(i18n.global.t('firewall.addressFormatError'));
                    return;
                }
            }
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
                    drawerVisiable.value = false;
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
                drawerVisiable.value = false;
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
