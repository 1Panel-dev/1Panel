<template>
    <el-row :gutter="20">
        <el-col :span="8" :offset="2">
            <el-form ref="defaultForm" label-position="top" :model="defaultModel" :rules="rules" :loading="loading">
                <el-form-item :label="$t('website.defaultDoc')" prop="index">
                    <el-input
                        v-model="defaultModel.index"
                        type="textarea"
                        :autosize="{ minRows: 8, maxRows: 20 }"
                    ></el-input>
                </el-form-item>
            </el-form>
            <el-button type="primary" @click="submit(defaultForm)" :loading="loading">
                {{ $t('commons.button.save') }}
            </el-button>
        </el-col>
    </el-row>
</template>

<script lang="ts" setup>
import { WebSite } from '@/api/interface/website';
import { GetNginxConfig, UpdateNginxConfig } from '@/api/modules/website';
import { Rules } from '@/global/form-rules';
import { ElMessage, FormInstance } from 'element-plus';
import { computed, onMounted, ref } from 'vue';
import i18n from '@/lang';

const props = defineProps({
    id: {
        type: Number,
        default: 0,
    },
});
const websiteId = computed(() => {
    return Number(props.id);
});
const defaultForm = ref<FormInstance>();
let rules = ref({
    index: [Rules.requiredInput],
});
let defaultModel = ref({
    index: '',
});
let req = ref({
    operate: 'update',
    scope: 'index',
    websiteId: websiteId.value,
    params: {},
});

let loading = ref(false);

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        req.value.params = defaultModel.value;
        loading.value = true;
        UpdateNginxConfig(req.value)
            .then(() => {
                ElMessage.success(i18n.global.t('commons.msg.updateSuccess'));
                search(req.value);
            })
            .finally(() => {
                loading.value = false;
            });
    });
};

const search = (req: WebSite.NginxConfigReq) => {
    loading.value = true;
    GetNginxConfig(req)
        .then((res) => {
            if (res.data && res.data.length > 0) {
                const indexParam = res.data[0];
                let values = '';
                for (const param of indexParam.params) {
                    values = values + param + '\n';
                }
                defaultModel.value.index = values;
            }
        })
        .finally(() => {
            loading.value = false;
        });
};

onMounted(() => {
    search(req.value);
});
</script>
