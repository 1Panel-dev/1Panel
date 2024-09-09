<template>
    <el-form :model="params" :rules="variablesRules" ref="phpFormRef" label-position="top" v-loading="loading">
        <el-row v-loading="loading">
            <el-col :span="22" :offset="1">
                <el-form-item :label="$t('runtime.operateMode')" prop="pm">
                    <el-select v-model="params.pm">
                        <el-option :label="$t('runtime.dynamic')" :value="'dynamic'"></el-option>
                        <el-option :label="$t('runtime.static')" :value="'static'"></el-option>
                        <el-option :label="$t('runtime.ondemand')" :value="'ondemand'"></el-option>
                    </el-select>
                    <span class="input-help">
                        <el-text v-if="params.pm == 'dynamic'">{{ $t('runtime.dynamicHelper') }}</el-text>
                        <el-text v-if="params.pm == 'static'">{{ $t('runtime.staticHelper') }}</el-text>
                        <el-text v-if="params.pm == 'ondemand'">{{ $t('runtime.ondemandHelper') }}</el-text>
                    </span>
                </el-form-item>
                <el-form-item label="max_children" prop="pm.max_children">
                    <el-input clearable v-model.number="params['pm.max_children']" maxlength="15"></el-input>
                    <span class="input-help">
                        {{ $t('runtime.max_children') }}
                    </span>
                </el-form-item>
                <el-form-item label="start_servers" prop="pm.start_servers">
                    <el-input clearable v-model.number="params['pm.start_servers']" maxlength="15"></el-input>
                    <span class="input-help">
                        {{ $t('runtime.start_servers') }}
                    </span>
                </el-form-item>
                <el-form-item label="min_spare_servers" prop="pm.min_spare_servers">
                    <el-input clearable v-model.number="params['pm.min_spare_servers']" maxlength="15"></el-input>
                    <span class="input-help">
                        {{ $t('runtime.min_spare_servers') }}
                    </span>
                </el-form-item>
                <el-form-item label="max_spare_servers" prop="pm.max_spare_servers">
                    <el-input clearable v-model.number="params['pm.max_spare_servers']" maxlength="15"></el-input>
                    <span class="input-help">
                        {{ $t('runtime.max_spare_servers') }}
                    </span>
                </el-form-item>
                <el-form-item>
                    <el-button type="primary" @click="onSaveStart(phpFormRef)">
                        {{ $t('commons.button.save') }}
                    </el-button>
                </el-form-item>
            </el-col>
        </el-row>
    </el-form>
</template>

<script lang="ts" setup>
import { GetFPMConfig, UpdateFPMConfig } from '@/api/modules/runtime';
import { checkNumberRange, Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { FormInstance } from 'element-plus';
import { computed, onMounted, reactive, ref } from 'vue';

const props = defineProps({
    id: {
        type: Number,
        default: 0,
    },
});

const id = computed(() => {
    return props.id;
});
const loading = ref(false);
const phpFormRef = ref();
const initData = () => {
    return {
        pm: 'dynamic',
        'pm.max_children': 150,
        'pm.start_servers': 10,
        'pm.min_spare_servers': 10,
        'pm.max_spare_servers': 30,
    };
};
const params = reactive(initData());
const variablesRules = reactive({
    pm: [Rules.requiredSelect],
    'pm.max_children': [checkNumberRange(0, 5000)],
    'pm.start_servers': [checkNumberRange(0, 99999)],
    'pm.min_spare_servers': [checkNumberRange(0, 99999)],
    'pm.max_spare_servers': [checkNumberRange(0, 99999)],
});

const get = () => {
    loading.value = true;
    GetFPMConfig(id.value)
        .then((res) => {
            const resParams = res.data.params;
            params['pm'] = resParams['pm'];
            params['pm.max_children'] = Number(resParams['pm.max_children']);
            params['pm.start_servers'] = Number(resParams['pm.start_servers']);
            params['pm.min_spare_servers'] = Number(resParams['pm.min_spare_servers']);
            params['pm.max_spare_servers'] = Number(resParams['pm.max_spare_servers']);
        })
        .finally(() => {
            loading.value = false;
        });
};

const onSaveStart = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        const action = await ElMessageBox.confirm(
            i18n.global.t('runtime.phpConfigHelper'),
            i18n.global.t('database.confChange'),
            {
                confirmButtonText: i18n.global.t('commons.button.confirm'),
                cancelButtonText: i18n.global.t('commons.button.cancel'),
                type: 'info',
            },
        );
        if (action === 'confirm') {
            loading.value = true;
            submit();
        }
    });
};

const submit = async () => {
    loading.value = true;
    UpdateFPMConfig({ id: id.value, params: params })
        .then(() => {
            MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
            get();
        })
        .finally(() => {
            loading.value = false;
        });
};

onMounted(() => {
    get();
});
</script>
