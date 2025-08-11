<template>
    <el-row :gutter="20" v-loading="loading">
        <el-col :xs="24" :sm="18" :md="8" :lg="8" :xl="8">
            <el-form ref="websiteForm" label-position="right" label-width="80px" :model="form" :rules="rules">
                <el-form-item :label="$t('website.primaryDomain')" prop="primaryDomain">
                    <el-input v-model="form.primaryDomain"></el-input>
                </el-form-item>
                <el-form-item :label="$t('website.alias')" prop="primaryDomain">
                    <el-input v-model="form.alias" disabled></el-input>
                </el-form-item>
                <el-form-item :label="$t('commons.table.group')" prop="webSiteGroupID">
                    <el-select v-model="form.webSiteGroupId">
                        <el-option
                            v-for="(group, index) in groups"
                            :key="index"
                            :label="group.name == 'Default' ? $t('commons.table.default') : group.name"
                            :value="group.id"
                        ></el-option>
                    </el-select>
                </el-form-item>
                <el-form-item :label="$t('website.remark')" prop="remark">
                    <el-input v-model="form.remark"></el-input>
                </el-form-item>
                <el-form-item prop="IPV6">
                    <el-checkbox v-model="form.IPV6" :label="$t('website.ipv6')" size="large" />
                </el-form-item>
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
import { getWebsite, updateWebsite } from '@/api/modules/website';
import { Rules } from '@/global/form-rules';
import { computed, onMounted, reactive, ref } from 'vue';
import { FormInstance } from 'element-plus';
import i18n from '@/lang';
import { MsgError, MsgSuccess } from '@/utils/message';
import { getAgentGroupList } from '@/api/modules/group';
import { Group } from '@/api/interface/group';

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
});
const rules = ref({
    primaryDomain: [Rules.requiredInput],
    webSiteGroupId: [Rules.requiredSelect],
});
const groups = ref<Group.GroupInfo[]>([]);

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
        updateWebsite(form)
            .then(() => {
                MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
                search();
            })
            .finally(() => {
                loading.value = false;
            });
    });
};
const search = async () => {
    const res = await getAgentGroupList('website');
    groups.value = res.data;

    getWebsite(websiteId.value).then((res) => {
        form.primaryDomain = res.data.primaryDomain;
        form.remark = res.data.remark;
        form.webSiteGroupId = res.data.webSiteGroupId;
        form.IPV6 = res.data.IPV6;
        form.alias = res.data.alias;
        form.favorite = res.data.favorite;
    });
};

onMounted(() => {
    search();
});
</script>
