<template>
    <DrawerPro v-model="open" :header="$t('website.addDomain')" @close="handleClose">
        <el-form ref="domainForm" label-position="top" :model="create">
            <DomainCreate v-model:form="create" @gengerate="domainForm.clearValidate()"></DomainCreate>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit(domainForm)" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { createDomain } from '@/api/modules/website';
import i18n from '@/lang';
import { FormInstance } from 'element-plus';
import { ref } from 'vue';
import { MsgSuccess } from '@/utils/message';
import DomainCreate from '@/views/website/website/domain-create/index.vue';

const domainForm = ref<FormInstance>();

const initDomain = () => ({
    domain: '',
    port: 80,
    ssl: false,
});

const open = ref(false);
const loading = ref(false);
const create = ref({
    websiteID: 0,
    domains: [initDomain()],
    domainStr: '',
});

const em = defineEmits(['close']);
const handleClose = () => {
    domainForm.value?.resetFields();
    open.value = false;
    em('close', false);
};

const acceptParams = async (websiteId: number) => {
    create.value.websiteID = Number(websiteId);
    create.value.domains = [initDomain()];
    create.value.domainStr = '';
    open.value = true;
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        loading.value = true;
        createDomain(create.value)
            .then(() => {
                MsgSuccess(i18n.global.t('commons.msg.createSuccess'));
                handleClose();
            })
            .finally(() => {
                loading.value = false;
            });
    });
};

defineExpose({
    acceptParams,
});
</script>
