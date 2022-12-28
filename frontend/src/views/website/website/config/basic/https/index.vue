<template>
    <el-row :gutter="20">
        <el-col :span="10" :offset="1">
            <el-form
                ref="httpsForm"
                label-position="left"
                label-width="auto"
                :model="form"
                :rules="rules"
                :loading="loading"
            >
                <el-form-item prop="enable">
                    <el-checkbox v-model="form.enable">
                        {{ $t('website.enableHTTPS') }}
                    </el-checkbox>
                </el-form-item>
                <el-form-item :label="$t('website.HTTPConfig')" prop="httpConfig">
                    <el-select v-model="form.httpConfig" style="width: 240px">
                        <el-option :label="$t('website.HTTPToHTTPS')" :value="'HTTPToHTTPS'"></el-option>
                        <el-option :label="$t('website.HTTPAlso')" :value="'HTTPAlso'"></el-option>
                        <el-option :label="$t('website.HTTPSOnly')" :value="'HTTPSOnly'"></el-option>
                    </el-select>
                </el-form-item>
                <el-form-item :label="$t('website.ssl')" prop="type">
                    <el-select v-model="form.type" @change="changeType(form.type)">
                        <el-option :label="$t('website.oldSSL')" :value="'existed'"></el-option>
                        <el-option :label="$t('website.manualSSL')" :value="'manual'"></el-option>
                        <!-- <el-option :label="'自动生成证书'" :value="'auto'"></el-option> -->
                    </el-select>
                </el-form-item>
                <el-form-item :label="$t('website.select')" prop="websiteSSLId" v-if="form.type === 'existed'">
                    <el-select
                        v-model="form.websiteSSLId"
                        :placeholder="$t('website.selectSSL')"
                        @change="changeSSl(form.websiteSSLId)"
                    >
                        <el-option
                            v-for="(ssl, index) in ssls"
                            :key="index"
                            :label="ssl.primaryDomain"
                            :value="ssl.id"
                        ></el-option>
                    </el-select>
                </el-form-item>
                <div v-if="form.type === 'manual'">
                    <el-form-item :label="$t('website.privateKey')" prop="privateKey">
                        <el-input v-model="form.privateKey" :rows="6" type="textarea" />
                    </el-form-item>
                    <el-form-item :label="$t('website.certificate')" prop="certificate">
                        <el-input v-model="form.certificate" :rows="6" type="textarea" />
                    </el-form-item>
                </div>
                <el-form-item :label="' '" v-if="websiteSSL && websiteSSL.id > 0">
                    <el-descriptions :column="3" border direction="vertical">
                        <el-descriptions-item :label="$t('website.primaryDomain')">
                            {{ websiteSSL.primaryDomain }}
                        </el-descriptions-item>
                        <el-descriptions-item :label="$t('website.otherDomains')">
                            {{ websiteSSL.otherDomains }}
                        </el-descriptions-item>
                        <el-descriptions-item :label="$t('website.expireDate')">
                            {{ dateFromat(1, 1, websiteSSL.expireDate) }}
                        </el-descriptions-item>
                    </el-descriptions>
                </el-form-item>
                <el-form-item>
                    <el-button type="primary" @click="submit(httpsForm)" :loading="loading">
                        {{ $t('commons.button.save') }}
                    </el-button>
                </el-form-item>
            </el-form>
        </el-col>
    </el-row>
</template>
<script lang="ts" setup>
import { Website } from '@/api/interface/website';
import { GetHTTPSConfig, ListSSL, UpdateHTTPSConfig } from '@/api/modules/website';
import { ElMessage, FormInstance } from 'element-plus';
import { computed, onMounted, reactive, ref } from 'vue';
import { dateFromat } from '@/utils/util';
import i18n from '@/lang';
import { Rules } from '@/global/form-rules';

const props = defineProps({
    id: {
        type: Number,
        default: 0,
    },
});
const id = computed(() => {
    return props.id;
});
const httpsForm = ref<FormInstance>();
let form = reactive({
    enable: false,
    websiteId: id.value,
    websiteSSLId: undefined,
    type: 'existed',
    privateKey: '',
    certificate: '',
    httpConfig: 'HTTPToHTTPS',
});
let loading = ref(false);
const ssls = ref();
let websiteSSL = ref();
let rules = ref({
    type: [Rules.requiredSelect],
    privateKey: [Rules.requiredInput],
    certificate: [Rules.requiredInput],
    websiteSSLId: [Rules.requiredSelect],
    httpConfig: [Rules.requiredSelect],
});

const listSSL = () => {
    ListSSL({}).then((res) => {
        ssls.value = res.data;
        changeSSl(form.websiteSSLId);
    });
};

const changeSSl = (sslid: number) => {
    const res = ssls.value.filter((element: Website.SSL) => {
        return element.id == sslid;
    });
    websiteSSL.value = res[0];
};

const changeType = (type: string) => {
    if (type != 'existed') {
        websiteSSL.value = {};
        form.websiteSSLId = undefined;
    }
};

const get = () => {
    GetHTTPSConfig(id.value).then((res) => {
        console.log(res);
        if (res.data) {
            form.enable = res.data.enable;
            form.httpConfig = res.data.httpConfig;
        }
        if (res.data?.SSL && res.data?.SSL.id > 0) {
            form.websiteSSLId = res.data.SSL.id;
            websiteSSL.value = res.data.SSL;
        }
        listSSL();
    });
};
const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        loading.value = true;
        form.websiteId = id.value;
        UpdateHTTPSConfig(form)
            .then(() => {
                ElMessage.success(i18n.global.t('commons.msg.updateSuccess'));
            })
            .finally(() => {
                loading.value = false;
            });
    });
};

onMounted(() => {
    get();
});
</script>
