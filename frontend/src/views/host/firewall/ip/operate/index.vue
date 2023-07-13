<template>
    <el-drawer v-model="drawerVisiable" :destroy-on-close="true" :close-on-click-modal="false" size="50%">
        <template #header>
            <DrawerHeader :header="title" :back="handleClose" />
        </template>
        <div v-loading="loading">
            <el-form ref="formRef" label-position="top" :model="dialogData.rowData" :rules="rules">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item :label="$t('firewall.address')" prop="address">
                            <el-input
                                :disabled="dialogData.title === 'edit'"
                                :autosize="{ minRows: 3, maxRows: 6 }"
                                type="textarea"
                                clearable
                                v-model.trim="dialogData.rowData!.address"
                            />
                            <span class="input-help">{{ $t('firewall.addressHelper1') }}</span>
                            <span class="input-help">{{ $t('firewall.addressHelper2') }}</span>
                        </el-form-item>
                        <el-form-item :label="$t('firewall.strategy')" prop="strategy">
                            <el-radio-group v-model="dialogData.rowData!.strategy">
                                <el-radio label="accept">{{ $t('firewall.allow') }}</el-radio>
                                <el-radio label="drop">{{ $t('firewall.deny') }}</el-radio>
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
import { operateIPRule, updateAddrRule } from '@/api/modules/host';
import { checkIpV4V6, deepCopy } from '@/utils/util';

const loading = ref();
const oldRule = ref<Host.RuleIP>();

interface DialogProps {
    title: string;
    rowData?: Host.RuleIP;
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
        let ips = [];
        if (dialogData.value.rowData.address.indexOf(',') !== -1) {
            ips = dialogData.value.rowData.address.split(',');
        } else if (dialogData.value.rowData.address.indexOf('/') !== -1) {
            ips.push(dialogData.value.rowData.address.split('/')[0]);
        } else {
            ips.push(dialogData.value.rowData.address);
        }
        for (const ip of ips) {
            if (checkIpV4V6(ip)) {
                MsgError(i18n.global.t('firewall.addressFormatError'));
                return;
            }
        }
        loading.value = true;
        if (dialogData.value.title === 'create') {
            await operateIPRule(dialogData.value.rowData)
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
        await updateAddrRule({ oldRule: oldRule.value, newRule: dialogData.value.rowData })
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
