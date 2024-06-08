<template>
    <div v-loading="loading">
        <el-row :gutter="20" v-loading="loading">
            <el-col :xs="24" :sm="18" :md="16" :lg="16" :xl="16">
                <el-form
                    :model="form"
                    :rules="rules"
                    ref="leechRef"
                    label-position="right"
                    label-width="180px"
                    class="moblie-form"
                >
                    <el-form-item :label="$t('website.enableOrNot')">
                        <el-switch v-model="form.enable" @change="changeEnable"></el-switch>
                    </el-form-item>
                    <div v-if="form.enable">
                        <el-form-item :label="$t('website.extends')" prop="extends">
                            <el-input v-model="form.extends" type="text"></el-input>
                        </el-form-item>
                        <el-form-item :label="$t('website.browserCache')" prop="cache">
                            <el-switch v-model="form.cache" />
                        </el-form-item>
                        <el-form-item :label="$t('website.cacheTime')" prop="cacheTime" v-if="form.cache">
                            <el-input v-model.number="form.cacheTime" maxlength="15">
                                <template #append>
                                    <el-select v-model="form.cacheUint" class="w-s-button p-w-100">
                                        <el-option
                                            v-for="(unit, index) in Units"
                                            :key="index"
                                            :label="unit.label"
                                            :value="unit.value"
                                        ></el-option>
                                    </el-select>
                                </template>
                            </el-input>
                        </el-form-item>
                        <el-form-item :label="$t('website.noneRef')" prop="noneRef">
                            <el-switch v-model="form.noneRef" />
                        </el-form-item>
                        <el-form-item :label="$t('website.accessDomain')" prop="domains">
                            <el-input v-model="form.domains" type="textarea" :rows="6"></el-input>
                        </el-form-item>
                        <el-form-item :label="$t('website.leechReturn')" prop="return">
                            <el-input v-model="form.return" type="text" :maxlength="35"></el-input>
                        </el-form-item>
                    </div>
                </el-form>
                <el-button type="primary" @click="submit(leechRef, true)" :disabled="loading" v-if="form.enable">
                    {{ $t('commons.button.save') }}
                </el-button>
            </el-col>
        </el-row>
    </div>
</template>

<script setup lang="ts">
import { GetAntiLeech, ListDomains, UpdateAntiLeech } from '@/api/modules/website';
import { Rules, checkNumberRange } from '@/global/form-rules';
import { FormInstance } from 'element-plus';
import { computed, onMounted, reactive } from 'vue';
import { ref } from 'vue';
import { Units } from '@/global/mimetype';
import { MsgSuccess, MsgError } from '@/utils/message';
import i18n from '@/lang';

const loading = ref(false);
const props = defineProps({
    id: {
        type: Number,
        default: 0,
    },
});
const id = computed(() => {
    return props.id;
});
const leechRef = ref<FormInstance>();
const resData = ref({
    enable: false,
});
const form = reactive({
    enable: false,
    cache: true,
    cacheTime: 30,
    cacheUint: 'd',
    extends: 'js,css,png,jpg,jpeg,gif,ico,bmp,swf,eot,svg,ttf,woff,woff2',
    return: '404',
    domains: '',
    noneRef: true,
    logEnable: false,
    blocked: true,
    serverNames: [],
    websiteID: 0,
});

const rules = ref({
    extends: [Rules.requiredInput, Rules.leechExts],
    cacheTime: [Rules.requiredInput, checkNumberRange(1, 65535)],
    return: [Rules.requiredInput],
    domains: [Rules.requiredInput],
});

const changeEnable = (enable: boolean) => {
    if (enable) {
        ListDomains(id.value)
            .then((res) => {
                const domains = res.data || [];
                let serverNameStr = '';
                for (const param of domains) {
                    serverNameStr = serverNameStr + param.domain + '\n';
                }
                form.domains = serverNameStr;
            })
            .finally(() => {});
    }
    if (resData.value.enable && !enable) {
        ElMessageBox.confirm(i18n.global.t('website.disableLeechHelper'), i18n.global.t('website.disableLeech'), {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'error',
            closeOnClickModal: false,
            beforeClose: async (action, instance, done) => {
                if (action !== 'confirm') {
                    form.enable = true;
                    done();
                } else {
                    instance.confirmButtonLoading = true;
                    update(enable);
                    done();
                }
            },
        }).then(() => {});
    }
};

const search = async () => {
    loading.value = true;
    const res = await GetAntiLeech({ websiteID: id.value });
    loading.value = false;
    if (!res.data.enable) {
        return;
    }
    resData.value = res.data;
    form.blocked = res.data.blocked;
    form.cache = res.data.cache;
    form.enable = res.data.enable;
    if (res.data.cache) {
        form.cacheTime = res.data.cacheTime;
        form.cacheUint = res.data.cacheUint;
    }
    form.extends = res.data.extends;
    form.return = res.data.return;
    form.logEnable = res.data.enable;
    form.noneRef = res.data.noneRef;

    const serverNames = res.data.serverNames;
    let serverNameStr = '';
    for (const param of serverNames) {
        serverNameStr = serverNameStr + param + '\n';
    }
    form.domains = serverNameStr;
};

const submit = async (formEl: FormInstance | undefined, enable: boolean) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        update(enable);
    });
};

const update = async (enable: boolean) => {
    if (enable) {
        form.serverNames = form.domains.split('\n');
    }
    if (!checkReturn()) {
        return;
    }
    form.enable = enable;
    loading.value = true;
    form.websiteID = id.value;
    await UpdateAntiLeech(form)
        .then(() => {
            MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
            search();
        })
        .finally(() => {
            loading.value = false;
        });
};

const checkReturn = (): boolean => {
    let returns = form.return.split(' ');
    if (returns[0]) {
        if (isHttpStatusCode(returns[0])) {
            return true;
        } else {
            MsgError(i18n.global.t('website.leechReturnError'));
            return false;
        }
    } else {
        return false;
    }
};

function isHttpStatusCode(input: string): boolean {
    const statusCodeRegex = /^[1-5][0-9]{2}$/;
    return statusCodeRegex.test(input);
}

onMounted(() => {
    search();
});
</script>
