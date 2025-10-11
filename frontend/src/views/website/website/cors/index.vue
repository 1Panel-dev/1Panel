<template>
    <div>
        <div class="flex justify-between items-center py-3" v-if="enableSize == 'large'">
            <div class="flex flex-col gap-1">
                <span class="font-medium">{{ $t('website.enableCors') }}</span>
            </div>
            <el-switch v-model="corsEnabled" size="large" @change="handleCorsChange" />
        </div>

        <el-form-item :label="$t('website.enableCors')" v-if="enableSize == 'small'">
            <el-switch v-model="corsEnabled" size="large" @change="handleCorsChange" />
        </el-form-item>

        <el-collapse-transition>
            <div v-if="corsEnabled" class="mt-4">
                <el-form-item :label="$t('website.allowOrigins')" prop="allowOrigins">
                    <el-input
                        v-model="corsConfig.allowOrigins"
                        type="textarea"
                        placeholder="*"
                        @input="emitUpdate"
                    ></el-input>
                </el-form-item>

                <el-form-item :label="$t('website.allowMethods')" prop="allowMethods">
                    <el-input
                        v-model="corsConfig.allowMethods"
                        type="textarea"
                        placeholder="GET,POST,OPTIONS,PUT,DELETE"
                        @input="emitUpdate"
                    ></el-input>
                </el-form-item>

                <el-form-item :label="$t('website.allowHeaders')" prop="allowHeaders">
                    <el-input v-model="corsConfig.allowHeaders" type="textarea" @input="emitUpdate"></el-input>
                </el-form-item>

                <el-form-item :label="$t('website.allowCredentials')" prop="allowCredentials">
                    <el-switch v-model="corsConfig.allowCredentials" @change="emitUpdate" />
                </el-form-item>

                <el-form-item :label="$t('website.preflight')" prop="preflight">
                    <el-switch v-model="corsConfig.preflight" @change="emitUpdate" />
                    <span class="input-help">{{ $t('website.preflightHleper') }}</span>
                </el-form-item>
            </div>
        </el-collapse-transition>
    </div>
</template>

<script lang="ts" setup>
import { ref, watch } from 'vue';

interface CorsConfig {
    allowOrigins: string;
    allowMethods: string;
    allowHeaders: string;
    allowCredentials: boolean;
    preflight: boolean;
}

interface Props {
    modelValue: boolean;
    config: CorsConfig;
    enableSize: string;
}

const props = withDefaults(defineProps<Props>(), {
    modelValue: false,
    config: () => ({
        allowOrigins: '*',
        allowMethods: 'GET,POST,OPTIONS,PUT,DELETE',
        allowHeaders: '',
        allowCredentials: false,
        preflight: true,
    }),
    enableSize: 'large',
});

const emit = defineEmits(['update:modelValue', 'update:config']);

const corsEnabled = ref(props.modelValue);
const corsConfig = ref<CorsConfig>({ ...props.config });

watch(
    () => props.modelValue,
    (val) => {
        corsEnabled.value = val;
    },
);

watch(
    () => props.config,
    (val) => {
        corsConfig.value = { ...val };
    },
    { deep: true },
);

const handleCorsChange = (enabled: boolean) => {
    emit('update:modelValue', enabled);

    if (enabled) {
        corsConfig.value = {
            allowOrigins: '*',
            allowMethods: 'GET,POST,OPTIONS,PUT,DELETE',
            allowHeaders: '',
            allowCredentials: false,
            preflight: true,
        };
    } else {
        corsConfig.value = {
            allowOrigins: '',
            allowMethods: '',
            allowHeaders: '',
            allowCredentials: false,
            preflight: true,
        };
    }
    emitUpdate();
};

const emitUpdate = () => {
    emit('update:config', corsConfig.value);
};
</script>
