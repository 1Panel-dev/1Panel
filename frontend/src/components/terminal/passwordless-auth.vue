<template>
    <div class="passwordless-auth">
        <el-form-item :label="$t('terminal.bootstrapAuth')">
            <el-radio-group v-model="bootstrapMode" @change="emit('change')">
                <el-radio value="password">{{ $t('terminal.passwordMode') }}</el-radio>
                <el-radio value="key">{{ $t('terminal.keyMode') }}</el-radio>
            </el-radio-group>
        </el-form-item>
        <el-form-item v-if="bootstrapMode === 'password'" :label="$t('commons.login.password')">
            <el-input
                v-model="password"
                @input="emit('change')"
                type="password"
                show-password
                clearable
                autocomplete="new-password"
            />
        </el-form-item>
        <template v-else>
            <el-form-item :label="$t('terminal.key')">
                <el-input v-model="privateKey" @input="emit('change')" type="textarea" :rows="4" />
            </el-form-item>
            <el-form-item :label="$t('terminal.keyPassword')">
                <el-input v-model="passPhrase" @input="emit('change')" type="password" show-password clearable />
            </el-form-item>
        </template>
    </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { Host } from '@/api/interface/host';

const props = defineProps<{
    connection: Host.PasswordlessConfig;
}>();
const emit = defineEmits(['change']);
let pendingTarget: string | null = null;
const target = computed(() =>
    JSON.stringify([props.connection.addr, props.connection.port, props.connection.user, !!props.connection.useProxy]),
);
watch(
    target,
    () => {
        pendingTarget = null;
        emit('change');
    },
    { flush: 'sync' },
);
const bootstrapMode = ref(props.connection.privateKey ? 'key' : 'password');
const password = ref(props.connection.password || '');
const privateKey = ref(props.connection.privateKey || '');
const passPhrase = ref(props.connection.passPhrase || '');

const getTestParams = () => {
    pendingTarget = target.value;
    return {
        configurePasswordless: true,
        bootstrapAuth: bootstrapMode.value,
        password: bootstrapMode.value === 'password' ? password.value : '',
        privateKey: bootstrapMode.value === 'key' ? privateKey.value : '',
        passPhrase: bootstrapMode.value === 'key' ? passPhrase.value : '',
    };
};

const onTestSuccess = () => {
    if (pendingTarget !== target.value || props.connection.authMode !== 'passwordless') return;
    pendingTarget = null;
    password.value = '';
    privateKey.value = '';
    passPhrase.value = '';
};

defineExpose({ getTestParams, onTestSuccess });
</script>

<style scoped lang="scss">
.passwordless-auth {
    margin-bottom: 20px;
    padding: 16px;
    background: var(--el-fill-color-light);
    border: 1px solid var(--el-border-color-lighter);
    border-radius: var(--el-border-radius-base);

    :deep(.el-form-item) {
        margin-bottom: 16px;
    }

    :deep(.el-form-item:last-child) {
        margin-bottom: 0;
    }
}
</style>
