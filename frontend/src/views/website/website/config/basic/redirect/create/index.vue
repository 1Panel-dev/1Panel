<template>
    <el-drawer
        v-model="open"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        size="40%"
        :before-close="handleClose"
    >
        <template #header>
            <DrawerHeader :header="$t('commons.button.' + redirect.operate)" :back="handleClose" />
        </template>
        <el-row v-loading="loading">
            <el-col :span="22" :offset="1">
                <el-form ref="redirectForm" label-position="top" :model="redirect" :rules="rules">
                    <el-form-item :label="$t('commons.table.name')" prop="name">
                        <el-input
                            v-model.trim="redirect.name"
                            :disabled="redirect.operate === 'edit' || redirect.type == '404'"
                        ></el-input>
                    </el-form-item>
                    <el-form-item :label="$t('commons.table.type')" prop="type">
                        <el-select
                            v-model="redirect.type"
                            @change="changeType(redirect.type)"
                            :disabled="redirect.operate === 'edit'"
                        >
                            <el-option :label="$t('website.domain')" :value="'domain'"></el-option>
                            <el-option :label="$t('website.path')" :value="'path'"></el-option>
                            <el-option :label="'404'" :value="'404'"></el-option>
                        </el-select>
                    </el-form-item>
                    <el-form-item :label="$t('website.redirectWay')" prop="redirect">
                        <el-select v-model="redirect.redirect">
                            <el-option :label="'301'" :value="'301'"></el-option>
                            <el-option :label="'302'" :value="'302'"></el-option>
                        </el-select>
                        <span class="input-help">
                            {{ $t('website.redirectHelper') }}
                        </span>
                    </el-form-item>
                    <el-form-item :label="$t('website.path')" prop="path" v-if="redirect.type == 'path'">
                        <el-input v-model.trim="redirect.path"></el-input>
                    </el-form-item>
                    <el-form-item :label="$t('website.domain')" prop="domains" v-if="redirect.type == 'domain'">
                        <el-select v-model="redirect.domains" multiple>
                            <el-option
                                v-for="(item, index) in domains"
                                :key="index"
                                :value="item.domain"
                                :label="item.domain"
                            />
                        </el-select>
                    </el-form-item>
                    <el-form-item :label="$t('website.redirectRoot')" prop="redirectRoot" v-if="redirect.type == '404'">
                        <el-switch v-model="redirect.redirectRoot"></el-switch>
                    </el-form-item>
                    <el-form-item :label="$t('website.targetURL')" prop="target" v-if="!redirect.redirectRoot">
                        <el-input v-model.trim="redirect.target"></el-input>
                    </el-form-item>
                    <el-form-item :label="$t('website.keepPath')" prop="keepPath" v-if="redirect.type !== '404'">
                        <el-switch v-model="redirect.keepPath"></el-switch>
                    </el-form-item>
                </el-form>
            </el-col>
        </el-row>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit(redirectForm)" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-drawer>
</template>

<script lang="ts" setup>
import DrawerHeader from '@/components/drawer-header/index.vue';
import { ListDomains, OperateRedirectConfig, GetRedirectConfig } from '@/api/modules/website';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { FormInstance } from 'element-plus';
import { ref } from 'vue';
import { MsgSuccess } from '@/utils/message';
import { Website } from '@/api/interface/website';

const redirectForm = ref<FormInstance>();
const rules = ref({
    name: [Rules.requiredInput, Rules.appName],
    type: [Rules.requiredSelect],
    redirect: [Rules.requiredSelect],
    domains: [Rules.requiredSelect],
    target: [Rules.requiredInput],
    path: [Rules.requiredInput],
});
const open = ref(false);
const loading = ref(false);

const initData = (): Website.RedirectConfig => ({
    websiteID: 0,
    operate: 'create',
    enable: true,
    name: '',
    domains: [],
    keepPath: true,
    type: 'domain',
    redirect: '301',
    target: 'http://',
    redirectRoot: false,
});
let redirect = ref(initData());
const em = defineEmits(['close']);
const handleClose = () => {
    redirectForm.value?.resetFields();
    open.value = false;
    em('close', false);
};
const domains = ref([]);

const acceptParams = (redirectParam: Website.RedirectConfig) => {
    if (redirectParam.operate == 'edit') {
        redirect.value = redirectParam;
    } else {
        redirect.value = initData();
        redirect.value.websiteID = redirectParam.websiteID;
    }
    domains.value = [];
    getDomains();
    open.value = true;
};

const changeType = (type: string) => {
    redirectForm.value?.clearValidate('name');
    if (type != '404') {
        redirect.value.name = '';
        redirect.value.redirectRoot = false;
    } else {
        redirect.value.name = '404';
    }
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        loading.value = true;
        OperateRedirectConfig(redirect.value)
            .then(() => {
                if (redirect.value.operate == 'create') {
                    MsgSuccess(i18n.global.t('commons.msg.createSuccess'));
                } else {
                    MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
                }
                handleClose();
            })
            .finally(() => {
                loading.value = false;
            });
    });
};

const getDomains = async () => {
    try {
        loading.value = true;
        const res = await GetRedirectConfig({ websiteID: redirect.value.websiteID });
        let oldDomains = [];
        if (res.data) {
            for (const old of res.data) {
                if (old.type == 'domain') {
                    oldDomains = oldDomains.concat(old.domains);
                }
            }
        }
        ListDomains(redirect.value.websiteID)
            .then((domainRes) => {
                if (domainRes.data) {
                    if (oldDomains.length > 0) {
                        for (const data of domainRes.data) {
                            if (oldDomains.indexOf(data.domain) > -1) {
                                continue;
                            }
                            domains.value.push(data);
                        }
                    } else {
                        domains.value = domainRes.data || [];
                    }
                }
            })
            .finally(() => {});
    } catch (error) {
    } finally {
        loading.value = false;
    }
};

defineExpose({
    acceptParams,
});
</script>
