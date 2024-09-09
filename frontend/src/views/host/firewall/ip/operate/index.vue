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
            <el-form ref="formRef" label-position="top" @submit.prevent :model="dialogData.rowData" :rules="rules">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item :label="$t('firewall.address')" prop="address">
                            <el-input
                                :disabled="dialogData.title === 'edit'"
                                :rows="3"
                                type="textarea"
                                clearable
                                v-model.trim="dialogData.rowData!.address"
                            />
                            <span class="input-help">{{ $t('firewall.addressHelper1') }}</span>
                            <span class="input-help">{{ $t('firewall.addressHelper2') }}</span>
                        </el-form-item>
                        <el-form-item :label="$t('firewall.strategy')" prop="strategy">
                            <el-radio-group v-model="dialogData.rowData!.strategy">
                                <el-radio value="accept">{{ $t('firewall.allow') }}</el-radio>
                                <el-radio value="drop">{{ $t('firewall.deny') }}</el-radio>
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
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { MsgSuccess } from '@/utils/message';
import { Host } from '@/api/interface/host';
import { operateIPRule, updateAddrRule } from '@/api/modules/host';
import { checkCidr, checkCidrV6, checkIpV4V6, deepCopy } from '@/utils/util';

const loading = ref();
const oldRule = ref<Host.RuleIP>();

interface DialogProps {
    title: string;
    rowData?: Host.RuleIP;
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
    title.value = i18n.global.t('firewall.' + dialogData.value.title);
    drawerVisible.value = true;
};
const emit = defineEmits<{ (e: 'search'): void }>();

const handleClose = () => {
    drawerVisible.value = false;
};

const rules = reactive({
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
        loading.value = true;
        if (dialogData.value.title === 'create') {
            await operateIPRule(dialogData.value.rowData)
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
        await updateAddrRule({ oldRule: oldRule.value, newRule: dialogData.value.rowData })
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
