<template>
    <el-row :gutter="20">
        <el-col :span="8" :offset="2">
            <el-form
                ref="websiteForm"
                label-position="right"
                label-width="80px"
                :model="form"
                :rules="rules"
                :loading="loading"
            >
                <el-form-item :label="$t('website.primaryDomain')" prop="primaryDomain">
                    <el-input v-model="form.primaryDomain"></el-input>
                </el-form-item>
                <el-form-item :label="$t('website.group')" prop="webSiteGroupID">
                    <el-select v-model="form.webSiteGroupId">
                        <el-option
                            v-for="(group, index) in groups"
                            :key="index"
                            :label="group.name"
                            :value="group.id"
                        ></el-option>
                    </el-select>
                </el-form-item>
                <el-form-item :label="$t('website.remark')" prop="remark">
                    <el-input v-model="form.remark"></el-input>
                </el-form-item>
                <el-form-item>
                    <el-button type="primary" @click="submit(websiteForm)" :loading="loading">
                        {{ $t('commons.button.save') }}
                    </el-button>
                </el-form-item>
            </el-form>
        </el-col>
    </el-row>
</template>

<script lang="ts" setup>
import { WebSite } from '@/api/interface/website';
import { GetWebsite, UpdateWebsite } from '@/api/modules/website';
import { Rules } from '@/global/form-rules';
import { computed, onMounted, reactive, ref } from 'vue';
import { ListGroups } from '@/api/modules/website';
import { ElMessage, FormInstance } from 'element-plus';
import i18n from '@/lang';

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
let loading = ref(false);
let form = reactive({
    id: websiteId.value,
    primaryDomain: '',
    remark: '',
    webSiteGroupId: 0,
});
let rules = ref({
    primaryDomain: [Rules.requiredInput],
    webSiteGroupId: [Rules.requiredSelect],
});
let groups = ref<WebSite.Group[]>([]);

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        loading.value = true;
        UpdateWebsite(form)
            .then(() => {
                ElMessage.success(i18n.global.t('commons.msg.updateSuccess'));
                search();
            })
            .finally(() => {
                loading.value = false;
            });
    });
};
const search = () => {
    ListGroups().then((res) => {
        groups.value = res.data;
        GetWebsite(websiteId.value).then((res) => {
            // form.id = res.data.id;
            form.primaryDomain = res.data.primaryDomain;
            form.remark = res.data.remark;
            form.webSiteGroupId = res.data.webSiteGroupId;
        });
    });
};

onMounted(() => {
    search();
});
</script>
