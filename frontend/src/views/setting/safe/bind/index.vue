<template>
    <div>
        <el-drawer
            v-model="drawerVisible"
            :destroy-on-close="true"
            :close-on-click-modal="false"
            :close-on-press-escape="false"
            size="30%"
        >
            <template #header>
                <DrawerHeader :header="$t('setting.bindInfo')" :back="handleClose" />
            </template>
            <el-form ref="formRef" label-position="top" :model="form" @submit.prevent v-loading="loading">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item :label="$t('setting.ipv6')" prop="ipv6" :rules="Rules.requiredSelect">
                            <el-radio-group style="width: 100%" v-model="form.ipv6" @change="onChangeMode()">
                                <el-radio value="enable">{{ $t('commons.button.enable') }}</el-radio>
                                <el-radio value="disable">{{ $t('commons.button.disable') }}</el-radio>
                            </el-radio-group>
                        </el-form-item>
                        <el-form-item
                            v-if="form.ipv6 === 'disable'"
                            :label="$t('setting.bindAddress')"
                            prop="bindAddress"
                            :rules="Rules.ip"
                        >
                            <el-checkbox v-model="form.input" :label="$t('container.input')" />
                            <el-select v-if="!form.input" clearable v-model="form.bindAddress">
                                <el-option value="0.0.0.0" label="0.0.0.0"></el-option>
                                <div v-for="item in interfaceOptions" :key="item">
                                    <el-option v-if="item.indexOf(':') === -1" :value="item" :label="item" />
                                </div>
                            </el-select>
                            <el-input v-else clearable v-model="form.bindAddress"></el-input>
                        </el-form-item>
                        <el-form-item v-else :label="$t('setting.bindAddress')" prop="bindAddress" :rules="Rules.ipV6">
                            <el-checkbox v-model="form.input" :label="$t('container.input')" />
                            <el-select v-if="!form.input" clearable v-model="form.bindAddress">
                                <el-option value="::" label="::"></el-option>
                                <div v-for="item in interfaceOptions" :key="item">
                                    <el-option v-if="item.indexOf(':') !== -1" :value="item" :label="item" />
                                </div>
                            </el-select>
                            <el-input v-else clearable v-model="form.bindAddress"></el-input>
                        </el-form-item>
                    </el-col>
                </el-row>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button :disabled="loading" type="primary" @click="onSavePort(formRef)">
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
import { updateBindInfo, loadInterfaceAddr } from '@/api/modules/setting';
import { ElMessageBox, FormInstance } from 'element-plus';
import { Rules } from '@/global/form-rules';
import { GlobalStore } from '@/store';
import DrawerHeader from '@/components/drawer-header/index.vue';
const globalStore = GlobalStore();

interface DialogProps {
    ipv6: string;
    bindAddress: string;
}
const drawerVisible = ref();
const loading = ref();
const interfaceOptions = ref();

const form = reactive({
    ipv6: '',
    bindAddress: '',
    input: false,
});

const formRef = ref<FormInstance>();

const acceptParams = (params: DialogProps): void => {
    form.ipv6 = params.ipv6;
    form.bindAddress = params.bindAddress;
    loadInterface();
    drawerVisible.value = true;
};

const loadInterface = async () => {
    const res = await loadInterfaceAddr();
    interfaceOptions.value = res.data || [];
};

const onChangeMode = () => {
    form.bindAddress = form.ipv6 === 'enable' ? '::' : '0.0.0.0';
};

const onSavePort = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        ElMessageBox.confirm(i18n.global.t('setting.bindInfoHelper'), i18n.global.t('setting.bindInfo'), {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        }).then(async () => {
            loading.value = true;
            await updateBindInfo(form.ipv6, form.bindAddress)
                .then(() => {
                    loading.value = false;
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                    globalStore.isLogin = false;
                    let href = window.location.href;
                    let address = '';
                    if (globalStore.entrance) {
                        address = href.replaceAll('settings/safe', globalStore.entrance);
                    } else {
                        address = href.replaceAll('settings/safe', 'login');
                    }
                    window.open(address, '_self');
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

defineExpose({
    acceptParams,
});
</script>
