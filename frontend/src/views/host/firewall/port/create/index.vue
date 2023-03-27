<template>
    <el-drawer v-model="drawerVisiable" :destroy-on-close="true" :close-on-click-modal="false" size="50%">
        <template #header>
            <DrawerHeader :header="$t('firewall.portRule')" :back="handleClose" />
        </template>
        <el-form ref="formRef" label-position="top" :model="dialogData.rowData" :rules="rules">
            <el-row type="flex" justify="center">
                <el-col :span="22">
                    <el-form-item :label="$t('firewall.protocol')" prop="protocol">
                        <el-select style="width: 100%" v-model="dialogData.rowData!.protocol">
                            <el-option value="tcp" label="tcp" />
                            <el-option value="udp" label="udp" />
                            <el-option value="tcp/udp" label="tcp/udp" />
                        </el-select>
                    </el-form-item>

                    <el-form-item :label="$t('firewall.port')" prop="port">
                        <el-input clearable v-model.trim="dialogData.rowData!.port" />
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
                        <el-input v-model="dialogData.rowData!.address" />
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
import { MsgSuccess } from '@/utils/message';
import { Host } from '@/api/interface/host';
import { operatePortRule } from '@/api/modules/host';

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
    title.value = i18n.global.t('commons.button.' + dialogData.value.title);
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
        if (!dialogData.value.rowData) return;
        if (dialogData.value.title === 'create') {
            await operatePortRule(dialogData.value.rowData);
        }

        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        emit('search');
        drawerVisiable.value = false;
    });
};

defineExpose({
    acceptParams,
});
</script>
