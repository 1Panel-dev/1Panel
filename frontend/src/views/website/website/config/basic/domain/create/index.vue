<template>
    <el-drawer
        v-model="open"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        :title="$t('website.addDomain')"
        size="40%"
        :before-close="handleClose"
    >
        <template #header>
            <DrawerHeader :header="$t('website.addDomain')" :back="handleClose" />
        </template>

        <el-row v-loading="loading">
            <el-col :span="22" :offset="1">
                <el-form ref="domainForm" label-position="top" :model="domain" :rules="rules">
                    <el-form-item :label="$t('website.domain')" prop="domains">
                        <el-input
                            type="textarea"
                            :rows="3"
                            v-model="domain.domains"
                            :placeholder="$t('website.domainHelper')"
                        ></el-input>
                    </el-form-item>
                </el-form>
            </el-col>
        </el-row>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit(domainForm)" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-drawer>
</template>

<script lang="ts" setup>
import DrawerHeader from '@/components/drawer-header/index.vue';
import { CreateDomain } from '@/api/modules/website';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { FormInstance } from 'element-plus';
import { ref } from 'vue';
import { MsgSuccess } from '@/utils/message';

const domainForm = ref<FormInstance>();

const rules = ref({
    domains: [Rules.requiredInput],
});

const open = ref(false);
const loading = ref(false);
const domain = ref({
    websiteID: 0,
    domains: '',
});

const em = defineEmits(['close']);
const handleClose = () => {
    domainForm.value?.resetFields();
    open.value = false;
    em('close', false);
};

const acceptParams = async (websiteId: number) => {
    domain.value.websiteID = Number(websiteId);
    open.value = true;
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        loading.value = true;
        CreateDomain(domain.value)
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
