<template>
    <el-row>
        <el-col :xs="24" :sm="18" :md="18" :lg="14" :xl="14">
            <el-form
                ref="corsForm"
                label-position="right"
                label-width="150px"
                :model="corsSetting"
                :rules="rules"
                v-loading="loading"
            >
                <CorsSetting
                    v-model="corsSetting.cors"
                    :config="{
                        allowOrigins: corsSetting.allowOrigins,
                        allowMethods: corsSetting.allowMethods,
                        allowHeaders: corsSetting.allowHeaders,
                        allowCredentials: corsSetting.allowCredentials,
                        preflight: corsSetting.preflight,
                    }"
                    enable-size="small"
                    @update:config="updateonfig"
                ></CorsSetting>

                <el-form-item>
                    <el-button type="primary" @click="submit()">
                        {{ $t('commons.button.save') }}
                    </el-button>
                </el-form-item>
            </el-form>
        </el-col>
    </el-row>
</template>

<script lang="ts" setup>
import { getCorsConfig, updateCorsConfig } from '@/api/modules/website';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import CorsSetting from '@/views/website/website/cors/index.vue';
import { FormInstance } from 'element-plus';

const props = defineProps({
    id: {
        type: Number,
        default: 0,
    },
});

const rules = ref({
    allowOrigins: [Rules.requiredInput],
});

const corsForm = ref<FormInstance>();
const loading = ref(false);

const corsSetting = ref({
    cors: false,
    allowOrigins: '*',
    allowMethods: 'GET,POST,OPTIONS,PUT,DELETE',
    allowHeaders: '',
    allowCredentials: false,
    preflight: true,
    websiteID: props.id,
});

const updateonfig = (config: any) => {
    corsSetting.value.allowOrigins = config.allowOrigins;
    corsSetting.value.allowMethods = config.allowMethods;
    corsSetting.value.allowHeaders = config.allowHeaders;
    corsSetting.value.allowCredentials = config.allowCredentials;
    corsSetting.value.preflight = config.preflight;
};

const submit = async () => {
    const isValid = await corsForm.value?.validate();
    if (!isValid) return;

    corsSetting.value.websiteID = props.id;
    try {
        loading.value = true;
        await updateCorsConfig(corsSetting.value);
        MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
    } catch (error) {
        return;
    } finally {
        loading.value = false;
    }
};

const getCors = () => {
    getCorsConfig(props.id).then((res: any) => {
        if (res.data) {
            corsSetting.value = res.data;
        }
    });
};

onMounted(() => {
    getCors();
});
</script>
