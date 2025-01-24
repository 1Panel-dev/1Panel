<template>
    <el-row :gutter="20" v-loading="loading">
        <el-col :xs="24" :sm="14" :md="14" :lg="10" :xl="8">
            <el-alert :closable="false">
                {{ $t('website.ipFromHelper') }}
                <div>
                    {{ $t('website.ipFromExample1') }}
                </div>
                <div>
                    {{ $t('website.ipFromExample2') }}
                </div>
                <div>
                    {{ $t('website.ipFromExample3') }}
                </div>
            </el-alert>
            <el-form
                v-loading="loading"
                @submit.prevent
                ref="realIPForm"
                label-position="right"
                label-width="100px"
                :model="req"
                :rules="rules"
                :validate-on-rule-change="false"
            >
                <el-form-item :label="$t('commons.button.start')" prop="open">
                    <el-switch v-model="req.open"></el-switch>
                </el-form-item>
                <div v-if="req.open">
                    <el-form-item :label="$t('website.ipFrom')" prop="ipFrom">
                        <el-input
                            type="textarea"
                            :rows="10"
                            clearable
                            v-model="req.ipFrom"
                            :placeholder="$t('website.wafInputHelper')"
                        ></el-input>
                        <span class="input-help">
                            {{ $t('website.wafInputHelper') }}
                        </span>
                    </el-form-item>
                    <el-form-item label="IP Header" prop="ipHeader">
                        <el-select v-model="req.ipHeader">
                            <el-option :label="$t('website.other')" key="other" value="other"></el-option>
                            <el-option
                                v-for="item in ['X-Forwarded-For', 'X-Real-IP', 'CF-Connecting-IP']"
                                :key="item"
                                :label="item"
                                :value="item"
                            />
                        </el-select>
                    </el-form-item>

                    <el-form-item prop="ipOther" v-if="req.ipHeader === 'other'">
                        <el-input type="text" v-model.trim="req.ipOther" />
                    </el-form-item>
                </div>
                <el-form-item>
                    <el-button type="primary" @click="submit(realIPForm)" :loading="loading">
                        {{ $t('commons.button.save') }}
                    </el-button>
                </el-form-item>
            </el-form>
        </el-col>
    </el-row>
</template>

<script setup lang="ts">
import { getRealIPConfig, updateRealIPConfig } from '@/api/modules/website';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { FormInstance } from 'element-plus';

const loading = ref(false);
const realIPForm = ref<FormInstance>();
const props = defineProps({
    id: {
        type: Number,
        default: 0,
    },
});

const req = reactive({
    websiteID: 0,
    open: false,
    ipFrom: '127.0.0.1',
    ipHeader: 'X-Real-IP',
    ipOther: '',
});
const rules = {
    ipFrom: [Rules.requiredInput],
    ipHeader: [Rules.requiredSelect],
    ipOther: [Rules.requiredInput],
};

const get = () => {
    getRealIPConfig(props.id).then((res) => {
        req.open = res.data.open;
        if (res.data.open) {
            req.ipFrom = res.data.ipFrom;
            req.ipHeader = res.data.ipHeader;
            req.ipOther = res.data.ipOther;
        }
    });
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate(async (valid) => {
        if (!valid) {
            return;
        }
        req.websiteID = props.id;
        try {
            await updateRealIPConfig(req);
            MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
        } catch (error) {}
    });
};

onMounted(() => {
    get();
});
</script>
