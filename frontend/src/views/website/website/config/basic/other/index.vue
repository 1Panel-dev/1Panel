<template>
    <el-row :gutter="20" v-loading="loading">
        <el-col :xs="24" :sm="24" :md="18" :lg="14" :xl="8">
            <el-form ref="websiteForm" label-position="right" label-width="80px" :model="form" :rules="rules">
                <el-form-item :label="$t('commons.table.name')" prop="primaryDomain">
                    <el-input v-model="form.primaryDomain"></el-input>
                </el-form-item>
                <el-form-item :label="$t('website.alias')" prop="alias">
                    <el-input v-model="form.alias" disabled></el-input>
                </el-form-item>
                <GroupSelect
                    v-model="form.webSiteGroupId"
                    :prop="'webSiteGroupId'"
                    :groupType="'website'"
                ></GroupSelect>
                <el-form-item :label="$t('website.remark')" prop="remark">
                    <el-input v-model="form.remark"></el-input>
                </el-form-item>
                <el-form-item prop="IPV6">
                    <el-checkbox v-model="form.IPV6" :label="$t('website.ipv6')" size="large" />
                </el-form-item>
                <div v-if="form.type === 'webflow'">
                    <el-divider content-position="left">Webflow</el-divider>
                    <el-form-item :label="$t('website.webflowURL')" prop="webflowURL">
                        <el-input v-model="form.webflowURL"></el-input>
                    </el-form-item>
                    <el-form-item :label="$t('website.webflowType')" prop="webflowType">
                        <el-radio-group v-model="form.webflowType">
                            <el-radio :label="'proxy'" :value="'proxy'">
                                {{ $t('website.webflowProxy') }}
                            </el-radio>
                            <el-radio :label="'static'" :value="'static'">
                                {{ $t('website.webflowStatic') }}
                            </el-radio>
                        </el-radio-group>
                    </el-form-item>
                    <el-form-item v-if="form.webflowType === 'static'">
                        <el-button type="primary" @click="sync()">
                            {{ $t('commons.button.sync') }}
                        </el-button>
                    </el-form-item>
                </div>

                <el-form-item>
                    <el-button type="primary" @click="submit(websiteForm)" :disabled="loading">
                        {{ $t('commons.button.save') }}
                    </el-button>
                </el-form-item>
            </el-form>
        </el-col>
    </el-row>
</template>

<script lang="ts" setup>
import GroupSelect from '@/views/website/website/components/group/index.vue';

import { getWebsite, updateWebsite, syncWebflow, updateWebflow } from '@/api/modules/website';
import { Rules } from '@/global/form-rules';
import { computed, onMounted, reactive, ref } from 'vue';
import { FormInstance } from 'element-plus';
import i18n from '@/lang';
import { MsgError, MsgSuccess } from '@/utils/message';

const websiteForm = ref<FormInstance>();
const props = defineProps({
    id: {
        type: Number,
        default: 0,
    },
});
const websiteId = computed(() => {
    return Number(props.id);
});
const loading = ref(false);
const form = reactive({
    id: websiteId.value,
    primaryDomain: '',
    remark: '',
    webSiteGroupId: 0,
    IPV6: false,
    alias: '',
    favorite: false,
    type: '',
    webflowURL: '',
    webflowType: '',
});
const rules = ref({
    primaryDomain: [Rules.requiredInput, Rules.linuxName],
    webSiteGroupId: [Rules.requiredSelect],
    webflowURL: [Rules.paramHttp],
});

const sync = async () => {
    loading.value = true;
    try {
        await syncWebflow({ websiteID: websiteId.value });
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    } finally {
        loading.value = false;
    }
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        if (form.remark && form.remark.length > 128) {
            MsgError(i18n.global.t('commons.rule.length128Err'));
            return;
        }
        loading.value = true;
        try {
            await updateWebsite(form);
            if (form.type === 'webflow') {
                await updateWebflow({ id: form.id, webflowURL: form.webflowURL, webflowType: form.webflowType });
            }
            MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
            search();
        } finally {
            loading.value = false;
        }
    });
};
const search = async () => {
    getWebsite(websiteId.value).then((res) => {
        form.primaryDomain = res.data.primaryDomain;
        form.remark = res.data.remark;
        form.webSiteGroupId = res.data.webSiteGroupId;
        form.IPV6 = res.data.IPV6;
        form.alias = res.data.alias;
        form.favorite = res.data.favorite;
        form.type = res.data.type;
        form.webflowURL = res.data.webflowURL;
        form.webflowType = res.data.webflowType;
    });
};

onMounted(() => {
    search();
});
</script>
