<template>
    <DrawerPro
        v-model="drawerVisible"
        :header="$t('commons.button.bind')"
        :resource="licenseName"
        :back="handleClose"
        size="small"
    >
        <el-form ref="formRef" label-position="top" :model="form" @submit.prevent v-loading="loading">
            <el-form-item :label="$t('setting.bindNode')" prop="nodeID" :rules="Rules.requiredSelect">
                <el-select filterable v-model="form.nodeID" style="width: 100%">
                    <el-option v-for="item in freeNodes" :key="item.id" :label="item.name" :value="item.id" />
                </el-select>
            </el-form-item>
        </el-form>
        <template #footer>
            <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
            <el-button :disabled="loading" type="primary" @click="onBind(formRef)">
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DrawerPro>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { bindLicense, listNodeOptions } from '@/api/modules/setting';
import { FormInstance } from 'element-plus';
import { GlobalStore } from '@/store';
import { Rules } from '@/global/form-rules';
const globalStore = GlobalStore();

interface DialogProps {
    licenseName: string;
    licenseID: number;
}
const drawerVisible = ref();
const loading = ref();
const licenseName = ref();
const freeNodes = ref([]);

const form = reactive({
    nodeID: null,
    licenseID: null,
});

const formRef = ref<FormInstance>();

const acceptParams = (params: DialogProps): void => {
    licenseName.value = params.licenseName;
    form.licenseID = params.licenseID;
    loadNodes();
    drawerVisible.value = true;
};

const onBind = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        loading.value = true;
        await bindLicense(form.licenseID, form.nodeID)
            .then(() => {
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                globalStore.isProductPro = false;
                globalStore.themeConfig.isGold = false;
                window.location.reload();
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

const loadNodes = async () => {
    if (!globalStore.isMasterProductPro) {
        freeNodes.value = [{ id: 0, name: i18n.global.t('terminal.local') }];
        return;
    }
    await listNodeOptions()
        .then((res) => {
            freeNodes.value = res.data || [];
        })
        .catch(() => {
            freeNodes.value = [];
        });
};

const handleClose = () => {
    drawerVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>
