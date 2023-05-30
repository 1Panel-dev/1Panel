<template>
    <el-row :gutter="20" v-loading="loading">
        <el-col :xs="24" :sm="18" :md="8" :lg="8" :xl="8">
            <el-form ref="limitForm" label-position="right" :model="form" :rules="rules" label-width="100px">
                <el-form-item prop="enable" :label="$t('website.enableOrNot')">
                    <el-switch v-model="enable" @change="changeEnable"></el-switch>
                </el-form-item>
                <el-form-item :label="$t('website.limit')">
                    <el-select v-model="ruleKey" @change="changeRule(ruleKey)">
                        <el-option :label="$t('website.current')" :value="'current'"></el-option>
                        <el-option
                            v-for="(limit, index) in limitRules"
                            :key="index"
                            :label="limit.key"
                            :value="limit.key"
                        ></el-option>
                    </el-select>
                </el-form-item>
                <el-form-item :label="$t('website.perserver')" prop="perserver">
                    <el-input v-model.number="form.perserver" maxlength="15"></el-input>
                    <span class="input-help">{{ $t('website.perserverHelper') }}</span>
                </el-form-item>
                <el-form-item :label="$t('website.perip')" prop="perip">
                    <el-input v-model.number="form.perip" maxlength="15"></el-input>
                    <span class="input-help">{{ $t('website.peripHelper') }}</span>
                </el-form-item>
                <el-form-item :label="$t('website.rate')" prop="rate">
                    <el-input v-model.number="form.rate" maxlength="15"></el-input>
                    <span class="input-help">{{ $t('website.rateHelper') }}</span>
                </el-form-item>
            </el-form>
            <el-button type="primary" @click="submit(limitForm)" :disabled="loading">
                <span v-if="enable">{{ $t('commons.button.save') }}</span>
                <span v-else>{{ $t('commons.button.saveAndEnable') }}</span>
            </el-button>
        </el-col>
    </el-row>
</template>

<script lang="ts" setup>
import { checkNumberRange, Rules } from '@/global/form-rules';
import { Website } from '@/api/interface/website';
import { GetNginxConfig, UpdateNginxConfig } from '@/api/modules/website';
import { FormInstance } from 'element-plus';
import { computed, onMounted, reactive, ref } from 'vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';

const props = defineProps({
    id: {
        type: Number,
        default: 0,
    },
});
const websiteId = computed(() => {
    return Number(props.id);
});
let rules = reactive({
    perserver: [Rules.requiredInput, checkNumberRange(1, 65535)],
    perip: [Rules.requiredInput, checkNumberRange(1, 65535)],
    rate: [Rules.requiredInput, checkNumberRange(1, 99999999)],
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
let scopeReq = reactive({
    scope: 'limit-conn',
    websiteId: websiteId.value,
});
let enable = ref(false);
let loading = ref(false);

const limitRules = [
    { key: i18n.global.t('website.blog'), values: [300, 25, 512] },
    { key: i18n.global.t('website.imageSite'), values: [200, 10, 1024] },
    { key: i18n.global.t('website.downloadSite'), values: [50, 3, 2048] },
    { key: i18n.global.t('website.shopSite'), values: [500, 10, 2048] },
    { key: i18n.global.t('website.doorSite'), values: [400, 15, 1024] },
    { key: i18n.global.t('website.qiteSite'), values: [60, 10, 512] },
    { key: i18n.global.t('website.videoSite'), values: [150, 4, 1024] },
];

let ruleKey = ref('');

const search = (scopeReq: Website.NginxScopeReq) => {
    loading.value = true;
    GetNginxConfig(scopeReq)
        .then((res) => {
            ruleKey.value = 'current';
            if (res.data) {
                enable.value = res.data.enable;
                if (res.data.enable == false) {
                    req.operate = 'add';
                }
                for (const param of res.data.params) {
                    if (param.name === 'limit_conn') {
                        if (param.params[0] === 'perserver' && param.params[1]) {
                            form.perserver = Number(param.params[1].match(/\d+/g));
                        }
                        if (param.params[0] === 'perip' && param.params[1]) {
                            form.perip = Number(param.params[1].match(/\d+/g));
                        }
                    }
                    if (param.name === 'limit_rate' && param.params[0]) {
                        form.rate = Number(param.params[0].match(/\d+/g));
                    }
                }
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
        if (req.operate === 'add') {
            enable.value = true;
        }
        UpdateNginxConfig(req)
            .then(() => {
                MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
                search(req);
            })
            .finally(() => {
                if (req.operate === 'add') {
                    enable.value = false;
                }
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

const changeRule = (key: string) => {
    limitRules.forEach((limit) => {
        if (limit.key === key) {
            form.perserver = limit.values[0];
            form.perip = limit.values[1];
            form.rate = limit.values[2];
        }
    });
};

onMounted(() => {
    search(scopeReq);
});
</script>
