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
                <DrawerHeader :header="$t('setting.defaultNetwork')" :back="handleClose" />
            </template>
            <el-form ref="formRef" label-position="top" :model="form" @submit.prevent v-loading="loading">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item
                            :label="$t('setting.defaultNetwork')"
                            prop="defaultNetwork"
                            :rules="Rules.requiredSelect"
                        >
                            <el-select v-model="form.defaultNetwork" filterable>
                                <el-option
                                    v-for="item in netOptions"
                                    :key="item"
                                    :label="item == 'all' ? $t('commons.table.all') : item"
                                    :value="item"
                                />
                            </el-select>
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
import { updateSetting } from '@/api/modules/setting';
import { FormInstance } from 'element-plus';
import { Rules } from '@/global/form-rules';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { getNetworkOptions } from '@/api/modules/host';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();

const emit = defineEmits<{ (e: 'search'): void }>();

interface DialogProps {
    defaultNetwork: string;
}
const drawerVisible = ref();
const loading = ref();
const netOptions = ref();

const form = reactive({
    defaultNetwork: '',
});

const formRef = ref<FormInstance>();

const acceptParams = (params: DialogProps): void => {
    form.defaultNetwork = params.defaultNetwork;
    loadNetworkOptions();
    drawerVisible.value = true;
};

const loadNetworkOptions = async () => {
    const res = await getNetworkOptions();
    netOptions.value = res.data;
};

const onSave = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        await updateSetting({ key: 'DefaultNetwork', value: form.defaultNetwork })
            .then(async () => {
                globalStore.setDefaultNetwork(form.defaultNetwork);
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                loading.value = false;
                drawerVisible.value = false;
                emit('search');
                return;
            })
            .catch(() => {
                loading.value = false;
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
