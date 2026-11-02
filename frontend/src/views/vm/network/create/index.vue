<template>
    <DrawerPro
        v-model="visible"
        :header="$t('commons.button.create')"
        size="large"
        :auto-close="!loading"
        @close="resetForm"
    >
        <el-form ref="formRef" v-loading="bridgeLoading" label-position="top" :model="form" :rules="rules">
            <el-form-item :label="$t('commons.table.name')" prop="name">
                <el-input v-model.trim="form.name" />
            </el-form-item>
            <el-form-item :label="$t('commons.table.type')" prop="type">
                <el-radio-group v-model="form.type">
                    <el-radio-button label="nat">NAT</el-radio-button>
                    <el-radio-button label="bridge">Bridge</el-radio-button>
                    <el-radio-button label="macvtap">Macvtap</el-radio-button>
                </el-radio-group>
                <span class="input-help">
                    {{ $t(`xpack.vm.${form.type}NetworkHelper`) }}
                </span>
            </el-form-item>
            <el-form-item v-if="form.type === 'bridge'" :label="$t('xpack.vm.bridgeName')" prop="bridgeName">
                <el-select v-model="form.bridgeName" class="full-width" filterable default-first-option>
                    <el-option
                        v-for="item in bridgeOptions"
                        :key="item.name"
                        :label="item.description ? `${item.name} (${item.description})` : item.name"
                        :value="item.name"
                        :disabled="!item.recommended"
                    >
                        <span>{{ item.name }}</span>
                        <span v-if="item.description" class="option-description">{{ item.description }}</span>
                    </el-option>
                </el-select>
                <span class="input-help">{{ $t('xpack.vm.bridgeNameHelper') }}</span>
            </el-form-item>
            <el-form-item v-if="form.type === 'macvtap'" :label="$t('xpack.vm.sourceInterface')" prop="sourceInterface">
                <el-select v-model="form.sourceInterface" class="full-width" filterable>
                    <el-option
                        v-for="item in interfaceOptions"
                        :key="item.name"
                        :label="item.addresses.length ? `${item.name} (${item.addresses.join(', ')})` : item.name"
                        :value="item.name"
                    />
                </el-select>
                <span class="input-help">{{ $t('xpack.vm.macvtapInterfaceHelper') }}</span>
            </el-form-item>
            <template v-if="form.type === 'nat'">
                <el-form-item :label="$t('xpack.vm.networkCIDR')" prop="cidr">
                    <el-input v-model.trim="form.cidr" />
                    <span class="input-help">{{ $t('xpack.vm.natCIDRHelper') }}</span>
                </el-form-item>
                <el-form-item :label="$t('xpack.vm.gateway')" prop="gateway">
                    <el-input v-model.trim="form.gateway" />
                </el-form-item>
                <el-form-item :label="$t('xpack.vm.dhcpStart')" prop="dhcpStart">
                    <el-input v-model.trim="form.dhcpStart" />
                </el-form-item>
                <el-form-item :label="$t('xpack.vm.dhcpEnd')" prop="dhcpEnd">
                    <el-input v-model.trim="form.dhcpEnd" />
                </el-form-item>
            </template>
            <el-form-item>
                <el-checkbox v-model="form.autoStart">{{ $t('xpack.vm.resourceAutoStart') }}</el-checkbox>
            </el-form-item>
            <el-form-item :label="$t('commons.table.description')" prop="description">
                <el-input v-model="form.description" type="textarea" :rows="3" />
            </el-form-item>
        </el-form>
        <template #footer>
            <el-button :disabled="loading" @click="visible = false">{{ $t('commons.button.cancel') }}</el-button>
            <el-button v-permission type="primary" :disabled="loading || bridgeLoading" @click="submit">
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DrawerPro>
    <div v-if="visible && loading" v-loading="loading" class="vm-drawer-loading-mask"></div>
</template>

<script setup lang="ts">
import { reactive, ref, watch } from 'vue';
import { type FormInstance, type FormRules } from 'element-plus';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { VM } from '@/api/interface/vm';
import { createVMNetwork, listVMNetworkBridges, listVMNetworkInterfaces } from '@/api/modules/vm';
import { Rules } from '@/global/form-rules';

const emit = defineEmits(['search']);
const visible = ref(false);
const loading = ref(false);
const bridgeLoading = ref(false);
const formRef = ref<FormInstance>();
const bridgeOptions = ref<VM.NetworkBridgeItem[]>([]);
const interfaceOptions = ref<VM.NetworkInterfaceItem[]>([]);

const form = reactive<VM.NetworkCreate & { id: number }>({
    id: 0,
    name: '',
    type: 'nat',
    bridgeName: '',
    sourceInterface: '',
    cidr: '192.168.100.0/24',
    gateway: '192.168.100.1',
    dhcpStart: '192.168.100.2',
    dhcpEnd: '192.168.100.254',
    autoStart: true,
    description: '',
});

const validateAvailableBridgeName = (_rule: unknown, value: string, callback: (error?: Error) => void) => {
    if (form.type !== 'bridge' || !value) {
        callback();
        return;
    }
    const bridge = bridgeOptions.value.find((item) => item.name === value);
    if (bridge && !bridge.recommended) {
        callback(new Error(i18n.global.t('xpack.vm.bridgeName') + ' ' + i18n.global.t('commons.status.used')));
        return;
    }
    callback();
};

const rules = reactive<FormRules>({
    name: [Rules.requiredInput, Rules.vmNetwork],
    value: [{ required: true, message: i18n.global.t('commons.rule.requiredInput'), trigger: 'blur' }],
    bridgeName: [
        Rules.requiredInput,
        Rules.vmNetwork,
        { validator: validateAvailableBridgeName, trigger: ['blur', 'change'] },
    ],
    sourceInterface: [Rules.requiredSelect],
    cidr: [Rules.requiredInput],
    gateway: [Rules.requiredInput, Rules.ipv4],
    dhcpStart: [Rules.requiredInput, Rules.ipv4],
    dhcpEnd: [Rules.requiredInput, Rules.ipv4],
});

const acceptParams = () => {
    resetForm();
    visible.value = true;
};

const loadBridgeOptions = async () => {
    bridgeLoading.value = true;
    await listVMNetworkBridges()
        .then((res) => {
            bridgeOptions.value = res.data || [];
        })
        .catch(() => {
            bridgeOptions.value = [];
        })
        .finally(() => {
            bridgeLoading.value = false;
        });
};

const loadInterfaceOptions = async () => {
    interfaceOptions.value = [];
    bridgeLoading.value = true;
    await listVMNetworkInterfaces()
        .then((res) => {
            interfaceOptions.value = res.data || [];
        })
        .catch(() => {
            interfaceOptions.value = [];
        })
        .finally(() => {
            bridgeLoading.value = false;
        });
};

const submit = async () => {
    if (!formRef.value) return;
    await formRef.value.validate();
    loading.value = true;
    await createVMNetwork({ ...form })
        .then(() => {
            visible.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        })
        .finally(() => {
            emit('search');
            loading.value = false;
        });
};

const resetForm = () => {
    form.id = 0;
    form.name = '';
    form.type = 'nat';
    form.bridgeName = '';
    form.sourceInterface = '';
    form.cidr = '192.168.100.0/24';
    form.gateway = '192.168.100.1';
    form.dhcpStart = '192.168.100.2';
    form.dhcpEnd = '192.168.100.254';
    form.autoStart = true;
    form.description = '';
    bridgeOptions.value = [];
    interfaceOptions.value = [];
    formRef.value?.clearValidate();
};

watch(
    () => form.type,
    (type) => {
        form.bridgeName = '';
        form.sourceInterface = '';
        formRef.value?.clearValidate();
        if (type === 'bridge') {
            loadBridgeOptions();
        } else if (type === 'macvtap') {
            loadInterfaceOptions();
        }
    },
);

defineExpose({ acceptParams });
</script>

<style scoped lang="scss">
.option-description {
    color: var(--el-text-color-secondary);
    float: right;
    font-size: 12px;
}

.vm-drawer-loading-mask {
    position: fixed;
    top: 0;
    right: 0;
    bottom: 0;
    z-index: 3000;
    width: 50%;
    background: var(--el-mask-color);
}
</style>
