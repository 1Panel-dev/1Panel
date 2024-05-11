<template>
    <el-drawer
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        :destroy-on-close="true"
        v-model="drawerVisible"
        size="30%"
    >
        <template #header>
            <Header :header="$t('toolbox.fail2ban.' + form.operate + 'IP')" :back="handleClose"></Header>
        </template>

        <el-form ref="formRef" label-position="top" :model="form" @submit.prevent :rules="rules" v-loading="loading">
            <el-row type="flex" justify="center">
                <el-col :span="22">
                    <el-form-item :label="$t('toolbox.fail2ban.' + form.operate + 'IP')" prop="ips">
                        <el-input :rows="25" type="textarea" clearable v-model="form.ips" />
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
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { operatorFail2banSSHD, searchFail2ban } from '@/api/modules/toolbox';
import Header from '@/components/drawer-header/index.vue';
import { FormInstance } from 'element-plus';
import { MsgSuccess } from '@/utils/message';
import i18n from '@/lang';
import { checkCidr, checkIpV4V6 } from '@/utils/util';

const emit = defineEmits<{ (e: 'search'): void }>();

const drawerVisible = ref(false);
const loading = ref();

const form = reactive({
    operate: 'ignore',
    ips: '',
});

const rules = reactive({
    ips: [{ validator: checkIPs, trigger: 'blur' }],
});
function checkIPs(rule: any, value: any, callback: any) {
    if (form.ips !== '') {
        let addr = form.ips.split('\n');
        for (const item of addr) {
            if (item === '') {
                continue;
            }
            if (item.indexOf('/') !== -1) {
                if (checkCidr(item)) {
                    return callback(new Error(i18n.global.t('firewall.addressFormatError')));
                }
            } else if (checkIpV4V6(item)) {
                return callback(new Error(i18n.global.t('firewall.addressFormatError')));
            }
        }
    }
    callback();
}

interface DialogProps {
    operate: string;
}
const formRef = ref<FormInstance>();
const acceptParams = (params: DialogProps): void => {
    form.operate = params.operate;
    form.ips = '';
    drawerVisible.value = true;
    search();
};

const search = () => {
    let params = {
        status: form.operate,
    };
    searchFail2ban(params).then((res) => {
        let dataItem = res.data || [];
        form.ips = dataItem.join('\n');
    });
};

const onSave = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        ElMessageBox.confirm(
            i18n.global.t('toolbox.fail2ban.' + form.operate + 'Helper'),
            i18n.global.t('commons.button.set'),
            {
                confirmButtonText: i18n.global.t('commons.button.confirm'),
                cancelButtonText: i18n.global.t('commons.button.cancel'),
                type: 'info',
            },
        ).then(async () => {
            let ipItems = form.ips.split('\n');
            let ips = [];
            for (const item of ipItems) {
                if (item !== '') {
                    ips.push(item);
                }
            }
            let param = {
                operate: form.operate,
                ips: ips,
            };
            loading.value = true;
            await operatorFail2banSSHD(param)
                .then(async () => {
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                    loading.value = false;
                    drawerVisible.value = false;
                    emit('search');
                })
                .catch(() => {
                    loading.value = false;
                });
        });
    });
};

const handleClose = () => {
    drawerVisible.value = false;
};
defineExpose({ acceptParams });
</script>
