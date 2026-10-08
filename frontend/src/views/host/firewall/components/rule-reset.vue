<template>
    <DialogPro v-model="visible" class="firewall-rule-dialog" :title="$t('commons.button.reset')">
        <template #content>
            <div class="firewall-rule-dialog__content">
                <el-alert :title="message" type="warning" :closable="false" show-icon />
                <el-radio-group v-model="backup" class="firewall-rule-dialog__choices">
                    <el-radio :value="true" border>
                        <span class="firewall-rule-dialog__choice-title">{{ $t('firewall.resetWithBackup') }}</span>
                        <span class="firewall-rule-dialog__choice-description">
                            {{ $t('firewall.resetWithBackupHelper') }}
                        </span>
                    </el-radio>
                    <el-radio :value="false" border>
                        <span class="firewall-rule-dialog__choice-title">{{ $t('firewall.resetOnly') }}</span>
                        <span class="firewall-rule-dialog__choice-description">
                            {{ $t('firewall.resetOnlyHelper') }}
                        </span>
                    </el-radio>
                </el-radio-group>
                <div class="firewall-rule-dialog__confirmation">
                    <label for="firewall-reset-confirmation">
                        {{ $t('commons.msg.operateConfirm') }}
                        <code>{{ provider }}</code>
                    </label>
                    <el-input
                        id="firewall-reset-confirmation"
                        v-model="confirmation"
                        :placeholder="provider"
                        autocomplete="off"
                        @keyup.enter="submit"
                    />
                </div>
            </div>
        </template>
        <template #footer>
            <el-button @click="visible = false">{{ $t('commons.button.cancel') }}</el-button>
            <el-button type="primary" :disabled="confirmation !== provider" @click="submit">
                {{ $t('commons.button.reset') }}
            </el-button>
        </template>
    </DialogPro>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import DialogPro from '@/components/dialog-pro/index.vue';

const emit = defineEmits<{ (event: 'confirm', backup: boolean): void }>();
const visible = ref(false);
const provider = ref('');
const message = ref('');
const backup = ref(true);
const confirmation = ref('');

const acceptParams = (params: { provider: string; message: string }) => {
    provider.value = params.provider;
    message.value = params.message;
    backup.value = true;
    confirmation.value = '';
    visible.value = true;
};

const submit = () => {
    if (!visible.value || !provider.value || confirmation.value !== provider.value) return;
    visible.value = false;
    emit('confirm', backup.value);
};

defineExpose({ acceptParams });
</script>

<style lang="scss" src="./rule-dialog.scss"></style>
