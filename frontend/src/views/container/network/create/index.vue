<template>
    <el-drawer
        v-model="drawerVisible"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        size="50%"
    >
        <template #header>
            <DrawerHeader :header="$t('container.createNetwork')" :back="handleClose" />
        </template>
        <el-form ref="formRef" label-position="top" v-loading="loading" :model="form" :rules="rules" label-width="80px">
            <el-row type="flex" justify="center">
                <el-col :span="22">
                    <el-form-item :label="$t('container.networkName')" prop="name">
                        <el-input clearable v-model.trim="form.name" />
                    </el-form-item>
                    <el-form-item :label="$t('container.driver')" prop="driver">
                        <el-select v-model="form.driver">
                            <el-option label="bridge" value="bridge" />
                            <el-option label="ipvlan" value="ipvlan" />
                            <el-option label="macvlan" value="macvlan" />
                            <el-option label="overlay" value="overlay" />
                        </el-select>
                    </el-form-item>

                    <el-checkbox v-model="form.ipv4">IPv4</el-checkbox>
                    <div v-if="form.ipv4">
                        <el-row type="flex" justify="center" :gutter="20">
                            <el-col :span="12">
                                <el-form-item :label="$t('container.subnet')" prop="subnet">
                                    <el-input placeholder="172.16.10.0/24" clearable v-model.trim="form.subnet" />
                                </el-form-item>
                            </el-col>
                            <el-col :span="12">
                                <el-form-item :label="$t('container.gateway')" prop="gateway">
                                    <el-input placeholder="172.16.10.12" clearable v-model.trim="form.gateway" />
                                </el-form-item>
                            </el-col>
                            <el-col :span="12">
                                <el-form-item :label="$t('container.scope')" prop="scope">
                                    <el-input placeholder="172.16.10.0/16" clearable v-model.trim="form.scope" />
                                </el-form-item>
                            </el-col>
                            <el-col :span="12"></el-col>
                        </el-row>
                        <el-form-item :label="$t('container.auxAddress')" prop="scopeV6">
                            <el-table :data="form.auxAddress" v-if="form.auxAddress.length !== 0">
                                <el-table-column :label="$t('container.label')" min-width="100">
                                    <template #default="{ row }">
                                        <el-input placeholder="my-router" v-model="row.key" />
                                    </template>
                                </el-table-column>
                                <el-table-column label="IP" min-width="150">
                                    <template #default="{ row }">
                                        <el-input placeholder="172.16.10.13" v-model="row.value" />
                                    </template>
                                </el-table-column>
                                <el-table-column min-width="40">
                                    <template #default="scope">
                                        <el-button link type="primary" @click="handleV4Delete(scope.$index)">
                                            {{ $t('commons.button.delete') }}
                                        </el-button>
                                    </template>
                                </el-table-column>
                            </el-table>
                            <el-button class="mt-2" @click="handleV4Add()">
                                {{ $t('commons.button.add') }}
                            </el-button>
                        </el-form-item>
                    </div>

                    <el-checkbox class="mb-4" v-model="form.ipv6">IPv6</el-checkbox>
                    <div v-if="form.ipv6">
                        <el-row type="flex" justify="center" :gutter="20">
                            <el-col :span="12">
                                <el-form-item :label="$t('container.subnet')" prop="subnetV6">
                                    <el-input placeholder="2408:400e::/48" clearable v-model.trim="form.subnetV6" />
                                </el-form-item>
                            </el-col>
                            <el-col :span="12">
                                <el-form-item :label="$t('container.gateway')" prop="gatewayV6">
                                    <el-input placeholder="2408:400e::1" clearable v-model.trim="form.gatewayV6" />
                                </el-form-item>
                            </el-col>
                            <el-col :span="12">
                                <el-form-item :label="$t('container.scope')" prop="scopeV6">
                                    <el-input placeholder="2408:400e::/64" clearable v-model.trim="form.scopeV6" />
                                </el-form-item>
                            </el-col>
                            <el-col :span="12"></el-col>
                        </el-row>
                        <el-form-item :label="$t('container.auxAddress')" prop="scopeV6">
                            <el-table :data="form.auxAddressV6" v-if="form.auxAddressV6.length !== 0">
                                <el-table-column :label="$t('container.label')" min-width="100">
                                    <template #default="{ row }">
                                        <el-input placeholder="my-router" v-model="row.key" />
                                    </template>
                                </el-table-column>
                                <el-table-column label="IP" min-width="150">
                                    <template #default="{ row }">
                                        <el-input placeholder="2408:400e::3" v-model="row.value" />
                                    </template>
                                </el-table-column>
                                <el-table-column min-width="40">
                                    <template #default="scope">
                                        <el-button link type="primary" @click="handleV6Delete(scope.$index)">
                                            {{ $t('commons.button.delete') }}
                                        </el-button>
                                    </template>
                                </el-table-column>
                            </el-table>
                            <el-button class="mt-2" @click="handleV6Add()">
                                {{ $t('commons.button.add') }}
                            </el-button>
                        </el-form-item>
                    </div>

                    <el-form-item :label="$t('container.option')" prop="optionStr">
                        <el-input
                            type="textarea"
                            :placeholder="$t('container.tagHelper')"
                            :rows="3"
                            v-model="form.optionStr"
                        />
                    </el-form-item>
                    <el-form-item :label="$t('container.tag')" prop="labelStr">
                        <el-input
                            type="textarea"
                            :placeholder="$t('container.tagHelper')"
                            :rows="3"
                            v-model="form.labelStr"
                        />
                    </el-form-item>
                </el-col>
            </el-row>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button :disabled="loading" @click="drawerVisible = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button :disabled="loading" type="primary" @click="onSubmit(formRef)">
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
import { createNetwork } from '@/api/modules/container';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { MsgSuccess } from '@/utils/message';
import { checkIp, checkIpV6 } from '@/utils/util';

const loading = ref(false);

const drawerVisible = ref(false);
const form = reactive({
    name: '',
    labelStr: '',
    labels: [] as Array<string>,
    optionStr: '',
    options: [] as Array<string>,
    driver: '',
    ipv4: true,
    subnet: '',
    gateway: '',
    scope: '',
    auxAddress: [],
    ipv6: false,
    subnetV6: '',
    gatewayV6: '',
    scopeV6: '',
    auxAddressV6: [],
});

const acceptParams = (): void => {
    form.name = '';
    form.labelStr = '';
    form.labels = [];
    form.optionStr = '';
    form.options = [];
    form.driver = 'bridge';
    form.ipv4 = true;
    form.subnet = '';
    form.gateway = '';
    form.scope = '';
    form.auxAddress = [];
    form.ipv6 = false;
    form.subnetV6 = '';
    form.gatewayV6 = '';
    form.scopeV6 = '';
    form.auxAddressV6 = [];
    drawerVisible.value = true;
};
const emit = defineEmits<{ (e: 'search'): void }>();

const handleClose = () => {
    drawerVisible.value = false;
};

const rules = reactive({
    name: [Rules.requiredInput],
    driver: [Rules.requiredSelect],
    subnet: [{ validator: checkCidr, trigger: 'blur' }, Rules.requiredInput],
    gateway: [{ validator: checkGateway, trigger: 'blur' }],
    scope: [{ validator: checkCidr, trigger: 'blur' }],
    subnetV6: [{ validator: checkFixedCidrV6, trigger: 'blur' }, Rules.requiredInput],
    gatewayV6: [{ validator: checkGatewayV6, trigger: 'blur' }],
    scopeV6: [{ validator: checkFixedCidrV6, trigger: 'blur' }],
});

function checkGateway(rule: any, value: any, callback: any) {
    if (value === '') {
        callback();
    }
    if (checkIp(value)) {
        return callback(new Error(i18n.global.t('commons.rule.formatErr')));
    }
    callback();
}

function checkGatewayV6(rule: any, value: any, callback: any) {
    if (value === '') {
        callback();
    }
    if (checkIpV6(value)) {
        return callback(new Error(i18n.global.t('commons.rule.formatErr')));
    }
    callback();
}

function checkCidr(rule: any, value: any, callback: any) {
    if (value === '') {
        callback();
    }
    const reg =
        /^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(?:\/([0-9]|[1-2][0-9]|3[0-2]))$/;
    if (!reg.test(value)) {
        return callback(new Error(i18n.global.t('commons.rule.formatErr')));
    }
    callback();
}

function checkFixedCidrV6(rule: any, value: any, callback: any) {
    if (value === '') {
        callback();
    }
    if (!form.subnetV6 || form.subnetV6.indexOf('/') === -1) {
        return callback(new Error(i18n.global.t('commons.rule.formatErr')));
    }
    if (checkIpV6(form.subnetV6.split('/')[0])) {
        return callback(new Error(i18n.global.t('commons.rule.formatErr')));
    }
    const reg = /^(?:[1-9]|[1-9][0-9]|1[0-1][0-9]|12[0-8])$/;
    if (!reg.test(form.subnetV6.split('/')[1])) {
        return callback(new Error(i18n.global.t('commons.rule.formatErr')));
    }
    callback();
}

const handleV4Add = () => {
    let item = {
        key: '',
        value: '',
    };
    form.auxAddress.push(item);
};
const handleV4Delete = (index: number) => {
    form.auxAddress.splice(index, 1);
};

const handleV6Add = () => {
    let item = {
        key: '',
        value: '',
    };
    form.auxAddressV6.push(item);
};
const handleV6Delete = (index: number) => {
    form.auxAddressV6.splice(index, 1);
};

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();
const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        if (form.labelStr !== '') {
            form.labels = form.labelStr.split('\n');
        }
        if (form.optionStr !== '') {
            form.options = form.optionStr.split('\n');
        }
        loading.value = true;
        await createNetwork(form)
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
