<template>
    <DrawerPro
        v-model="open"
        :header="$t('commons.button.bind') + $t('menu.website')"
        :resource="agentName"
        size="736"
        @close="handleClose"
    >
        <el-form ref="formRef" v-loading="loading" label-position="top" :model="form" :rules="rules" @submit.prevent>
            <el-form-item :label="$t('menu.website')" prop="websiteId">
                <el-select v-model="form.websiteId" clearable filterable>
                    <el-option
                        v-for="website in websites"
                        :key="website.id"
                        :label="website.primaryDomain"
                        :value="website.id"
                    />
                </el-select>
            </el-form-item>
        </el-form>
        <template #footer>
            <el-button @click="open = false" :disabled="loading">
                {{ $t('commons.button.cancel') }}
            </el-button>
            <el-button type="primary" @click="submit" :disabled="loading || !hasManagePermission">
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DrawerPro>
</template>

<script setup lang="ts">
import { FormInstance, FormRules } from 'element-plus';
import { reactive, ref } from 'vue';
import { useMenuManagePermission } from '@/composables/useMenuManagePermission';
import { AI } from '@/api/interface/ai';
import { Website } from '@/api/interface/website';
import { bindAgentWebsite } from '@/api/modules/ai';
import { getWebsiteOptions } from '@/api/modules/website';
import { Rules } from '@/global/form-rules';
import { MsgSuccess } from '@/utils/message';
import i18n from '@/lang';

const open = ref(false);
const loading = ref(false);
const { hasManagePermission } = useMenuManagePermission();
const agentName = ref('');
const websites = ref<Website.WebsiteOption[]>([]);
const formRef = ref<FormInstance>();
const form = reactive<AI.AgentWebsiteBindReq>({
    agentId: 0,
    websiteId: 0,
});
const rules = reactive<FormRules>({
    websiteId: [Rules.requiredSelectBusiness],
});

const emit = defineEmits<{ (e: 'success'): void }>();

const acceptParams = async (row: AI.AgentItem) => {
    form.agentId = row.id;
    form.websiteId = 0;
    agentName.value = row.name;
    open.value = true;
    loading.value = true;
    try {
        const res = await getWebsiteOptions({ types: ['proxy', 'static'] });
        websites.value = res.data || [];
        form.websiteId = websites.value[0]?.id || 0;
    } finally {
        loading.value = false;
    }
};

const handleClose = () => {
    websites.value = [];
    form.agentId = 0;
    form.websiteId = 0;
    agentName.value = '';
};

const submit = async () => {
    if (!formRef.value) {
        return;
    }
    await formRef.value.validate();
    loading.value = true;
    try {
        await bindAgentWebsite(form);
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        open.value = false;
        emit('success');
    } finally {
        loading.value = false;
    }
};

defineExpose({
    acceptParams,
});
</script>
