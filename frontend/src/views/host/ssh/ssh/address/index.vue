<template>
    <div>
        <el-drawer
            v-model="drawerVisible"
            :destroy-on-close="true"
            @close="handleClose"
            :close-on-click-modal="false"
            :close-on-press-escape="false"
            size="30%"
        >
            <template #header>
                <DrawerHeader :header="$t('ssh.listenAddress')" :back="handleClose" />
            </template>
            <el-form
                ref="formRef"
                label-position="top"
                :rules="rules"
                :model="form"
                @submit.prevent
                v-loading="loading"
            >
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-alert class="common-prompt" :closable="false" type="warning">
                            <template #default>
                                <span>
                                    {{ $t('ssh.listenHelper', [form.port]) }}
                                </span>
                            </template>
                        </el-alert>
                        <el-form-item label="IPv4" prop="listenAddressV4">
                            <el-checkbox
                                v-model="form.ipv4All"
                                @change="form.listenAddressV4 = form.ipv4All ? '0.0.0.0' : form.listenAddressV4"
                            >
                                {{ $t('setting.bindAll') }}
                            </el-checkbox>
                            <el-input :disabled="form.ipv4All" clearable v-model="form.listenAddressV4"></el-input>
                        </el-form-item>
                        <el-form-item label="IPv6" prop="listenAddressV6">
                            <el-checkbox
                                v-model="form.ipv6All"
                                @change="form.listenAddressV6 = form.ipv6All ? '::' : form.listenAddressV6"
                            >
                                {{ $t('setting.bindAll') }}
                            </el-checkbox>
                            <el-input :disabled="form.ipv6All" clearable v-model="form.listenAddressV6"></el-input>
                        </el-form-item>
                    </el-col>
                </el-row>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button :disabled="loading" type="primary" @click="onSave(formRef)">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-drawer>
    </div>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { ElMessageBox, FormInstance } from 'element-plus';
import { updateSSH } from '@/api/modules/host';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { checkIp, checkIpV6 } from '@/utils/util';

const emit = defineEmits<{ (e: 'search'): void }>();

interface DialogProps {
    port: number;
    address: string;
}
const drawerVisible = ref();
const loading = ref();

const form = reactive({
    port: 22,
    ipv4All: false,
    ipv6All: false,
    listenAddressV4: '',
    listenAddressV6: '',
});

const rules = reactive({
    listenAddressV4: [{ validator: checkIPv4, trigger: 'blur' }],
    listenAddressV6: [{ validator: checkIPv6, trigger: 'blur' }],
});

function checkIPv4(rule: any, value: any, callback: any) {
    if (value === '') {
        callback();
    }
    if (checkIp(value)) {
        return callback(new Error(i18n.global.t('commons.rule.ip')));
    }
    callback();
}

function checkIPv6(rule: any, value: any, callback: any) {
    if (value === '') {
        callback();
    }
    if (checkIpV6(value)) {
        return callback(new Error(i18n.global.t('commons.rule.ip')));
    }
    callback();
}

const formRef = ref<FormInstance>();

const acceptParams = (params: DialogProps): void => {
    let items = params.address.split(',');
    for (const item of items) {
        if (item.indexOf(':') !== -1) {
            form.listenAddressV6 = item;
            form.ipv6All = item === '::';
            continue;
        }
        form.listenAddressV4 = item;
        form.ipv4All = item === '0.0.0.0';
    }
    form.port = params.port;
    drawerVisible.value = true;
};

const onSave = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        let itemAddr = [];
        if (form.listenAddressV4 !== '') {
            itemAddr.push(form.listenAddressV4);
        }
        if (form.listenAddressV6 !== '') {
            itemAddr.push(form.listenAddressV6);
        }
        let addr =
            itemAddr.join(',') === '' || itemAddr.join(',') === '0.0.0.0,::'
                ? i18n.global.t('ssh.allV4V6', [form.port])
                : itemAddr.join(',');
        ElMessageBox.confirm(
            i18n.global.t('ssh.sshChangeHelper', [i18n.global.t('ssh.listenAddress'), addr]),
            i18n.global.t('ssh.sshChange'),
            {
                confirmButtonText: i18n.global.t('commons.button.confirm'),
                cancelButtonText: i18n.global.t('commons.button.cancel'),
                type: 'info',
            },
        )
            .then(async () => {
                let params = {
                    key: 'ListenAddress',
                    oldValue: '',
                    newValue: itemAddr.join(','),
                };
                loading.value = true;
                await updateSSH(params)
                    .then(() => {
                        loading.value = false;
                        handleClose();
                        emit('search');
                        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                    })
                    .catch(() => {
                        loading.value = false;
                    });
            })
            .catch(() => {
                emit('search');
            });
    });
};

const handleClose = () => {
    drawerVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>
