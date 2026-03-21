<template>
    <div>
        <ComplexTable :pagination-config="paginationConfig" @search="search" :data="data">
            <template #toolbar>
                <el-button type="primary" @click="onCreate">{{ $t('mail.createAccount') }}</el-button>
            </template>
            <el-table-column :label="$t('mail.email')" prop="email" min-width="200" show-overflow-tooltip />
            <el-table-column :label="$t('mail.username')" prop="username" min-width="120" />
            <el-table-column :label="$t('mail.quota')" prop="quota" min-width="100">
                <template #default="{ row }">
                    {{ row.quota === 0 ? 'Unlimited' : row.quota + ' MB' }}
                </template>
            </el-table-column>
            <el-table-column :label="$t('mail.status')" prop="status" min-width="80">
                <template #default="{ row }">
                    <el-tag v-if="row.status === 'active'" type="success" size="small">{{ $t('mail.enabled') }}</el-tag>
                    <el-tag v-else type="info" size="small">{{ $t('mail.disabled') }}</el-tag>
                </template>
            </el-table-column>
            <fu-table-operations :ellipsis="10" :buttons="buttons" :label="$t('commons.table.operate')" fixed="right" fix />
        </ComplexTable>

        <!-- Create dialog -->
        <DialogPro v-model="createOpen" :title="$t(isEdit ? 'mail.editAccount' : 'mail.createAccount')" @close="handleClose">
            <el-row>
                <el-col :span="22" :offset="1">
                    <el-form ref="formRef" label-position="top" :model="form" :rules="rules">
                        <el-form-item v-if="!isEdit" :label="$t('mail.username')" prop="username">
                            <el-input v-model.trim="form.username" placeholder="user">
                                <template #append>@{{ props.domainName }}</template>
                            </el-input>
                        </el-form-item>
                        <el-form-item :label="isEdit ? $t('mail.newPassword') : $t('mail.password')" :prop="isEdit ? '' : 'password'">
                            <el-input v-model="form.password" type="password" show-password />
                        </el-form-item>
                        <el-form-item :label="$t('mail.quota')" prop="quota">
                            <el-input-number v-model="form.quota" :min="0" :max="102400" :step="100" />
                            <span class="input-help ml-2">{{ $t('mail.quotaHelper') }}</span>
                        </el-form-item>
                    </el-form>
                </el-col>
            </el-row>
            <template #footer>
                <el-button @click="handleClose" :disabled="formLoading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit(formRef)" :loading="formLoading">{{ $t('commons.button.confirm') }}</el-button>
            </template>
        </DialogPro>

        <OpDialog ref="opRef" @search="search()" />
    </div>
</template>

<script lang="ts" setup>
import { onMounted, reactive, ref } from 'vue';
import { FormInstance } from 'element-plus';
import { searchMailAccounts, createMailAccount, updateMailAccount, deleteMailAccount } from '@/api/modules/mail';
import { Mail } from '@/api/interface/mail';
import { Rules } from '@/global/form-rules';
import { MsgSuccess } from '@/utils/message';
import i18n from '@/lang';

const props = defineProps<{ domainId: number; domainName: string }>();

const data = ref<Mail.AccountRes[]>([]);
const createOpen = ref(false);
const isEdit = ref(false);
const formLoading = ref(false);
const formRef = ref<FormInstance>();
const opRef = ref();

const form = ref({ id: 0, username: '', password: '', quota: 0 });
const rules = ref({
    username: [Rules.requiredInput],
    password: [Rules.requiredInput],
});

const paginationConfig = reactive({
    cacheSizeKey: 'mail-account-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('mail-account-page-size')) || 20,
    total: 0,
});

const buttons = [
    {
        label: i18n.global.t('commons.button.edit'),
        click: (row: Mail.AccountRes) => onEdit(row),
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: Mail.AccountRes) => onDelete(row),
    },
];

const search = async () => {
    const res = await searchMailAccounts({
        domainID: props.domainId,
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
    });
    data.value = res.data.items || [];
    paginationConfig.total = res.data.total;
};

const onCreate = () => {
    isEdit.value = false;
    form.value = { id: 0, username: '', password: '', quota: 0 };
    createOpen.value = true;
};

const onEdit = (row: Mail.AccountRes) => {
    isEdit.value = true;
    form.value = { id: row.id, username: row.username, password: '', quota: row.quota };
    createOpen.value = true;
};

const onDelete = (row: Mail.AccountRes) => {
    opRef.value.acceptParams({
        title: i18n.global.t('mail.deleteAccount'),
        names: [row.email],
        msg: i18n.global.t('mail.deleteAccountHelper'),
        api: deleteMailAccount,
        params: { id: row.id },
    });
};

const handleClose = () => { createOpen.value = false; formRef.value?.resetFields(); };

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) return;
        formLoading.value = true;
        const action = isEdit.value
            ? updateMailAccount({ id: form.value.id, password: form.value.password || undefined, quota: form.value.quota })
            : createMailAccount({ domainID: props.domainId, username: form.value.username, password: form.value.password, quota: form.value.quota });
        action
            .then(() => {
                MsgSuccess(i18n.global.t(isEdit.value ? 'commons.msg.updateSuccess' : 'commons.msg.createSuccess'));
                handleClose();
                search();
            })
            .finally(() => { formLoading.value = false; });
    });
};

onMounted(() => search());
</script>
