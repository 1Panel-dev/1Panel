<template>
    <DialogPro v-model="open" :title="$t(isEdit ? 'dns.editRecord' : 'dns.createRecord')" @close="handleClose">
        <el-row>
            <el-col :span="22" :offset="1">
                <el-form ref="formRef" label-position="top" :model="form" :rules="rules">
                    <el-form-item :label="$t('dns.name')" prop="name">
                        <el-input v-model.trim="form.name" :placeholder="'@ or subdomain'">
                            <template #append>.{{ props.zoneName }}</template>
                        </el-input>
                        <span class="input-help">Use @ for the zone apex</span>
                    </el-form-item>
                    <el-form-item :label="$t('dns.recordType')" prop="type">
                        <el-select v-model="form.type" style="width: 100%" :disabled="isEdit">
                            <el-option v-for="t in recordTypes" :key="t" :label="t" :value="t" />
                        </el-select>
                    </el-form-item>
                    <el-form-item :label="$t('dns.content')" prop="content">
                        <el-input
                            v-model.trim="form.content"
                            :placeholder="contentPlaceholder"
                        />
                    </el-form-item>
                    <el-form-item :label="$t('dns.ttl')" prop="ttl">
                        <el-input-number v-model="form.ttl" :min="60" :max="86400" :step="60" />
                    </el-form-item>
                    <el-form-item
                        v-if="form.type === 'MX' || form.type === 'SRV'"
                        :label="$t('dns.priority')"
                        prop="priority"
                    >
                        <el-input-number v-model="form.priority" :min="0" :max="65535" />
                    </el-form-item>
                    <el-form-item v-if="isEdit" :label="$t('dns.status')">
                        <el-switch
                            v-model="form.disabled"
                            :active-text="$t('dns.disabled')"
                            :inactive-text="$t('dns.enabled')"
                        />
                    </el-form-item>
                </el-form>
            </el-col>
        </el-row>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button type="primary" @click="submit(formRef)" :loading="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DialogPro>
</template>

<script lang="ts" setup>
import { computed, ref } from 'vue';
import { FormInstance } from 'element-plus';
import { createDnsRecord, updateDnsRecord } from '@/api/modules/dns';
import { Rules } from '@/global/form-rules';
import { MsgSuccess } from '@/utils/message';
import i18n from '@/lang';

const props = defineProps<{
    zoneId: number;
    zoneName: string;
    defaultTtl: number;
}>();

const open = ref(false);
const loading = ref(false);
const isEdit = ref(false);
const formRef = ref<FormInstance>();

const recordTypes = ['A', 'AAAA', 'CNAME', 'MX', 'TXT', 'NS', 'SRV', 'CAA'];

const form = ref({
    id: 0,
    name: '',
    type: 'A',
    content: '',
    ttl: 3600,
    priority: 10,
    disabled: false,
});

const rules = ref({
    name: [Rules.requiredInput],
    type: [Rules.requiredSelect],
    content: [Rules.requiredInput],
});

const contentPlaceholder = computed(() => {
    switch (form.value.type) {
        case 'A':
            return '192.168.1.1';
        case 'AAAA':
            return '2001:db8::1';
        case 'CNAME':
            return 'target.example.com.';
        case 'MX':
            return 'mail.example.com.';
        case 'TXT':
            return '"v=spf1 include:example.com ~all"';
        case 'NS':
            return 'ns1.example.com.';
        case 'SRV':
            return '0 443 target.example.com.';
        case 'CAA':
            return '0 issue "letsencrypt.org"';
        default:
            return '';
    }
});

const emit = defineEmits(['search']);

interface RecordProps {
    mode: string;
    form?: any;
}

const acceptParams = (params: RecordProps) => {
    isEdit.value = params.mode === 'edit';
    if (isEdit.value && params.form) {
        // Extract subdomain from FQDN
        let name = params.form.name;
        if (name === props.zoneName) {
            name = '@';
        } else if (name.endsWith('.' + props.zoneName)) {
            name = name.slice(0, -(props.zoneName.length + 1));
        }
        form.value = {
            id: params.form.id,
            name: name,
            type: params.form.type,
            content: params.form.content,
            ttl: params.form.ttl,
            priority: params.form.priority || 10,
            disabled: params.form.disabled || false,
        };
    } else {
        form.value = {
            id: 0,
            name: '',
            type: 'A',
            content: '',
            ttl: props.defaultTtl || 3600,
            priority: 10,
            disabled: false,
        };
    }
    open.value = true;
};

const handleClose = () => {
    open.value = false;
    formRef.value?.resetFields();
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) return;
        loading.value = true;

        if (isEdit.value) {
            updateDnsRecord({
                id: form.value.id,
                name: form.value.name,
                type: form.value.type,
                content: form.value.content,
                ttl: form.value.ttl,
                priority: form.value.priority,
                disabled: form.value.disabled,
            })
                .then(() => {
                    MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
                    handleClose();
                    emit('search');
                })
                .finally(() => {
                    loading.value = false;
                });
        } else {
            createDnsRecord({
                zoneID: props.zoneId,
                name: form.value.name,
                type: form.value.type,
                content: form.value.content,
                ttl: form.value.ttl,
                priority: form.value.priority,
            })
                .then(() => {
                    MsgSuccess(i18n.global.t('commons.msg.createSuccess'));
                    handleClose();
                    emit('search');
                })
                .finally(() => {
                    loading.value = false;
                });
        }
    });
};

defineExpose({
    acceptParams,
});
</script>
