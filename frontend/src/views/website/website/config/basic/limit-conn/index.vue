<template>
    <el-row :gutter="20">
        <el-col :span="8" :offset="2">
            <el-checkbox v-model="enable" @change="changeEnable">{{ $t('website.limtHelper') }}</el-checkbox>
            <el-form ref="limitForm" label-position="left" :model="form" :rules="rules" :loading="loading">
                <el-form-item :label="$t('website.perserver')" prop="perserver">
                    <el-input v-model="form.perserver"></el-input>
                    <span class="input-help">{{ $t('website.perserverHelper') }}</span>
                </el-form-item>
                <el-form-item :label="$t('website.perip')" prop="perip">
                    <el-input v-model="form.perip"></el-input>
                    <span class="input-help">{{ $t('website.peripHelper') }}</span>
                </el-form-item>
                <el-form-item :label="$t('website.rate')" prop="rate">
                    <el-input v-model="form.rate"></el-input>
                    <span class="input-help">{{ $t('website.rateHelper') }}</span>
                </el-form-item>
            </el-form>
            <el-button type="primary" @click="submit(limitForm)" :loading="loading">
                {{ $t('commons.button.save') }}
            </el-button>
        </el-col>
    </el-row>
</template>

<script lang="ts" setup>
import { Rules } from '@/global/form-rules';
import { WebSite } from '@/api/interface/website';
import { GetNginxConfig, UpdateNginxConfig } from '@/api/modules/website';
import { ElMessage, FormInstance } from 'element-plus';
import { computed, onMounted, reactive, ref } from 'vue';
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
let rules = ref({
    perserver: [Rules.requiredInput],
    perip: [Rules.requiredInput],
    rate: [Rules.requiredInput],
});
const limitForm = ref<FormInstance>();
let form = reactive({
    perserver: 300,
    perip: 25,
    rate: 512,
});
let req = reactive({
    operate: 'update',
    scope: 'limit-conn',
    websiteId: websiteId.value,
    params: [{}],
});
let enable = ref(false);
let loading = ref(false);

const search = (req: WebSite.NginxConfigReq) => {
    loading.value = true;
    GetNginxConfig(req)
        .then((res) => {
            if (res.data && res.data.length > 0) {
                enable.value = true;
                for (const param of res.data) {
                    if (param.name === 'limit_conn') {
                        if (param.secondKey === 'perserver') {
                            form.perserver = Number(param.params[1]);
                        }
                        if (param.secondKey === 'perip') {
                            form.perip = Number(param.params[1]);
                        }
                    }
                    if (param.name === 'limit_rate') {
                        form.rate = Number(param.params[0].match(/\d+/g));
                    }
                }
            } else {
                enable.value = false;
                req.operate = 'add';
            }
        })
        .finally(() => {
            loading.value = false;
        });
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        loading.value = true;
        let params = [
            {
                limit_conn: 'perserver ' + String(form.perserver),
            },
            {
                limit_conn: 'perip ' + String(form.perip),
            },
            {
                limit_rate: String(form.rate) + 'k',
            },
        ];
        req.params = params;
        UpdateNginxConfig(req)
            .then(() => {
                ElMessage.success(i18n.global.t('commons.msg.updateSuccess'));
                search(req);
            })
            .finally(() => {
                loading.value = false;
            });
    });
};

const changeEnable = () => {
    if (!enable.value) {
        req.operate = 'delete';
    } else {
        req.operate = 'add';
    }
    submit(limitForm.value);
};

onMounted(() => {
    search(req);
});
</script>
