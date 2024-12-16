<template>
    <div>
        <el-drawer
            v-model="drawerVisible"
            :destroy-on-close="true"
            :close-on-click-modal="false"
            :close-on-press-escape="false"
            @close="handleClose"
            size="30%"
        >
            <template #header>
                <DrawerHeader header="IPv6" :back="handleClose" />
            </template>
            <el-alert class="common-prompt" :closable="false" type="warning">
                <template #default>
                    <span class="input-help">
                        {{ $t('container.ipv6Helper') }}
                        <el-link
                            style="font-size: 12px; margin-left: 5px"
                            icon="Position"
                            @click="toDoc()"
                            type="primary"
                        >
                            {{ $t('firewall.quickJump') }}
                        </el-link>
                    </span>
                </template>
            </el-alert>

            <el-form :model="form" ref="formRef" :rules="rules" v-loading="loading" label-position="top">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item prop="fixedCidrV6" :label="$t('container.subnet')">
                            <el-input v-model="form.fixedCidrV6" />
                            <span class="input-help">{{ $t('container.ipv6CidrHelper') }}</span>
                        </el-form-item>
                        <el-form-item>
                            <el-checkbox v-model="showMore" :label="$t('app.advanced')" />
                        </el-form-item>
                        <div v-if="showMore">
                            <el-form-item prop="ip6Tables" label="ip6tables">
                                <el-switch v-model="form.ip6Tables"></el-switch>
                                <span class="input-help">{{ $t('container.ipv6TablesHelper') }}</span>
                            </el-form-item>
                            <el-form-item prop="experimental" label="experimental">
                                <el-switch v-model="form.experimental"></el-switch>
                                <span class="input-help">{{ $t('container.experimentalHelper') }}</span>
                            </el-form-item>
                        </div>
                    </el-col>
                </el-row>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="handleClose">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button :disabled="loading" type="primary" @click="onSave(formRef)">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-drawer>

        <ConfirmDialog ref="confirmDialogRef" @confirm="onSubmitSave"></ConfirmDialog>
    </div>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { FormInstance } from 'element-plus';
import { updateIpv6Option } from '@/api/modules/container';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { checkIpV6 } from '@/utils/util';
import { GlobalStore } from '@/store';

const globalStore = GlobalStore();
const loading = ref();
const drawerVisible = ref();
const confirmDialogRef = ref();
const formRef = ref();
const showMore = ref(true);

interface DialogProps {
    fixedCidrV6: string;
    ip6Tables: boolean;
    experimental: boolean;
}

const form = reactive({
    fixedCidrV6: '',
    ip6Tables: false,
    experimental: false,
});
const rules = reactive({
    fixedCidrV6: [{ validator: checkFixedCidrV6, trigger: 'blur', required: true }],
});

function checkFixedCidrV6(rule: any, value: any, callback: any) {
    if (!form.fixedCidrV6 || form.fixedCidrV6.indexOf('/') === -1) {
        return callback(new Error(i18n.global.t('commons.rule.formatErr')));
    }
    if (checkIpV6(form.fixedCidrV6.split('/')[0])) {
        return callback(new Error(i18n.global.t('commons.rule.formatErr')));
    }
    const reg = /^(?:[1-9]|[1-9][0-9]|1[0-1][0-9]|12[0-8])$/;
    if (!reg.test(form.fixedCidrV6.split('/')[1])) {
        return callback(new Error(i18n.global.t('commons.rule.formatErr')));
    }
    callback();
}

const toDoc = () => {
    window.open(globalStore.docsUrl + '/user_manual/containers/setting/', '_blank', 'noopener,noreferrer');
};

const emit = defineEmits<{ (e: 'search'): void }>();

const acceptParams = (params: DialogProps): void => {
    form.fixedCidrV6 = params.fixedCidrV6;
    form.ip6Tables = params.ip6Tables;
    form.experimental = params.experimental;
    drawerVisible.value = true;
};

const onSave = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        let params = {
            header: i18n.global.t('database.confChange'),
            operationInfo: i18n.global.t('database.restartNowHelper'),
            submitInputInfo: i18n.global.t('database.restartNow'),
        };
        confirmDialogRef.value!.acceptParams(params);
    });
};

const onSubmitSave = async () => {
    loading.value = true;
    await updateIpv6Option(form.fixedCidrV6, form.ip6Tables, form.experimental)
        .then(() => {
            loading.value = false;
            drawerVisible.value = false;
            emit('search');
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        })
        .catch(() => {
            loading.value = false;
        });
};

const handleClose = () => {
    emit('search');
    drawerVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>
